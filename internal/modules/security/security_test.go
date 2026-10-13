package security

import (
	"testing"
)

type mockSecClient struct {
	disabledCount int
	addedRules    int
	addedPaths    []string
	addedData     []map[string]interface{}
}

func (m *mockSecClient) Connect(h, u, p string, port int, ssl bool) (string, error) { return "mock", nil }
func (m *mockSecClient) Disconnect() error                                         { return nil }
func (m *mockSecClient) IsConnected() bool                                         { return true }
func (m *mockSecClient) GetIdentity() string                                       { return "mock" }
func (m *mockSecClient) Get(path string) ([]map[string]interface{}, error) {
	if path == "ip/service" {
		return []map[string]interface{}{
			{".id": "*1", "name": "api", "disabled": "no"},
			{".id": "*2", "name": "telnet", "disabled": "no"},
			{".id": "*3", "name": "winbox", "disabled": "no"},
		}, nil
	}
	return nil, nil
}
func (m *mockSecClient) GetOne(path string) (map[string]interface{}, error) { return nil, nil }
func (m *mockSecClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.addedRules++
	m.addedPaths = append(m.addedPaths, path)
	m.addedData = append(m.addedData, d)
	return nil, nil
}
func (m *mockSecClient) Set(path, id string, d map[string]interface{}) error {
	m.disabledCount++
	return nil
}
func (m *mockSecClient) Remove(path, id string) error                        { return nil }
func (m *mockSecClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockSecClient) RemoveByComment(path, comment string) (int, error)       { return 0, nil }
func (m *mockSecClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockSecClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestDisableInsecureServices(t *testing.T) {
	mock := &mockSecClient{}
	cfg := NewConfigurator(mock, func(string) {})
	count, err := cfg.DisableInsecureServices([]string{"api", "telnet"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 disabled services, got %d", count)
	}
}

func TestConfigureFirewallFilters(t *testing.T) {
	mock := &mockSecClient{}
	cfg := NewConfigurator(mock, func(string) {})
	err := cfg.ConfigureFirewallFilters([]string{"pppoe-wan1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.addedRules < 6 {
		t.Fatalf("expected >= 6 rules, got %d", mock.addedRules)
	}
}

func TestConfigureBruteForceProtection(t *testing.T) {
	mock := &mockSecClient{}
	cfg := NewConfigurator(mock, func(string) {})
	err := cfg.ConfigureBruteForceProtection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundScript := false
	foundScheduler := false
	foundDropRule := false

	for i, path := range mock.addedPaths {
		if path == "system/script" {
			if name, ok := mock.addedData[i]["name"].(string); ok && name == BruteForceScriptName {
				foundScript = true
			}
		}
		if path == "system/scheduler" {
			if interval, ok := mock.addedData[i]["interval"].(string); ok && interval == "5m" {
				foundScheduler = true
			}
		}
		if path == "ip/firewall/filter" {
			if action, ok := mock.addedData[i]["action"].(string); ok && action == "drop" {
				if list, ok := mock.addedData[i]["src-address-list"].(string); ok && list == BlacklistAddressList {
					foundDropRule = true
				}
			}
		}
	}

	if !foundScript {
		t.Fatalf("expected script '%s' was not added", BruteForceScriptName)
	}
	if !foundScheduler {
		t.Fatalf("expected scheduler '%s' was not added", BruteForceSchedulerName)
	}
	if !foundDropRule {
		t.Fatalf("expected firewall drop rule for blacklist was not added")
	}
}

