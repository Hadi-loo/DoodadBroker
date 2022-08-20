package main

import (
	"context"
	"log"
	"sync"
	"therealbroker/api/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const address = "localhost:8080"

func main() {

	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			connection, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer connection.Close()

			client := proto.NewBrokerClient(connection)
			request := &proto.PublishRequest{
				Subject:           "test",
				Body:              "message",
				ExpirationSeconds: 10,
			}
			// context1, _ := context.WithTimeout(context.Background(), 10*time.Second)
			// sub, _ := client.Subscribe(context1, &proto.SubscribeRequest{Subject: "test"})
			context2, _ := context.WithTimeout(context.Background(), 10*time.Second)
			for j := 0; j < 500; j++ {
				_, _ = client.Publish(context2, request)
			}
			wg.Done()
			// msg, err := sub.Recv()
			// fmt.Println(msg, err)
		}()
	}

	wg.Wait()
	// time.Sleep(20 * time.Second)
}
