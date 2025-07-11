// Package dbus provides D-Bus communication with wpa_supplicant for 802.1X authentication.
// It implements the SupplicantAPI interface to manage network interfaces, configure
// authentication parameters, and monitor connection status through the D-Bus system.
package dbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// D-Bus interface and path constants for wpa_supplicant
const supplicantInterface = "fi.w1.wpa_supplicant1"
const supplicantPath = dbus.ObjectPath("/fi/w1/wpa_supplicant1")

// D-Bus interface constants for network management
const (
	interfaceInterface = supplicantInterface + ".Interface"
	networkInterface   = supplicantInterface + ".Network"
)

// SupplicantClient provides D-Bus communication with wpa_supplicant.
// It handles the creation and management of network interfaces, configuration
// of authentication parameters, and monitoring of connection status.
type SupplicantClient struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

// NewSupplicantClient creates a new SupplicantClient instance and establishes
// a connection to the system D-Bus with access to wpa_supplicant.
//
// Returns an error if the D-Bus connection cannot be established.
func NewSupplicantClient() (*SupplicantClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %v", err)
	}
	obj := conn.Object(supplicantInterface, supplicantPath)
	return &SupplicantClient{conn: conn, obj: obj}, nil
}

// CreateInterface creates a new network interface in wpa_supplicant for the specified
// interface name. The interface is configured for wired (Ethernet) connections.
//
// The method sets up the interface with:
//   - Interface name (ifname)
//   - Driver type (wired)
//   - Empty config file (configuration will be added via D-Bus)
//
// Returns the D-Bus object path of the created interface or an error.
func (s *SupplicantClient) CreateInterface(ifname string) (dbus.ObjectPath, error) {
	props := map[string]dbus.Variant{
		"Ifname":     dbus.MakeVariant(ifname),
		"Driver":     dbus.MakeVariant("wired"),
		"ConfigFile": dbus.MakeVariant(""),
	}
	var path dbus.ObjectPath
	err := s.obj.Call(supplicantInterface+".CreateInterface", 0, props).Store(&path)
	if err != nil {
		return "", fmt.Errorf("CreateInterface failed: %v", err)
	}
	return path, nil
}

// RemoveInterface removes a network interface from wpa_supplicant.
// This method disconnects the interface and cleans up associated resources.
//
// Returns an error if the removal operation fails.
func (s *SupplicantClient) RemoveInterface(path dbus.ObjectPath) error {
	call := s.obj.Call(supplicantInterface+".RemoveInterface", 0, path)
	return call.Err
}

// GetInterfacePathByName retrieves the D-Bus object path of an existing interface
// by its name. This method queries all available interfaces and matches by name.
//
// Returns the object path of the interface or an error if not found.
func (s *SupplicantClient) GetInterfacePathByName(ifname string) (dbus.ObjectPath, error) {
	var paths []dbus.ObjectPath
	err := s.obj.Call(supplicantInterface+".GetInterfacePaths", 0).Store(&paths)
	if err != nil {
		return "", err
	}
	for _, p := range paths {
		obj := s.conn.Object(supplicantInterface, p)
		prop, err := obj.GetProperty(interfaceInterface + ".Ifname")
		if err == nil && prop.Value().(string) == ifname {
			return p, nil
		}
	}
	return "", fmt.Errorf("interface %s not found", ifname)
}

// AddNetwork adds a network configuration to a wpa_supplicant interface.
// The configuration map contains key-value pairs that define the authentication
// parameters (EAP type, identity, credentials, certificates, etc.).
//
// The method converts the configuration map to D-Bus variants and sends
// the AddNetwork call to wpa_supplicant.
//
// Returns the D-Bus object path of the created network or an error.
func (s *SupplicantClient) AddNetwork(ifacePath dbus.ObjectPath, config map[string]string) (dbus.ObjectPath, error) {
	obj := s.conn.Object(supplicantInterface, ifacePath)
	networkProps := make(map[string]dbus.Variant)
	for k, v := range config {
		networkProps[k] = dbus.MakeVariant(v)
	}
	var netPath dbus.ObjectPath
	err := obj.Call(interfaceInterface+".AddNetwork", 0, networkProps).Store(&netPath)
	if err != nil {
		return "", fmt.Errorf("AddNetwork failed: %v", err)
	}
	return netPath, nil
}

// SelectNetwork activates a network configuration on an interface.
// This method tells wpa_supplicant to attempt authentication using
// the specified network configuration.
//
// Returns an error if the network selection fails.
func (s *SupplicantClient) SelectNetwork(ifacePath, networkPath dbus.ObjectPath) error {
	obj := s.conn.Object(supplicantInterface, ifacePath)
	return obj.Call(interfaceInterface+".SelectNetwork", 0, networkPath).Err
}

// DisconnectNetwork disconnects the current network on an interface.
// This method terminates the 802.1X authentication session and
// disconnects from the network.
//
// Returns an error if the disconnection fails.
func (s *SupplicantClient) DisconnectNetwork(ifacePath dbus.ObjectPath) error {
	obj := s.conn.Object(supplicantInterface, ifacePath)
	return obj.Call(interfaceInterface+".Disconnect", 0).Err
}

// RawConnection returns the underlying D-Bus connection object.
// This method is primarily used for testing and advanced D-Bus operations
// that require direct access to the connection.
func (s *SupplicantClient) RawConnection() *dbus.Conn {
	return s.conn
}

// Close closes the D-Bus connection and releases associated resources.
// This method should be called when the client is no longer needed
// to prevent resource leaks.
func (s *SupplicantClient) Close() {
	s.conn.Close()
}
