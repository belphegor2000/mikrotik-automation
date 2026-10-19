package queue

import (
	"fmt"
	"strconv"
	"strings"

	"mikrotik-automation/internal/api"
)

const AutoQueueComment = "[MK-AUTO]"

// QueueItem represents a Simple Queue entry in RouterOS.
type QueueItem struct {
	ID            string `json:".id,omitempty"`
	Name          string `json:"name"`
	Target        string `json:"target"`
	MaxLimit      string `json:"max-limit"` // "upload/download" format e.g. "10M/20M"
	UploadLimit   string `json:"-"`
	DownloadLimit string `json:"-"`
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

// FormatRateLimit normalizes user inputs into RouterOS rate format (e.g. "10" -> "10M", "10M" -> "10M", "0" -> "0").
func FormatRateLimit(val string) string {
	s := strings.TrimSpace(val)
	if s == "" || s == "0" || strings.EqualFold(s, "unlimited") || strings.EqualFold(s, "không giới hạn") {
		return "0"
	}
	// If purely digits, assume Mbps
	if _, err := strconv.Atoi(s); err == nil {
		return s + "M"
	}
	// If already ends with k, K, m, M, g, G
	lastChar := strings.ToLower(s[len(s)-1:])
	if lastChar == "k" || lastChar == "m" || lastChar == "g" {
		return strings.ToUpper(s[:len(s)-1]) + strings.ToUpper(lastChar)
	}
	return s
}

// FormatHumanRate converts numeric bytes or rate strings into clean readable rate (e.g. 10000000 -> 10M, 0 -> Unlimited).
func FormatHumanRate(val string) string {
	s := strings.TrimSpace(val)
	if s == "" || s == "0" {
		return "Unlimited"
	}
	num, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		if num >= 1000000000 {
			return fmt.Sprintf("%dG", num/1000000000)
		}
		if num >= 1000000 {
			return fmt.Sprintf("%dM", num/1000000)
		}
		if num >= 1000 {
			return fmt.Sprintf("%dk", num/1000)
		}
		return fmt.Sprintf("%d", num)
	}
	return s
}

// ParseMaxLimit splits "upload/download" into readable parts.
func ParseMaxLimit(maxLimit string) (upload, download string) {
	parts := strings.Split(maxLimit, "/")
	if len(parts) == 2 {
		return FormatHumanRate(parts[0]), FormatHumanRate(parts[1])
	}
	if len(parts) == 1 && parts[0] != "" {
		return FormatHumanRate(parts[0]), FormatHumanRate(parts[0])
	}
	return "Unlimited", "Unlimited"
}

// ListQueues fetches all Simple Queues from RouterOS.
func (c *Configurator) ListQueues() ([]QueueItem, error) {
	if c.client == nil || !c.client.IsConnected() {
		return nil, fmt.Errorf("chưa kết nối đến router")
	}

	rawItems, err := c.client.Get("queue/simple")
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách queue: %w", err)
	}

	result := make([]QueueItem, 0, len(rawItems))
	for _, item := range rawItems {
		id, _ := item[".id"].(string)
		name, _ := item["name"].(string)
		target, _ := item["target"].(string)
		maxLimit, _ := item["max-limit"].(string)
		comment, _ := item["comment"].(string)

		up, down := ParseMaxLimit(maxLimit)

		result = append(result, QueueItem{
			ID:            id,
			Name:          name,
			Target:        target,
			MaxLimit:      maxLimit,
			UploadLimit:   up,
			DownloadLimit: down,
			Comment:       comment,
		})
	}
	return result, nil
}

// AddQueue creates or updates a Simple Queue on RouterOS.
func (c *Configurator) AddQueue(item QueueItem) error {
	if c.client == nil || !c.client.IsConnected() {
		return fmt.Errorf("chưa kết nối đến router")
	}

	name := strings.TrimSpace(item.Name)
	if name == "" {
		return fmt.Errorf("tên Queue không được để trống")
	}

	target := strings.TrimSpace(item.Target)
	if target == "" {
		return fmt.Errorf("địa chỉ Target không được để trống")
	}

	up := FormatRateLimit(item.UploadLimit)
	down := FormatRateLimit(item.DownloadLimit)
	maxLimit := fmt.Sprintf("%s/%s", up, down)

	comment := strings.TrimSpace(item.Comment)
	if comment == "" {
		comment = AutoQueueComment
	} else if !strings.Contains(comment, AutoQueueComment) {
		comment = AutoQueueComment + " " + comment
	}

	data := map[string]interface{}{
		"name":      name,
		"target":    target,
		"max-limit": maxLimit,
		"comment":   comment,
	}

	c.log(fmt.Sprintf("▶ Đang tạo Simple Queue '%s' cho Target: %s (Max: %s/s)...", name, target, maxLimit))
	_, err := c.client.Add("queue/simple", data)
	if err != nil {
		return fmt.Errorf("lỗi thêm queue '%s': %w", name, err)
	}
	c.log(fmt.Sprintf("✓ Đã tạo thành công Simple Queue '%s' (Upload: %s, Download: %s)", name, up, down))
	return nil
}

// DeleteQueue removes a Simple Queue by ID or name.
func (c *Configurator) DeleteQueue(id string) error {
	if c.client == nil || !c.client.IsConnected() {
		return fmt.Errorf("chưa kết nối đến router")
	}

	c.log(fmt.Sprintf("▶ Đang xoá Simple Queue (ID: %s)...", id))
	err := c.client.Remove("queue/simple", id)
	if err != nil {
		return fmt.Errorf("lỗi xoá queue '%s': %w", id, err)
	}
	c.log(fmt.Sprintf("✓ Đã xoá Simple Queue (ID: %s) thành công.", id))
	return nil
}

const (
	CakeQueueName     = "Smart-CAKE-QoS"
	CakeQueueNameWAN1 = "Smart-CAKE-WAN1"
	CakeQueueNameWAN2 = "Smart-CAKE-WAN2"
	CakeTypeRx        = "cake-rx"
	CakeTypeTx        = "cake-tx"
)

// EnsureCakeQueueTypes creates the cake-rx and cake-tx queue types in RouterOS 7.
func (c *Configurator) EnsureCakeQueueTypes() error {
	types, err := c.client.Get("queue/type")
	if err != nil {
		return err
	}

	hasRx := false
	hasTx := false
	for _, t := range types {
		name, _ := t["name"].(string)
		if name == CakeTypeRx {
			hasRx = true
		}
		if name == CakeTypeTx {
			hasTx = true
		}
	}

	if !hasRx {
		c.log("  ▶ Tạo Queue Type 'cake-rx' (Diffserv4, Triple-Isolate, RTT 100ms)...")
		_, _ = c.client.Add("queue/type", map[string]interface{}{
			"name":          CakeTypeRx,
			"kind":          "cake",
			"cake-diffserv": "diffserv4",
			"cake-flowmode": "triple-isolate",
			"cake-rtt":      "100ms",
			"cake-nat":      "yes",
			"cake-memlimit": "32M",
		})
	}

	if !hasTx {
		c.log("  ▶ Tạo Queue Type 'cake-tx' (Diffserv4, Triple-Isolate, RTT 100ms)...")
		_, _ = c.client.Add("queue/type", map[string]interface{}{
			"name":          CakeTypeTx,
			"kind":          "cake",
			"cake-diffserv": "diffserv4",
			"cake-flowmode": "triple-isolate",
			"cake-rtt":      "100ms",
			"cake-nat":      "yes",
			"cake-memlimit": "32M",
		})
	}
	return nil
}

// ConfigureCakeQoS creates or updates the Smart CAKE QoS Simple Queue at 95% line rate for Single WAN.
func (c *Configurator) ConfigureCakeQoS(downMbps, upMbps float64) error {
	if c.client == nil || !c.client.IsConnected() {
		return fmt.Errorf("chưa kết nối đến router")
	}
	if downMbps <= 0 || upMbps <= 0 {
		return fmt.Errorf("tốc độ gói cước Down/Up phải lớn hơn 0")
	}

	c.log(fmt.Sprintf("▶ Đang cấu hình Smart CAKE QoS (Gói cước: Down %.0fM / Up %.0fM)...", downMbps, upMbps))

	// Step 1: Ensure cake queue types
	if err := c.EnsureCakeQueueTypes(); err != nil {
		c.log(fmt.Sprintf("  ⚠ Cảnh báo kiểm tra queue type: %v", err))
	}

	// Clean up any dual-wan queues if present
	_ = c.removeQueueByName(CakeQueueNameWAN1)
	_ = c.removeQueueByName(CakeQueueNameWAN2)

	// Step 2: Compute 95% bandwidth limit
	down95k := int(downMbps * 0.95 * 1000)
	up95k := int(upMbps * 0.95 * 1000)
	maxLimit := fmt.Sprintf("%dk/%dk", up95k, down95k)
	c.log(fmt.Sprintf("  • Giới hạn 95%% chống Bufferbloat: Up %d Kbps, Down %d Kbps (Max-Limit: %s)", up95k, down95k, maxLimit))

	comment := fmt.Sprintf("%s Smart CAKE QoS (Gói %.0fM/%.0fM - Giới hạn 95%%)", AutoQueueComment, upMbps, downMbps)

	if err := c.upsertCakeQueue(CakeQueueName, "0.0.0.0/0", maxLimit, comment); err != nil {
		return fmt.Errorf("tạo CAKE QoS thất bại: %w", err)
	}

	c.log(fmt.Sprintf("✓ Đã kích hoạt thành công Smart CAKE QoS! (Triệt tiêu Bufferbloat, ưu tiên Gaming/VoIP, 95%% = %s)", maxLimit))
	return nil
}

// ConfigureCakeQoSDual creates or updates 2 separate Smart CAKE QoS Simple Queues for Dual WAN lines.
func (c *Configurator) ConfigureCakeQoSDual(wan1Down, wan1Up, wan2Down, wan2Up float64, wan1Iface, wan2Iface string) error {
	if c.client == nil || !c.client.IsConnected() {
		return fmt.Errorf("chưa kết nối đến router")
	}
	if wan1Down <= 0 || wan1Up <= 0 || wan2Down <= 0 || wan2Up <= 0 {
		return fmt.Errorf("tốc độ gói cước WAN 1 và WAN 2 phải lớn hơn 0")
	}

	if wan1Iface == "" {
		wan1Iface = "pppoe-wan1"
	}
	if wan2Iface == "" {
		wan2Iface = "pppoe-wan2"
	}

	c.log(fmt.Sprintf("▶ Đang cấu hình Smart CAKE QoS cho Dual WAN (WAN1: %.0fM/%.0fM, WAN2: %.0fM/%.0fM)...", wan1Down, wan1Up, wan2Down, wan2Up))

	if err := c.EnsureCakeQueueTypes(); err != nil {
		c.log(fmt.Sprintf("  ⚠ Cảnh báo kiểm tra queue type: %v", err))
	}

	// Remove single-wan queue if present to avoid conflict
	_ = c.removeQueueByName(CakeQueueName)

	// Calculate 95% limits
	down1_95 := int(wan1Down * 0.95 * 1000)
	up1_95 := int(wan1Up * 0.95 * 1000)
	maxLimit1 := fmt.Sprintf("%dk/%dk", up1_95, down1_95)

	down2_95 := int(wan2Down * 0.95 * 1000)
	up2_95 := int(wan2Up * 0.95 * 1000)
	maxLimit2 := fmt.Sprintf("%dk/%dk", up2_95, down2_95)

	comment1 := fmt.Sprintf("%s Smart CAKE QoS WAN1 (Gói %.0fM/%.0fM - 95%%)", AutoQueueComment, wan1Up, wan1Down)
	comment2 := fmt.Sprintf("%s Smart CAKE QoS WAN2 (Gói %.0fM/%.0fM - 95%%)", AutoQueueComment, wan2Up, wan2Down)

	c.log(fmt.Sprintf("  • Cấu hình WAN 1 (%s): Up %dK, Down %dK (Max: %s)...", wan1Iface, up1_95, down1_95, maxLimit1))
	if err := c.upsertCakeQueue(CakeQueueNameWAN1, wan1Iface, maxLimit1, comment1); err != nil {
		return fmt.Errorf("cấu hình CAKE WAN1 thất bại: %w", err)
	}

	c.log(fmt.Sprintf("  • Cấu hình WAN 2 (%s): Up %dK, Down %dK (Max: %s)...", wan2Iface, up2_95, down2_95, maxLimit2))
	if err := c.upsertCakeQueue(CakeQueueNameWAN2, wan2Iface, maxLimit2, comment2); err != nil {
		return fmt.Errorf("cấu hình CAKE WAN2 thất bại: %w", err)
	}

	c.log(fmt.Sprintf("✓ Đã kích hoạt Smart CAKE QoS cho cả 2 đường ISP! (WAN1: %s, WAN2: %s)", maxLimit1, maxLimit2))
	return nil
}

func (c *Configurator) removeQueueByName(targetName string) error {
	queues, err := c.client.Get("queue/simple")
	if err != nil {
		return err
	}
	for _, q := range queues {
		name, _ := q["name"].(string)
		if name == targetName {
			if id, ok := q[".id"].(string); ok && id != "" {
				_ = c.client.Remove("queue/simple", id)
			}
		}
	}
	return nil
}

func (c *Configurator) upsertCakeQueue(name, target, maxLimit, comment string) error {
	queues, err := c.client.Get("queue/simple")
	var existingID string
	if err == nil {
		for _, q := range queues {
			n, _ := q["name"].(string)
			if n == name {
				existingID, _ = q[".id"].(string)
				break
			}
		}
	}

	if existingID != "" {
		return c.client.Set("queue/simple", existingID, map[string]interface{}{
			"target":    target,
			"max-limit": maxLimit,
			"queue":     fmt.Sprintf("%s/%s", CakeTypeTx, CakeTypeRx),
			"comment":   comment,
		})
	}

	_, err = c.client.Add("queue/simple", map[string]interface{}{
		"name":      name,
		"target":    target,
		"queue":     fmt.Sprintf("%s/%s", CakeTypeTx, CakeTypeRx),
		"max-limit": maxLimit,
		"comment":   comment,
	})
	return err
}

// RemoveCakeQoS deletes all Smart-CAKE-QoS queues (Single or Dual WAN).
func (c *Configurator) RemoveCakeQoS() error {
	if c.client == nil || !c.client.IsConnected() {
		return fmt.Errorf("chưa kết nối đến router")
	}

	queues, err := c.client.Get("queue/simple")
	if err != nil {
		return err
	}

	c.log("▶ Đang gỡ bỏ Smart CAKE QoS...")
	for _, q := range queues {
		name, _ := q["name"].(string)
		if name == CakeQueueName || name == CakeQueueNameWAN1 || name == CakeQueueNameWAN2 {
			if id, ok := q[".id"].(string); ok && id != "" {
				_ = c.client.Remove("queue/simple", id)
			}
		}
	}
	c.log("✓ Đã gỡ bỏ Smart CAKE QoS thành công.")
	return nil
}

// GetCakeStatus checks if Smart CAKE QoS is currently deployed.
func (c *Configurator) GetCakeStatus() (active bool, maxLimit string, err error) {
	if c.client == nil || !c.client.IsConnected() {
		return false, "", nil
	}
	queues, err := c.client.Get("queue/simple")
	if err != nil {
		return false, "", err
	}
	hasWAN1, hasWAN2, hasSingle := false, false, false
	var limWAN1, limWAN2, limSingle string
	for _, q := range queues {
		name, _ := q["name"].(string)
		ml, _ := q["max-limit"].(string)
		if name == CakeQueueName {
			hasSingle = true
			limSingle = ml
		} else if name == CakeQueueNameWAN1 {
			hasWAN1 = true
			limWAN1 = ml
		} else if name == CakeQueueNameWAN2 {
			hasWAN2 = true
			limWAN2 = ml
		}
	}
	if hasWAN1 && hasWAN2 {
		return true, fmt.Sprintf("Dual WAN [WAN1: %s | WAN2: %s]", limWAN1, limWAN2), nil
	}
	if hasSingle {
		return true, limSingle, nil
	}
	if hasWAN1 {
		return true, fmt.Sprintf("WAN1 (%s)", limWAN1), nil
	}
	return false, "", nil
}

