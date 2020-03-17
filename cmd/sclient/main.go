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
	delete       uint32
	closeMonitor chan struct{}
	close        chan struct{}
}

func main() {
	rand.Seed(time.Now().Unix())
	e := EchoClient{fname: "bob", lname: "smith", closeMonitor: make(chan struct{}), close: make(chan struct{})}

	ctx, _ := context.WithCancel(context.Background())
	// dail server, this context not helpful unless i need to call cancel on conn
	conn, err := grpc.DialContext(ctx, ":50051", grpc.WithInsecure())
	if err != nil {
		log.Errorf("can not connect with server %v", err)
		os.Exit(1)
	}

	e.startReading(ctx, conn)
	go e.console()

	// wait for inner channel message from EchoClient read
WAITING:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("---- tope ---- ")
		case <-e.closeMonitor:
			fmt.Println("timeout")
			// time for stream to close out
			time.Sleep(time.Millisecond * 100)
			break WAITING
		}
	}

}

// startReading takes a connection and creates a gRPC service and calls rpc streaming function
func (e *EchoClient) startReading(transportContext context.Context, inconn *grpc.ClientConn) {
	go func() {

		// create service
		echoService := pb.NewEchoClient(inconn)

		// create the stream on Echo retry forever
		stream, err := echoService.Echo(transportContext)
		for err != nil {
			log.Errorf("open stream error %v", err)
			time.Sleep(time.Second) // slow down rate of retry
			stream, err = echoService.Echo(transportContext)
		}

		// this one calls cancel on stream error
		streamContext := stream.Context()

		// restart thread grpc connection if it drops
		go func() {
			// wait for thread to die
			newctx, _ := context.WithCancel(context.Background())

		QUIT:
			select {

			case <-streamContext.Done():
				// if shutting down, exit w/o restart
				if atomic.LoadUint32(&e.delete) == 1 {
					inconn.Close()
					e.closeMonitor <- struct{}{}
					break QUIT
				}

				// start restart of read - duplicate of main()
				newconn, err := grpc.DialContext(newctx, ":50051", grpc.WithInsecure())
				if err != nil {
					log.Errorf("dial can't connect to server: %s", err)
				}
				e.startReading(transportContext, newconn)

				// called by external channel message to kill running process
			case <-e.close:
				fmt.Println("Calling cancel")
				// kill running stream
				inconn.Close()
				e.closeMonitor <- struct{}{}
			}
		}()

		// just read the stream
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Errorf("EOF received")
				break
			}
			if err != nil {
				log.Errorf("second error: %s\n", err)
				break
			}
			echoReply := resp.GetEreply()
			log.Printf("new echo message [%d] \n", echoReply)
		}
	}()

}

// call shutdown EchoClient from outside
func (e *EchoClient) shutdownAll() error {

	if !atomic.CompareAndSwapUint32(&e.delete, 0, 1) {
		return errAlreadyInShutdown
	}
	e.close <- struct{}{}
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

		if strings.Compare("q", text) == 0 {
			fmt.Println("Shutting down =================")

			err := e.shutdownAll()
			if err != nil {
				log.Errorf("Error shutting down : %s\n", err)
			}
		}
	}
}
