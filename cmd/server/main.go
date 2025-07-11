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

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	service := grpcapi.NewDot1xService()
	pb.RegisterDot1xManagerServer(s, service)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

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
