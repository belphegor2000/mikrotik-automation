package vlan

import (
	"encoding/binary"
	"fmt"
	"net"

	"mikrotik-automation/internal/api"
)

const (
	AppComment     = "[MK-AUTO]"
	BridgeName     = "bridgeVLAN"
	DNSServers     = "8.8.8.8,1.1.1.1"
	LanAddressList = "LAN-Address"
)

type VLANEntry struct {
	VLANID int    `json:"vlan_id"`
	Name   string `json:"name"`
	CIDR   string `json:"cidr"`
}

func (v VLANEntry) InterfaceName() string {
	if v.VLANID == 1 {
		return BridgeName
	}
	return fmt.Sprintf("vlan%d", v.VLANID)
}

func (v VLANEntry) calculateIPs() (gateway, poolStart, poolEnd, network string, prefixLen int, err error) {
	ip, ipNet, err := net.ParseCIDR(v.CIDR)
	if err != nil {
		return "", "", "", "", 0, err
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", "", "", "", 0, fmt.Errorf("chỉ hỗ trợ IPv4")
	}

	ones, bits := ipNet.Mask.Size()
	prefixLen = ones
	if bits != 32 || ones > 30 {
		return "", "", "", "", 0, fmt.Errorf("subnet quá nhỏ hoặc không hợp lệ")
	}

	netInt := binary.BigEndian.Uint32(ipNet.IP.To4())
	maskInt := binary.BigEndian.Uint32(net.IP(ipNet.Mask).To4())
	bcastInt := netInt | (^maskInt)

	gwInt := netInt + 1
	startInt := netInt + 10
	if startInt >= bcastInt {
		startInt = gwInt + 1
	}
	endInt := bcastInt - 1

	gwIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(gwIP, gwInt)

	startIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(startIP, startInt)

	endIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(endIP, endInt)

	return gwIP.String(), startIP.String(), endIP.String(), ipNet.IP.String(), prefixLen, nil
}

func (v VLANEntry) Gateway() string {
	gw, _, _, _, _, _ := v.calculateIPs()
	return gw
}

func (v VLANEntry) DHCPPoolStart() string {
	_, start, _, _, _, _ := v.calculateIPs()
	return start
}

func (v VLANEntry) DHCPPoolEnd() string {
	_, _, end, _, _, _ := v.calculateIPs()
	return end
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

func (c *Configurator) Configure(vlans []VLANEntry) error {
	if len(vlans) == 0 {
		return fmt.Errorf("danh sách VLAN trống")
	}

	// 1. Tạo VLAN interfaces (ngoại trừ VLAN 1)
	c.log("  ▶ [1/9] Tạo VLAN interfaces trên bridgeVLAN...")
	for _, v := range vlans {
		if v.VLANID == 1 {
			continue
		}
		ifName := v.InterfaceName()
		c.log(fmt.Sprintf("    + Tạo interface '%s' (vlan-id=%d)...", ifName, v.VLANID))
		_, err := c.client.Add("interface/vlan", map[string]interface{}{
			"name":      ifName,
			"vlan-id":   v.VLANID,
			"interface": BridgeName,
			"comment":   AppComment,
		})
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Không thể tạo vlan interface %s: %v", ifName, err))
		}
	}

	// 2. Gán IP Address cho từng VLAN
	c.log("  ▶ [2/9] Gán IP Address cho các VLAN / Bridge...")
	for _, v := range vlans {
		gw, _, _, _, prefix, err := v.calculateIPs()
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi subnet VLAN %d: %v", v.VLANID, err))
			continue
		}
		ifName := v.InterfaceName()
		addr := fmt.Sprintf("%s/%d", gw, prefix)
		c.log(fmt.Sprintf("    + Gán %s trên interface '%s'...", addr, ifName))
		_, err = c.client.Add("ip/address", map[string]interface{}{
			"address":   addr,
			"interface": ifName,
			"comment":   AppComment,
		})
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi gán IP: %v", err))
		}
	}

	// 3. Tạo DHCP IP Pools
	c.log("  ▶ [3/9] Tạo DHCP IP Pools...")
	for _, v := range vlans {
		_, start, end, _, _, err := v.calculateIPs()
		if err != nil {
			continue
		}
		poolName := fmt.Sprintf("pool-vlan%d", v.VLANID)
		poolRange := fmt.Sprintf("%s-%s", start, end)
		c.log(fmt.Sprintf("    + Tạo pool '%s' (%s)...", poolName, poolRange))
		_, err = c.client.Add("ip/pool", map[string]interface{}{
			"name":    poolName,
			"ranges":  poolRange,
			"comment": AppComment,
		})
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi tạo pool: %v", err))
		}
	}

	// 4. Tạo DHCP Servers
	c.log("  ▶ [4/9] Tạo DHCP Servers...")
	for _, v := range vlans {
		ifName := v.InterfaceName()
		srvName := fmt.Sprintf("dhcp-vlan%d", v.VLANID)
		poolName := fmt.Sprintf("pool-vlan%d", v.VLANID)
		c.log(fmt.Sprintf("    + Tạo DHCP server '%s' trên '%s'...", srvName, ifName))
		_, err := c.client.Add("ip/dhcp-server", map[string]interface{}{
			"name":         srvName,
			"interface":    ifName,
			"address-pool": poolName,
			"lease-time":   "12h",
			"disabled":     "no",
			"comment":      AppComment,
		})
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi tạo DHCP server: %v", err))
		}
	}

	// 5. Cấu hình DHCP Networks
	c.log("  ▶ [5/9] Cấu hình DHCP Networks...")
	for _, v := range vlans {
		gw, _, _, _, _, err := v.calculateIPs()
		if err != nil {
			continue
		}
		c.log(fmt.Sprintf("    + Tạo DHCP network %s (gw=%s)...", v.CIDR, gw))
		_, err = c.client.Add("ip/dhcp-server/network", map[string]interface{}{
			"address":    v.CIDR,
			"gateway":    gw,
			"dns-server": DNSServers,
			"comment":    AppComment,
		})
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi tạo DHCP network: %v", err))
		}
	}

	// 6. Cấu hình DNS Server
	c.log(fmt.Sprintf("  ▶ [6/9] Cấu hình Router DNS (%s, allow-remote=yes)...", DNSServers))
	err := c.client.Set("ip/dns", "", map[string]interface{}{
		"servers":               DNSServers,
		"allow-remote-requests": "yes",
	})
	if err != nil {
		c.log(fmt.Sprintf("    ⚠ Lỗi cấu hình DNS: %v", err))
	} else {
		c.log("    ✓ Cấu hình DNS thành công.")
	}

	// 7. Thêm dải mạng VLAN vào Firewall Address List 'Lan-Address'
	c.log(fmt.Sprintf("  ▶ [7/9] Thêm mạng VLAN vào Address List '%s'...", LanAddressList))
	for _, v := range vlans {
		_, _, _, network, prefixLen, err := v.calculateIPs()
		if err != nil {
			continue
		}
		netCIDR := fmt.Sprintf("%s/%d", network, prefixLen)
		c.log(fmt.Sprintf("    + Thêm %s vào list '%s'...", netCIDR, LanAddressList))
		_, err = c.client.Add("ip/firewall/address-list", map[string]interface{}{
			"list":    LanAddressList,
			"address": netCIDR,
			"comment": AppComment,
		})
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi thêm address-list %s: %v", netCIDR, err))
		}
	}

	// 8. Cấu hình Bridge VLAN Table
	c.log("  ▶ [8/9] Cấu hình Bridge VLAN Table...")
	for _, v := range vlans {
		c.log(fmt.Sprintf("    + Thêm vlan-id %d vào bridge '%s'...", v.VLANID, BridgeName))
		data := map[string]interface{}{
			"bridge":   BridgeName,
			"vlan-ids": v.VLANID,
			"comment":  AppComment,
		}
		if v.VLANID != 1 {
			data["tagged"] = BridgeName
		}
		_, err := c.client.Add("interface/bridge/vlan", data)
		if err != nil {
			c.log(fmt.Sprintf("    ⚠ Lỗi cấu hình bridge vlan table: %v", err))
		}
	}

	// 9. Bật vlan-filtering=yes trên bridge
	c.log(fmt.Sprintf("  ▶ [9/9] Bật vlan-filtering=yes trên '%s'...", BridgeName))
	bridges, err := c.client.Get("interface/bridge")
	if err == nil {
		for _, b := range bridges {
			name, _ := b["name"].(string)
			if name == BridgeName {
				id, _ := b[".id"].(string)
				err := c.client.Set("interface/bridge", id, map[string]interface{}{
					"vlan-filtering": "yes",
				})
				if err != nil {
					c.log(fmt.Sprintf("    ⚠ Lỗi bật vlan-filtering: %v", err))
				} else {
					c.log("    ✓ Đã kích hoạt vlan-filtering=yes trên bridgeVLAN.")
				}
				break
			}
		}
	}

	c.log("✓ Hoàn tất cấu hình toàn bộ VLAN và hệ thống mạng!")
	return nil
}

// AddLanAddress adds a specific network CIDR to the Lan-Address address list
func (c *Configurator) AddLanAddress(netCIDR string) error {
	_, err := c.client.Add("ip/firewall/address-list", map[string]interface{}{
		"list":    LanAddressList,
		"address": netCIDR,
		"comment": AppComment,
	})
	return err
}
