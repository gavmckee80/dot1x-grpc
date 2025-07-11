package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/gavmckee80/dot1x-grpc/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewDot1xManagerClient(conn)
	var wg sync.WaitGroup

	for i := 1; i <= 8; i++ {
		iface := fmt.Sprintf("eth%d", i)
		wg.Add(1)
		go func(iface string) {
			defer wg.Done()
			disconnect(client, iface)
		}(iface)
	}

	wg.Wait()
	log.Println("All interfaces disconnected.")
}

func disconnect(client pb.Dot1xManagerClient, iface string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Disconnect(ctx, &pb.InterfaceRequest{Interface: iface})
	if err != nil {
		log.Printf("[%s] Error: %v", iface, err)
		return
	}
	log.Printf("[%s] %v - %s", iface, resp.Success, resp.Message)
}
