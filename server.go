package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jskoven/Dissys-fault-tolerant-client2server-template/replication"

	"google.golang.org/grpc"
)

var (
	timestamp time.Time
	timeLimit time.Time
)

func main() {
	//Setting log output
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Eror on opening file: %s", err)
	}
	log.SetOutput(f)

	/*The part below is responsible for making each server listen on their own port. They need to be
	started with "go run server.go 0" up to 2, to make them run on the correct ports.*/
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 9000
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	log.Printf("Starting server...")
	if err != nil {
		log.Printf("Failed to listen on port :%d, Error: %s", ownPort, err)
	}
	log.Printf("Listening on port :%d", ownPort)

	grpcserver := grpc.NewServer()
	server := ServerStruct{}
	server.id = ownPort
	replication.RegisterReplicationServer(grpcserver, &server)

	err = grpcserver.Serve(listener)
	if err != nil {
		log.Printf("Replica #%d:  Failed to serve with listener, error: %s\n", server.id, err)

	}

}

type ServerStruct struct {
	mux sync.Mutex
	replication.UnimplementedReplicationServer
	id int32
}

/*
The grpc method below must contain all this (context, etc) to work. The exact function header
can be found in the grpc.pb.go file, and almost just copy pasted in since it is auto-generated
from the .proto file.
*/
func (b *ServerStruct) Send(ctx context.Context, message *replication.Package) (answer *replication.Package, err error) {
	messageString := message.Message
	answerPackage := &replication.Package{}

	fmt.Printf("Received message from client: %s \n", messageString)
	answerPackage.Message = "We've received your message, thank you!"
	return answerPackage, nil
}
