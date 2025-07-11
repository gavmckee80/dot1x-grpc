// Package dbus provides D-Bus communication with wpa_supplicant for 802.1X authentication.
// This file defines the SupplicantAPI interface that abstracts the D-Bus operations
// for managing network interfaces and authentication configurations.
package dbus

import "github.com/godbus/dbus/v5"

// SupplicantAPI defines the interface for interacting with wpa_supplicant via D-Bus.
// This interface abstracts the D-Bus operations needed to manage 802.1X authentication
// on network interfaces, providing a clean API for the business logic layer.
//
// The interface includes methods for:
//   - Interface management (create, remove, lookup)
//   - Network configuration (add, select, disconnect)
//   - Resource cleanup (close connection)
//
// Implementations of this interface should handle the low-level D-Bus communication
// with wpa_supplicant, converting between Go types and D-Bus variants as needed.
type SupplicantAPI interface {
	// CreateInterface creates a new network interface in wpa_supplicant.
	// Returns the D-Bus object path of the created interface.
	CreateInterface(ifname string) (dbus.ObjectPath, error)

	// RemoveInterface removes a network interface from wpa_supplicant.
	// This disconnects the interface and cleans up associated resources.
	RemoveInterface(path dbus.ObjectPath) error

	// GetInterfacePathByName retrieves the D-Bus object path of an existing interface.
	// Returns the object path or an error if the interface is not found.
	GetInterfacePathByName(ifname string) (dbus.ObjectPath, error)

	// AddNetwork adds a network configuration to a wpa_supplicant interface.
	// The configuration map contains authentication parameters (EAP type, credentials, etc.).
	// Returns the D-Bus object path of the created network.
	AddNetwork(ifacePath dbus.ObjectPath, config map[string]string) (dbus.ObjectPath, error)

	// SelectNetwork activates a network configuration on an interface.
	// This tells wpa_supplicant to attempt authentication using the specified configuration.
	SelectNetwork(ifacePath, networkPath dbus.ObjectPath) error

	// DisconnectNetwork disconnects the current network on an interface.
	// This terminates the 802.1X authentication session.
	DisconnectNetwork(ifacePath dbus.ObjectPath) error

	// Close closes the D-Bus connection and releases associated resources.
	// This method should be called when the client is no longer needed.
	Close()
}
