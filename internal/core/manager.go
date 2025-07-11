// Package core provides the business logic for 802.1X authentication management.
// It handles the configuration of network interfaces, credential management,
// and communication with the D-Bus layer for wpa_supplicant integration.
package core

import (
	"errors"
	"fmt"
	"os"
	"time"

	godbus "github.com/godbus/dbus/v5"

	"github.com/gavmckee80/dot1x-grpc/internal/dbus"
	pb "github.com/gavmckee80/dot1x-grpc/proto"
)

// InterfaceManager handles 802.1X authentication configuration for network interfaces.
// It provides methods to configure, monitor, and disconnect interfaces using
// the underlying D-Bus client to communicate with wpa_supplicant.
type InterfaceManager struct {
	client     dbus.SupplicantAPI
	interfaces map[string]godbus.ObjectPath
	tempFiles  []string
}

// NewInterfaceManager creates a new InterfaceManager instance with a default
// D-Bus client connection to wpa_supplicant.
//
// Returns an error if the D-Bus connection cannot be established.
func NewInterfaceManager() (*InterfaceManager, error) {
	client, err := dbus.NewSupplicantClient()
	if err != nil {
		return nil, err
	}
	return &InterfaceManager{
		client:     client,
		interfaces: make(map[string]godbus.ObjectPath),
	}, nil
}

// NewInterfaceManagerWithClient creates a new InterfaceManager instance with
// a custom D-Bus client. This is primarily used for testing with mock clients.
func NewInterfaceManagerWithClient(c dbus.SupplicantAPI) *InterfaceManager {
	return &InterfaceManager{
		client:     c,
		interfaces: make(map[string]godbus.ObjectPath),
	}
}

// Configure sets up 802.1X authentication for a network interface based on
// the provided configuration request.
//
// The method validates the request parameters and handles different EAP types:
//   - EAP-PEAP/TTLS: Requires identity, password, and phase2 authentication
//   - EAP-TLS: Requires CA certificate, client certificate, and private key
//
// Returns a Dot1XConfigResponse indicating success or failure with details.
func (m *InterfaceManager) Configure(req *pb.Dot1XConfigRequest) (*pb.Dot1XConfigResponse, error) {
	// Validate EAP type
	if req.EapType == pb.EapType_EAP_UNKNOWN {
		return &pb.Dot1XConfigResponse{Success: false, Message: "Invalid EAP type"}, nil
	}

	// Validate required identity
	if req.Identity == "" {
		return &pb.Dot1XConfigResponse{Success: false, Message: "Identity is required"}, nil
	}

	// Validate TLS credentials for EAP-TLS
	if req.EapType == pb.EapType_EAP_TLS {
		if len(req.CaCert) == 0 || len(req.ClientCert) == 0 || len(req.PrivateKey) == 0 {
			return &pb.Dot1XConfigResponse{Success: false, Message: "TLS credentials missing"}, nil
		}
	}

	// Get or create interface path
	ifacePath, err := m.client.GetInterfacePathByName(req.Interface)
	if err != nil {
		ifacePath, err = m.client.CreateInterface(req.Interface)
		if err != nil {
			return &pb.Dot1XConfigResponse{Success: false, Message: err.Error()}, nil
		}
	}
	m.interfaces[req.Interface] = ifacePath

	// Build wpa_supplicant configuration
	cfg := map[string]string{
		"eap":         req.EapType.String()[4:], // Strip EAP_ prefix
		"identity":    req.Identity,
		"key_mgmt":    "IEEE8021X",
		"eapol_flags": "0",
	}

	// Add password and phase2 auth for PEAP/TTLS
	if req.EapType == pb.EapType_EAP_PEAP || req.EapType == pb.EapType_EAP_TTLS {
		cfg["password"] = req.Password
		cfg["phase2"] = fmt.Sprintf("auth=%s", req.Phase2Auth)
	}

	// Handle TLS certificate files for EAP-TLS
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

	// Add network configuration to wpa_supplicant
	netPath, err := m.client.AddNetwork(ifacePath, cfg)
	if err != nil {
		return &pb.Dot1XConfigResponse{Success: false, Message: err.Error()}, nil
	}

	// Select the configured network
	err = m.client.SelectNetwork(ifacePath, netPath)
	if err != nil {
		return &pb.Dot1XConfigResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.Dot1XConfigResponse{Success: true, Message: "Configured"}, nil
}

// Disconnect terminates the 802.1X authentication session for the specified interface.
// It removes the interface from the managed interfaces list and disconnects
// the network in wpa_supplicant.
//
// Returns a DisconnectResponse indicating success or failure.
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

// Shutdown performs cleanup operations when the service is shutting down.
// It removes all temporary certificate files and disconnects all managed interfaces.
func (m *InterfaceManager) Shutdown() {
	// Clean up temporary certificate files
	for _, path := range m.tempFiles {
		os.Remove(path)
	}

	// Remove all managed interfaces
	for _, iface := range m.interfaces {
		m.client.RemoveInterface(iface)
	}

	// Close the D-Bus connection
	m.client.Close()
}

// writeTempFile writes the provided content to a temporary file with the given filename.
// The file is created in /tmp with a unique timestamp prefix and restrictive permissions.
//
// Returns the full path to the created file or an error if the operation fails.
func writeTempFile(content []byte, filename string) (string, error) {
	tmpPath := fmt.Sprintf("/tmp/%d_%s", time.Now().UnixNano(), filename)
	if err := os.WriteFile(tmpPath, content, 0600); err != nil {
		return "", errors.New("failed to write temp file: " + err.Error())
	}
	return tmpPath, nil
}
