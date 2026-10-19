package nat

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"mikrotik-automation/internal/api"
)

const (
	AppComment       = "[MK-AUTO]"
	PFCommentPrefix  = "[MK-AUTO] PF: "
	HairpinComment   = "[MK-AUTO] Hairpin NAT"
)

type PortForwardRule struct {
	ID            string `json:".id,omitempty"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol"` // "tcp" or "udp"
	DstPort       string `json:"dst_port"` // "80", "8080", "8000-8010"
	ToAddresses   string `json:"to_addresses"`
	ToPorts       string `json:"to_ports"`
	WANInterface  string `json:"in_interface,omitempty"` // "pppoe-wan1", "pppoe-wan2", "ether1", or ""
	EnableHairpin bool   `json:"enable_hairpin"`
	Comment       string `json:"comment,omitempty"`
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

// ValidatePort validates single port, comma-separated ports, or port ranges (e.g. "80", "80,443", "8000-8010").
func ValidatePort(portStr string) error {
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		return fmt.Errorf("cổng (port) không được để trống")
	}
	parts := strings.Split(portStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.Contains(p, "-") {
			rangeParts := strings.Split(p, "-")
			if len(rangeParts) != 2 {
				return fmt.Errorf("dải cổng '%s' không đúng định dạng min-max", p)
			}
			p1, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			p2, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 != nil || err2 != nil || p1 < 1 || p1 > 65535 || p2 < 1 || p2 > 65535 || p1 > p2 {
				return fmt.Errorf("dải cổng '%s' không hợp lệ (1-65535)", p)
			}
		} else {
			num, err := strconv.Atoi(p)
			if err != nil || num < 1 || num > 65535 {
				return fmt.Errorf("cổng '%s' không hợp lệ (1-65535)", p)
			}
		}
	}
	return nil
}

// DeriveSubnet takes an IP (or CIDR) and returns the /24 CIDR network string.
func DeriveSubnet(ipStr string) (string, error) {
	ipStr = strings.TrimSpace(ipStr)
	if strings.Contains(ipStr, "/") {
		_, ipNet, err := net.ParseCIDR(ipStr)
		if err != nil {
			return "", err
		}
		return ipNet.String(), nil
	}
	parsed := net.ParseIP(ipStr)
	if parsed == nil {
		return "", fmt.Errorf("địa chỉ IP '%s' không hợp lệ", ipStr)
	}
	v4 := parsed.To4()
	if v4 == nil {
		return "", fmt.Errorf("chỉ hỗ trợ địa chỉ IPv4")
	}
	// Default to /24 class subnet
	mask := net.CIDRMask(24, 32)
	network := v4.Mask(mask)
	return fmt.Sprintf("%s/24", network.String()), nil
}

// ListPortForwards retrieves all Port Forwarding rules created by MK-AUTO.
func (c *Configurator) ListPortForwards() ([]PortForwardRule, error) {
	rules, err := c.client.Get("ip/firewall/nat")
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách NAT: %w", err)
	}

	var results []PortForwardRule
	for _, r := range rules {
		chain, _ := r["chain"].(string)
		comment, _ := r["comment"].(string)

		if chain == "dstnat" && strings.HasPrefix(comment, PFCommentPrefix) {
			id, _ := r[".id"].(string)
			proto, _ := r["protocol"].(string)
			dstPort, _ := r["dst-port"].(string)
			toAddr, _ := r["to-addresses"].(string)
			toPorts, _ := r["to-ports"].(string)
			name := strings.TrimPrefix(comment, PFCommentPrefix)
			inIface, _ := r["in-interface"].(string)

			results = append(results, PortForwardRule{
				ID:           id,
				Name:         name,
				Protocol:     proto,
				DstPort:      dstPort,
				ToAddresses:  toAddr,
				ToPorts:      toPorts,
				WANInterface: inIface,
				Comment:      comment,
			})
		}
	}
	return results, nil
}

// AddPortForward creates a new dst-nat rule and optionally configures Hairpin NAT.
func (c *Configurator) AddPortForward(rule PortForwardRule) error {
	rule.Name = strings.TrimSpace(rule.Name)
	rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
	rule.DstPort = strings.TrimSpace(rule.DstPort)
	rule.ToAddresses = strings.TrimSpace(rule.ToAddresses)
	rule.ToPorts = strings.TrimSpace(rule.ToPorts)
	rule.WANInterface = strings.TrimSpace(rule.WANInterface)

	if rule.Name == "" {
		return fmt.Errorf("tên dịch vụ mở cổng không được để trống")
	}
	if rule.Protocol != "tcp" && rule.Protocol != "udp" {
		rule.Protocol = "tcp"
	}
	if err := ValidatePort(rule.DstPort); err != nil {
		return fmt.Errorf("cổng ngoài (Dst. Port) không hợp lệ: %w", err)
	}
	if parsed := net.ParseIP(rule.ToAddresses); parsed == nil || parsed.To4() == nil {
		return fmt.Errorf("địa chỉ IP nội bộ '%s' không phải IPv4 hợp lệ", rule.ToAddresses)
	}
	if rule.ToPorts != "" {
		if err := ValidatePort(rule.ToPorts); err != nil {
			return fmt.Errorf("cổng đích nội bộ (To Ports) không hợp lệ: %w", err)
		}
	}

	natData := map[string]interface{}{
		"chain":        "dstnat",
		"action":       "dst-nat",
		"protocol":     rule.Protocol,
		"dst-port":     rule.DstPort,
		"to-addresses": rule.ToAddresses,
		"comment":      fmt.Sprintf("%s%s", PFCommentPrefix, rule.Name),
	}
	if rule.WANInterface != "" && !strings.Contains(strings.ToLower(rule.WANInterface), "tất cả") && !strings.EqualFold(rule.WANInterface, "all") {
		natData["in-interface"] = rule.WANInterface
	} else {
		natData["dst-address-type"] = "local"
	}
	if rule.ToPorts != "" && rule.ToPorts != rule.DstPort {
		natData["to-ports"] = rule.ToPorts
	}

	c.log(fmt.Sprintf("▶ Đang mở cổng %s (%s %s ➔ %s:%s)...",
		rule.Name, rule.Protocol, rule.DstPort, rule.ToAddresses,
		func() string {
			if rule.ToPorts != "" {
				return rule.ToPorts
			}
			return rule.DstPort
		}()))

	_, err := c.client.Add("ip/firewall/nat", natData)
	if err != nil {
		return fmt.Errorf("không thể thêm luật dst-nat: %w", err)
	}
	c.log(fmt.Sprintf("✓ Đã tạo thành công luật Mở Port cho: %s", rule.Name))

	if rule.EnableHairpin {
		subnet, err := DeriveSubnet(rule.ToAddresses)
		if err == nil && subnet != "" {
			_ = c.EnsureHairpinNAT(subnet)
		}
	}

	return nil
}

// DeletePortForward removes a Port Forwarding rule by its RouterOS .id.
func (c *Configurator) DeletePortForward(id string) error {
	if id == "" {
		return fmt.Errorf("ID luật NAT không hợp lệ")
	}
	err := c.client.Remove("ip/firewall/nat", id)
	if err != nil {
		return fmt.Errorf("lỗi xoá luật NAT: %w", err)
	}
	c.log(fmt.Sprintf("✓ Đã xoá thành công luật NAT (ID: %s)", id))
	return nil
}

// EnsureHairpinNAT ensures a masquerade rule exists for the specified LAN subnet.
func (c *Configurator) EnsureHairpinNAT(subnet string) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		return nil
	}

	rules, err := c.client.Get("ip/firewall/nat")
	if err == nil {
		for _, r := range rules {
			chain, _ := r["chain"].(string)
			action, _ := r["action"].(string)
			src, _ := r["src-address"].(string)
			dst, _ := r["dst-address"].(string)

			if chain == "srcnat" && action == "masquerade" && src == subnet && dst == subnet {
				c.log(fmt.Sprintf("  • Hairpin NAT cho dải %s đã tồn tại, bỏ qua tạo mới.", subnet))
				return nil
			}
		}
	}

	c.log(fmt.Sprintf("▶ Đang tạo luật Hairpin NAT cho subnet: %s...", subnet))
	data := map[string]interface{}{
		"chain":       "srcnat",
		"action":      "masquerade",
		"src-address": subnet,
		"dst-address": subnet,
		"comment":     fmt.Sprintf("%s - %s", HairpinComment, subnet),
	}

	_, err = c.client.Add("ip/firewall/nat", data)
	if err != nil {
		c.log(fmt.Sprintf("⚠ Cảnh báo: Lỗi cấu hình Hairpin NAT (%s): %v", subnet, err))
		return err
	}
	c.log(fmt.Sprintf("✓ Đã cấu hình Hairpin NAT thành công cho dải mạng: %s", subnet))
	return nil
}

// DetectLANSubnets retrieves all local LAN subnets configured on Bridge or VLAN interfaces.
func (c *Configurator) DetectLANSubnets() ([]string, error) {
	addrs, err := c.client.Get("ip/address")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var subnets []string

	for _, a := range addrs {
		iface, _ := a["interface"].(string)
		addrStr, _ := a["address"].(string)

		// Exclude WAN interfaces
		if strings.HasPrefix(iface, "pppoe") || iface == "ether1" || iface == "ether2" {
			continue
		}

		if addrStr != "" {
			sub, err := DeriveSubnet(addrStr)
			if err == nil && !seen[sub] {
				seen[sub] = true
				subnets = append(subnets, sub)
			}
		}
	}
	return subnets, nil
}

// EnableHairpinAllSubnets detects all LAN subnets (and any provided custom subnets) and applies Hairpin NAT.
func (c *Configurator) EnableHairpinAllSubnets(customSubnets ...string) (int, error) {
	seen := make(map[string]bool)
	var finalSubnets []string

	for _, s := range customSubnets {
		s = strings.TrimSpace(s)
		if s != "" {
			sub, err := DeriveSubnet(s)
			if err == nil && !seen[sub] {
				seen[sub] = true
				finalSubnets = append(finalSubnets, sub)
			}
		}
	}

	detected, err := c.DetectLANSubnets()
	if err == nil {
		for _, s := range detected {
			if !seen[s] {
				seen[s] = true
				finalSubnets = append(finalSubnets, s)
			}
		}
	}

	if len(finalSubnets) == 0 {
		finalSubnets = []string{"192.168.88.0/24"}
	}

	count := 0
	for _, sub := range finalSubnets {
		if err := c.EnsureHairpinNAT(sub); err == nil {
			count++
		}
	}
	return count, nil
}

