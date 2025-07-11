package dbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const supplicantInterface = "fi.w1.wpa_supplicant1"
const supplicantPath = dbus.ObjectPath("/fi/w1/wpa_supplicant1")

const (
	interfaceInterface = supplicantInterface + ".Interface"
	networkInterface   = supplicantInterface + ".Network"
)

type SupplicantClient struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

func NewSupplicantClient() (*SupplicantClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %v", err)
	}
	obj := conn.Object(supplicantInterface, supplicantPath)
	return &SupplicantClient{conn: conn, obj: obj}, nil
}

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

func (s *SupplicantClient) RemoveInterface(path dbus.ObjectPath) error {
	call := s.obj.Call(supplicantInterface+".RemoveInterface", 0, path)
	return call.Err
}

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

func (s *SupplicantClient) SelectNetwork(ifacePath, networkPath dbus.ObjectPath) error {
	obj := s.conn.Object(supplicantInterface, ifacePath)
	return obj.Call(interfaceInterface+".SelectNetwork", 0, networkPath).Err
}

func (s *SupplicantClient) DisconnectNetwork(ifacePath dbus.ObjectPath) error {
	obj := s.conn.Object(supplicantInterface, ifacePath)
	return obj.Call(interfaceInterface+".Disconnect", 0).Err
}

func (s *SupplicantClient) RawConnection() *dbus.Conn {
	return s.conn
}

func (s *SupplicantClient) Close() {
	s.conn.Close()
}
