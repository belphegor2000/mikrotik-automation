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
	VLANs     []vlan.VLANEntry
	OnConnect []func()
	mu        sync.RWMutex
}

func NewAppState(logPanel *LogPanel) *AppState {
	return &AppState{
		Client:   api.NewRESTClient(),
		Proto:    "REST",
		LogPanel: logPanel,
		VLANs:    make([]vlan.VLANEntry, 0),
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
