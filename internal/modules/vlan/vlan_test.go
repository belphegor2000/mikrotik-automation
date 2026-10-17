package vlan

import (
	"testing"
)

func TestVLANCalculation(t *testing.T) {
	v := VLANEntry{
		VLANID: 10,
		Name:   "users",
		CIDR:   "192.168.10.0/24",
	}

	if v.InterfaceName() != "vlan10" {
		t.Fatalf("expected vlan10, got %s", v.InterfaceName())
	}

	gw := v.Gateway()
	if gw != "192.168.10.1" {
		t.Fatalf("expected gateway 192.168.10.1, got %s", gw)
	}

	start := v.DHCPPoolStart()
	if start != "192.168.10.10" {
		t.Fatalf("expected pool start 192.168.10.10, got %s", start)
	}

	end := v.DHCPPoolEnd()
	if end != "192.168.10.254" {
		t.Fatalf("expected pool end 192.168.10.254, got %s", end)
	}
}

func TestVLAN1Native(t *testing.T) {
	v := VLANEntry{
		VLANID: 1,
		Name:   "management",
		CIDR:   "192.168.1.0/24",
	}

	if v.InterfaceName() != BridgeName {
		t.Fatalf("expected %s, got %s", BridgeName, v.InterfaceName())
	}
}

type mockVlanClient struct {
	addedPaths []string
	addedData  []map[string]interface{}
}

func (m *mockVlanClient) Connect(h, u, p string, port int, ssl bool) (string, error) { return "mock", nil }
func (m *mockVlanClient) Disconnect() error                                         { return nil }
func (m *mockVlanClient) IsConnected() bool                                         { return true }
func (m *mockVlanClient) GetIdentity() string                                       { return "mock" }
func (m *mockVlanClient) Get(path string) ([]map[string]interface{}, error)        { return nil, nil }
func (m *mockVlanClient) GetOne(path string) (map[string]interface{}, error)       { return nil, nil }
func (m *mockVlanClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.addedPaths = append(m.addedPaths, path)
	m.addedData = append(m.addedData, d)
	return nil, nil
}
func (m *mockVlanClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockVlanClient) Remove(path, id string) error                        { return nil }
func (m *mockVlanClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockVlanClient) RemoveByComment(path, comment string) (int, error)       { return 0, nil }
func (m *mockVlanClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockVlanClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestConfigureVLANWithLanAddress(t *testing.T) {
	mock := &mockVlanClient{}
	cfg := NewConfigurator(mock, func(string) {})
	vlans := []VLANEntry{
		{VLANID: 10, Name: "users", CIDR: "192.168.10.1/24"},
		{VLANID: 20, Name: "guests", CIDR: "192.168.20.1/24"},
	}
	err := cfg.Configure(vlans)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundLanAddr := false
	for _, d := range mock.addedData {
		if list, ok := d["list"].(string); ok && list == "LAN-Address" {
			if addr, ok := d["address"].(string); ok && addr == "192.168.10.0/24" {
				foundLanAddr = true
				break
			}
		}
	}
	if !foundLanAddr {
		t.Fatalf("expected LAN-Address entry for 192.168.10.0/24 not found in addedData")
	}
}

func TestConfigureVLANBridgeTrunking(t *testing.T) {
	mock := &mockVlanClient{}
	cfg := NewConfigurator(mock, func(string) {})
	vlans := []VLANEntry{
		{VLANID: 10, Name: "users", CIDR: "192.168.10.1/24"},
	}
	err := cfg.Configure(vlans)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundTagged := false
	for _, d := range mock.addedData {
		if _, ok := d["vlan-ids"]; ok {
			if tagged, ok := d["tagged"].(string); ok && tagged == BridgeName {
				foundTagged = true
				break
			}
		}
	}
	if !foundTagged {
		t.Fatalf("expected tagged entry for VLAN table not found")
	}
}

