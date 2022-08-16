package main

import (
	"context"
	"fmt"
	"log"
	"therealbroker/api/proto"
	"time"

	"google.golang.org/grpc"
)

const address = "localhost:8080"

func main() {

	connection, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer connection.Close()

	client := proto.NewBrokerClient(connection)
	request := &proto.PublishRequest{
		Subject:           "test",
		Body:              []byte("test message"),
		ExpirationSeconds: 10,
	}
	context1, _ := context.WithTimeout(context.Background(), 10*time.Second)
	sub, _ := client.Subscribe(context1, &proto.SubscribeRequest{Subject: "test"})
	context2, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, _ = client.Publish(context2, request)
	msg, err := sub.Recv()
	fmt.Println(msg, err)
}
