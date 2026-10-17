package bridge

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strings"

	"mikrotik-automation/internal/api"
)

const (
	AppComment     = "[MK-AUTO]"
	BridgeName     = "bridgeVLAN"
	DefaultDNS     = "8.8.8.8,1.1.1.1"
	LanAddressList = "LAN-Address"
)

// BridgeConfigParams defines parameters for full bridge configuration.
type BridgeConfigParams struct {
	WANCount   int
	BridgeIP   string // Any CIDR or IP e.g. "192.168.88.1/24", "10.0.0.1/24", "172.16.1.1/24"
	EnableDHCP bool   // Enable DHCP server on bridge
	DNSServers string // DNS servers separated by commas
}

type Configurator struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewConfigurator(client api.MikrotikClient, logFn func(string)) *Configurator {
	return &Configurator{client: client, logFn: logFn}
}

func (b *Configurator) log(msg string) {
	if b.logFn != nil {
		b.logFn(msg)
	}
}

// ParseBridgeCIDR parses any IPv4 CIDR or plain IP address and calculates
// routerIP, gateway, startIP, endIP, network, and prefixLen.
func ParseBridgeCIDR(input string) (routerIP, gateway, startIP, endIP, network string, prefixLen int, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", "", "", "", 0, fmt.Errorf("địa chỉ IP bridge không được để trống")
	}

	// Auto-append /24 if no prefix mask provided
	if !strings.Contains(input, "/") {
		input = input + "/24"
	}

	ip, ipNet, err := net.ParseCIDR(input)
	if err != nil {
		return "", "", "", "", "", 0, fmt.Errorf("CIDR không hợp lệ (%s): %w", input, err)
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", "", "", "", "", 0, fmt.Errorf("chỉ hỗ trợ định dạng IPv4")
	}

	ones, bits := ipNet.Mask.Size()
	prefixLen = ones
	if bits != 32 || ones < 8 || ones > 30 {
		return "", "", "", "", "", 0, fmt.Errorf("prefix mask /%d không hợp lệ (hỗ trợ /8 đến /30)", ones)
	}

	netInt := binary.BigEndian.Uint32(ipNet.IP.To4())
	maskInt := binary.BigEndian.Uint32(net.IP(ipNet.Mask).To4())
	bcastInt := netInt | (^maskInt)

	firstUsable := netInt + 1
	lastUsable := bcastInt - 1

	inputInt := binary.BigEndian.Uint32(ipv4)
	routerInt := inputInt
	// If input was network address or broadcast address, use first usable IP
	if inputInt <= netInt || inputInt >= bcastInt {
		routerInt = firstUsable
	}

	rIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(rIP, routerInt)
	routerIP = rIP.String()
	gateway = routerIP

	// DHCP pool range calculation
	startInt := netInt + 10
	if startInt >= bcastInt || startInt <= routerInt {
		startInt = routerInt + 1
	}
	if startInt > lastUsable {
		startInt = firstUsable
		if startInt == routerInt && routerInt+1 <= lastUsable {
			startInt = routerInt + 1
		}
	}
	endInt := lastUsable

	sIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(sIP, startInt)
	startIP = sIP.String()

	eIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(eIP, endInt)
	endEndIP := eIP.String()

	network = ipNet.IP.String()
	return routerIP, gateway, startIP, endEndIP, network, prefixLen, nil
}

func (b *Configurator) GetAvailablePorts(pppoeCount int) ([]string, error) {
	wanPorts := map[string]bool{"ether1": true}
	if pppoeCount == 2 {
		wanPorts["ether2"] = true
	}

	ifaces, err := b.client.Get("interface/ethernet")
	if err != nil {
		return nil, err
	}

	var available []string
	for _, iface := range ifaces {
		name, _ := iface["name"].(string)
		if name != "" && !wanPorts[name] {
			available = append(available, name)
		}
	}
	sort.Strings(available)
	return available, nil
}

// CreateBridge creates bridgeVLAN and moves available ports into it (legacy helper).
func (b *Configurator) CreateBridge(pppoeCount int) ([]string, error) {
	ports, _, err := b.ConfigureBridge(BridgeConfigParams{
		WANCount:   pppoeCount,
		BridgeIP:   "192.168.88.1/24",
		EnableDHCP: false,
	})
	return ports, err
}

// ConfigureBridge creates/updates bridgeVLAN, assigns IP, moves ports, sets DHCP Server & DNS.
// Returns the list of added ports, the configured routerIP, and any error.
func (b *Configurator) ConfigureBridge(p BridgeConfigParams) ([]string, string, error) {
	bridgeIP := p.BridgeIP
	if bridgeIP == "" {
		bridgeIP = "192.168.88.1/24"
	}
	dnsServers := p.DNSServers
	if strings.TrimSpace(dnsServers) == "" {
		dnsServers = DefaultDNS
	}

	routerIP, gateway, poolStart, poolEnd, network, prefixLen, err := ParseBridgeCIDR(bridgeIP)
	if err != nil {
		return nil, "", err
	}

	// 1. Tạo bridgeVLAN nếu chưa có
	b.log(fmt.Sprintf("  ▶ [1/6] Kiểm tra & Tạo bridge '%s'...", BridgeName))
	bridges, _ := b.client.Get("interface/bridge")
	bridgeExists := false
	for _, br := range bridges {
		if name, _ := br["name"].(string); name == BridgeName {
			bridgeExists = true
			break
		}
	}
	if !bridgeExists {
		_, err := b.client.Add("interface/bridge", map[string]interface{}{
			"name":           BridgeName,
			"vlan-filtering": "no",
			"comment":        AppComment,
		})
		if err != nil {
			return nil, "", fmt.Errorf("tạo bridge thất bại: %w", err)
		}
		b.log(fmt.Sprintf("    ✓ Bridge '%s' đã tạo thành công.", BridgeName))
	} else {
		b.log(fmt.Sprintf("    ✓ Bridge '%s' đã tồn tại.", BridgeName))
	}

	// 2. Chuyển các port khả dụng vào bridgeVLAN
	ports, err := b.GetAvailablePorts(p.WANCount)
	if err != nil {
		return nil, "", err
	}

	var addedPorts []string
	if len(ports) > 0 {
		b.log(fmt.Sprintf("  ▶ [2/6] Đồng bộ %d port vào bridge '%s'...", len(ports), BridgeName))
		// Lấy danh sách port hiện tại để gỡ khỏi bridge khác nếu cần
		curPorts, _ := b.client.Get("interface/bridge/port")
		portMap := make(map[string]map[string]interface{})
		for _, cp := range curPorts {
			if iface, _ := cp["interface"].(string); iface != "" {
				portMap[iface] = cp
			}
		}

		for _, port := range ports {
			if existing, found := portMap[port]; found {
				curBr, _ := existing["bridge"].(string)
				if curBr == BridgeName {
					addedPorts = append(addedPorts, port)
					continue
				}
				// Port đang ở bridge khác (ví dụ bridge mặc định defconf) -> gỡ ra
				if id, ok := existing[".id"].(string); ok && id != "" {
					_ = b.client.Remove("interface/bridge/port", id)
				}
			}

			_, err := b.client.Add("interface/bridge/port", map[string]interface{}{
				"bridge":    BridgeName,
				"interface": port,
				"comment":   AppComment,
			})
			if err != nil {
				b.log(fmt.Sprintf("    ⚠ Không thể thêm port %s: %v", port, err))
			} else {
				b.log(fmt.Sprintf("    ✓ Đã gán port %s vào %s", port, BridgeName))
				addedPorts = append(addedPorts, port)
			}
		}
	}

	// 3. Gán IP Address cho bridgeVLAN
	fullCIDR := fmt.Sprintf("%s/%d", routerIP, prefixLen)
	b.log(fmt.Sprintf("  ▶ [3/6] Gán IP %s cho interface '%s'...", fullCIDR, BridgeName))
	addrs, _ := b.client.Get("ip/address")
	ipAlreadyOnBridge := false
	var oldDefconfIDs []string

	for _, a := range addrs {
		addrStr, _ := a["address"].(string)
		iface, _ := a["interface"].(string)
		id, _ := a[".id"].(string)

		if iface == BridgeName && addrStr == fullCIDR {
			ipAlreadyOnBridge = true
		} else if iface != BridgeName && (addrStr == fullCIDR || strings.HasPrefix(addrStr, "192.168.88.1/")) {
			// Lưu lại ID của IP cũ trên bridge defconf để dọn dẹp sau khi gán IP mới
			if id != "" {
				oldDefconfIDs = append(oldDefconfIDs, id)
			}
		}
	}

	if !ipAlreadyOnBridge {
		_, err := b.client.Add("ip/address", map[string]interface{}{
			"address":   fullCIDR,
			"interface": BridgeName,
			"comment":   AppComment,
		})
		if err != nil {
			b.log(fmt.Sprintf("    ⚠ Gán IP thất bại: %v", err))
		} else {
			b.log(fmt.Sprintf("    ✓ Đã gán thành công IP %s cho %s.", fullCIDR, BridgeName))
		}
	} else {
		b.log(fmt.Sprintf("    ✓ IP %s đã tồn tại trên %s.", fullCIDR, BridgeName))
	}

	// 4. Cấu hình DHCP Server cho Bridge
	if p.EnableDHCP {
		b.log("  ▶ [4/6] Cấu hình DHCP Server cho Bridge...")
		poolName := "pool-bridge"
		poolRange := fmt.Sprintf("%s-%s", poolStart, poolEnd)
		b.log(fmt.Sprintf("    + Tạo/cập nhật IP pool '%s' (%s)...", poolName, poolRange))
		_, _ = b.client.Add("ip/pool", map[string]interface{}{
			"name":    poolName,
			"ranges":  poolRange,
			"comment": AppComment,
		})

		srvName := "dhcp-bridge"
		b.log(fmt.Sprintf("    + Cấu hình DHCP server '%s' trên '%s'...", srvName, BridgeName))
		_, _ = b.client.Add("ip/dhcp-server", map[string]interface{}{
			"name":         srvName,
			"interface":    BridgeName,
			"address-pool": poolName,
			"lease-time":   "12h",
			"disabled":     "no",
			"comment":      AppComment,
		})

		netAddr := fmt.Sprintf("%s/%d", network, prefixLen)
		b.log(fmt.Sprintf("    + Cấu hình DHCP Network (%s, Gateway: %s)...", netAddr, gateway))
		_, _ = b.client.Add("ip/dhcp-server/network", map[string]interface{}{
			"address":    netAddr,
			"gateway":    gateway,
			"dns-server": dnsServers,
			"comment":    AppComment,
		})
	}

	// 5. Cấu hình DNS Server
	b.log(fmt.Sprintf("  ▶ [5/6] Cấu hình DNS Server (%s)...", dnsServers))
	_ = b.client.Set("ip/dns", "", map[string]interface{}{
		"allow-remote-requests": "yes",
		"servers":               dnsServers,
	})

	// Thêm vào Firewall Address-List "LAN-Address"
	netAddr := fmt.Sprintf("%s/%d", network, prefixLen)
	b.log(fmt.Sprintf("    + Thêm %s vào danh sách firewall '%s'...", netAddr, LanAddressList))
	_, _ = b.client.Add("ip/firewall/address-list", map[string]interface{}{
		"list":    LanAddressList,
		"address": netAddr,
		"comment": AppComment,
	})

	// 6. Dọn dẹp IP cũ trên interface khác để tránh xung đột
	b.log("  ▶ [6/6] Dọn dẹp IP và DHCP cũ nếu có...")
	for _, oldID := range oldDefconfIDs {
		_ = b.client.Remove("ip/address", oldID)
	}

	b.log(fmt.Sprintf("  ✓ Hoàn tất cấu hình Bridge LAN! IP Router: %s (Subnet: %s)", routerIP, fullCIDR))
	return addedPorts, routerIP, nil
}
