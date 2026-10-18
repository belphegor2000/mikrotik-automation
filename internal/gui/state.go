package gui

import (
	"sync"

	"mikrotik-automation/internal/api"
	"mikrotik-automation/internal/modules/vlan"
)

type AppState struct {
	Client    api.MikrotikClient
	Proto     string // "REST" or "SSH"
	LogPanel  *LogPanel
	vlans     []vlan.VLANEntry
	OnConnect []func()
	mu        sync.RWMutex
}

func NewAppState(logPanel *LogPanel) *AppState {
	return &AppState{
		Client:   api.NewRESTClient(),
		Proto:    "REST",
		LogPanel: logPanel,
		vlans:    make([]vlan.VLANEntry, 0),
	}
}

func (s *AppState) Log(msg string) {
	if s.LogPanel != nil {
		s.LogPanel.Log(msg)
	}
}

func (s *AppState) SetClient(c api.MikrotikClient, proto string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Client = c
	s.Proto = proto
}

func (s *AppState) GetClient() api.MikrotikClient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Client
}

// GetVLANs returns a thread-safe copy of the current VLAN entries.
func (s *AppState) GetVLANs() []vlan.VLANEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]vlan.VLANEntry, len(s.vlans))
	copy(result, s.vlans)
	return result
}

// SetVLANs safely updates the stored VLAN entries under write lock.
func (s *AppState) SetVLANs(entries []vlan.VLANEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entries == nil {
		s.vlans = make([]vlan.VLANEntry, 0)
		return
	}
	s.vlans = make([]vlan.VLANEntry, len(entries))
	copy(s.vlans, entries)
}
