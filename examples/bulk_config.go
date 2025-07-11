//go:build examples

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

	client := pb.NewDot1XManagerClient(conn)
	var wg sync.WaitGroup

	for i := 1; i <= 8; i++ {
		iface := fmt.Sprintf("eth%d", i)
		wg.Add(1)
		go func(iface string) {
			defer wg.Done()
			configure(client, iface)
		}(iface)
	}

	wg.Wait()
	log.Println("All interfaces configured.")
}

func configure(client pb.Dot1XManagerClient, iface string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.Dot1XConfigRequest{
		Interface:  iface,
		EapType:    pb.EapType_EAP_PEAP,
		Identity:   "testuser",
		Password:   "testpass",
		Phase2Auth: "mschapv2",
	}

	resp, err := client.ConfigureInterface(ctx, req)
	if err != nil {
		log.Printf("[%s] Error: %v", iface, err)
		return
	}
	log.Printf("[%s] %v - %s", iface, resp.Success, resp.Message)
}
