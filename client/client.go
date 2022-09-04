package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"therealbroker/api/proto"
	"time"

	"google.golang.org/grpc"
)

const address = "192.168.70.191:30470"

func main() {

	wg := sync.WaitGroup{}
	// connection, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	connection, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer connection.Close()

	request := &proto.PublishRequest{
		Subject:           "test2",
		Body:              "a",
		ExpirationSeconds: 0,
	}

	// context1, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// sub, _ := client.Subscribe(context1, &proto.SubscribeRequest{Subject: "test"})
	context2, _ := context.WithTimeout(context.Background(), 1000*time.Second)

	for k := 0; k < 10; k++ {

		client := proto.NewBrokerClient(connection)
		wg.Add(100)

		for i := 0; i < 100; i++ {

			go func() {
				for j := 0; j < 1000; j++ {
					_, err = client.Publish(context2, request)
					if err != nil {
						fmt.Println(err)
					}
				}
				// msg, err := sub.Recv()
				// fmt.Println(msg, err)
				wg.Done()
			}()
		}
	}

	wg.Wait()
	// time.Sleep(40 * time.Second)
}
