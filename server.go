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
	/*Listener made. We've not been told a lot about the listener, just stick to what is
	made below, it works.*/
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	log.Printf("Starting server...")
	if err != nil {
		log.Printf("Failed to listen on port :%d, Error: %s", ownPort, err)
	}
	log.Printf("Listening on port :%d", ownPort)

	/*GRPC relevant stuff. The server struct is created and a the NewServer() grpc method (auto generated)
	is called. These two are then put together in the RegisterReplicationServer(grpcserver, &server)
	function, and afterwards made to serve with the Serve(listener) function. */
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
	/*The mutex lock is not necessary for the server to work, but is useful to make sure that
	critical sections are not messed up when working with replication.*/
	mux sync.Mutex
	/*This needs to be in the struct, for GRPC to know it's a server.*/
	replication.UnimplementedReplicationServer
	/*ID is purely for the template to be able to tell the client which server answered.
	It is not needed by GRPC.*/
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
	answerPackage.Message = fmt.Sprintf("Server #%d received your message, thank you!", b.id)
	return answerPackage, nil
}
