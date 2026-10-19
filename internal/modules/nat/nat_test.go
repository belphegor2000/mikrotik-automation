package nat

import (
	"fmt"
	"testing"
)

type mockNATClient struct {
	natRules   []map[string]interface{}
	addrRules  []map[string]interface{}
	addedRules int
	removedIDs []string
}

func (m *mockNATClient) Connect(h, u, p string, port int, ssl bool) (string, error) {
	return "mock", nil
}
func (m *mockNATClient) Disconnect() error                 { return nil }
func (m *mockNATClient) IsConnected() bool                 { return true }
func (m *mockNATClient) GetIdentity() string               { return "mock" }
func (m *mockNATClient) Get(path string) ([]map[string]interface{}, error) {
	if path == "ip/firewall/nat" {
		return m.natRules, nil
	}
	if path == "ip/address" {
		return m.addrRules, nil
	}
	return nil, nil
}
func (m *mockNATClient) GetOne(path string) (map[string]interface{}, error) { return nil, nil }
func (m *mockNATClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.addedRules++
	d[".id"] = fmt.Sprintf("*%d", len(m.natRules)+1)
	m.natRules = append(m.natRules, d)
	return d, nil
}
func (m *mockNATClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockNATClient) Remove(path, id string) error {
	m.removedIDs = append(m.removedIDs, id)
	for i, r := range m.natRules {
		if r[".id"] == id {
			m.natRules = append(m.natRules[:i], m.natRules[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockNATClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockNATClient) RemoveByComment(path, comment string) (int, error)       { return 0, nil }
func (m *mockNATClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockNATClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestValidatePort(t *testing.T) {
	valid := []string{"80", "443", "8080", "1-65535", "8000-8080", "80,443,8080"}
	for _, p := range valid {
		if err := ValidatePort(p); err != nil {
			t.Errorf("expected valid port %s, got: %v", p, err)
		}
	}

	invalid := []string{"", "0", "65536", "-1", "abc", "8080-8000", "80-", "-80"}
	for _, p := range invalid {
		if err := ValidatePort(p); err == nil {
			t.Errorf("expected error for port %s, but got none", p)
		}
	}
}

func TestDeriveSubnet(t *testing.T) {
	sub, err := DeriveSubnet("192.168.88.50")
	if err != nil || sub != "192.168.88.0/24" {
		t.Fatalf("expected 192.168.88.0/24, got %s (err: %v)", sub, err)
	}

	sub2, err2 := DeriveSubnet("10.10.20.1/24")
	if err2 != nil || sub2 != "10.10.20.0/24" {
		t.Fatalf("expected 10.10.20.0/24, got %s (err: %v)", sub2, err2)
	}
}

func TestAddAndListPortForwards(t *testing.T) {
	mock := &mockNATClient{}
	cfg := NewConfigurator(mock, func(string) {})

	rule := PortForwardRule{
		Name:          "Web Server",
		Protocol:      "tcp",
		DstPort:       "8080",
		ToAddresses:   "192.168.88.100",
		ToPorts:       "80",
		EnableHairpin: true,
	}

	err := cfg.AddPortForward(rule)
	if err != nil {
		t.Fatalf("unexpected error adding port forward: %v", err)
	}

	// Should have added dst-nat rule + hairpin nat rule
	if mock.addedRules != 2 {
		t.Fatalf("expected 2 added rules (dst-nat + hairpin), got %d", mock.addedRules)
	}

	// List rules
	items, err := cfg.ListPortForwards()
	if err != nil {
		t.Fatalf("unexpected error listing rules: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 PF rule, got %d", len(items))
	}
	if items[0].Name != "Web Server" || items[0].DstPort != "8080" {
		t.Fatalf("mismatched rule content: %+v", items[0])
	}

	// Delete
	err = cfg.DeletePortForward(items[0].ID)
	if err != nil {
		t.Fatalf("unexpected error deleting rule: %v", err)
	}
	if len(mock.removedIDs) != 1 {
		t.Fatalf("expected 1 removed ID, got %d", len(mock.removedIDs))
	}
}

func TestEnsureHairpinNAT(t *testing.T) {
	mock := &mockNATClient{}
	cfg := NewConfigurator(mock, func(string) {})

	// First time: should add
	err := cfg.EnsureHairpinNAT("192.168.88.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedRules != 1 {
		t.Fatalf("expected 1 added rule, got %d", mock.addedRules)
	}

	// Second time: should skip since already in mock
	err = cfg.EnsureHairpinNAT("192.168.88.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedRules != 1 {
		t.Fatalf("expected still 1 rule (skipped duplicate), got %d", mock.addedRules)
	}
}

func TestDetectLANSubnets(t *testing.T) {
	mock := &mockNATClient{
		addrRules: []map[string]interface{}{
			{"interface": "bridgeVLAN", "address": "192.168.88.1/24"},
			{"interface": "vlan10", "address": "192.168.10.1/24"},
			{"interface": "pppoe-wan1", "address": "14.162.20.5/32"},
			{"interface": "ether1", "address": "192.168.1.2/24"},
		},
	}
	cfg := NewConfigurator(mock, func(string) {})
	subnets, err := cfg.DetectLANSubnets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subnets) != 2 {
		t.Fatalf("expected 2 LAN subnets, got %d: %v", len(subnets), subnets)
	}
}
