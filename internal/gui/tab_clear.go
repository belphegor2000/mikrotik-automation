package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/cleaner"
)

type TabClear struct {
	state     *AppState
	window    fyne.Window
	container *fyne.Container
	btnClear  *widget.Button
}

func NewTabClear(state *AppState, w fyne.Window) *TabClear {
	tc := &TabClear{state: state, window: w}

	tc.btnClear = widget.NewButton("🗑️ Xoá Cấu Hình Tự Động (15 Bước)", func() {
		tc.confirmClear()
	})
	tc.btnClear.Importance = widget.DangerImportance

	return tc
}

func (tc *TabClear) BuildDesktopContainer() fyne.CanvasObject {
	desc := "⚠️ Chức năng này sẽ dọn dẹp và xoá an toàn tất cả cấu hình do công cụ tạo ra theo 15 bước tuần tự:\n" +
		"  1. DHCP Leases & Networks\n" +
		"  2. DHCP Servers & IP Pools\n" +
		"  3. IP Addresses của VLAN & Bridge\n" +
		"  4. Bridge VLAN table & VLAN interfaces\n" +
		"  5. Bridge bridgeVLAN và các port gán kèm\n" +
		"  6. PPPoE WAN clients (pppoe-wan1, pppoe-wan2)\n" +
		"  7. Firewall Filter, Mangle và NAT rules có tag [MK-AUTO]\n" +
		"  8. Routing Tables và Custom Routes\n" +
		"  9. IPSec VPN Policies, Identities, Peers, Profiles\n\n" +
		"Các cấu hình mặc định khác của router sẽ được giữ nguyên an toàn."

	card := container.NewVBox(
		widget.NewLabelWithStyle("🗑️ Dọn Dẹp & Xoá Cấu Hình RouterOS 7", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(desc),
		widget.NewSeparator(),
		container.NewPadded(tc.btnClear),
	)
	return card
}

func (tc *TabClear) BuildMobileContainer() fyne.CanvasObject {
	desc := "⚠️ Dọn dẹp an toàn 15 bước các cấu hình do tool tạo:\n" +
		"• DHCP, IP Pools, VLANs, Bridge\n" +
		"• PPPoE WAN clients & PCC Mangle\n" +
		"• Firewall [MK-AUTO] & Routing Tables\n" +
		"• IPSec VPN Policies & Peers\n\n" +
		"Các cấu hình gốc của RouterOS luôn an toàn."

	return container.NewVBox(
		widget.NewLabelWithStyle("🗑️ Dọn Dẹp Cấu Hình", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(desc),
		widget.NewSeparator(),
		container.NewPadded(tc.btnClear),
	)
}

func (tc *TabClear) Container() fyne.CanvasObject {
	return tc.BuildDesktopContainer()
}

func (tc *TabClear) confirmClear() {
	c := tc.state.GetClient()
	if !c.IsConnected() {
		tc.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	dialog.ShowConfirm("Xác Nhận Xoá Cấu Hình", "Bạn có chắc chắn muốn chạy quy trình xoá 15 bước không?", func(ok bool) {
		if !ok {
			return
		}
		tc.btnClear.Disable()
		tc.btnClear.SetText("⏳ Đang dọn dẹp...")
		go func() {
			defer fyne.Do(func() {
				tc.btnClear.Enable()
				tc.btnClear.SetText("🗑️ Xoá Cấu Hình Tự Động (15 Bước)")
			})
			cln := cleaner.NewCleaner(c, tc.state.Log)
			cln.CleanAll()
		}()
	}, tc.window)
}
