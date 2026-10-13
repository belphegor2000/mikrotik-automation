package pppoe

import (
	"fmt"
	"mikrotik-automation/internal/api"
)

const (
	AppComment    = "[MK-AUTO]"
	PPPoEWAN1     = "pppoe-wan1"
	PPPoEWAN2     = "pppoe-wan2"
	ConnMarkWAN1  = "conn-wan1"
	ConnMarkWAN2  = "conn-wan2"
	RouteMarkWAN1 = "to_wan1"
	RouteMarkWAN2 = "to_wan2"
	RTWAN1         = "to_wan1"
	RTWAN2         = "to_wan2"
	LanAddressList = "LAN-Address"
)

type Configurator struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewConfigurator(client api.MikrotikClient, logFn func(string)) *Configurator {
	return &Configurator{client: client, logFn: logFn}
}

func (p *Configurator) log(msg string) {
	if p.logFn != nil {
		p.logFn(msg)
	}
}

func (p *Configurator) ConfigureSingle(user, password string) error {
	p.log("  ▶ Tạo PPPoE client trên ether1...")
	_, err := p.client.Add("interface/pppoe-client", map[string]interface{}{
		"name":              PPPoEWAN1,
		"interface":         "ether1",
		"user":              user,
		"password":          password,
		"add-default-route": "yes",
		"use-peer-dns":      "yes",
		"disabled":          "no",
		"comment":           AppComment,
	})
	if err != nil {
		return fmt.Errorf("tạo PPPoE client thất bại: %w", err)
	}
	p.log("    ✓ Tạo 'pppoe-wan1' thành công.")

	p.log("  ▶ Tạo NAT masquerade...")
	_, err = p.client.Add("ip/firewall/nat", map[string]interface{}{
		"chain":         "srcnat",
		"out-interface": PPPoEWAN1,
		"action":        "masquerade",
		"comment":       AppComment,
	})
	if err != nil {
		return fmt.Errorf("tạo NAT masquerade thất bại: %w", err)
	}
	p.log("    ✓ NAT masquerade cho pppoe-wan1 đã tạo.")
	return nil
}

func (p *Configurator) ConfigureDualPCC(user1, pass1, user2, pass2, bridgeIface string) error {
	if bridgeIface == "" {
		bridgeIface = "bridgeVLAN"
	}
	p.log("  ▶ Tạo PPPoE client WAN1 (ether1)...")
	_, err := p.client.Add("interface/pppoe-client", map[string]interface{}{
		"name":              PPPoEWAN1,
		"interface":         "ether1",
		"user":              user1,
		"password":          pass1,
		"add-default-route": "no",
		"use-peer-dns":      "yes",
		"disabled":          "no",
		"comment":           AppComment,
	})
	if err != nil {
		return err
	}

	p.log("  ▶ Tạo PPPoE client WAN2 (ether2)...")
	_, err = p.client.Add("interface/pppoe-client", map[string]interface{}{
		"name":              PPPoEWAN2,
		"interface":         "ether2",
		"user":              user2,
		"password":          pass2,
		"add-default-route": "no",
		"use-peer-dns":      "yes",
		"disabled":          "no",
		"comment":           AppComment,
	})
	if err != nil {
		return err
	}

	p.log("  ▶ Tạo Routing Tables (ROS 7)...")
	for _, rt := range []string{RTWAN1, RTWAN2} {
		_, _ = p.client.Add("routing/table", map[string]interface{}{
			"name":    rt,
			"fib":     true,
			"comment": AppComment,
		})
	}
	p.log("    ✓ Đã tạo routing tables 'to_wan1' và 'to_wan2'.")

	p.log("  ▶ Tạo Mangle rules cho PCC load balancing...")
	mangleRules := []map[string]interface{}{
		// Input mark connections
		{
			"chain":             "input",
			"in-interface":      PPPoEWAN1,
			"action":            "mark-connection",
			"new-connection-mark": ConnMarkWAN1,
			"passthrough":       "yes",
			"comment":           AppComment,
		},
		{
			"chain":             "input",
			"in-interface":      PPPoEWAN2,
			"action":            "mark-connection",
			"new-connection-mark": ConnMarkWAN2,
			"passthrough":       "yes",
			"comment":           AppComment,
		},
		// Output mark routing
		{
			"chain":           "output",
			"connection-mark": ConnMarkWAN1,
			"action":          "mark-routing",
			"new-routing-mark": RouteMarkWAN1,
			"passthrough":     "no",
			"comment":         AppComment,
		},
		{
			"chain":           "output",
			"connection-mark": ConnMarkWAN2,
			"action":          "mark-routing",
			"new-routing-mark": RouteMarkWAN2,
			"passthrough":     "no",
			"comment":         AppComment,
		},
		// Accept inter-LAN traffic between subnets in Lan-Address before PCC
		{
			"chain":            "prerouting",
			"src-address-list": LanAddressList,
			"dst-address-list": LanAddressList,
			"action":           "accept",
			"comment":          AppComment,
		},
		// Prerouting PCC rules
		{
			"chain":             "prerouting",
			"in-interface":      bridgeIface,
			"dst-address-type":  "!local",
			"per-connection-classifier": "both-addresses-and-ports:2/0",
			"action":            "mark-connection",
			"new-connection-mark": ConnMarkWAN1,
			"passthrough":       "yes",
			"comment":           AppComment,
		},
		{
			"chain":             "prerouting",
			"in-interface":      bridgeIface,
			"dst-address-type":  "!local",
			"per-connection-classifier": "both-addresses-and-ports:2/1",
			"action":            "mark-connection",
			"new-connection-mark": ConnMarkWAN2,
			"passthrough":       "yes",
			"comment":           AppComment,
		},
		// Prerouting mark routing
		{
			"chain":           "prerouting",
			"in-interface":    bridgeIface,
			"connection-mark": ConnMarkWAN1,
			"action":          "mark-routing",
			"new-routing-mark": RouteMarkWAN1,
			"passthrough":     "no",
			"comment":         AppComment,
		},
		{
			"chain":           "prerouting",
			"in-interface":    bridgeIface,
			"connection-mark": ConnMarkWAN2,
			"action":          "mark-routing",
			"new-routing-mark": RouteMarkWAN2,
			"passthrough":     "no",
			"comment":         AppComment,
		},
	}

	for _, rule := range mangleRules {
		_, _ = p.client.Add("ip/firewall/mangle", rule)
	}
	p.log(fmt.Sprintf("    ✓ Đã tạo %d Mangle rules.", len(mangleRules)))

	p.log("  ▶ Tạo Default Routes...")
	routes := []map[string]interface{}{
		{
			"dst-address":   "0.0.0.0/0",
			"gateway":       PPPoEWAN1,
			"routing-table": RTWAN1,
			"distance":      1,
			"comment":       AppComment,
		},
		{
			"dst-address":   "0.0.0.0/0",
			"gateway":       PPPoEWAN2,
			"routing-table": RTWAN2,
			"distance":      1,
			"comment":       AppComment,
		},
		{
			"dst-address":   "0.0.0.0/0",
			"gateway":       PPPoEWAN1,
			"routing-table": "main",
			"distance":      1,
			"comment":       AppComment,
		},
		{
			"dst-address":   "0.0.0.0/0",
			"gateway":       PPPoEWAN2,
			"routing-table": "main",
			"distance":      2,
			"comment":       AppComment,
		},
	}
	for _, r := range routes {
		_, _ = p.client.Add("ip/route", r)
	}
	p.log(fmt.Sprintf("    ✓ Đã tạo %d Routes.", len(routes)))

	p.log("  ▶ Tạo NAT masquerade cho 2 đường WAN...")
	for _, wan := range []string{PPPoEWAN1, PPPoEWAN2} {
		_, _ = p.client.Add("ip/firewall/nat", map[string]interface{}{
			"chain":         "srcnat",
			"out-interface": wan,
			"action":        "masquerade",
			"comment":       AppComment,
		})
	}
	p.log("    ✓ Đã tạo NAT masquerade cho cả 2 đường.")
	return nil
}
