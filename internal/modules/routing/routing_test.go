package routing

import (
	"fmt"
	"testing"
)

type mockPBRClient struct {
	tables       []map[string]interface{}
	routingRules []map[string]interface{}
	mangleRules  []map[string]interface{}
	addressLists []map[string]interface{}
	added        []string
}

func (m *mockPBRClient) Connect(h, u, p string, port int, ssl bool) (string, error) {
	return "mock", nil
}
func (m *mockPBRClient) Disconnect() error                 { return nil }
func (m *mockPBRClient) IsConnected() bool                 { return true }
func (m *mockPBRClient) GetIdentity() string               { return "mock" }
func (m *mockPBRClient) Get(path string) ([]map[string]interface{}, error) {
	switch path {
	case "routing/table":
		return m.tables, nil
	case "routing/rule":
		return m.routingRules, nil
	case "ip/firewall/mangle":
		return m.mangleRules, nil
	case "ip/firewall/address-list":
		return m.addressLists, nil
	}
	return nil, nil
}
func (m *mockPBRClient) GetOne(path string) (map[string]interface{}, error) { return nil, nil }
func (m *mockPBRClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.added = append(m.added, path)
	switch path {
	case "routing/table":
		d[".id"] = fmt.Sprintf("*tbl%d", len(m.tables)+1)
		m.tables = append(m.tables, d)
	case "routing/rule":
		d[".id"] = fmt.Sprintf("*rr%d", len(m.routingRules)+1)
		m.routingRules = append(m.routingRules, d)
	case "ip/firewall/mangle":
		d[".id"] = fmt.Sprintf("*mg%d", len(m.mangleRules)+1)
		m.mangleRules = append(m.mangleRules, d)
	case "ip/firewall/address-list":
		d[".id"] = fmt.Sprintf("*al%d", len(m.addressLists)+1)
		m.addressLists = append(m.addressLists, d)
	}
	return d, nil
}
func (m *mockPBRClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockPBRClient) Remove(path, id string) error {
	switch path {
	case "routing/rule":
		for i, r := range m.routingRules {
			if r[".id"] == id {
				m.routingRules = append(m.routingRules[:i], m.routingRules[i+1:]...)
				break
			}
		}
	case "ip/firewall/address-list":
		for i, r := range m.addressLists {
			if r[".id"] == id {
				m.addressLists = append(m.addressLists[:i], m.addressLists[i+1:]...)
				break
			}
		}
	}
	return nil
}
func (m *mockPBRClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockPBRClient) RemoveByComment(path, comment string) (int, error)       { return 0, nil }
func (m *mockPBRClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockPBRClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestPBRVLANSubnetRule(t *testing.T) {
	mock := &mockPBRClient{}
	cfg := NewConfigurator(mock, func(string) {})

	err := cfg.AddVLANSubnetRule("VLAN Kế toán", "192.168.10.0/24", "WAN1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have added routing rule & mangle rule
	if len(mock.routingRules) != 1 {
		t.Fatalf("expected 1 routing rule, got %d", len(mock.routingRules))
	}
	if mock.routingRules[0]["table"] != "to_wan1" {
		t.Fatalf("expected table to_wan1, got %v", mock.routingRules[0]["table"])
	}

	rules, err := cfg.ListPBRRules()
	if err != nil {
		t.Fatalf("list rules error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 listed rule, got %d", len(rules))
	}
	if rules[0].Name != "VLAN Kế toán" || rules[0].TargetWAN != "WAN1" {
		t.Fatalf("mismatched listed rule: %+v", rules[0])
	}

	// Test delete
	err = cfg.DeletePBRRule(rules[0])
	if err != nil {
		t.Fatalf("delete rule error: %v", err)
	}
	if len(mock.routingRules) != 0 {
		t.Fatalf("expected 0 routing rules after delete")
	}
}

func TestPBRDomainIPRule(t *testing.T) {
	mock := &mockPBRClient{}
	cfg := NewConfigurator(mock, func(string) {})

	err := cfg.AddDomainIPRule("Netflix Stream", "netflix.com", "WAN2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.addressLists) != 1 {
		t.Fatalf("expected 1 address list item, got %d", len(mock.addressLists))
	}
	if mock.addressLists[0]["list"] != AddressListWAN2 {
		t.Fatalf("expected list %s, got %v", AddressListWAN2, mock.addressLists[0]["list"])
	}

	rules, err := cfg.ListPBRRules()
	if err != nil {
		t.Fatalf("list rules error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 listed rule, got %d", len(rules))
	}
	if rules[0].Destination != "netflix.com" || rules[0].TargetWAN != "WAN2" {
		t.Fatalf("mismatched listed rule: %+v", rules[0])
	}

	// Delete
	err = cfg.DeletePBRRule(rules[0])
	if err != nil {
		t.Fatalf("delete rule error: %v", err)
	}
	if len(mock.addressLists) != 0 {
		t.Fatalf("expected 0 address list items after delete")
	}
}
