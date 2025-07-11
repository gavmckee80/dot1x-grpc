package core

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/gavmckee80/dot1x-grpc/internal/dbus"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
)

type InterfaceManager struct {
	client     dbus.SupplicantAPI
	interfaces map[string]dbus.ObjectPath
	tempFiles  []string
}

func NewInterfaceManager() (*InterfaceManager, error) {
	client, err := dbus.NewSupplicantClient()
	if err != nil {
		return nil, err
	}
	return &InterfaceManager{
		client:     client,
		interfaces: make(map[string]dbus.ObjectPath),
	}, nil
}

func NewInterfaceManagerWithClient(c dbus.SupplicantAPI) *InterfaceManager {
	return &InterfaceManager{
		client:     c,
		interfaces: make(map[string]dbus.ObjectPath),
	}
}

func (m *InterfaceManager) Configure(req *pb.Dot1xConfigRequest) (*pb.Dot1xConfigResponse, error) {
	if req.EapType == pb.EapType_EAP_UNKNOWN {
		return &pb.Dot1xConfigResponse{Success: false, Message: "Invalid EAP type"}, nil
	}
	if req.Identity == "" {
		return &pb.Dot1xConfigResponse{Success: false, Message: "Identity is required"}, nil
	}

	if req.EapType == pb.EapType_EAP_TLS {
		if len(req.CaCert) == 0 || len(req.ClientCert) == 0 || len(req.PrivateKey) == 0 {
			return &pb.Dot1xConfigResponse{Success: false, Message: "TLS credentials missing"}, nil
		}
	}

	ifacePath, err := m.client.GetInterfacePathByName(req.Interface)
	if err != nil {
		ifacePath, err = m.client.CreateInterface(req.Interface)
		if err != nil {
			return &pb.Dot1xConfigResponse{Success: false, Message: err.Error()}, nil
		}
	}
	m.interfaces[req.Interface] = ifacePath

	cfg := map[string]string{
		"eap":         req.EapType.String()[4:], // Strip EAP_ prefix
		"identity":    req.Identity,
		"key_mgmt":    "IEEE8021X",
		"eapol_flags": "0",
	}

	if req.EapType == pb.EapType_EAP_PEAP || req.EapType == pb.EapType_EAP_TTLS {
		cfg["password"] = req.Password
		cfg["phase2"] = fmt.Sprintf("auth=%s", req.Phase2Auth)
	}

	if req.EapType == pb.EapType_EAP_TLS {
		caPath, err := writeTempFile(req.CaCert, "ca.pem")
		if err != nil {
			return nil, err
		}
		clientCert, err := writeTempFile(req.ClientCert, "client.pem")
		if err != nil {
			return nil, err
		}
		privateKey, err := writeTempFile(req.PrivateKey, "key.pem")
		if err != nil {
			return nil, err
		}
		m.tempFiles = append(m.tempFiles, caPath, clientCert, privateKey)

		cfg["ca_cert"] = caPath
		cfg["client_cert"] = clientCert
		cfg["private_key"] = privateKey

		if req.PrivateKeyPassword != "" {
			cfg["private_key_passwd"] = req.PrivateKeyPassword
		}
	}

	netPath, err := m.client.AddNetwork(ifacePath, cfg)
	if err != nil {
		return &pb.Dot1xConfigResponse{Success: false, Message: err.Error()}, nil
	}
	err = m.client.SelectNetwork(ifacePath, netPath)
	if err != nil {
		return &pb.Dot1xConfigResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.Dot1xConfigResponse{Success: true, Message: "Configured"}, nil
}

func (m *InterfaceManager) Disconnect(req *pb.InterfaceRequest) (*pb.DisconnectResponse, error) {
	ifacePath, ok := m.interfaces[req.Interface]
	if !ok {
		return &pb.DisconnectResponse{Success: false, Message: "Interface not managed"}, nil
	}
	err := m.client.DisconnectNetwork(ifacePath)
	if err != nil {
		return &pb.DisconnectResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.DisconnectResponse{Success: true, Message: "Disconnected"}, nil
}

func (m *InterfaceManager) Shutdown() {
	for _, path := range m.tempFiles {
		os.Remove(path)
	}
	for _, iface := range m.interfaces {
		m.client.RemoveInterface(iface)
	}
	m.client.Close()
}

func writeTempFile(content []byte, filename string) (string, error) {
	tmpPath := fmt.Sprintf("/tmp/%d_%s", time.Now().UnixNano(), filename)
	if err := os.WriteFile(tmpPath, content, 0600); err != nil {
		return "", errors.New("failed to write temp file: " + err.Error())
	}
	return tmpPath, nil
}
