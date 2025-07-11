package grpc

import (
	"context"
	"log"
	"time"

	"github.com/gavmckee80/dot1x-grpc/internal/core"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
)

type Dot1xService struct {
	pb.UnimplementedDot1xManagerServer
	manager *core.InterfaceManager
}

func NewDot1xService() *Dot1xService {
	manager, err := core.NewInterfaceManager()
	if err != nil {
		log.Fatalf("Failed to create interface manager: %v", err)
	}
	return &Dot1xService{manager: manager}
}

func (s *Dot1xService) ConfigureInterface(ctx context.Context, req *pb.Dot1xConfigRequest) (*pb.Dot1xConfigResponse, error) {
	select {
	case <-ctx.Done():
		log.Println("[WARN] ConfigureInterface canceled")
		return nil, ctx.Err()
	default:
	}

	start := time.Now()
	resp, err := s.manager.Configure(req)
	log.Printf("[INFO] Configure %s (%s) in %s: %s", req.Interface, req.EapType.String(), time.Since(start), resp.Message)
	return resp, err
}

func (s *Dot1xService) Disconnect(ctx context.Context, req *pb.InterfaceRequest) (*pb.DisconnectResponse, error) {
	select {
	case <-ctx.Done():
		log.Println("[WARN] Disconnect canceled")
		return nil, ctx.Err()
	default:
	}
	log.Printf("[INFO] Disconnect %s", req.Interface)
	return s.manager.Disconnect(req)
}

func (s *Dot1xService) GetStatus(ctx context.Context, req *pb.InterfaceRequest) (*pb.InterfaceStatus, error) {
	return &pb.InterfaceStatus{
		Interface: req.Interface,
		Status:    "mock-status",
		EapState:  "mock-success",
		LastEvent: "connected",
		Timestamp: time.Now().Unix(),
		IpAddress: "192.0.2.1",
	}, nil
}

func (s *Dot1xService) StreamStatus(req *pb.InterfaceRequest, stream pb.Dot1xManager_StreamStatusServer) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case t := <-ticker.C:
			err := stream.Send(&pb.InterfaceStatus{
				Interface: req.Interface,
				Status:    "streaming",
				EapState:  "authenticating",
				LastEvent: "EAP-STARTED",
				Timestamp: t.Unix(),
				IpAddress: "",
			})
			if err != nil {
				return err
			}
		}
	}
}

func (s *Dot1xService) Shutdown() {
	log.Println("[INFO] Shutting down Dot1x service...")
	s.manager.Shutdown()
}
