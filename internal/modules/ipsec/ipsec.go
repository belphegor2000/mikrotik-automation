package ipsec

import (
	"fmt"
	"net"
	"strings"

	"mikrotik-automation/internal/api"
)

const (
	AppComment        = "[MK-AUTO]"
	LanAddressList    = "LAN-Address"
	RemoteAddressList = "Remote-Address"
)

type IPSecConfig struct {
	PeerAddress   string   `json:"peer_address"`
	PSK           string   `json:"psk"`
	LocalSubnets  []string `json:"local_subnets"`
	RemoteSubnets []string `json:"remote_subnets"`
	IsPassivePeer bool     `json:"is_passive_peer"`
	LocalSubnet   string   `json:"local_subnet,omitempty"`
	RemoteSubnet  string   `json:"remote_subnet,omitempty"`
}

type Configurator struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewConfigurator(client api.MikrotikClient, logFn func(string)) *Configurator {
	return &Configurator{client: client, logFn: logFn}
}

func (c *Configurator) log(msg string) {
	if c.logFn != nil {
		c.logFn(msg)
	}
}

// ParseAndNormalizeSubnets splits input by commas, semicolons, or newlines,
// parses and normalizes each subnet to network CIDR (e.g. 192.168.10.1/24 -> 192.168.10.0/24),
// and removes duplicates.
func ParseAndNormalizeSubnets(input string) ([]string, error) {
	replacer := strings.NewReplacer(";", ",", "\n", ",", "\r", ",", " ", "")
	cleaned := replacer.Replace(input)
	tokens := strings.Split(cleaned, ",")

	seen := make(map[string]bool)
	var result []string

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		ip, ipNet, err := net.ParseCIDR(token)
		if err != nil {
			return nil, fmt.Errorf("dải mạng không hợp lệ '%s': %w", token, err)
		}
		if ip.To4() == nil {
			return nil, fmt.Errorf("chỉ hỗ trợ IPv4: '%s'", token)
		}
		ones, _ := ipNet.Mask.Size()
		normalized := fmt.Sprintf("%s/%d", ipNet.IP.String(), ones)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("không tìm thấy dải mạng hợp lệ")
	}
	return result, nil
}

// DetectLocalSubnets scans all IP addresses on bridgeVLAN and vlan* interfaces
func (c *Configurator) DetectLocalSubnets() ([]string, error) {
	addrs, err := c.client.Get("ip/address")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var result []string

	for _, a := range addrs {
		iface, _ := a["interface"].(string)
		addr, _ := a["address"].(string)
		if iface == "bridgeVLAN" || strings.HasPrefix(iface, "vlan") {
			_, ipNet, err := net.ParseCIDR(addr)
			if err == nil && ipNet != nil {
				ones, _ := ipNet.Mask.Size()
				netCIDR := fmt.Sprintf("%s/%d", ipNet.IP.String(), ones)
				if !seen[netCIDR] {
					seen[netCIDR] = true
					result = append(result, netCIDR)
				}
			}
		}
	}
	if len(result) == 0 {
		return []string{"192.168.10.0/24"}, nil
	}
	return result, nil
}

// DetectLocalSubnet returns the first detected local subnet (for backward compatibility)
func (c *Configurator) DetectLocalSubnet() (string, error) {
	subnets, err := c.DetectLocalSubnets()
	if err != nil || len(subnets) == 0 {
		return "192.168.10.0/24", err
	}
	return subnets[0], nil
}

// GetCloudDDNS retrieves the MikroTik Cloud DDNS hostname of this router
func (c *Configurator) GetCloudDDNS() (string, error) {
	_ = c.client.Set("ip/cloud", "", map[string]interface{}{
		"ddns-enabled": "yes",
	})
	cloud, err := c.client.GetOne("ip/cloud")
	if err != nil {
		return "", err
	}
	if dns, ok := cloud["dns-name"].(string); ok && dns != "" {
		return dns, nil
	}
	return "", fmt.Errorf("chưa có dns-name từ MikroTik Cloud")
}

func (c *Configurator) Configure(cfg IPSecConfig) error {
	localList := cfg.LocalSubnets
	if len(localList) == 0 && cfg.LocalSubnet != "" {
		parsed, err := ParseAndNormalizeSubnets(cfg.LocalSubnet)
		if err == nil {
			localList = parsed
		} else {
			localList = []string{cfg.LocalSubnet}
		}
	}
	remoteList := cfg.RemoteSubnets
	if len(remoteList) == 0 && cfg.RemoteSubnet != "" {
		parsed, err := ParseAndNormalizeSubnets(cfg.RemoteSubnet)
		if err == nil {
			remoteList = parsed
		} else {
			remoteList = []string{cfg.RemoteSubnet}
		}
	}

	if len(localList) == 0 {
		return fmt.Errorf("chưa có dải mạng nội bộ (Local Subnets)")
	}
	if len(remoteList) == 0 {
		return fmt.Errorf("chưa có dải mạng đầu xa (Remote Subnets)")
	}

	peerAddr := strings.TrimSpace(cfg.PeerAddress)
	if cfg.IsPassivePeer || peerAddr == "" {
		peerAddr = "0.0.0.0/0"
	}

	profileName := "profile-site2site"
	proposalName := "proposal-site2site"
	peerName := "peer-site2site"

	// 1. Profile với DPD Fast Reconnect tối ưu cho IP động
	c.log("  ▶ [1/8] Tạo IPSec Profile (IKEv2, AES-256, SHA256, DPD Fast Reconnect)...")
	_, err := c.client.Add("ip/ipsec/profile", map[string]interface{}{
		"name":                 profileName,
		"hash-algorithm":       "sha256",
		"enc-algorithm":        "aes-256",
		"dh-group":             "modp2048",
		"nat-traversal":        "yes",
		"dpd-interval":         "10s",
		"dpd-maximum-failures": 3,
		"comment":              AppComment,
	})
	if err != nil {
		c.log(fmt.Sprintf("    ⚠ Không thể tạo Profile: %v", err))
	}

	// 2. Proposal
	c.log("  ▶ [2/8] Tạo IPSec Proposal (AES-256-CBC, SHA256, PFS)...")
	_, err = c.client.Add("ip/ipsec/proposal", map[string]interface{}{
		"name":            proposalName,
		"auth-algorithms": "sha256",
		"enc-algorithms":  "aes-256-cbc",
		"pfs-group":       "modp2048",
		"comment":         AppComment,
	})
	if err != nil {
		c.log(fmt.Sprintf("    ⚠ Không thể tạo Proposal: %v", err))
	}

	// 3. Peer (Hỗ trợ Static IP, DDNS FQDN, hoặc Passive Responder 0.0.0.0/0)
	peerDesc := peerAddr
	if peerAddr == "0.0.0.0/0" {
		peerDesc = "0.0.0.0/0 (Passive Responder - Chờ đầu xa kết nối)"
	}
	c.log(fmt.Sprintf("  ▶ [3/8] Tạo IPSec Peer tới '%s'...", peerDesc))
	peerData := map[string]interface{}{
		"name":          peerName,
		"address":       peerAddr,
		"profile":       profileName,
		"exchange-mode": "ike2",
		"comment":       AppComment,
	}
	if cfg.IsPassivePeer || peerAddr == "0.0.0.0/0" {
		peerData["passive"] = "yes"
		peerData["send-initial-contact"] = "no"
	}
	_, err = c.client.Add("ip/ipsec/peer", peerData)
	if err != nil {
		c.log(fmt.Sprintf("    ⚠ Không thể tạo Peer: %v", err))
	}

	// 4. Identity
	c.log("  ▶ [4/8] Tạo IPSec Identity (Pre-Shared Key)...")
	_, err = c.client.Add("ip/ipsec/identity", map[string]interface{}{
		"peer":        peerName,
		"auth-method": "pre-shared-key",
		"secret":      cfg.PSK,
		"comment":     AppComment,
	})
	if err != nil {
		c.log(fmt.Sprintf("    ⚠ Không thể tạo Identity: %v", err))
	}

	// 5. Mở Firewall Filter cho IKEv2 & NAT-T & ESP
	c.log("  ▶ [5/8] Mở cổng Firewall Filter cho IKEv2 (UDP 500, 4500 & ESP)...")
	_, _ = c.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":    "input",
		"action":   "accept",
		"protocol": "udp",
		"dst-port": "500,4500",
		"comment":  AppComment + " Allow IKEv2 & NAT-T",
	})
	_, _ = c.client.Add("ip/firewall/filter", map[string]interface{}{
		"chain":    "input",
		"action":   "accept",
		"protocol": "ipsec-esp",
		"comment":  AppComment + " Allow IPSec ESP",
	})

	// 6. Cập nhật Address List (LAN-Address & Remote-Address)
	c.log(fmt.Sprintf("  ▶ [6/8] Cập nhật Address Lists ('%s' & '%s')...", LanAddressList, RemoteAddressList))
	for _, local := range localList {
		c.log(fmt.Sprintf("    + Thêm %s vào list '%s'...", local, LanAddressList))
		_, _ = c.client.Add("ip/firewall/address-list", map[string]interface{}{
			"list":    LanAddressList,
			"address": local,
			"comment": AppComment,
		})
	}
	for _, remote := range remoteList {
		c.log(fmt.Sprintf("    + Thêm %s vào list '%s'...", remote, RemoteAddressList))
		_, _ = c.client.Add("ip/firewall/address-list", map[string]interface{}{
			"list":    RemoteAddressList,
			"address": remote,
			"comment": AppComment,
		})
	}

	// 7. Mangle & NAT Bypass rules sử dụng LAN-Address & Remote-Address
	c.log(fmt.Sprintf("  ▶ [7/8] Tạo Mangle & NAT Bypass rules sử dụng '%s' & '%s'...", LanAddressList, RemoteAddressList))
	// Mangle Bypass PCC (Chiều đi: LAN-Address -> Remote-Address)
	_, _ = c.client.Add("ip/firewall/mangle", map[string]interface{}{
		"chain":            "prerouting",
		"src-address-list": LanAddressList,
		"dst-address-list": RemoteAddressList,
		"action":           "accept",
		"comment":          AppComment + " IPSec Bypass PCC",
	})
	// Mangle Bypass PCC (Chiều về: Remote-Address -> LAN-Address)
	_, _ = c.client.Add("ip/firewall/mangle", map[string]interface{}{
		"chain":            "prerouting",
		"src-address-list": RemoteAddressList,
		"dst-address-list": LanAddressList,
		"action":           "accept",
		"comment":          AppComment + " IPSec Bypass PCC",
	})
	// NAT Bypass (LAN-Address -> Remote-Address)
	_, _ = c.client.Add("ip/firewall/nat", map[string]interface{}{
		"chain":            "srcnat",
		"src-address-list": LanAddressList,
		"dst-address-list": RemoteAddressList,
		"action":           "accept",
		"comment":          AppComment + " IPSec NAT Bypass",
	})
	c.log("    ✓ Đã tạo 2 Mangle Bypass rules và 1 NAT Bypass rule.")

	// 8. IPSec Policies cho toàn bộ cặp (Local, Remote)
	c.log(fmt.Sprintf("  ▶ [8/8] Tạo IPSec Policies (%d dải nội bộ x %d dải đầu xa)...", len(localList), len(remoteList)))
	policyCount := 0
	for _, local := range localList {
		for _, remote := range remoteList {
			_, err := c.client.Add("ip/ipsec/policy", map[string]interface{}{
				"peer":        peerName,
				"src-address": local,
				"dst-address": remote,
				"tunnel":      "yes",
				"action":      "encrypt",
				"proposal":    proposalName,
				"comment":     AppComment,
			})
			if err == nil {
				policyCount++
			} else {
				c.log(fmt.Sprintf("    ⚠ Không thể tạo Policy (%s <-> %s): %v", local, remote, err))
			}
		}
	}
	c.log(fmt.Sprintf("    ✓ Đã tạo %d IPSec Policies.", policyCount))

	c.log(fmt.Sprintf("✓ Cấu hình hoàn tất! Đường hầm IPSec kết nối %d VLAN nội bộ và %d VLAN đầu xa.", len(localList), len(remoteList)))
	return nil
}
