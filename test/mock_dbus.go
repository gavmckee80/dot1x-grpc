package test

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

type MockSupplicant struct {
	Created []string
}

func (m *MockSupplicant) CreateInterface(ifname string) (dbus.ObjectPath, error) {
	m.Created = append(m.Created, ifname)
	return dbus.ObjectPath("/mock/" + ifname), nil
}

func (m *MockSupplicant) RemoveInterface(path dbus.ObjectPath) error {
	return nil
}

func (m *MockSupplicant) GetInterfacePathByName(ifname string) (dbus.ObjectPath, error) {
	if ifname == "fail" {
		return "", errors.New("not found")
	}
	return dbus.ObjectPath("/mock/" + ifname), nil
}

func (m *MockSupplicant) AddNetwork(_ dbus.ObjectPath, _ map[string]string) (dbus.ObjectPath, error) {
	return dbus.ObjectPath("/mock/net"), nil
}

func (m *MockSupplicant) SelectNetwork(_, _ dbus.ObjectPath) error {
	return nil
}

func (m *MockSupplicant) DisconnectNetwork(_ dbus.ObjectPath) error {
	return nil
}

func (m *MockSupplicant) Close() {}
