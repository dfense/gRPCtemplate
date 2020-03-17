package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	pb "github.com/dfense/protobufModels"

	"time"

	"google.golang.org/grpc"
)

var (
	errAlreadyInShutdown = errors.New("shutdown already in progress")
)

type EchoClient struct {
	fname        string
	lname        string
	conn         *grpc.ClientConn
	stream       pb.Echo_EchoClient
	delete       uint32
	closeMonitor chan struct{}
}

func main() {
	rand.Seed(time.Now().Unix())
	e := EchoClient{fname: "bob", lname: "smith", closeMonitor: make(chan struct{})}

	ctx := context.Background()

	var err error
	e.conn, e.stream, err = getStream(ctx)
	for err != nil {
		// stall and try again
		log.Println("trying getStream")
		e.conn, e.stream, err = getStream(ctx)
	}

	ctx = e.stream.Context()
	if err != nil {
		fmt.Errorf("error opening stream: %s", err)
	}

	go e.monitor(ctx)

	e.startReading()
	go e.console()

WAITING:
	for {
		select {
		case <-e.closeMonitor:
			fmt.Println("timeout")
			// time for stream to close out
			time.Sleep(time.Millisecond * 100)
			break WAITING
		}
	}

}

func (e *EchoClient) startReading() {
	go func() {

		defer e.stream.CloseSend()
		defer e.conn.Close()

		for {
			log.Println("setting up Recv()")
			resp, err := e.stream.Recv()
			if err == io.EOF {
				log.Println("EOF received")
				break
			}
			if err != nil {
				log.Printf("second error: %s\n", err)
				break
			}
			echoReply := resp.GetEreply()
			log.Printf("new echo reply %d received\n", echoReply)
		}
		// select {
		// case <-ctx.Done():
		// 	log.Println("inside startReading-context")
		// default:
		// 	log.Println("inside startReading-default")
		// }
		log.Println("leaving second loop")
	}()
}

func getStream(ctx context.Context) (*grpc.ClientConn, pb.Echo_EchoClient, error) {
	// dail server
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Errorf("can not connect with server %v", err)
		return nil, nil, err
	}

	log.Info("made it past dial")

	// create stream
	// client := pb.NewGreeterClient(conn)
	echoService := pb.NewEchoClient(conn)
	//stream, err := client2.Echo(context.Background())
	stream, err := echoService.Echo(ctx)
	if err != nil {
		log.Errorf("openn stream error %v", err)
		return nil, nil, err
	}
	return conn, stream, nil
}

func (e *EchoClient) monitor(ctx context.Context) {

FINISH:
	for {

		select {
		case <-ctx.Done():
			fmt.Println("done triggered in monitor")

			// are we in shutdown ?
			if atomic.LoadUint32(&e.delete) == 1 {
				e.closeMonitor <- struct{}{}
				break FINISH
			}

			var err error
			ctx = context.Background()
			e.conn, e.stream, err = getStream(ctx)
			for err != nil {
				// stall and try again
				time.Sleep(time.Second * 1)
				log.Errorf("trying getStream: %s", err)
				e.conn, e.stream, err = getStream(ctx)
			}
			fmt.Println("reconnected stream")

			ctx = e.stream.Context()
			e.startReading()

		case <-e.closeMonitor:
			fmt.Println("leaving")
			break FINISH
		}
	}

}

func (e *EchoClient) shutdownAll() error {

	if !atomic.CompareAndSwapUint32(&e.delete, 0, 1) {
		return errAlreadyInShutdown
	}

	// force close
	// stream has no effect
	// e.stream.CloseSend()
	e.conn.Close()
	return nil

}

// simple console to test shutdown
func (e *EchoClient) console() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Simple Shell")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("hi", text) == 0 {
			fmt.Println("hello, Yourself")
		}
		if strings.Compare("q", text) == 0 {
			fmt.Println("Shutting down =================")

			err := e.shutdownAll()
			if err != nil {
				fmt.Printf("Error shutting down : %s\n", err)
			}
		}
		fmt.Printf("<= %s\n", text)

	}

}
