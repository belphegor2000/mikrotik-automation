package dashboard

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"mikrotik-automation/internal/api"
)

type SystemInfo struct {
	Identity   string  `json:"identity"`
	BoardName  string  `json:"board_name"`
	Version    string  `json:"version"`
	Uptime     string  `json:"uptime"`
	CPULoad    int     `json:"cpu_load"`
	CPUCount   int     `json:"cpu_count"`
	FreeMemMB  float64 `json:"free_mem_mb"`
	TotalMemMB float64 `json:"total_mem_mb"`
	UsedMemMB  float64 `json:"used_mem_mb"`
	MemPct     float64 `json:"mem_pct"`
	FreeHDDMB  float64 `json:"free_hdd_mb"`
	TotalHDDMB float64 `json:"total_hdd_mb"`
	UsedHDDMB  float64 `json:"used_hdd_mb"`
	HDDPct     float64 `json:"hdd_pct"`
}

type InterfaceItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Running  bool   `json:"running"`
	Disabled bool   `json:"disabled"`
	Comment  string `json:"comment"`
	MAC      string `json:"mac"`
}

type Service struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewService(client api.MikrotikClient, logFn func(string)) *Service {
	return &Service{client: client, logFn: logFn}
}

func (s *Service) log(msg string) {
	if s.logFn != nil {
		s.logFn(msg)
	}
}

func ParseBytesValue(val interface{}) float64 {
	if val == nil {
		return 0
	}
	valStr := strings.TrimSpace(fmt.Sprintf("%v", val))
	if n, err := strconv.ParseInt(valStr, 10, 64); err == nil {
		return math.Round(float64(n)/(1024*1024)*10) / 10
	}

	re := regexp.MustCompile(`^([\d.]+)\s*([KkMmGgTt]?[iI]?[Bb])?$`)
	match := re.FindStringSubmatch(valStr)
	if len(match) > 1 {
		num, _ := strconv.ParseFloat(match[1], 64)
		unit := "B"
		if len(match) > 2 && match[2] != "" {
			unit = strings.ToUpper(match[2])
		}
		if strings.Contains(unit, "G") {
			return math.Round(num*1024*10) / 10
		} else if strings.Contains(unit, "M") {
			return math.Round(num*10) / 10
		} else if strings.Contains(unit, "K") {
			return math.Round(num/1024*10) / 10
		}
		return math.Round(num/(1024*1024)*10) / 10
	}
	return 0
}

func (s *Service) GetSystemInfo() (*SystemInfo, error) {
	res, err := s.client.GetOne("system/resource")
	if err != nil {
		return nil, err
	}
	ident, _ := s.client.GetOne("system/identity")

	name, _ := ident["name"].(string)
	if name == "" {
		name = "RouterOS"
	}

	board, _ := res["board-name"].(string)
	if board == "" {
		board, _ = res["platform"].(string)
	}
	if board == "" {
		board = "MikroTik Router"
	}

	ver, _ := res["version"].(string)
	uptime, _ := res["uptime"].(string)

	cpuStr := strings.ReplaceAll(fmt.Sprintf("%v", res["cpu-load"]), "%", "")
	cpuLoad, _ := strconv.Atoi(strings.TrimSpace(cpuStr))

	cpuCountStr := fmt.Sprintf("%v", res["cpu-count"])
	cpuCount, _ := strconv.Atoi(cpuCountStr)
	if cpuCount <= 0 {
		cpuCount = 1
	}

	freeMem := ParseBytesValue(res["free-memory"])
	totalMem := ParseBytesValue(res["total-memory"])
	usedMem := math.Max(0, totalMem-freeMem)
	memPct := 0.0
	if totalMem > 0 {
		memPct = math.Round((usedMem/totalMem)*1000) / 10
	}

	freeHDD := ParseBytesValue(res["free-hdd-space"])
	totalHDD := ParseBytesValue(res["total-hdd-space"])
	usedHDD := math.Max(0, totalHDD-freeHDD)
	hddPct := 0.0
	if totalHDD > 0 {
		hddPct = math.Round((usedHDD/totalHDD)*1000) / 10
	}

	return &SystemInfo{
		Identity:   name,
		BoardName:  board,
		Version:    ver,
		Uptime:     uptime,
		CPULoad:    cpuLoad,
		CPUCount:   cpuCount,
		FreeMemMB:  freeMem,
		TotalMemMB: totalMem,
		UsedMemMB:  usedMem,
		MemPct:     memPct,
		FreeHDDMB:  freeHDD,
		TotalHDDMB: totalHDD,
		UsedHDDMB:  usedHDD,
		HDDPct:     hddPct,
	}, nil
}

func (s *Service) GetInterfaces() ([]InterfaceItem, error) {
	raw, err := s.client.Get("interface")
	if err != nil {
		return nil, err
	}

	var items []InterfaceItem
	for _, iface := range raw {
		name, _ := iface["name"].(string)
		if name == "" {
			continue
		}
		runningStr := strings.ToLower(fmt.Sprintf("%v", iface["running"]))
		isRunning := runningStr == "true" || runningStr == "yes" || runningStr == "1"

		disabledStr := strings.ToLower(fmt.Sprintf("%v", iface["disabled"]))
		isDisabled := disabledStr == "true" || disabledStr == "yes" || disabledStr == "1"

		ifType, _ := iface["type"].(string)
		cmt, _ := iface["comment"].(string)
		mac, _ := iface["mac-address"].(string)

		items = append(items, InterfaceItem{
			Name:     name,
			Type:     ifType,
			Running:  isRunning,
			Disabled: isDisabled,
			Comment:  cmt,
			MAC:      mac,
		})
	}
	return items, nil
}

func (s *Service) RebootRouter() error {
	s.log("⚠️ Đang gửi lệnh khởi động lại (Reboot) đến Router...")
	_, err := s.client.ExecRaw(":execute \"/system reboot\"")
	if err != nil {
		_, err = s.client.Run("system/reboot", map[string]interface{}{})
	}
	if err != nil {
		return err
	}
	s.log("✓ Đã gửi lệnh Reboot thành công! Router đang khởi động lại.")
	return nil
}
