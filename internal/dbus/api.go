package dbus

import "github.com/godbus/dbus/v5"

// SupplicantAPI defines the interface implemented by SupplicantClient.
type SupplicantAPI interface {
	CreateInterface(ifname string) (dbus.ObjectPath, error)
	RemoveInterface(path dbus.ObjectPath) error
	GetInterfacePathByName(ifname string) (dbus.ObjectPath, error)
	AddNetwork(ifacePath dbus.ObjectPath, config map[string]string) (dbus.ObjectPath, error)
	SelectNetwork(ifacePath, networkPath dbus.ObjectPath) error
	DisconnectNetwork(ifacePath dbus.ObjectPath) error
	Close()
}
