// Package main provides the gRPC server entry point for the 802.1X authentication service.
// This server exposes a gRPC API for managing 802.1X authentication on Linux Ethernet
// interfaces, interfacing with wpa_supplicant via D-Bus.
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcapi "github.com/gavmckee80/dot1x-grpc/internal/grpc"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
	"google.golang.org/grpc"
)

// main initializes and starts the gRPC server for 802.1X authentication management.
// The server:
//   - Listens on port 50051 for gRPC connections
//   - Registers the Dot1XManager service
//   - Handles graceful shutdown on SIGINT/SIGTERM signals
//   - Cleans up resources when shutting down
func main() {
	// Create TCP listener on the default gRPC port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Initialize gRPC server and register the 802.1X service
	s := grpc.NewServer()
	service := grpcapi.NewDot1xService()
	pb.RegisterDot1XManagerServer(s, service)

	// Set up signal handling for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Start graceful shutdown goroutine
	go func() {
		<-sig
		log.Println("Shutting down...")
		s.GracefulStop()
		service.Shutdown()
		os.Exit(0)
	}()

	log.Println("gRPC server listening on :50051")
	s.Serve(lis)
}
