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
