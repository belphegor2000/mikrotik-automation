package bridge

import (
	"testing"
)

func TestParseBridgeCIDR(t *testing.T) {
	tests := []struct {
		input       string
		wantRouter  string
		wantGateway string
		wantStart   string
		wantEnd     string
		wantNet     string
		wantPrefix  int
		wantErr     bool
	}{
		{
			input:       "192.168.88.1/24",
			wantRouter:  "192.168.88.1",
			wantGateway: "192.168.88.1",
			wantStart:   "192.168.88.10",
			wantEnd:     "192.168.88.254",
			wantNet:     "192.168.88.0",
			wantPrefix:  24,
			wantErr:     false,
		},
		{
			input:       "10.0.0.1/24",
			wantRouter:  "10.0.0.1",
			wantGateway: "10.0.0.1",
			wantStart:   "10.0.0.10",
			wantEnd:     "10.0.0.254",
			wantNet:     "10.0.0.0",
			wantPrefix:  24,
			wantErr:     false,
		},
		{
			input:       "172.16.5.1/24",
			wantRouter:  "172.16.5.1",
			wantGateway: "172.16.5.1",
			wantStart:   "172.16.5.10",
			wantEnd:     "172.16.5.254",
			wantNet:     "172.16.5.0",
			wantPrefix:  24,
			wantErr:     false,
		},
		{
			input:       "10.20.0.0/16",
			wantRouter:  "10.20.0.1",
			wantGateway: "10.20.0.1",
			wantStart:   "10.20.0.10",
			wantEnd:     "10.20.255.254",
			wantNet:     "10.20.0.0",
			wantPrefix:  16,
			wantErr:     false,
		},
		{
			// Without /24 suffix
			input:       "10.10.10.1",
			wantRouter:  "10.10.10.1",
			wantGateway: "10.10.10.1",
			wantStart:   "10.10.10.10",
			wantEnd:     "10.10.10.254",
			wantNet:     "10.10.10.0",
			wantPrefix:  24,
			wantErr:     false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "invalid-ip",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			rIP, gw, start, end, net, prefix, err := ParseBridgeCIDR(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseBridgeCIDR(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if rIP != tt.wantRouter {
				t.Errorf("routerIP = %s, want %s", rIP, tt.wantRouter)
			}
			if gw != tt.wantGateway {
				t.Errorf("gateway = %s, want %s", gw, tt.wantGateway)
			}
			if start != tt.wantStart {
				t.Errorf("startIP = %s, want %s", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("endIP = %s, want %s", end, tt.wantEnd)
			}
			if net != tt.wantNet {
				t.Errorf("network = %s, want %s", net, tt.wantNet)
			}
			if prefix != tt.wantPrefix {
				t.Errorf("prefixLen = %d, want %d", prefix, tt.wantPrefix)
			}
		})
	}
}
