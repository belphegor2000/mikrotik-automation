package orchestrator

import (
	"fmt"
	"strings"

	"mikrotik-automation/internal/api"
	"mikrotik-automation/internal/modules/bridge"
	"mikrotik-automation/internal/modules/pppoe"
	"mikrotik-automation/internal/modules/vlan"
)

type NetworkDeployParams struct {
	WANCount int
	User1    string
	Pass1    string
	User2    string
	Pass2    string
	VLANs    []vlan.VLANEntry
}

type Orchestrator struct {
	client api.MikrotikClient
	logFn  func(string)
}

func NewOrchestrator(client api.MikrotikClient, logFn func(string)) *Orchestrator {
	return &Orchestrator{client: client, logFn: logFn}
}

func (o *Orchestrator) log(msg string) {
	if o.logFn != nil {
		o.logFn(msg)
	}
}

func (o *Orchestrator) DeployNetwork(p NetworkDeployParams) error {
	o.log("╔" + strings.Repeat("═", 48) + "╗")
	o.log("║  🚀 BẮT ĐẦU TRIỂN KHAI TOÀN DIỆN MẠNG ROUTEROS 7 ║")
	o.log("╚" + strings.Repeat("═", 48) + "╝")

	// Step 1: PPPoE WAN
	o.log("\n▶ [BƯỚC 1/3] CẤU HÌNH PPPoE WAN...")
	pppoeCfg := pppoe.NewConfigurator(o.client, o.logFn)
	if p.WANCount == 1 {
		if err := pppoeCfg.ConfigureSingle(p.User1, p.Pass1); err != nil {
			return fmt.Errorf("bước 1 (PPPoE 1 WAN) thất bại: %w", err)
		}
	} else {
		if err := pppoeCfg.ConfigureDualPCC(p.User1, p.Pass1, p.User2, p.Pass2, "bridgeVLAN"); err != nil {
			return fmt.Errorf("bước 1 (PPPoE 2 WAN PCC) thất bại: %w", err)
		}
	}
	o.log(fmt.Sprintf("  ✓ [BƯỚC 1/3] Cấu hình PPPoE (%d đường) thành công!", p.WANCount))

	// Step 2: Bridge
	o.log("\n▶ [BƯỚC 2/3] TẠO BRIDGE bridgeVLAN...")
	bridgeCfg := bridge.NewConfigurator(o.client, o.logFn)
	ports, err := bridgeCfg.CreateBridge(p.WANCount)
	if err != nil {
		return fmt.Errorf("bước 2 (Tạo Bridge) thất bại: %w", err)
	}
	o.log(fmt.Sprintf("  ✓ [BƯỚC 2/3] Bridge bridgeVLAN đã tạo (%d ports: %s)!", len(ports), strings.Join(ports, ", ")))

	// Step 3: VLANs
	o.log(fmt.Sprintf("\n▶ [BƯỚC 3/3] CẤU HÌNH %d VLANs & DHCP...", len(p.VLANs)))
	vlanCfg := vlan.NewConfigurator(o.client, o.logFn)
	if err := vlanCfg.Configure(p.VLANs); err != nil {
		return fmt.Errorf("bước 3 (Cấu hình VLANs) thất bại: %w", err)
	}
	o.log(fmt.Sprintf("  ✓ [BƯỚC 3/3] Hoàn tất cấu hình %d VLAN!", len(p.VLANs)))

	o.log("\n" + strings.Repeat("★", 50))
	o.log("🎉 CHÚC MỪNG! TOÀN BỘ HỆ THỐNG MẠNG ĐÃ ĐƯỢC TRIỂN KHAI HOÀN HẢO!")
	o.log(strings.Repeat("★", 50))
	return nil
}
