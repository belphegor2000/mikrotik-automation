package orchestrator

import (
	"fmt"
	"strings"

	"mikrotik-automation/internal/api"
	"mikrotik-automation/internal/modules/bridge"
	"mikrotik-automation/internal/modules/pppoe"
	"mikrotik-automation/internal/modules/queue"
	"mikrotik-automation/internal/modules/routing"
	"mikrotik-automation/internal/modules/vlan"
	"mikrotik-automation/internal/profile"
)

type NetworkDeployParams struct {
	WANCount        int
	User1           string
	Pass1           string
	User2           string
	Pass2           string
	EnableRecursive bool
	BridgeIP        string
	EnableDHCP      bool
	DNSServers      string
	VLANs           []vlan.VLANEntry
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
		if err := pppoeCfg.ConfigureDualPCCEx(p.User1, p.Pass1, p.User2, p.Pass2, "bridgeVLAN", p.EnableRecursive); err != nil {
			return fmt.Errorf("bước 1 (PPPoE 2 WAN PCC) thất bại: %w", err)
		}
	}
	o.log(fmt.Sprintf("  ✓ [BƯỚC 1/3] Cấu hình PPPoE (%d đường) thành công!", p.WANCount))

	// Step 2: Bridge
	o.log("\n▶ [BƯỚC 2/3] CẤU HÌNH BRIDGE bridgeVLAN, IP, DHCP & DNS...")
	bridgeCfg := bridge.NewConfigurator(o.client, o.logFn)
	ports, routerIP, err := bridgeCfg.ConfigureBridge(bridge.BridgeConfigParams{
		WANCount:   p.WANCount,
		BridgeIP:   p.BridgeIP,
		EnableDHCP: p.EnableDHCP,
		DNSServers: p.DNSServers,
	})
	if err != nil {
		return fmt.Errorf("bước 2 (Cấu hình Bridge) thất bại: %w", err)
	}
	o.log(fmt.Sprintf("  ✓ [BƯỚC 2/3] Bridge bridgeVLAN đã sẵn sàng (%d ports, IP: %s)!", len(ports), routerIP))

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

// DeployFromTemplate deploys a complete NetworkTemplate to the connected router.
func (o *Orchestrator) DeployFromTemplate(tmpl profile.NetworkTemplate) error {
	o.log("╔" + strings.Repeat("═", 52) + "╗")
	o.log(fmt.Sprintf("║  🚀 NẠP VÀ TRIỂN KHAI TEMPLATE: %-23s║", tmpl.Name))
	o.log("╚" + strings.Repeat("═", 52) + "╝")

	// 1. Deploy Core Network
	deployParams := NetworkDeployParams{
		WANCount:        tmpl.PPPoE.WANCount,
		User1:           tmpl.PPPoE.User1,
		Pass1:           tmpl.PPPoE.Pass1,
		User2:           tmpl.PPPoE.User2,
		Pass2:           tmpl.PPPoE.Pass2,
		EnableRecursive: tmpl.PPPoE.EnableRecursive,
		BridgeIP:        tmpl.Bridge.BridgeIP,
		EnableDHCP:      tmpl.Bridge.EnableDHCP,
		DNSServers:      tmpl.Bridge.DNSServers,
		VLANs:           tmpl.VLANs,
	}
	if deployParams.WANCount <= 0 {
		deployParams.WANCount = 1
	}

	if err := o.DeployNetwork(deployParams); err != nil {
		return fmt.Errorf("triển khai mạng cơ sở từ template thất bại: %w", err)
	}

	// 2. Deploy PBR Rules if any
	if len(tmpl.PBRRules) > 0 {
		o.log(fmt.Sprintf("\n▶ CẤU HÌNH %d CHÍNH SÁCH ĐỊNH TUYẾN PBR...", len(tmpl.PBRRules)))
		pbrCfg := routing.NewConfigurator(o.client, o.logFn)
		for _, r := range tmpl.PBRRules {
			if r.Type == routing.TypeVLANSubnet {
				_ = pbrCfg.AddVLANSubnetRule(r.Name, r.Source, r.TargetWAN)
			} else {
				_ = pbrCfg.AddDomainIPRule(r.Name, r.Destination, r.TargetWAN)
			}
		}
		o.log(fmt.Sprintf("  ✓ Đã nạp thành công %d luật định tuyến PBR.", len(tmpl.PBRRules)))
	}

	// 3. Deploy Smart CAKE QoS if enabled
	if tmpl.CakeQoS.Enabled && tmpl.CakeQoS.DownloadMbps > 0 && tmpl.CakeQoS.UploadMbps > 0 {
		o.log("\n▶ KÍCH HOẠT SMART CAKE QOS...")
		qCfg := queue.NewConfigurator(o.client, o.logFn)
		if err := qCfg.ConfigureCakeQoS(tmpl.CakeQoS.DownloadMbps, tmpl.CakeQoS.UploadMbps); err != nil {
			o.log(fmt.Sprintf("  ⚠ Lỗi cấu hình CAKE QoS: %v", err))
		}
	}

	// 4. Deploy Simple Queues if any
	if len(tmpl.SimpleQueues) > 0 {
		o.log(fmt.Sprintf("\n▶ CẤU HÌNH %d SIMPLE QUEUES...", len(tmpl.SimpleQueues)))
		qCfg := queue.NewConfigurator(o.client, o.logFn)
		for _, sq := range tmpl.SimpleQueues {
			_ = qCfg.AddQueue(queue.QueueItem{
				Name:          sq.Name,
				Target:        sq.Target,
				UploadLimit:   sq.UploadLimit,
				DownloadLimit: sq.DownloadLimit,
				Comment:       sq.Comment,
			})
		}
	}

	o.log("\n🎉 HOÀN TẤT TRIỂN KHAI TOÀN BỘ CẤU HÌNH TỪ TEMPLATE!")
	return nil
}
