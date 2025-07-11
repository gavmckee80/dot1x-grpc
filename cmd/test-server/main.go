// Package main provides a test gRPC server with reflection enabled for development and testing.
// This server uses mock D-Bus clients to avoid requiring a real D-Bus connection.
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gavmckee80/dot1x-grpc/internal/core"
	grpcapi "github.com/gavmckee80/dot1x-grpc/internal/grpc"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
	"github.com/gavmckee80/dot1x-grpc/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// main initializes and starts a test gRPC server with reflection enabled.
// This server uses mock D-Bus clients for testing without requiring a real D-Bus connection.
func main() {
	// Create TCP listener on the default gRPC port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Initialize gRPC server with mock D-Bus client
	s := grpc.NewServer()

	// Create manager with mock D-Bus client
	mockClient := &test.MockSupplicant{}
	manager := core.NewInterfaceManagerWithClient(mockClient)
	service := grpcapi.NewDot1xServiceWithManager(manager)

	pb.RegisterDot1XManagerServer(s, service)

	// Enable gRPC reflection for service discovery and debugging
	reflection.Register(s)

	// Set up signal handling for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Start graceful shutdown goroutine
	go func() {
		<-sig
		log.Println("Shutting down test server...")
		s.GracefulStop()
		service.Shutdown()
		os.Exit(0)
	}()

	log.Println("Test gRPC server listening on :50051")
	log.Println("gRPC reflection enabled - use grpcurl to explore the API")
	log.Println("Example: grpcurl -plaintext localhost:50051 list")
	s.Serve(lis)
}
