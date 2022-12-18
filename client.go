package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	//If the github link shows an error, remember to do "go mod tidy" to make go actually
	//get it. If it still shows an error, give it a bit of time. Sometimes it takes a while to get it down.
	//Also remember to change the module in the go.mod file. It needs to be a link to the correct repo,
	//or it wont work.
	"github.com/jskoven/Dissys-fault-tolerant-client2server-template/replication"

	"google.golang.org/grpc"
)

type client struct {
	//The client struct needs to contain the line below, in order for GRPC to know
	//that the struct is in fact a client.
	replication.ReplicationClient
	//A map for containing the client connections to the servers, allowing for a distributed
	//fault tolerant system. This template does not use a frontend, but simply connects the
	//client to all the servers. This isn't quite correct, but it works and was accepted for the handin
	//as long as it is mentioned.
	replicas map[int32]replication.ReplicationClient
}

func main() {
	//Setting log output
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Eror on opening file: %s", err)
	}
	log.SetOutput(f)

	//Client struct
	c := &client{
		replicas: make(map[int32]replication.ReplicationClient),
	}

	/*For loop to connect the client to each of the servers. This template needs 3 servers to run, with
	the correct ports (9000 to 9002) to function. It saves the servers to the client structs map.*/
	for i := 0; i < 3; i++ {
		port := int32(9000) + int32(i)
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}

		defer conn.Close()

		server := replication.NewReplicationClient(conn)
		/*Although we are instantiating a "client" above, it should not be seen as we are starting
		3 other clients. Rather, our client (this file) is starting 3 client connections to each'
		server.*/
		c.replicas[port] = server
		//Server being saved to client map
		log.Printf("Client %s connected to port %d \n", port)
		fmt.Println()
	}

	fmt.Println("Type a string to send to server.")
	for {
		/*Simple for loop to send messages to server*/
		Scanner := bufio.NewScanner(os.Stdin)
		Scanner.Scan()
		MessageToBeSent := strings.ToLower(Scanner.Text())
		c.sendMessage(MessageToBeSent)

	}

}

func (c *client) sendMessage(messageToSend string) {

	bp := replication.Package{
		Message: messageToSend,
	}
	for index, element := range c.replicas {
		if element != nil {
			/*Conf is in this case also a package. It is simply the return type chosen in the
			.proto file. The actual send method, and it's functionality, is defined in the server.go file*/
			conf, err := element.Send(context.Background(), &bp)
			/*The template has no functionality to check whether the servers answer the same thing,
			it simply takes the answer from the last server.*/
			if err != nil {
				log.Printf("## Replica number %d is down, skipping it.##\n", index)
				/*If a server doesn't respond, an error is thrown. We then just skip that server, and
				print the error. This is what makes the program fault tolerant, as it will continue
				to work as long as one of the servers answer.*/
			} else {
				/*Currently just set to print whatever the server returns.*/
				fmt.Println(conf.Message)
			}
		}
	}
}
