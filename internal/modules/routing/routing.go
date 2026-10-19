package routing

import (
	"fmt"
	"net"
	"strings"

	"mikrotik-automation/internal/api"
)

const (
	AppComment       = "[MK-AUTO]"
	PBRCommentPrefix = "[MK-AUTO] PBR: "
	RTWAN1           = "to_wan1"
	RTWAN2           = "to_wan2"
	AddressListWAN1  = "PBR-to-WAN1"
	AddressListWAN2  = "PBR-to-WAN2"

	TypeVLANSubnet = "VLAN_SUBNET"
	TypeDomainIP   = "DOMAIN_IP"
)

type PBRRule struct {
	ID          string `json:".id,omitempty"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "VLAN_SUBNET" or "DOMAIN_IP"
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	TargetWAN   string `json:"target_wan"` // "WAN1" or "WAN2"
	Comment     string `json:"comment,omitempty"`
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

// EnsureRoutingTables checks and creates RouterOS 7 routing tables to_wan1 and to_wan2 if not present.
func (c *Configurator) EnsureRoutingTables() error {
	tables, err := c.client.Get("routing/table")
	if err != nil {
		return err
	}

	hasWAN1 := false
	hasWAN2 := false
	for _, t := range tables {
		name, _ := t["name"].(string)
		if name == RTWAN1 {
			hasWAN1 = true
		}
		if name == RTWAN2 {
			hasWAN2 = true
		}
	}

	if !hasWAN1 {
		_, _ = c.client.Add("routing/table", map[string]interface{}{
			"name":    RTWAN1,
			"fib":     true,
			"comment": AppComment,
		})
	}
	if !hasWAN2 {
		_, _ = c.client.Add("routing/table", map[string]interface{}{
			"name":    RTWAN2,
			"fib":     true,
			"comment": AppComment,
		})
	}
	return nil
}

// AddVLANSubnetRule creates a Policy Routing rule binding a LAN/VLAN subnet strictly to WAN1 or WAN2.
func (c *Configurator) AddVLANSubnetRule(name, subnet, targetWAN string) error {
	name = strings.TrimSpace(name)
	subnet = strings.TrimSpace(subnet)
	targetWAN = strings.ToUpper(strings.TrimSpace(targetWAN))

	if name == "" {
		return fmt.Errorf("tên chính sách PBR không được để trống")
	}
	if subnet == "" {
		return fmt.Errorf("địa chỉ mạng (subnet CIDR) không được để trống")
	}

	// Validate CIDR or IP
	if !strings.Contains(subnet, "/") {
		parsedIP := net.ParseIP(subnet)
		if parsedIP == nil || parsedIP.To4() == nil {
			return fmt.Errorf("địa chỉ '%s' không phải IPv4 hợp lệ", subnet)
		}
		subnet = subnet + "/32"
	} else {
		_, _, err := net.ParseCIDR(subnet)
		if err != nil {
			return fmt.Errorf("dải mạng CIDR '%s' không hợp lệ: %w", subnet, err)
		}
	}

	table := RTWAN1
	if targetWAN == "WAN2" {
		table = RTWAN2
	}

	_ = c.EnsureRoutingTables()

	comment := fmt.Sprintf("%s%s (%s -> %s)", PBRCommentPrefix, name, subnet, targetWAN)
	c.log(fmt.Sprintf("▶ Đang cấu hình định tuyến chính sách cho %s: Subnet %s ➔ %s...", name, subnet, targetWAN))

	// 1. Create RouterOS 7 /routing/rule
	ruleData := map[string]interface{}{
		"src-address": subnet,
		"action":      "lookup",
		"table":       table,
		"comment":     comment,
	}
	_, err := c.client.Add("routing/rule", ruleData)
	if err != nil {
		c.log(fmt.Sprintf("  ⚠ Không thể thêm /routing/rule (%v), chuyển sang cấu hình qua Firewall Mangle...", err))
	}

	// 2. Also create Mangle prerouting mark-routing rule to guarantee packet marking
	mangleData := map[string]interface{}{
		"chain":            "prerouting",
		"src-address":      subnet,
		"dst-address-type": "!local",
		"action":           "mark-routing",
		"new-routing-mark": table,
		"passthrough":      "no",
		"comment":          comment,
	}
	_, errMangle := c.client.Add("ip/firewall/mangle", mangleData)
	if err != nil && errMangle != nil {
		return fmt.Errorf("thêm luật PBR thất bại: %v, mangle: %v", err, errMangle)
	}

	c.log(fmt.Sprintf("✓ Đã thiết lập thành công định tuyến chính sách PBR cho %s (Subnet %s ➔ %s)", name, subnet, targetWAN))
	return nil
}

// AddDomainIPRule creates a destination-based routing rule (e.g. netflix.com, bank IP) to WAN1 or WAN2.
func (c *Configurator) AddDomainIPRule(name, destination, targetWAN string) error {
	name = strings.TrimSpace(name)
	destination = strings.TrimSpace(destination)
	targetWAN = strings.ToUpper(strings.TrimSpace(targetWAN))

	if name == "" {
		return fmt.Errorf("tên chính sách PBR không được để trống")
	}
	if destination == "" {
		return fmt.Errorf("tên miền hoặc IP đích không được để trống")
	}

	table := RTWAN1
	listName := AddressListWAN1
	if targetWAN == "WAN2" {
		table = RTWAN2
		listName = AddressListWAN2
	}

	_ = c.EnsureRoutingTables()

	comment := fmt.Sprintf("%s%s (%s -> %s)", PBRCommentPrefix, name, destination, targetWAN)
	c.log(fmt.Sprintf("▶ Đang cấu hình định tuyến dịch vụ %s (%s) ➔ %s...", name, destination, targetWAN))

	// 1. Add destination domain or IP to Firewall Address-List
	addrData := map[string]interface{}{
		"list":    listName,
		"address": destination,
		"comment": comment,
	}
	_, err := c.client.Add("ip/firewall/address-list", addrData)
	if err != nil {
		return fmt.Errorf("lỗi thêm vào address-list: %w", err)
	}

	// 2. Ensure Mangle rule for this address-list exists
	c.ensureMangleForAddressList(listName, table, targetWAN)

	c.log(fmt.Sprintf("✓ Đã định tuyến thành công %s (%s) đi qua cổng %s", name, destination, targetWAN))
	return nil
}

func (c *Configurator) ensureMangleForAddressList(listName, table, targetWAN string) {
	rules, err := c.client.Get("ip/firewall/mangle")
	if err == nil {
		for _, r := range rules {
			chain, _ := r["chain"].(string)
			dstList, _ := r["dst-address-list"].(string)
			if chain == "prerouting" && dstList == listName {
				return // Already exists
			}
		}
	}

	data := map[string]interface{}{
		"chain":            "prerouting",
		"dst-address-list": listName,
		"dst-address-type": "!local",
		"action":           "mark-routing",
		"new-routing-mark": table,
		"passthrough":      "no",
		"comment":          fmt.Sprintf("%sRoute Address-List %s -> %s", PBRCommentPrefix, listName, targetWAN),
	}
	_, _ = c.client.Add("ip/firewall/mangle", data)
}

// ListPBRRules retrieves all configured PBR rules from RouterOS.
func (c *Configurator) ListPBRRules() ([]PBRRule, error) {
	var result []PBRRule

	// 1. Check /routing/rule
	rrList, err := c.client.Get("routing/rule")
	if err == nil {
		for _, item := range rrList {
			comment, _ := item["comment"].(string)
			if strings.HasPrefix(comment, PBRCommentPrefix) {
				id, _ := item[".id"].(string)
				src, _ := item["src-address"].(string)
				tbl, _ := item["table"].(string)
				cleanName := strings.TrimPrefix(comment, PBRCommentPrefix)
				if idx := strings.Index(cleanName, " ("); idx != -1 {
					cleanName = cleanName[:idx]
				}

				targetWAN := "WAN1"
				if tbl == RTWAN2 {
					targetWAN = "WAN2"
				}

				result = append(result, PBRRule{
					ID:        id,
					Name:      cleanName,
					Type:      TypeVLANSubnet,
					Source:    src,
					TargetWAN: targetWAN,
					Comment:   comment,
				})
			}
		}
	}

	// 2. Check /ip/firewall/address-list
	addrList, err := c.client.Get("ip/firewall/address-list")
	if err == nil {
		for _, item := range addrList {
			comment, _ := item["comment"].(string)
			list, _ := item["list"].(string)
			if strings.HasPrefix(comment, PBRCommentPrefix) && (list == AddressListWAN1 || list == AddressListWAN2) {
				id, _ := item[".id"].(string)
				addr, _ := item["address"].(string)
				cleanName := strings.TrimPrefix(comment, PBRCommentPrefix)
				if idx := strings.Index(cleanName, " ("); idx != -1 {
					cleanName = cleanName[:idx]
				}

				targetWAN := "WAN1"
				if list == AddressListWAN2 {
					targetWAN = "WAN2"
				}

				result = append(result, PBRRule{
					ID:          id,
					Name:        cleanName,
					Type:        TypeDomainIP,
					Destination: addr,
					TargetWAN:   targetWAN,
					Comment:     comment,
				})
			}
		}
	}

	return result, nil
}

// DeletePBRRule removes a PBR rule from RouterOS.
func (c *Configurator) DeletePBRRule(rule PBRRule) error {
	c.log(fmt.Sprintf("▶ Đang xoá chính sách định tuyến '%s'...", rule.Name))

	if rule.Type == TypeVLANSubnet {
		if rule.ID != "" {
			_ = c.client.Remove("routing/rule", rule.ID)
		}
		// Also clean corresponding Mangle rule if exists
		if rule.Comment != "" {
			_, _ = c.client.RemoveByComment("ip/firewall/mangle", rule.Comment)
		}
	} else if rule.Type == TypeDomainIP {
		if rule.ID != "" {
			_ = c.client.Remove("ip/firewall/address-list", rule.ID)
		}
	}

	c.log(fmt.Sprintf("✓ Đã xoá chính sách định tuyến '%s' thành công.", rule.Name))
	return nil
}
