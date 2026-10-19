package pppoe

import (
	"testing"
)

type mockClient struct {
	added     []string
	addedData []map[string]interface{}
}

func (m *mockClient) Connect(h, u, p string, port int, ssl bool) (string, error) { return "mock", nil }
func (m *mockClient) Disconnect() error                                         { return nil }
func (m *mockClient) IsConnected() bool                                         { return true }
func (m *mockClient) GetIdentity() string                                       { return "mock" }
func (m *mockClient) Get(path string) ([]map[string]interface{}, error)        { return nil, nil }
func (m *mockClient) GetOne(path string) (map[string]interface{}, error)       { return nil, nil }
func (m *mockClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.added = append(m.added, path)
	m.addedData = append(m.addedData, d)
	return nil, nil
}
func (m *mockClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockClient) Remove(path, id string) error                        { return nil }
func (m *mockClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockClient) RemoveByComment(path, comment string) (int, error)       { return 0, nil }
func (m *mockClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestConfigureSingle(t *testing.T) {
	mock := &mockClient{}
	cfg := NewConfigurator(mock, func(string) {})
	err := cfg.ConfigureSingle("user1", "pass1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.added) != 2 {
		t.Fatalf("expected 2 additions (pppoe + nat), got %d", len(mock.added))
	}
}

func TestConfigureDualPCC(t *testing.T) {
	mock := &mockClient{}
	cfg := NewConfigurator(mock, func(string) {})
	err := cfg.ConfigureDualPCC("user1", "pass1", "user2", "pass2", "bridgeVLAN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 pppoe + 2 tables + 9 mangle (with Lan-Address accept) + 4 routes + 2 nat = 19 additions
	if len(mock.added) < 10 {
		t.Fatalf("expected >= 10 additions for dual pcc, got %d", len(mock.added))
	}

	foundAcceptRule := false
	for _, d := range mock.addedData {
		if d["chain"] == "prerouting" && d["src-address-list"] == "LAN-Address" && d["dst-address-list"] == "LAN-Address" && d["action"] == "accept" {
			foundAcceptRule = true
			break
		}
	}
	if !foundAcceptRule {
		t.Fatalf("expected accept rule for LAN-Address not found in mangle rules")
	}
}

func TestConfigureDualPCCRecursive(t *testing.T) {
	mock := &mockClient{}
	cfg := NewConfigurator(mock, func(string) {})
	err := cfg.ConfigureDualPCCEx("user1", "pass1", "user2", "pass2", "bridgeVLAN", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found8888 := false
	found1111 := false
	foundCheckGateway := false

	for _, d := range mock.addedData {
		dst, _ := d["dst-address"].(string)
		gw, _ := d["gateway"].(string)
		chk, _ := d["check-gateway"].(string)

		if dst == "8.8.8.8/32" && gw == PPPoEWAN1 {
			found8888 = true
		}
		if dst == "1.1.1.1/32" && gw == PPPoEWAN2 {
			found1111 = true
		}
		if dst == "0.0.0.0/0" && (gw == "8.8.8.8" || gw == "1.1.1.1") && chk == "ping" {
			foundCheckGateway = true
		}
	}

	if !found8888 {
		t.Fatalf("expected virtual hop route to 8.8.8.8/32 via WAN1")
	}
	if !found1111 {
		t.Fatalf("expected virtual hop route to 1.1.1.1/32 via WAN2")
	}
	if !foundCheckGateway {
		t.Fatalf("expected check-gateway=ping recursive default route")
	}
}

