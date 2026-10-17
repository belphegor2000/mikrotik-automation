package ipsec

import (
	"testing"
)

type mockIPSecClient struct {
	addedPaths []string
	addedData  []map[string]interface{}
}

func (m *mockIPSecClient) Connect(h, u, p string, port int, ssl bool) (string, error) { return "mock", nil }
func (m *mockIPSecClient) Disconnect() error                                         { return nil }
func (m *mockIPSecClient) IsConnected() bool                                         { return true }
func (m *mockIPSecClient) GetIdentity() string { return "mock" }
func (m *mockIPSecClient) Get(path string) ([]map[string]interface{}, error) {
	if path == "ip/address" {
		return []map[string]interface{}{
			{"interface": "pppoe-wan1", "address": "14.162.100.20/32"},
		}, nil
	}
	return nil, nil
}
func (m *mockIPSecClient) GetOne(path string) (map[string]interface{}, error) {
	if path == "ip/cloud" {
		return map[string]interface{}{
			"dns-name": "test123456.sn.mynetname.net",
		}, nil
	}
	return nil, nil
}
func (m *mockIPSecClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.addedPaths = append(m.addedPaths, path)
	m.addedData = append(m.addedData, d)
	return nil, nil
}
func (m *mockIPSecClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockIPSecClient) Remove(path, id string) error                        { return nil }
func (m *mockIPSecClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockIPSecClient) RemoveByComment(path, comment string) (int, error)       { return 0, nil }
func (m *mockIPSecClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockIPSecClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestParseAndNormalizeSubnets(t *testing.T) {
	// Normalization & deduplication
	input := "192.168.10.1/24, 192.168.20.5/24\n192.168.30.0/24; 192.168.10.0/24"
	subnets, err := ParseAndNormalizeSubnets(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subnets) != 3 {
		t.Fatalf("expected 3 unique subnets, got %d: %v", len(subnets), subnets)
	}
	if subnets[0] != "192.168.10.0/24" || subnets[1] != "192.168.20.0/24" || subnets[2] != "192.168.30.0/24" {
		t.Fatalf("unexpected subnets content: %v", subnets)
	}

	// Invalid input
	_, err = ParseAndNormalizeSubnets("invalid-ip")
	if err == nil {
		t.Fatalf("expected error for invalid ip, got nil")
	}
}

func TestConfigureMultiSubnet(t *testing.T) {
	mock := &mockIPSecClient{}
	cfg := NewConfigurator(mock, func(string) {})

	err := cfg.Configure(IPSecConfig{
		PeerAddress:   "branch.sn.mynetname.net",
		PSK:           "SecretKey123",
		LocalSubnets:  []string{"192.168.10.0/24", "192.168.20.0/24"},
		RemoteSubnets: []string{"172.16.1.0/24", "172.16.2.0/24", "172.16.3.0/24"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	policyCount := 0
	natCount := 0
	mangleCount := 0
	addrLanCount := 0
	addrRemoteCount := 0
	hasDPD := false
	hasFWFilter := false

	for i, path := range mock.addedPaths {
		data := mock.addedData[i]
		if path == "ip/firewall/address-list" {
			if list, ok := data["list"].(string); ok {
				if list == "LAN-Address" {
					addrLanCount++
				} else if list == "Remote-Address" {
					addrRemoteCount++
				}
			}
		}
		if path == "ip/ipsec/policy" {
			policyCount++
		}
		if path == "ip/firewall/nat" {
			if action, ok := data["action"].(string); ok && action == "accept" {
				if data["src-address-list"] == "LAN-Address" && data["dst-address-list"] == "Remote-Address" {
					natCount++
				}
			}
		}
		if path == "ip/firewall/mangle" {
			if action, ok := data["action"].(string); ok && action == "accept" {
				mangleCount++
			}
		}
		if path == "ip/ipsec/profile" {
			if dpd, ok := data["dpd-interval"].(string); ok && dpd == "10s" {
				hasDPD = true
			}
		}
		if path == "ip/firewall/filter" {
			hasFWFilter = true
		}
	}

	// 2 local in LAN-Address, 3 remote in Remote-Address
	if addrLanCount != 2 {
		t.Fatalf("expected 2 subnets in LAN-Address, got %d", addrLanCount)
	}
	if addrRemoteCount != 3 {
		t.Fatalf("expected 3 subnets in Remote-Address, got %d", addrRemoteCount)
	}
	// 2 Mangle rules (LAN-Address <-> Remote-Address)
	if mangleCount != 2 {
		t.Fatalf("expected 2 Mangle bypass rules, got %d", mangleCount)
	}
	// 1 NAT Bypass rule (LAN-Address -> Remote-Address)
	if natCount != 1 {
		t.Fatalf("expected 1 NAT bypass rule, got %d", natCount)
	}
	// 2 local x 3 remote = 6 policies
	if policyCount != 6 {
		t.Fatalf("expected 6 IPSec policies, got %d", policyCount)
	}
	if !hasDPD {
		t.Fatalf("expected DPD configured on profile")
	}
	if !hasFWFilter {
		t.Fatalf("expected firewall filter rules for IPSec")
	}
}

func TestConfigureDynamicPassivePeer(t *testing.T) {
	mock := &mockIPSecClient{}
	cfg := NewConfigurator(mock, func(string) {})

	err := cfg.Configure(IPSecConfig{
		PeerAddress:   "",
		PSK:           "SecretKey123",
		LocalSubnets:  []string{"192.168.10.0/24"},
		RemoteSubnets: []string{"172.16.1.0/24"},
		IsPassivePeer: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundPassivePeer := false
	foundLocalAddress := false
	for i, path := range mock.addedPaths {
		if path == "ip/ipsec/peer" {
			data := mock.addedData[i]
			if data["address"] == "0.0.0.0/0" && data["passive"] == "yes" {
				foundPassivePeer = true
				if data["local-address"] == "14.162.100.20" {
					foundLocalAddress = true
				}
			}
		}
	}
	if !foundPassivePeer {
		t.Fatalf("expected passive peer with address 0.0.0.0/0 not found")
	}
	if !foundLocalAddress {
		t.Fatalf("expected local-address 14.162.100.20 on passive peer, but was not found or mismatched")
	}
}

func TestGetCloudDDNS(t *testing.T) {
	mock := &mockIPSecClient{}
	cfg := NewConfigurator(mock, func(string) {})

	dns, err := cfg.GetCloudDDNS()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dns != "test123456.sn.mynetname.net" {
		t.Fatalf("expected test123456.sn.mynetname.net, got %s", dns)
	}
}
