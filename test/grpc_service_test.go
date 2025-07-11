package test

import (
	"context"
	"net"
	"testing"

	"github.com/gavmckee80/dot1x-grpc/internal/core"
	grpcapi "github.com/gavmckee80/dot1x-grpc/internal/grpc"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	manager := core.NewInterfaceManagerWithClient(&MockSupplicant{})
	service := grpcapi.NewDot1xServiceWithManager(manager)
	pb.RegisterDot1XManagerServer(s, service)
	go s.Serve(lis)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestConfigureInterfaceTLS(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewDot1XManagerClient(conn)

	req := &pb.Dot1XConfigRequest{
		Interface:  "eth1",
		EapType:    pb.EapType_EAP_TLS,
		Identity:   "testuser",
		CaCert:     []byte("CA CERT"),
		ClientCert: []byte("CLIENT CERT"),
		PrivateKey: []byte("PRIVATE KEY"),
	}
	resp, err := client.ConfigureInterface(ctx, req)
	if err != nil {
		t.Fatalf("ConfigureInterface error: %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected success, got failure: %s", resp.Message)
	}
}

func TestDisconnect(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewDot1XManagerClient(conn)

	_, _ = client.ConfigureInterface(ctx, &pb.Dot1XConfigRequest{
		Interface:  "eth2",
		EapType:    pb.EapType_EAP_PEAP,
		Identity:   "bob",
		Password:   "pass",
		Phase2Auth: "mschapv2",
	})
	resp, err := client.Disconnect(ctx, &pb.InterfaceRequest{Interface: "eth2"})
	if err != nil {
		t.Fatalf("Disconnect error: %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected disconnect success, got: %s", resp.Message)
	}
}
