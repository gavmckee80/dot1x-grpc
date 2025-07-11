package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/gavmckee80/dot1x-grpc/proto"
	"google.golang.org/grpc"
)

func main() {
	var (
		serverAddr = flag.String("server", "localhost:50051", "gRPC server address")
		iface      = flag.String("iface", "eth0", "interface to authenticate")
		eap        = flag.String("eap", "PEAP", "EAP method (PEAP, TLS, TTLS)")
		identity   = flag.String("id", "", "EAP identity")
		password   = flag.String("pass", "", "EAP password (if applicable)")
		phase2     = flag.String("phase2", "mschapv2", "Inner auth for PEAP/TTLS")
		disconnect = flag.Bool("disconnect", false, "disconnect interface")
		status     = flag.Bool("status", false, "get one-time status of interface")
		stream     = flag.Bool("stream", false, "stream live status updates")
	)
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewDot1XManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch {
	case *disconnect:
		resp, err := client.Disconnect(ctx, &pb.InterfaceRequest{Interface: *iface})
		if err != nil {
			log.Fatalf("Disconnect error: %v", err)
		}
		fmt.Printf("Disconnect result: %v - %s\n", resp.Success, resp.Message)
		return
	case *status:
		resp, err := client.GetStatus(ctx, &pb.InterfaceRequest{Interface: *iface})
		if err != nil {
			log.Fatalf("Status error: %v", err)
		}
		fmt.Printf("Status: %s\nEAP: %s\nLast: %s\nTimestamp: %d\n",
			resp.Status, resp.EapState, resp.LastEvent, resp.Timestamp)
		return
	case *stream:
		streamCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		stream, err := client.StreamStatus(streamCtx, &pb.InterfaceRequest{Interface: *iface})
		if err != nil {
			log.Fatalf("StreamStatus error: %v", err)
		}
		fmt.Println("Streaming live interface status. Ctrl+C to exit.")
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error reading stream: %v", err)
			}
			fmt.Printf("[%d] %s - %s (%s)\n",
				resp.Timestamp, resp.Interface, resp.Status, resp.EapState)
		}
		return
	}

	eapType := map[string]pb.EapType{
		"PEAP": pb.EapType_EAP_PEAP,
		"TLS":  pb.EapType_EAP_TLS,
		"TTLS": pb.EapType_EAP_TTLS,
	}[(*eap)]

	req := &pb.Dot1XConfigRequest{
		Interface:  *iface,
		EapType:    eapType,
		Identity:   *identity,
		Password:   *password,
		Phase2Auth: *phase2,
	}

	resp, err := client.ConfigureInterface(ctx, req)
	if err != nil {
		log.Fatalf("Configure error: %v", err)
	}
	fmt.Printf("Configure result: %v - %s\n", resp.Success, resp.Message)
}
