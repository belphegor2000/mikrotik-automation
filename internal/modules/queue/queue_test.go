package queue

import (
	"testing"
)

type mockQueueClient struct {
	items        []map[string]interface{}
	types        []map[string]interface{}
	added        map[string]interface{}
	removedID    string
	removeCalled bool
}

func (m *mockQueueClient) Connect(h, u, p string, port int, ssl bool) (string, error) { return "mock", nil }
func (m *mockQueueClient) Disconnect() error                                         { return nil }
func (m *mockQueueClient) IsConnected() bool                                         { return true }
func (m *mockQueueClient) GetIdentity() string                                       { return "mock" }
func (m *mockQueueClient) Get(path string) ([]map[string]interface{}, error) {
	if path == "queue/simple" {
		return m.items, nil
	}
	if path == "queue/type" {
		return m.types, nil
	}
	return nil, nil
}

func (m *mockQueueClient) GetOne(path string) (map[string]interface{}, error) { return nil, nil }
func (m *mockQueueClient) Add(path string, d map[string]interface{}) (map[string]interface{}, error) {
	m.added = d
	return d, nil
}
func (m *mockQueueClient) Set(path, id string, d map[string]interface{}) error { return nil }
func (m *mockQueueClient) Remove(path, id string) error {
	m.removedID = id
	m.removeCalled = true
	return nil
}
func (m *mockQueueClient) Run(path string, d map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockQueueClient) RemoveByComment(path, comment string) (int, error) {
	return 0, nil
}
func (m *mockQueueClient) RemoveAllMatching(path string, f map[string]string) (int, error) {
	return 0, nil
}
func (m *mockQueueClient) ExecRaw(cmd string) (string, error) { return "", nil }

func TestFormatRateLimit(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"10", "10M"},
		{"20M", "20M"},
		{"512k", "512K"},
		{"1g", "1G"},
		{"0", "0"},
		{"Unlimited", "0"},
		{"", "0"},
	}

	for _, tt := range tests {
		got := FormatRateLimit(tt.in)
		if got != tt.out {
			t.Errorf("FormatRateLimit(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

func TestParseMaxLimit(t *testing.T) {
	up, down := ParseMaxLimit("10000000/20000000")
	if up != "10M" || down != "20M" {
		t.Errorf("ParseMaxLimit bytes: got %s / %s, want 10M / 20M", up, down)
	}

	up2, down2 := ParseMaxLimit("5M/15M")
	if up2 != "5M" || down2 != "15M" {
		t.Errorf("ParseMaxLimit strings: got %s / %s, want 5M / 15M", up2, down2)
	}

	up3, down3 := ParseMaxLimit("0/0")
	if up3 != "Unlimited" || down3 != "Unlimited" {
		t.Errorf("ParseMaxLimit 0: got %s / %s, want Unlimited / Unlimited", up3, down3)
	}
}

func TestQueueOperations(t *testing.T) {
	mock := &mockQueueClient{
		items: []map[string]interface{}{
			{
				".id":       "*1",
				"name":      "queue-user1",
				"target":    "192.168.10.50/32",
				"max-limit": "5000000/10000000",
				"comment":   "[MK-AUTO]",
			},
		},
	}

	cfg := NewConfigurator(mock, func(string) {})

	// Test List
	queues, err := cfg.ListQueues()
	if err != nil {
		t.Fatalf("ListQueues failed: %v", err)
	}
	if len(queues) != 1 {
		t.Fatalf("Expected 1 queue, got %d", len(queues))
	}
	if queues[0].Name != "queue-user1" || queues[0].UploadLimit != "5M" || queues[0].DownloadLimit != "10M" {
		t.Errorf("Unexpected queue item values: %+v", queues[0])
	}

	// Test Add
	err = cfg.AddQueue(QueueItem{
		Name:          "VLAN20-Limit",
		Target:        "192.168.20.0/24",
		UploadLimit:   "10M",
		DownloadLimit: "30M",
	})
	if err != nil {
		t.Fatalf("AddQueue failed: %v", err)
	}
	if mock.added["name"] != "VLAN20-Limit" {
		t.Errorf("Expected added name VLAN20-Limit, got %v", mock.added["name"])
	}
	if mock.added["max-limit"] != "10M/30M" {
		t.Errorf("Expected added max-limit 10M/30M, got %v", mock.added["max-limit"])
	}
	if mock.added["comment"] != "[MK-AUTO]" {
		t.Errorf("Expected added comment [MK-AUTO], got %v", mock.added["comment"])
	}

	// Test Delete
	err = cfg.DeleteQueue("*1")
	if err != nil {
		t.Fatalf("DeleteQueue failed: %v", err)
	}
	if !mock.removeCalled || mock.removedID != "*1" {
		t.Errorf("Expected remove ID *1, got %s", mock.removedID)
	}
}

func TestCakeQoS(t *testing.T) {
	mock := &mockQueueClient{
		types: []map[string]interface{}{},
		items: []map[string]interface{}{},
	}
	cfg := NewConfigurator(mock, func(string) {})

	// Test 1: Configure CAKE QoS with 150M down, 150M up
	err := cfg.ConfigureCakeQoS(150, 150)
	if err != nil {
		t.Fatalf("ConfigureCakeQoS failed: %v", err)
	}

	// 95% of 150M = 142500k
	expectedMax := "142500k/142500k"
	if mock.added["name"] != CakeQueueName || mock.added["max-limit"] != expectedMax {
		t.Errorf("expected queue '%s' with max-limit '%s', got: %+v", CakeQueueName, expectedMax, mock.added)
	}

	// Test 2: Status check
	mock.items = append(mock.items, mock.added)
	active, limit, err := cfg.GetCakeStatus()
	if err != nil || !active || limit != expectedMax {
		t.Errorf("GetCakeStatus failed, active: %v, limit: %s", active, limit)
	}

	// Test 3: Remove
	mock.items[0][".id"] = "*cake1"
	err = cfg.RemoveCakeQoS()
	if err != nil {
		t.Fatalf("RemoveCakeQoS failed: %v", err)
	}
	if !mock.removeCalled || mock.removedID != "*cake1" {
		t.Errorf("expected remove called for *cake1")
	}

	// Test 4: Configure Dual WAN CAKE QoS (WAN1 150M/150M, WAN2 100M/100M)
	mock.items = []map[string]interface{}{}
	err = cfg.ConfigureCakeQoSDual(150, 150, 100, 100, "pppoe-wan1", "pppoe-wan2")
	if err != nil {
		t.Fatalf("ConfigureCakeQoSDual failed: %v", err)
	}

	// Simulate adding both queues
	mock.items = []map[string]interface{}{
		{".id": "*cakeW1", "name": CakeQueueNameWAN1, "max-limit": "142500k/142500k"},
		{".id": "*cakeW2", "name": CakeQueueNameWAN2, "max-limit": "95000k/95000k"},
	}
	activeDual, limitDual, err := cfg.GetCakeStatus()
	if err != nil || !activeDual || limitDual == "" {
		t.Errorf("GetCakeStatus Dual failed: active=%v, limit=%s", activeDual, limitDual)
	}
}

