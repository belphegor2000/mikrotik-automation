package cleaner

import (
	"fmt"
	"strings"

	"mikrotik-automation/internal/api"
)

const (
	AppComment = "[MK-AUTO]"
	BridgeName = "bridgeVLAN"
	PPPoEWAN1  = "pppoe-wan1"
	PPPoEWAN2  = "pppoe-wan2"
	RTWAN1     = "to_wan1"
	RTWAN2     = "to_wan2"
)

type Cleaner struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewCleaner(client api.MikrotikClient, logFn func(string)) *Cleaner {
	return &Cleaner{client: client, logFn: logFn}
}

func (c *Cleaner) log(msg string) {
	if c.logFn != nil {
		c.logFn(msg)
	}
}

func (c *Cleaner) CleanAll() {
	steps := []struct {
		name string
		fn   func() (int, error)
	}{
		{"Xoá DHCP Leases...", c.cleanDHCPLeases},
		{"Xoá DHCP Networks...", c.cleanDHCPNetworks},
		{"Xoá DHCP Servers...", c.cleanDHCPServers},
		{"Xoá DHCP Pools...", c.cleanDHCPPools},
		{"Xoá IP Addresses (VLAN/Bridge)...", c.cleanIPAddresses},
		{"Xoá Bridge VLAN table...", c.cleanBridgeVLANs},
		{"Xoá VLAN interfaces...", c.cleanVLANInterfaces},
		{"Xoá Bridge ports & Bridge...", c.cleanBridge},
		{"Xoá PPPoE clients...", c.cleanPPPoE},
		{"Xoá Firewall Filter rules...", c.cleanFirewallFilters},
		{"Xoá Firewall Mangle rules...", c.cleanMangle},
		{"Xoá Firewall NAT rules...", c.cleanNAT},
		{"Xoá Firewall Address-List...", c.cleanAddressList},
		{"Xoá System Scripts (MK-AUTO)...", c.cleanScripts},
		{"Xoá System Schedulers (MK-AUTO)...", c.cleanSchedulers},
		{"Xoá IP Routes (custom)...", c.cleanRoutes},
		{"Xoá Routing Tables...", c.cleanRoutingTables},
		{"Xoá IPSec configuration...", c.cleanIPSec},
	}

	for _, step := range steps {
		c.log(fmt.Sprintf("  ▶ %s", step.name))
		count, err := step.fn()
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi: %v", err))
		} else {
			c.log(fmt.Sprintf("    ✓ Đã xoá %d mục.", count))
		}
	}
}

func (c *Cleaner) cleanDHCPLeases() (int, error) {
	servers, err := c.client.Get("ip/dhcp-server")
	if err != nil {
		return 0, err
	}
	appServers := make(map[string]bool)
	for _, s := range servers {
		if name, ok := s["name"].(string); ok && strings.HasPrefix(name, "dhcp-") {
			appServers[name] = true
		}
	}

	leases, err := c.client.Get("ip/dhcp-server/lease")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, l := range leases {
		srv, _ := l["server"].(string)
		if appServers[srv] {
			if id, ok := l[".id"].(string); ok && id != "" {
				if err := c.client.Remove("ip/dhcp-server/lease", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanDHCPNetworks() (int, error) {
	return c.client.RemoveByComment("ip/dhcp-server/network", AppComment)
}

func (c *Cleaner) cleanDHCPServers() (int, error) {
	servers, err := c.client.Get("ip/dhcp-server")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, s := range servers {
		if name, ok := s["name"].(string); ok && strings.HasPrefix(name, "dhcp-") {
			if id, ok := s[".id"].(string); ok && id != "" {
				if err := c.client.Remove("ip/dhcp-server", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanDHCPPools() (int, error) {
	pools, err := c.client.Get("ip/pool")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, p := range pools {
		if name, ok := p["name"].(string); ok && strings.HasPrefix(name, "pool-") {
			if id, ok := p[".id"].(string); ok && id != "" {
				if err := c.client.Remove("ip/pool", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanIPAddresses() (int, error) {
	addrs, err := c.client.Get("ip/address")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, a := range addrs {
		iface, _ := a["interface"].(string)
		cmt, _ := a["comment"].(string)
		if iface == BridgeName || strings.HasPrefix(iface, "vlan") || cmt == AppComment {
			if id, ok := a[".id"].(string); ok && id != "" {
				if err := c.client.Remove("ip/address", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanBridgeVLANs() (int, error) {
	vlans, err := c.client.Get("interface/bridge/vlan")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, v := range vlans {
		br, _ := v["bridge"].(string)
		if br == BridgeName {
			if id, ok := v[".id"].(string); ok && id != "" {
				if err := c.client.Remove("interface/bridge/vlan", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanVLANInterfaces() (int, error) {
	vlans, err := c.client.Get("interface/vlan")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, v := range vlans {
		iface, _ := v["interface"].(string)
		name, _ := v["name"].(string)
		if iface == BridgeName || strings.HasPrefix(name, "vlan") {
			if id, ok := v[".id"].(string); ok && id != "" {
				if err := c.client.Remove("interface/vlan", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanBridge() (int, error) {
	ports, err := c.client.Get("interface/bridge/port")
	count := 0
	if err == nil {
		for _, p := range ports {
			br, _ := p["bridge"].(string)
			if br == BridgeName {
				if id, ok := p[".id"].(string); ok && id != "" {
					if err := c.client.Remove("interface/bridge/port", id); err == nil {
						count++
					}
				}
			}
		}
	}

	bridges, err := c.client.Get("interface/bridge")
	if err == nil {
		for _, b := range bridges {
			name, _ := b["name"].(string)
			if name == BridgeName {
				if id, ok := b[".id"].(string); ok && id != "" {
					if err := c.client.Remove("interface/bridge", id); err == nil {
						count++
					}
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanPPPoE() (int, error) {
	clients, err := c.client.Get("interface/pppoe-client")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, cl := range clients {
		name, _ := cl["name"].(string)
		if name == PPPoEWAN1 || name == PPPoEWAN2 {
			if id, ok := cl[".id"].(string); ok && id != "" {
				if err := c.client.Remove("interface/pppoe-client", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanFirewallFilters() (int, error) {
	return c.client.RemoveByComment("ip/firewall/filter", AppComment)
}

func (c *Cleaner) cleanMangle() (int, error) {
	return c.client.RemoveByComment("ip/firewall/mangle", AppComment)
}

func (c *Cleaner) cleanNAT() (int, error) {
	return c.client.RemoveByComment("ip/firewall/nat", AppComment)
}

func (c *Cleaner) cleanRoutes() (int, error) {
	routes, err := c.client.Get("ip/route")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, r := range routes {
		cmt, _ := r["comment"].(string)
		rt, _ := r["routing-table"].(string)
		if cmt == AppComment || rt == RTWAN1 || rt == RTWAN2 {
			if id, ok := r[".id"].(string); ok && id != "" {
				if err := c.client.Remove("ip/route", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanRoutingTables() (int, error) {
	tables, err := c.client.Get("routing/table")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, t := range tables {
		name, _ := t["name"].(string)
		if name == RTWAN1 || name == RTWAN2 {
			if id, ok := t[".id"].(string); ok && id != "" {
				if err := c.client.Remove("routing/table", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanIPSec() (int, error) {
	count := 0
	n, _ := c.client.RemoveByComment("ip/ipsec/policy", AppComment)
	count += n
	n, _ = c.client.RemoveByComment("ip/ipsec/identity", AppComment)
	count += n
	n, _ = c.client.RemoveByComment("ip/ipsec/peer", AppComment)
	count += n
	n, _ = c.client.RemoveByComment("ip/ipsec/proposal", AppComment)
	count += n
	n, _ = c.client.RemoveByComment("ip/ipsec/profile", AppComment)
	count += n

	// Remove profile-site2site by name if comment was unsupported in RouterOS
	profiles, _ := c.client.Get("ip/ipsec/profile")
	for _, p := range profiles {
		if name, _ := p["name"].(string); name == "profile-site2site" || name == "ikev2-profile" {
			if id, ok := p[".id"].(string); ok && id != "" {
				if err := c.client.Remove("ip/ipsec/profile", id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

func (c *Cleaner) cleanAddressList() (int, error) {
	return c.client.RemoveByComment("ip/firewall/address-list", AppComment)
}

func (c *Cleaner) cleanScripts() (int, error) {
	return c.client.RemoveByComment("system/script", AppComment)
}

func (c *Cleaner) cleanSchedulers() (int, error) {
	return c.client.RemoveByComment("system/scheduler", AppComment)
}

