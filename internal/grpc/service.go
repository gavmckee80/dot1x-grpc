// Package grpc provides the gRPC service implementation for 802.1X authentication management.
// It implements the Dot1XManager service defined in the protobuf specification,
// handling client requests and delegating to the core business logic layer.
package grpc

import (
	"context"
	"log"
	"time"

	"github.com/gavmckee80/dot1x-grpc/internal/core"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
)

// Dot1xService implements the gRPC Dot1XManager service interface.
// It provides methods for configuring, monitoring, and disconnecting
// 802.1X authentication sessions on network interfaces.
type Dot1xService struct {
	pb.UnimplementedDot1XManagerServer
	manager *core.InterfaceManager
}

// NewDot1xService creates a new Dot1xService instance with a default
// InterfaceManager. This is the primary constructor for production use.
//
// Panics if the InterfaceManager cannot be created (e.g., D-Bus connection failure).
func NewDot1xService() *Dot1xService {
	manager, err := core.NewInterfaceManager()
	if err != nil {
		log.Fatalf("Failed to create interface manager: %v", err)
	}
	return &Dot1xService{manager: manager}
}

// NewDot1xServiceWithManager creates a service using the provided manager.
// This constructor is primarily used for testing with mock managers.
func NewDot1xServiceWithManager(m *core.InterfaceManager) *Dot1xService {
	return &Dot1xService{manager: m}
}

// ConfigureInterface configures 802.1X authentication for a network interface.
// This method handles the gRPC request, validates context cancellation,
// delegates to the core manager, and logs the operation results.
//
// The method supports all EAP types defined in the protobuf specification:
//   - EAP-PEAP: Protected EAP with MSCHAPv2
//   - EAP-TTLS: Tunneled TLS with inner authentication
//   - EAP-TLS: Certificate-based authentication
//   - EAP-FAST: Flexible Authentication via Secure Tunneling
//
// Returns a Dot1XConfigResponse with success/failure status and details.
func (s *Dot1xService) ConfigureInterface(ctx context.Context, req *pb.Dot1XConfigRequest) (*pb.Dot1XConfigResponse, error) {
	// Check for context cancellation before processing
	select {
	case <-ctx.Done():
		log.Println("[WARN] ConfigureInterface canceled")
		return nil, ctx.Err()
	default:
	}

	// Measure and log operation duration
	start := time.Now()
	resp, err := s.manager.Configure(req)
	log.Printf("[INFO] Configure %s (%s) in %s: %s", req.Interface, req.EapType.String(), time.Since(start), resp.Message)
	return resp, err
}

// Disconnect terminates the 802.1X authentication session for the specified interface.
// This method handles the gRPC request, validates context cancellation,
// and delegates to the core manager for the actual disconnection.
//
// Returns a DisconnectResponse indicating success or failure.
func (s *Dot1xService) Disconnect(ctx context.Context, req *pb.InterfaceRequest) (*pb.DisconnectResponse, error) {
	// Check for context cancellation before processing
	select {
	case <-ctx.Done():
		log.Println("[WARN] Disconnect canceled")
		return nil, ctx.Err()
	default:
	}

	log.Printf("[INFO] Disconnect %s", req.Interface)
	return s.manager.Disconnect(req)
}

// GetStatus retrieves the current status of a network interface.
// This is a mock implementation that returns static status information.
// In a production environment, this would query the actual interface state
// from wpa_supplicant via D-Bus.
//
// Returns an InterfaceStatus with current interface information.
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

// StreamStatus provides a real-time stream of interface status updates.
// This method implements server-side streaming gRPC, sending status updates
// every 3 seconds until the client disconnects or the context is canceled.
//
// The stream includes:
//   - Interface name and current status
//   - EAP authentication state
//   - Last authentication event
//   - Timestamp of the status update
//   - IP address (when available)
//
// The stream continues until the client closes the connection or the context is canceled.
func (s *Dot1xService) StreamStatus(req *pb.InterfaceRequest, stream pb.Dot1XManager_StreamStatusServer) error {
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

// Shutdown performs cleanup operations when the service is shutting down.
// It delegates to the core manager to clean up resources, including:
//   - Removing temporary certificate files
//   - Disconnecting managed interfaces
//   - Closing D-Bus connections
func (s *Dot1xService) Shutdown() {
	log.Println("[INFO] Shutting down Dot1x service...")
	s.manager.Shutdown()
}
