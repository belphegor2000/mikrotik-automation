package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/bridge"
	"mikrotik-automation/internal/modules/pppoe"
	"mikrotik-automation/internal/modules/vlan"
	"mikrotik-automation/internal/orchestrator"
)

// vlanRowLayout arranges a single row of VLAN inputs with proportional column widths.
type vlanRowLayout struct{}

func (l *vlanRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 5 {
		return
	}
	pad := float32(8)
	wID := float32(90)
	wGW := float32(140)
	wAct := float32(80)

	fixed := wID + wGW + wAct + 4*pad
	avail := size.Width - fixed
	if avail < 240 {
		avail = 240
	}
	wName := avail * 0.45
	wCIDR := avail * 0.55

	h := size.Height
	x := float32(0)

	// 0: VLAN ID
	objects[0].Move(fyne.NewPos(x, 0))
	objects[0].Resize(fyne.NewSize(wID, h))
	x += wID + pad

	// 1: Name
	objects[1].Move(fyne.NewPos(x, 0))
	objects[1].Resize(fyne.NewSize(wName, h))
	x += wName + pad

	// 2: CIDR
	objects[2].Move(fyne.NewPos(x, 0))
	objects[2].Resize(fyne.NewSize(wCIDR, h))
	x += wCIDR + pad

	// 3: Gateway
	objects[3].Move(fyne.NewPos(x, 0))
	objects[3].Resize(fyne.NewSize(wGW, h))
	x += wGW + pad

	// 4: Action (Delete)
	objects[4].Move(fyne.NewPos(x, 0))
	objects[4].Resize(fyne.NewSize(wAct, h))
}

func (l *vlanRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	wID := float32(90)
	wName := float32(140)
	wCIDR := float32(170)
	wGW := float32(140)
	wAct := float32(80)
	pad := float32(8)

	minW := wID + wName + wCIDR + wGW + wAct + 4*pad
	maxH := float32(36)
	for _, o := range objects {
		if o != nil && o.MinSize().Height > maxH {
			maxH = o.MinSize().Height
		}
	}
	return fyne.NewSize(minW, maxH)
}

type vlanInputRow struct {
	container  *fyne.Container
	entryID    *widget.Entry
	entryName  *widget.Entry
	entryCIDR  *widget.Entry
	lblGateway *widget.Label
	btnDelete  *widget.Button
}

type TabNetwork struct {
	state      *AppState
	window     fyne.Window
	radioWAN   *widget.RadioGroup
	entryUser1 *widget.Entry
	entryPass1 *widget.Entry
	entryUser2 *widget.Entry
	entryPass2 *widget.Entry
	wan1Box    *fyne.Container
	wan2Box    *fyne.Container

	// Sub-tab 2: Bridge LAN
	entryBridgeIP      *widget.Entry
	lblBridgeGW        *widget.Label
	lblBridgeDHCPRange *widget.Label
	chkBridgeDHCP      *widget.Check
	entryBridgeDNS     *widget.Entry
	lblBridgePorts     *widget.Label
	btnPreviewPort     *widget.Button
	viewBridge         *fyne.Container

	// Sub-tab 3: VLANs & Subnets (5 inline rows)
	vlanRowsBox   *fyne.Container
	vlanRows      []*vlanInputRow
	btnAddVLANRow *widget.Button
	btnClearVLAN  *widget.Button
	viewVLAN      *fyne.Container

	// Global controls
	subTabs      *container.AppTabs
	btnApplyStep *widget.Button
	btnDeployAll *widget.Button

	// Callbacks for host state and auto-reconnect
	GetConnectedHost func() string
	OnHostChanged    func(newHost string, onDone func(bool))
}

func NewTabNetwork(state *AppState, w fyne.Window) *TabNetwork {
	tn := &TabNetwork{state: state, window: w}

	// ── 1. PPPoE WAN Entries ─────────────────────────
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
	tn.entryBridgeIP = widget.NewEntry()
	tn.entryBridgeIP.SetText("192.168.88.1/24")
	tn.entryBridgeIP.SetPlaceHolder("vd: 192.168.88.1/24, 10.0.0.1/24, 172.16.1.1/24")

	tn.lblBridgeGW = widget.NewLabelWithStyle("Gateway Router: 192.168.88.1", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tn.lblBridgeDHCPRange = widget.NewLabel("Dải DHCP: 192.168.88.10 - 192.168.88.254")

	tn.entryBridgeIP.OnChanged = func(val string) {
		val = strings.TrimSpace(val)
		if val == "" {
			tn.lblBridgeGW.SetText("Gateway: -")
			tn.lblBridgeDHCPRange.SetText("Dải DHCP: -")
			return
		}
		_, gw, start, end, _, _, err := bridge.ParseBridgeCIDR(val)
		if err == nil {
			tn.lblBridgeGW.SetText(fmt.Sprintf("Gateway Router: %s", gw))
			tn.lblBridgeDHCPRange.SetText(fmt.Sprintf("Dải DHCP: %s - %s", start, end))
		} else {
			tn.lblBridgeGW.SetText("Gateway: (IP/CIDR chưa đúng)")
			tn.lblBridgeDHCPRange.SetText("Dải DHCP: (IP/CIDR chưa đúng)")
		}
	}

	tn.chkBridgeDHCP = widget.NewCheck("Tự động cấu hình DHCP Server cho Bridge", nil)
	tn.chkBridgeDHCP.SetChecked(true)

	tn.entryBridgeDNS = widget.NewEntry()
	tn.entryBridgeDNS.SetText("8.8.8.8, 1.1.1.1")
	tn.entryBridgeDNS.SetPlaceHolder("8.8.8.8, 1.1.1.1")

	tn.lblBridgePorts = widget.NewLabel("Nhấn 'Kiểm tra danh sách Port' để xem các cổng sẽ gán vào Bridge.")
	tn.btnPreviewPort = widget.NewButton("🔍 Kiểm tra danh sách Port", func() {
		tn.previewBridgePorts()
	})

	bridgeForm := widget.NewForm(
		widget.NewFormItem("Địa chỉ IP Bridge (CIDR):", tn.entryBridgeIP),
		widget.NewFormItem("Thông số tính toán:", container.NewVBox(tn.lblBridgeGW, tn.lblBridgeDHCPRange)),
		widget.NewFormItem("Dịch vụ DHCP Server:", tn.chkBridgeDHCP),
		widget.NewFormItem("DNS Servers:", tn.entryBridgeDNS),
	)

	descBridge := "• Tạo bridge trung tâm 'bridgeVLAN' và gán các cổng Ethernet còn lại.\n" +
		"• Hỗ trợ mọi dải IP tùy biến (10.x, 172.x, 192.168.x...). Tự tính Gateway và DHCP Pool.\n" +
		"• Nếu đổi IP khác IP hiện tại, hệ thống sẽ tự động chuyển kết nối sang IP mới."

	tn.viewBridge = container.NewVBox(
		widget.NewLabelWithStyle("Cấu Hình Bridge LAN (bridgeVLAN):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(descBridge),
		bridgeForm,
		container.NewHBox(tn.btnPreviewPort, tn.lblBridgePorts),
	)

	// ── 3. VLAN Sub-view (5 Inline Rows) ──────────────
	tn.vlanRowsBox = container.NewVBox()

	tn.btnAddVLANRow = widget.NewButton("➕ Thêm dòng VLAN", func() {
		tn.addVLANRow()
	})
	tn.btnAddVLANRow.Importance = widget.HighImportance

	tn.btnClearVLAN = widget.NewButton("🗑️ Xoá tất cả", func() {
		tn.clearAllVLANRows()
	})
	tn.btnClearVLAN.Importance = widget.DangerImportance

	vlanBar := container.NewHBox(tn.btnAddVLANRow, tn.btnClearVLAN)

	vlanHeader := container.New(&vlanRowLayout{},
		widget.NewLabelWithStyle("VLAN ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Tên VLAN", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Network (CIDR)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Gateway (Tự tính)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Thao tác", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	tn.initVLANRows()

	vlanScroll := container.NewVScroll(tn.vlanRowsBox)

	tn.viewVLAN = container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Quản Lý Phân Đoạn Mạng VLANs (5 dòng mặc định, tự tính Gateway):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			vlanBar,
			widget.NewSeparator(),
			vlanHeader,
		),
		nil, nil, nil,
		vlanScroll,
	)

	// Bottom action bar
	tn.btnApplyStep = widget.NewButton("⚙️ Áp dụng mục hiện tại", func() {
		tn.applyCurrentStep()
	})
	tn.btnDeployAll = widget.NewButton("🚀 Triển Khai Toàn Diện (WAN ➔ Bridge ➔ VLAN)", func() {
		tn.deployAll()
	})
	tn.btnDeployAll.Importance = widget.HighImportance

	return tn
}

func (tn *TabNetwork) createVLANRow(defaultID string, defaultName string, defaultCIDR string) *vlanInputRow {
	row := &vlanInputRow{
		entryID:    widget.NewEntry(),
		entryName:  widget.NewEntry(),
		entryCIDR:  widget.NewEntry(),
		lblGateway: widget.NewLabel("-"),
	}

	row.entryID.SetPlaceHolder(defaultID)
	row.entryName.SetPlaceHolder(defaultName)
	row.entryCIDR.SetPlaceHolder(defaultCIDR)

	row.btnDelete = widget.NewButton("🗑️ Xoá", func() {
		tn.removeOrClearRow(row)
	})

	row.entryCIDR.OnChanged = func(val string) {
		val = strings.TrimSpace(val)
		if val == "" {
			row.lblGateway.SetText("-")
			tn.syncVLANsQuietly()
			return
		}
		v := vlan.VLANEntry{CIDR: val}
		gw := v.Gateway()
		if gw != "" {
			row.lblGateway.SetText(gw)
		} else {
			row.lblGateway.SetText("(CIDR lỗi)")
		}
		tn.syncVLANsQuietly()
	}

	row.entryID.OnChanged = func(_ string) { tn.syncVLANsQuietly() }
	row.entryName.OnChanged = func(_ string) { tn.syncVLANsQuietly() }

	row.container = container.New(&vlanRowLayout{},
		row.entryID,
		row.entryName,
		row.entryCIDR,
		row.lblGateway,
		row.btnDelete,
	)
	return row
}

func (tn *TabNetwork) initVLANRows() {
	defaults := []struct {
		id, name, cidr string
	}{
		{"10", "users", "192.168.10.0/24"},
		{"20", "servers", "192.168.20.0/24"},
		{"30", "guests", "192.168.30.0/24"},
		{"40", "iot", "192.168.40.0/24"},
		{"50", "cameras", "192.168.50.0/24"},
	}

	tn.vlanRowsBox.Objects = nil
	tn.vlanRows = nil

	for i, d := range defaults {
		var r *vlanInputRow
		if i < len(tn.state.VLANs) {
			v := tn.state.VLANs[i]
			r = tn.createVLANRow(strconv.Itoa(v.VLANID), v.Name, v.CIDR)
			r.entryID.SetText(strconv.Itoa(v.VLANID))
			r.entryName.SetText(v.Name)
			r.entryCIDR.SetText(v.CIDR)
		} else {
			r = tn.createVLANRow(d.id, d.name, d.cidr)
		}
		tn.vlanRows = append(tn.vlanRows, r)
		tn.vlanRowsBox.Add(r.container)
	}

	for i := len(defaults); i < len(tn.state.VLANs); i++ {
		v := tn.state.VLANs[i]
		r := tn.createVLANRow(strconv.Itoa(v.VLANID), v.Name, v.CIDR)
		r.entryID.SetText(strconv.Itoa(v.VLANID))
		r.entryName.SetText(v.Name)
		r.entryCIDR.SetText(v.CIDR)
		tn.vlanRows = append(tn.vlanRows, r)
		tn.vlanRowsBox.Add(r.container)
	}
	tn.vlanRowsBox.Refresh()
}

func (tn *TabNetwork) addVLANRow() {
	idx := len(tn.vlanRows) + 1
	nextID := fmt.Sprintf("%d", idx*10)
	nextName := fmt.Sprintf("vlan%d", idx*10)
	nextCIDR := fmt.Sprintf("192.168.%d.0/24", idx*10)

	r := tn.createVLANRow(nextID, nextName, nextCIDR)
	tn.vlanRows = append(tn.vlanRows, r)
	tn.vlanRowsBox.Add(r.container)
	tn.vlanRowsBox.Refresh()
	tn.state.Log(fmt.Sprintf("+ Đã thêm dòng VLAN %d.", idx))
}

func (tn *TabNetwork) removeOrClearRow(row *vlanInputRow) {
	if len(tn.vlanRows) > 5 {
		newRows := make([]*vlanInputRow, 0, len(tn.vlanRows)-1)
		for _, r := range tn.vlanRows {
			if r != row {
				newRows = append(newRows, r)
			}
		}
		tn.vlanRows = newRows
		tn.vlanRowsBox.Remove(row.container)
		tn.vlanRowsBox.Refresh()
		tn.syncVLANsQuietly()
		tn.state.Log("Đã xoá 1 dòng VLAN.")
	} else {
		row.entryID.SetText("")
		row.entryName.SetText("")
		row.entryCIDR.SetText("")
		row.lblGateway.SetText("-")
		tn.syncVLANsQuietly()
		tn.state.Log("Đã làm sạch dòng VLAN.")
	}
}

func (tn *TabNetwork) clearAllVLANRows() {
	for i := len(tn.vlanRows) - 1; i >= 5; i-- {
		tn.vlanRowsBox.Remove(tn.vlanRows[i].container)
	}
	tn.vlanRows = tn.vlanRows[:5]
	for _, r := range tn.vlanRows {
		r.entryID.SetText("")
		r.entryName.SetText("")
		r.entryCIDR.SetText("")
		r.lblGateway.SetText("-")
	}
	tn.vlanRowsBox.Refresh()
	tn.state.VLANs = make([]vlan.VLANEntry, 0)
	tn.state.Log("Đã làm mới toàn bộ danh sách về 5 dòng trống.")
}

func (tn *TabNetwork) syncVLANsFromUI() ([]vlan.VLANEntry, error) {
	var result []vlan.VLANEntry
	seenIDs := make(map[int]bool)

	for i, r := range tn.vlanRows {
		idStr := strings.TrimSpace(r.entryID.Text)
		nameStr := strings.TrimSpace(r.entryName.Text)
		cidrStr := strings.TrimSpace(r.entryCIDR.Text)

		if idStr == "" && nameStr == "" && cidrStr == "" {
			continue
		}

		if idStr == "" {
			return nil, fmt.Errorf("dòng %d: VLAN ID không được để trống", i+1)
		}
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 || id > 4094 {
			return nil, fmt.Errorf("dòng %d: VLAN ID '%s' không hợp lệ (1-4094)", i+1, idStr)
		}
		if seenIDs[id] {
			return nil, fmt.Errorf("dòng %d: VLAN ID %d bị trùng lặp", i+1, id)
		}
		seenIDs[id] = true

		if nameStr == "" {
			return nil, fmt.Errorf("dòng %d: Tên VLAN không được để trống", i+1)
		}
		if cidrStr == "" {
			return nil, fmt.Errorf("dòng %d: Network CIDR không được để trống", i+1)
		}

		v := vlan.VLANEntry{
			VLANID: id,
			Name:   nameStr,
			CIDR:   cidrStr,
		}
		if v.Gateway() == "" {
			return nil, fmt.Errorf("dòng %d: Network CIDR '%s' không hợp lệ", i+1, cidrStr)
		}
		result = append(result, v)
	}

	tn.state.VLANs = result
	return result, nil
}

func (tn *TabNetwork) syncVLANsQuietly() {
	var result []vlan.VLANEntry
	seenIDs := make(map[int]bool)
	for _, r := range tn.vlanRows {
		idStr := strings.TrimSpace(r.entryID.Text)
		nameStr := strings.TrimSpace(r.entryName.Text)
		cidrStr := strings.TrimSpace(r.entryCIDR.Text)
		if idStr == "" || nameStr == "" || cidrStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 || id > 4094 || seenIDs[id] {
			continue
		}
		seenIDs[id] = true
		v := vlan.VLANEntry{VLANID: id, Name: nameStr, CIDR: cidrStr}
		if v.Gateway() != "" {
			result = append(result, v)
		}
	}
	tn.state.VLANs = result
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
			bridgeIP := strings.TrimSpace(tn.entryBridgeIP.Text)
			if bridgeIP == "" {
				bridgeIP = "192.168.88.1/24"
			}
			p := bridge.BridgeConfigParams{
				WANCount:   wanCount,
				BridgeIP:   bridgeIP,
				EnableDHCP: tn.chkBridgeDHCP.Checked,
				DNSServers: strings.TrimSpace(tn.entryBridgeDNS.Text),
			}
			cfg := bridge.NewConfigurator(c, tn.state.Log)
			_, routerIP, err := cfg.ConfigureBridge(p)
			if err != nil {
				tn.state.Log(fmt.Sprintf("✗ Lỗi cấu hình Bridge: %v", err))
				return
			}

			currentHost := ""
			if tn.GetConnectedHost != nil {
				currentHost = tn.GetConnectedHost()
			}

			// If router IP changed, trigger auto-reconnect
			if currentHost != "" && routerIP != "" && routerIP != currentHost {
				tn.state.Log(fmt.Sprintf("▶ Phát hiện IP Router đã đổi: %s ➔ %s. Bắt đầu tự động chuyển kết nối...", currentHost, routerIP))
				if tn.OnHostChanged != nil {
					tn.OnHostChanged(routerIP, func(success bool) {
						if success {
							tn.state.Log("✓ Hệ thống đã kết nối lại và hoạt động trên IP mới: " + routerIP)
						}
					})
				}
			}
		case "3. VLANs & Subnets":
			vlans, err := tn.syncVLANsFromUI()
			if err != nil {
				tn.state.Log(fmt.Sprintf("✗ %v", err))
				return
			}
			if len(vlans) == 0 {
				tn.state.Log("✗ Danh sách VLAN trống! Vui lòng điền ít nhất 1 dòng VLAN hợp lệ.")
				return
			}
			cfg := vlan.NewConfigurator(c, tn.state.Log)
			_ = cfg.Configure(vlans)
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

	vlans, err := tn.syncVLANsFromUI()
	if err != nil {
		tn.state.Log(fmt.Sprintf("✗ %v", err))
		return
	}
	if len(vlans) == 0 {
		tn.state.Log("✗ Danh sách VLAN trống! Vui lòng thêm ít nhất 1 VLAN hợp lệ.")
		return
	}

	bridgeIP := strings.TrimSpace(tn.entryBridgeIP.Text)
	if bridgeIP == "" {
		bridgeIP = "192.168.88.1/24"
	}

	currentHost := ""
	if tn.GetConnectedHost != nil {
		currentHost = tn.GetConnectedHost()
	}

	tn.btnDeployAll.Disable()
	tn.btnDeployAll.SetText("⏳ Đang triển khai toàn diện...")

	go func() {
		defer fyne.Do(func() {
			tn.btnDeployAll.Enable()
			tn.btnDeployAll.SetText("🚀 Triển Khai Toàn Diện (WAN ➔ Bridge ➔ VLAN)")
		})

		params := orchestrator.NetworkDeployParams{
			WANCount:   wanCount,
			User1:      tn.entryUser1.Text,
			Pass1:      tn.entryPass1.Text,
			User2:      tn.entryUser2.Text,
			Pass2:      tn.entryPass2.Text,
			BridgeIP:   bridgeIP,
			EnableDHCP: tn.chkBridgeDHCP.Checked,
			DNSServers: strings.TrimSpace(tn.entryBridgeDNS.Text),
			VLANs:      vlans,
		}
		orch := orchestrator.NewOrchestrator(c, tn.state.Log)
		_ = orch.DeployNetwork(params)

		// Check if Bridge IP changed from current host
		rIP, _, _, _, _, _, _ := bridge.ParseBridgeCIDR(bridgeIP)
		if currentHost != "" && rIP != "" && rIP != currentHost {
			tn.state.Log(fmt.Sprintf("▶ Phát hiện IP Router đã đổi: %s ➔ %s. Bắt đầu tự động chuyển kết nối...", currentHost, rIP))
			if tn.OnHostChanged != nil {
				tn.OnHostChanged(rIP, nil)
			}
		}
	}()
}
