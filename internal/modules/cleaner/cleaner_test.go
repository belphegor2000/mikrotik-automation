package cleaner

import (
	"testing"
)

type mockClnClient struct {
	removedCount int
}

func (m *mockClnClient) Connect(h, u, p string, port int, ssl bool) (string, error) { return "mock", nil }
func (m *mockClnClient) Disconnect() error                                         { return nil }
func (m *mockClnClient) IsConnected() bool                                         { return true }
func (m *mockClnClient) GetIdentity() string                                       { return "mock" }
func (m *mockClnClient) Get(path string) ([]map[string]interface{}, error) {
	if path == "ip/dhcp-server" {
		return []map[string]interface{}{
			{".id": "*1", "name": "dhcp-vlan10"},
		}, nil
	}
	if path == "ip/dhcp-server/lease" {
		return []map[string]interface{}{
			{".id": "*10", "server": "dhcp-vlan10"},
		}, nil
	}
	if path == "interface/bridge" {
		return []map[string]interface{}{
			{".id": "*b1", "name": "bridgeVLAN"},
		}, nil
	}
	return nil, nil
}
func (m *mockClnClient) GetOne(path string) (map[string]interface{}, error) { return nil, nil }
func (m *mockClnClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockClnClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockClnClient) Remove(path, id string) error {
	m.removedCount++
	return nil
}
func (m *mockClnClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockClnClient) RemoveByComment(path, comment string) (int, error) {
	m.removedCount++
	return 1, nil
}
func (m *mockClnClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockClnClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestCleanAll(t *testing.T) {
	mock := &mockClnClient{}
	logs := []string{}
	cln := NewCleaner(mock, func(msg string) { logs = append(logs, msg) })
	cln.CleanAll()

	if mock.removedCount == 0 {
		t.Fatalf("expected cleanup removals, got 0")
	}
	if len(logs) < 18 {
		t.Fatalf("expected >= 18 step logs, got %d", len(logs))
	}
}
