package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/bridge"
	"mikrotik-automation/internal/modules/pppoe"
	"mikrotik-automation/internal/modules/vlan"
	"mikrotik-automation/internal/orchestrator"
)

type TabNetwork struct {
	state          *AppState
	window         fyne.Window
	container      *fyne.Container
	radioWAN       *widget.RadioGroup
	entryUser1     *widget.Entry
	entryPass1     *widget.Entry
	entryUser2     *widget.Entry
	entryPass2     *widget.Entry
	lblBridgePorts *widget.Label
	btnPreviewPort *widget.Button
	vlanTable      *widget.Table
	btnApplyStep   *widget.Button
	btnDeployAll   *widget.Button
	subTabs        *container.AppTabs
	wan1Box        *fyne.Container
	wan2Box        *fyne.Container
	viewBridge     *fyne.Container
	viewVLAN       *fyne.Container
}

func NewTabNetwork(state *AppState, w fyne.Window) *TabNetwork {
	tn := &TabNetwork{state: state, window: w}

	// Initialize entries first before radio callback
	tn.entryUser1 = widget.NewEntry()
	tn.entryUser1.SetPlaceHolder("pppoe_username")
	tn.entryPass1 = widget.NewPasswordEntry()
	tn.entryPass1.SetPlaceHolder("mật khẩu")

	tn.entryUser2 = widget.NewEntry()
	tn.entryUser2.SetPlaceHolder("pppoe_username")
	tn.entryPass2 = widget.NewPasswordEntry()
	tn.entryPass2.SetPlaceHolder("mật khẩu")
	tn.entryUser2.Disable()
	tn.entryPass2.Disable()

	// ── 1. PPPoE WAN Sub-view ─────────────────────────
	tn.radioWAN = widget.NewRadioGroup([]string{"1 đường (ether1)", "2 đường (ether1 + ether2) + PCC"}, func(val string) {
		if tn.entryUser2 == nil || tn.entryPass2 == nil {
			return
		}
		isDual := strings.HasPrefix(val, "2")
		if isDual {
			tn.entryUser2.Enable()
			tn.entryPass2.Enable()
		} else {
			tn.entryUser2.Disable()
			tn.entryPass2.Disable()
		}
	})
	tn.radioWAN.SetSelected("1 đường (ether1)")

	tn.wan1Box = container.NewVBox(
		widget.NewLabelWithStyle("WAN 1 (ether1)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Tài khoản:"), tn.entryUser1,
		widget.NewLabel("Mật khẩu:"), tn.entryPass1,
	)

	tn.wan2Box = container.NewVBox(
		widget.NewLabelWithStyle("WAN 2 (ether2)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Tài khoản:"), tn.entryUser2,
		widget.NewLabel("Mật khẩu:"), tn.entryPass2,
	)

	// ── 2. Bridge Sub-view ────────────────────────────
	descBridge := "• Tạo bridge trung tâm 'bridgeVLAN' và gán các cổng Ethernet còn lại.\n" +
		"• Tự động loại trừ ether1 (nếu dùng 1 WAN) hoặc ether1 & ether2 (nếu dùng 2 WAN)."
	tn.lblBridgePorts = widget.NewLabel("Nhấn 'Kiểm tra danh sách Port' để xem các cổng sẽ gán vào Bridge.")

	tn.btnPreviewPort = widget.NewButton("🔍 Kiểm tra danh sách Port", func() {
		tn.previewBridgePorts()
	})

	tn.viewBridge = container.NewVBox(
		widget.NewLabelWithStyle("Cấu Hình Bridge LAN (bridgeVLAN):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(descBridge),
		tn.btnPreviewPort,
		tn.lblBridgePorts,
	)

	// ── 3. VLAN Sub-view ──────────────────────────────
	btnAddVLAN := widget.NewButton("➕ Thêm VLAN", func() {
		tn.showAddVLANDialog()
	})
	btnClearVLAN := widget.NewButton("🗑️ Xoá tất cả", func() {
		tn.state.VLANs = make([]vlan.VLANEntry, 0)
		tn.vlanTable.Refresh()
		tn.state.Log("Đã xoá danh sách VLAN.")
	})

	vlanBar := container.NewHBox(btnAddVLAN, btnClearVLAN)

	headers := []string{"VLAN ID", "Tên VLAN", "Network (CIDR)", "Gateway", "Thao tác"}
	tn.vlanTable = widget.NewTable(
		func() (int, int) {
			return len(tn.state.VLANs) + 1, 5
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				lbl.SetText(headers[id.Col])
				return
			}
			lbl.TextStyle = fyne.TextStyle{Bold: false}
			v := tn.state.VLANs[id.Row-1]
			switch id.Col {
			case 0:
				lbl.SetText(strconv.Itoa(v.VLANID))
			case 1:
				lbl.SetText(v.Name)
			case 2:
				lbl.SetText(v.CIDR)
			case 3:
				lbl.SetText(v.Gateway())
			case 4:
				lbl.SetText("Xoá")
			}
		},
	)
	tn.vlanTable.SetColumnWidth(0, 80)
	tn.vlanTable.SetColumnWidth(1, 140)
	tn.vlanTable.SetColumnWidth(2, 160)
	tn.vlanTable.SetColumnWidth(3, 140)
	tn.vlanTable.SetColumnWidth(4, 80)

	tn.viewVLAN = container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Quản Lý Phân Đoạn Mạng VLANs:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			vlanBar,
		),
		nil, nil, nil,
		tn.vlanTable,
	)

	// Bottom action bar
	tn.btnApplyStep = widget.NewButton("⚙️ Áp dụng mục hiện tại", func() {
		tn.applyCurrentStep()
	})
	tn.btnDeployAll = widget.NewButton("🚀 Triển Khai Toàn Diện (WAN ➔ Bridge ➔ VLAN)", func() {
		tn.deployAll()
	})

	return tn
}

func (tn *TabNetwork) BuildDesktopContainer() fyne.CanvasObject {
	viewPPPoE := container.NewVBox(
		widget.NewLabelWithStyle("Cấu Hình Đường Truyền PPPoE WAN:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tn.radioWAN,
		container.NewGridWithColumns(2, tn.wan1Box, tn.wan2Box),
		widget.NewLabel("Khi chọn 2 đường: Hệ thống tự động tạo Mangle PCC, Dual Default Routes & NAT masquerade."),
	)

	tn.subTabs = container.NewAppTabs(
		container.NewTabItem("1. PPPoE WAN", viewPPPoE),
		container.NewTabItem("2. Bridge LAN", tn.viewBridge),
		container.NewTabItem("3. VLANs & Subnets", tn.viewVLAN),
	)

	bottomBar := container.NewBorder(nil, nil, tn.btnApplyStep, tn.btnDeployAll)

	return container.NewBorder(
		widget.NewLabelWithStyle("🌐 Cấu Hình Mạng Toàn Diện", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		bottomBar,
		nil,
		nil,
		tn.subTabs,
	)
}

func (tn *TabNetwork) BuildMobileContainer() fyne.CanvasObject {
	viewPPPoE := container.NewVBox(
		widget.NewLabelWithStyle("Cấu Hình PPPoE WAN:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tn.radioWAN,
		tn.wan1Box,
		widget.NewSeparator(),
		tn.wan2Box,
	)

	mobileSubTabs := container.NewAppTabs(
		container.NewTabItem("1. WAN", viewPPPoE),
		container.NewTabItem("2. Bridge", tn.viewBridge),
		container.NewTabItem("3. VLAN", tn.viewVLAN),
	)

	bottomBar := container.NewVBox(
		container.NewPadded(tn.btnApplyStep),
		container.NewPadded(tn.btnDeployAll),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("🌐 Cấu Hình Mạng", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		mobileSubTabs,
		widget.NewSeparator(),
		bottomBar,
	)
}

func (tn *TabNetwork) Container() fyne.CanvasObject {
	return tn.BuildDesktopContainer()
}

func (tn *TabNetwork) previewBridgePorts() {
	c := tn.state.GetClient()
	if !c.IsConnected() {
		tn.state.Log("✗ Chưa kết nối đến router!")
		return
	}
	wanCount := 1
	if strings.HasPrefix(tn.radioWAN.Selected, "2") {
		wanCount = 2
	}

	tn.btnPreviewPort.Disable()
	go func() {
		b := bridge.NewConfigurator(c, tn.state.Log)
		ports, err := b.GetAvailablePorts(wanCount)
		fyne.Do(func() {
			tn.btnPreviewPort.Enable()
			if err != nil {
				tn.state.Log(fmt.Sprintf("✗ Lỗi kiểm tra ports: %v", err))
				return
			}
			tn.lblBridgePorts.SetText(fmt.Sprintf("Ports sẽ gán vào bridge: %s", strings.Join(ports, ", ")))
			tn.state.Log(fmt.Sprintf("Tìm thấy %d ports khả dụng: %s", len(ports), strings.Join(ports, ", ")))
		})
	}()
}

func (tn *TabNetwork) showAddVLANDialog() {
	entryID := widget.NewEntry()
	entryID.SetPlaceHolder("10")
	entryName := widget.NewEntry()
	entryName.SetPlaceHolder("users")
	entryCIDR := widget.NewEntry()
	entryCIDR.SetPlaceHolder("192.168.10.0/24")

	items := []*widget.FormItem{
		{Text: "VLAN ID (1-4094):", Widget: entryID},
		{Text: "Tên VLAN:", Widget: entryName},
		{Text: "Network CIDR:", Widget: entryCIDR},
	}

	dialog.ShowForm("Thêm VLAN Mới", "Thêm", "Huỷ", items, func(ok bool) {
		if !ok {
			return
		}
		id, err := strconv.Atoi(entryID.Text)
		if err != nil || id < 1 || id > 4094 {
			tn.state.Log("✗ VLAN ID không hợp lệ (1-4094)!")
			return
		}
		if entryName.Text == "" {
			tn.state.Log("✗ Tên VLAN không được để trống!")
			return
		}
		newVlan := vlan.VLANEntry{
			VLANID: id,
			Name:   entryName.Text,
			CIDR:   entryCIDR.Text,
		}
		tn.state.VLANs = append(tn.state.VLANs, newVlan)
		tn.vlanTable.Refresh()
		tn.state.Log(fmt.Sprintf("+ Đã thêm VLAN %d (%s) - %s", id, newVlan.Name, newVlan.CIDR))
	}, tn.window)
}

func (tn *TabNetwork) applyCurrentStep() {
	c := tn.state.GetClient()
	if !c.IsConnected() {
		tn.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	selected := tn.subTabs.Selected()
	if selected == nil {
		return
	}

	tn.btnApplyStep.Disable()
	go func() {
		defer fyne.Do(func() { tn.btnApplyStep.Enable() })

		switch selected.Text {
		case "1. PPPoE WAN":
			wanCount := 1
			if strings.HasPrefix(tn.radioWAN.Selected, "2") {
				wanCount = 2
			}
			cfg := pppoe.NewConfigurator(c, tn.state.Log)
			if wanCount == 1 {
				_ = cfg.ConfigureSingle(tn.entryUser1.Text, tn.entryPass1.Text)
			} else {
				_ = cfg.ConfigureDualPCC(tn.entryUser1.Text, tn.entryPass1.Text, tn.entryUser2.Text, tn.entryPass2.Text, "bridgeVLAN")
			}
		case "2. Bridge LAN":
			wanCount := 1
			if strings.HasPrefix(tn.radioWAN.Selected, "2") {
				wanCount = 2
			}
			cfg := bridge.NewConfigurator(c, tn.state.Log)
			_, _ = cfg.CreateBridge(wanCount)
		case "3. VLANs & Subnets":
			cfg := vlan.NewConfigurator(c, tn.state.Log)
			_ = cfg.Configure(tn.state.VLANs)
		}
	}()
}

func (tn *TabNetwork) deployAll() {
	c := tn.state.GetClient()
	if !c.IsConnected() {
		tn.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	wanCount := 1
	if strings.HasPrefix(tn.radioWAN.Selected, "2") {
		wanCount = 2
	}

	if tn.entryUser1.Text == "" || tn.entryPass1.Text == "" {
		tn.state.Log("✗ Vui lòng điền thông tin PPPoE WAN 1!")
		return
	}
	if wanCount == 2 && (tn.entryUser2.Text == "" || tn.entryPass2.Text == "") {
		tn.state.Log("✗ Vui lòng điền thông tin PPPoE WAN 2!")
		return
	}
	if len(tn.state.VLANs) == 0 {
		tn.state.Log("✗ Danh sách VLAN trống! Vui lòng thêm ít nhất 1 VLAN.")
		return
	}

	tn.btnDeployAll.Disable()
	tn.btnDeployAll.SetText("⏳ Đang triển khai toàn diện...")

	go func() {
		defer fyne.Do(func() {
			tn.btnDeployAll.Enable()
			tn.btnDeployAll.SetText("🚀 Triển Khai Toàn Diện (WAN ➔ Bridge ➔ VLAN)")
		})

		params := orchestrator.NetworkDeployParams{
			WANCount: wanCount,
			User1:    tn.entryUser1.Text,
			Pass1:    tn.entryPass1.Text,
			User2:    tn.entryUser2.Text,
			Pass2:    tn.entryPass2.Text,
			VLANs:    tn.state.VLANs,
		}
		orch := orchestrator.NewOrchestrator(c, tn.state.Log)
		_ = orch.DeployNetwork(params)
	}()
}
