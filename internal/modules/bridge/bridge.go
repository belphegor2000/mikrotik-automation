package bridge

import (
	"fmt"
	"sort"
	"strings"

	"mikrotik-automation/internal/api"
)

const (
	AppComment = "[MK-AUTO]"
	BridgeName = "bridgeVLAN"
)

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

func (b *Configurator) CreateBridge(pppoeCount int) ([]string, error) {
	b.log(fmt.Sprintf("  ▶ Tạo bridge '%s'...", BridgeName))
	_, err := b.client.Add("interface/bridge", map[string]interface{}{
		"name":           BridgeName,
		"vlan-filtering": "no",
		"comment":        AppComment,
	})
	if err != nil {
		return nil, fmt.Errorf("tạo bridge thất bại: %w", err)
	}
	b.log(fmt.Sprintf("    ✓ Bridge '%s' đã tạo (vlan-filtering=no).", BridgeName))

	ports, err := b.GetAvailablePorts(pppoeCount)
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		b.log("    ⚠ Không tìm thấy port nào để thêm vào bridge.")
		return nil, nil
	}

	b.log(fmt.Sprintf("  ▶ Thêm %d port vào bridge: %s", len(ports), strings.Join(ports, ", ")))
	var added []string
	for _, port := range ports {
		_, err := b.client.Add("interface/bridge/port", map[string]interface{}{
			"bridge":    BridgeName,
			"interface": port,
			"comment":   AppComment,
		})
		if err != nil {
			b.log(fmt.Sprintf("    ⚠ Không thể thêm %s vào bridge: %v", port, err))
		} else {
			b.log(fmt.Sprintf("    ✓ Đã thêm port %s vào bridge", port))
			added = append(added, port)
		}
	}
	return added, nil
}
