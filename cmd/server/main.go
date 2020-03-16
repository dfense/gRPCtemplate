/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/dfense/protobufModels"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

type eserver struct {
	name string
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// Ping returns a PongReply
func (eserver *eserver) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongReply, error) {
	log.Printf("Servername: %s\n", eserver.name)
	log.Printf("Received: %v\n", in.GetMessageId())
	return &pb.PongReply{MessageId: 200, PongCnt: 0}, nil
}

func (eserver eserver) Echo(srv pb.Echo_EchoServer) error {

	ctx, _ := context.WithCancel(srv.Context())
	for {
		fmt.Println("settting up for read")
		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			log.Println("exiting on ctx.Done()")
			return ctx.Err()
		default:
		}

		// receive data from stream
		// req, err := srv.Recv()
		// if err == io.EOF {
		// 	// return will close stream from server side
		// 	log.Println("exit")
		// 	return nil
		// }
		// if err != nil {
		// 	log.Printf("receive error %v", err)
		// 	continue
		// }

		// value := req.GetErequest()
		// fmt.Println(value)

		// update max and send it to stream
		value := int32(5)
		resp := pb.EchoReply{Ereply: value}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
		log.Printf("send new max=%d", value)
		// log.Println("calling cencel")
		// cancel()
		time.Sleep(500 * time.Millisecond)

	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &server{})
	pb.RegisterEchoServer(s, &eserver{name: "EchoServer"})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	s.GracefulStop()
}
