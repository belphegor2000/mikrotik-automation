package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/modules/nat"
	"mikrotik-automation/internal/modules/security"
)

type pfRowLayout struct{}

func (l *pfRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 6 {
		return
	}
	pad := float32(8)
	wProto := float32(60)
	wDstPort := float32(90)
	wTarget := float32(160)
	wWAN := float32(110)
	wAct := float32(75)

	fixed := wProto + wDstPort + wTarget + wWAN + wAct + 5*pad
	avail := size.Width - fixed
	if avail < 120 {
		avail = 120
	}
	wName := avail

	h := size.Height
	x := float32(0)

	// 0: Name
	objects[0].Move(fyne.NewPos(x, 0))
	objects[0].Resize(fyne.NewSize(wName, h))
	x += wName + pad

	// 1: Protocol
	objects[1].Move(fyne.NewPos(x, 0))
	objects[1].Resize(fyne.NewSize(wProto, h))
	x += wProto + pad

	// 2: Dst Port
	objects[2].Move(fyne.NewPos(x, 0))
	objects[2].Resize(fyne.NewSize(wDstPort, h))
	x += wDstPort + pad

	// 3: To IP:Port
	objects[3].Move(fyne.NewPos(x, 0))
	objects[3].Resize(fyne.NewSize(wTarget, h))
	x += wTarget + pad

	// 4: WAN Interface
	objects[4].Move(fyne.NewPos(x, 0))
	objects[4].Resize(fyne.NewSize(wWAN, h))
	x += wWAN + pad

	// 5: Action (Delete)
	objects[5].Move(fyne.NewPos(x, 0))
	objects[5].Resize(fyne.NewSize(wAct, h))
}

func (l *pfRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	wName := float32(130)
	wProto := float32(60)
	wDstPort := float32(90)
	wTarget := float32(160)
	wWAN := float32(110)
	wAct := float32(75)
	pad := float32(8)

	minW := wName + wProto + wDstPort + wTarget + wWAN + wAct + 5*pad
	maxH := float32(36)
	for _, o := range objects {
		if o != nil && o.MinSize().Height > maxH {
			maxH = o.MinSize().Height
		}
	}
	return fyne.NewSize(minW, maxH)
}

type TabSecurity struct {
	state         *AppState
	window        fyne.Window
	container     *fyne.Container
	lblMainTitle  *widget.Label
	subTabs       *container.AppTabs

	// Sub-tab 1: Firewall & Services
	lblSrvTitle   *widget.Label
	lblFWTitle    *widget.Label
	lblFWDesc     *widget.Label
	checks        map[string]*widget.Check
	btnDisable    *widget.Button
	btnFW         *widget.Button
	btnBruteForce *widget.Button
	serviceBoxes  *fyne.Container
	fwBoxes       *fyne.Container

	// Sub-tab 2: Port Forwarding & Hairpin NAT
	lblPFTitle      *widget.Label
	entryPFName     *widget.Entry
	selectPFProto   *widget.Select
	entryPFDstPort  *widget.Entry
	entryPFToIP     *widget.Entry
	entryPFToPort   *widget.Entry
	selectPFWAN     *widget.Select
	chkHairpin      *widget.Check
	btnAddPF        *widget.Button
	btnAutoHairpin  *widget.Button
	lblPFList       *widget.Label
	btnRefreshPF    *widget.Button
	pfHeaders       []*widget.Label
	pfList          *widget.List
	pfData          []nat.PortForwardRule
}

func NewTabSecurity(state *AppState, w fyne.Window) *TabSecurity {
	ts := &TabSecurity{
		state:  state,
		window: w,
		checks: make(map[string]*widget.Check),
		pfData: make([]nat.PortForwardRule, 0),
	}


	ts.lblMainTitle = widget.NewLabelWithStyle(i18n.T("sec_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// ── Sub-tab 1: Tường lửa & Dịch vụ ────────────────
	services := security.DefaultInsecureServices
	ts.lblSrvTitle = widget.NewLabelWithStyle(i18n.T("sec_srv_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	serviceBoxes := container.NewVBox(
		ts.lblSrvTitle,
	)

	for _, sv := range services {
		chk := widget.NewCheck(sv, nil)
		chk.SetChecked(true)
		ts.checks[sv] = chk
		serviceBoxes.Add(chk)
	}

	ts.btnDisable = widget.NewButton(i18n.T("sec_srv_btn"), func() {
		ts.disableServices()
	})
	ts.btnDisable.Importance = widget.WarningImportance

	ts.btnFW = widget.NewButton(i18n.T("sec_fw_btn"), func() {
		ts.applyFirewall()
	})
	ts.btnFW.Importance = widget.HighImportance

	ts.btnBruteForce = widget.NewButton(i18n.T("sec_bf_btn"), func() {
		ts.applyBruteForce()
	})
	ts.btnBruteForce.Importance = widget.HighImportance

	ts.lblFWDesc = widget.NewLabel(i18n.T("sec_fw_desc"))
	ts.lblFWDesc.Wrapping = fyne.TextWrapWord

	ts.lblFWTitle = widget.NewLabelWithStyle(i18n.T("sec_fw_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	ts.fwBoxes = container.NewVBox(
		ts.lblFWTitle,
		ts.lblFWDesc,
	)
	ts.serviceBoxes = serviceBoxes

	// ── Sub-tab 2: Port Forwarding & Hairpin NAT ──────
	ts.lblPFTitle = widget.NewLabelWithStyle(i18n.T("nat_form_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	ts.entryPFName = widget.NewEntry()
	ts.entryPFName.SetPlaceHolder("vd: Camera Sân, Web Server, NAS...")

	ts.selectPFProto = widget.NewSelect([]string{"TCP", "UDP"}, nil)
	ts.selectPFProto.SetSelected("TCP")

	ts.entryPFDstPort = widget.NewEntry()
	ts.entryPFDstPort.SetPlaceHolder("vd: 80, 8080, 8000-8010")

	ts.entryPFToIP = widget.NewEntry()
	ts.entryPFToIP.SetPlaceHolder("vd: 192.168.88.100")

	ts.entryPFToPort = widget.NewEntry()
	ts.entryPFToPort.SetPlaceHolder("vd: 80 (để trống nếu trùng port ngoài)")

	ts.selectPFWAN = widget.NewSelect([]string{"Tất cả (Mọi WAN)", "pppoe-wan1", "pppoe-wan2", "ether1"}, nil)
	ts.selectPFWAN.SetSelected("Tất cả (Mọi WAN)")

	ts.chkHairpin = widget.NewCheck(i18n.T("nat_chk_hairpin"), nil)
	ts.chkHairpin.SetChecked(true)

	ts.btnAddPF = widget.NewButton(i18n.T("nat_btn_add"), func() {
		ts.addPortForward()
	})
	ts.btnAddPF.Importance = widget.HighImportance

	ts.btnAutoHairpin = widget.NewButton(i18n.T("nat_btn_hairpin_all"), func() {
		ts.autoConfigureAllHairpin()
	})
	ts.btnAutoHairpin.Importance = widget.WarningImportance

	ts.lblPFList = widget.NewLabelWithStyle(i18n.T("nat_list_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ts.btnRefreshPF = widget.NewButton(i18n.T("nat_refresh_btn"), func() {
		ts.RefreshPF()
	})

	ts.pfHeaders = []*widget.Label{
		widget.NewLabelWithStyle(i18n.T("nat_col_name"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("nat_proto"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("nat_col_port"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("nat_col_target"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("nat_col_wan"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("nat_col_action"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}

	ts.pfList = widget.NewList(
		func() int {
			return len(ts.pfData)
		},
		func() fyne.CanvasObject {
			btnDel := widget.NewButton(i18n.T("nat_delete_btn"), nil)
			btnDel.Importance = widget.DangerImportance
			return container.New(&pfRowLayout{},
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				btnDel,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(ts.pfData) {
				return
			}
			item := ts.pfData[id]
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(item.Name)
			box.Objects[1].(*widget.Label).SetText(strings.ToUpper(item.Protocol))
			box.Objects[2].(*widget.Label).SetText(item.DstPort)

			targetStr := item.ToAddresses
			if item.ToPorts != "" && item.ToPorts != item.DstPort {
				targetStr = fmt.Sprintf("%s:%s", item.ToAddresses, item.ToPorts)
			}
			box.Objects[3].(*widget.Label).SetText(targetStr)

			wanStr := item.WANInterface
			if wanStr == "" {
				wanStr = "Tất cả"
			}
			box.Objects[4].(*widget.Label).SetText(wanStr)

			btnDel := box.Objects[5].(*widget.Button)
			btnDel.SetText(i18n.T("nat_delete_btn"))
			btnDel.OnTapped = func() {
				ts.confirmDeletePF(item)
			}
		},
	)

	ts.state.OnConnect = append(ts.state.OnConnect, func() {
		ts.RefreshPF()
	})

	return ts
}

func (ts *TabSecurity) BuildDesktopContainer() fyne.CanvasObject {
	// Sub-tab 1: Firewall & Services view
	buttonsRow := container.NewGridWithColumns(3, ts.btnDisable, ts.btnFW, ts.btnBruteForce)
	viewFirewall := container.NewVBox(
		container.NewGridWithColumns(2, ts.serviceBoxes, ts.fwBoxes),
		widget.NewSeparator(),
		buttonsRow,
	)

	// Sub-tab 2: Port Forwarding view
	formRow1 := container.NewGridWithColumns(3,
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("nat_name")), nil, ts.entryPFName),
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("nat_proto")), nil, ts.selectPFProto),
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("nat_dst_port")), nil, ts.entryPFDstPort),
	)
	formRow2 := container.NewGridWithColumns(3,
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("nat_to_ip")), nil, ts.entryPFToIP),
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("nat_to_port")), nil, ts.entryPFToPort),
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("nat_wan_iface")), nil, ts.selectPFWAN),
	)

	actionsBox := container.NewHBox(ts.chkHairpin, ts.btnAddPF, widget.NewSeparator(), ts.btnAutoHairpin)

	pfFormCard := container.NewVBox(
		ts.lblPFTitle,
		formRow1,
		formRow2,
		actionsBox,
	)

	listHeader := container.NewBorder(nil, nil, ts.lblPFList, ts.btnRefreshPF)
	columnHeaders := container.New(&pfRowLayout{},
		ts.pfHeaders[0],
		ts.pfHeaders[1],
		ts.pfHeaders[2],
		ts.pfHeaders[3],
		ts.pfHeaders[4],
		ts.pfHeaders[5],
	)

	pfTableCard := container.NewBorder(
		container.NewVBox(listHeader, widget.NewSeparator(), columnHeaders),
		nil, nil, nil,
		ts.pfList,
	)

	viewNAT := container.NewBorder(
		container.NewVBox(pfFormCard, widget.NewSeparator()),
		nil, nil, nil,
		pfTableCard,
	)

	ts.subTabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("sec_sub_firewall"), viewFirewall),
		container.NewTabItem(i18n.T("sec_sub_nat"), viewNAT),
	)

	return container.NewBorder(
		ts.lblMainTitle,
		nil, nil, nil,
		ts.subTabs,
	)
}

func (ts *TabSecurity) UpdateLanguage() {
	if ts.lblMainTitle != nil {
		ts.lblMainTitle.SetText(i18n.T("sec_title"))
	}
	if ts.lblSrvTitle != nil {
		ts.lblSrvTitle.SetText(i18n.T("sec_srv_title"))
	}
	if ts.btnDisable != nil {
		ts.btnDisable.SetText(i18n.T("sec_srv_btn"))
	}
	if ts.lblFWTitle != nil {
		ts.lblFWTitle.SetText(i18n.T("sec_fw_title"))
	}
	if ts.lblFWDesc != nil {
		ts.lblFWDesc.SetText(i18n.T("sec_fw_desc"))
	}
	if ts.btnFW != nil {
		ts.btnFW.SetText(i18n.T("sec_fw_btn"))
	}
	if ts.btnBruteForce != nil {
		ts.btnBruteForce.SetText(i18n.T("sec_bf_btn"))
	}
	if ts.subTabs != nil && len(ts.subTabs.Items) >= 2 {
		ts.subTabs.Items[0].Text = i18n.T("sec_sub_firewall")
		ts.subTabs.Items[1].Text = i18n.T("sec_sub_nat")
		ts.subTabs.Refresh()
	}
	if ts.lblPFTitle != nil {
		ts.lblPFTitle.SetText(i18n.T("nat_form_title"))
	}
	if ts.chkHairpin != nil {
		ts.chkHairpin.SetText(i18n.T("nat_chk_hairpin"))
	}
	if ts.btnAddPF != nil {
		ts.btnAddPF.SetText(i18n.T("nat_btn_add"))
	}
	if ts.btnAutoHairpin != nil {
		ts.btnAutoHairpin.SetText(i18n.T("nat_btn_hairpin_all"))
	}
	if ts.lblPFList != nil {
		ts.lblPFList.SetText(i18n.T("nat_list_title"))
	}
	if ts.btnRefreshPF != nil {
		ts.btnRefreshPF.SetText(i18n.T("nat_refresh_btn"))
	}
	if len(ts.pfHeaders) >= 6 {
		ts.pfHeaders[0].SetText(i18n.T("nat_col_name"))
		ts.pfHeaders[1].SetText(i18n.T("nat_proto"))
		ts.pfHeaders[2].SetText(i18n.T("nat_col_port"))
		ts.pfHeaders[3].SetText(i18n.T("nat_col_target"))
		ts.pfHeaders[4].SetText(i18n.T("nat_col_wan"))
		ts.pfHeaders[5].SetText(i18n.T("nat_col_action"))
	}
	if ts.pfList != nil {
		ts.pfList.Refresh()
	}
}

func (ts *TabSecurity) Container() fyne.CanvasObject {
	return ts.BuildDesktopContainer()
}

func (ts *TabSecurity) disableServices() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	var selected []string
	for name, chk := range ts.checks {
		if chk.Checked {
			selected = append(selected, name)
		}
	}

	ts.btnDisable.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnDisable.Enable() })
		sec := security.NewConfigurator(c, ts.state.Log)
		_, _ = sec.DisableInsecureServices(selected)
	}()
}

func (ts *TabSecurity) applyFirewall() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	ts.btnFW.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnFW.Enable() })
		sec := security.NewConfigurator(c, ts.state.Log)
		_ = sec.ConfigureFirewallFilters(nil)
	}()
}

func (ts *TabSecurity) applyBruteForce() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	ts.btnBruteForce.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnBruteForce.Enable() })
		sec := security.NewConfigurator(c, ts.state.Log)
		_ = sec.ConfigureBruteForceProtection()
	}()
}

func (ts *TabSecurity) addPortForward() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	name := strings.TrimSpace(ts.entryPFName.Text)
	proto := strings.ToLower(ts.selectPFProto.Selected)
	dstPort := strings.TrimSpace(ts.entryPFDstPort.Text)
	toIP := strings.TrimSpace(ts.entryPFToIP.Text)
	toPort := strings.TrimSpace(ts.entryPFToPort.Text)
	wan := strings.TrimSpace(ts.selectPFWAN.Selected)
	hairpin := ts.chkHairpin.Checked

	if name == "" || dstPort == "" || toIP == "" {
		ts.state.Log("⚠ Vui lòng nhập đầy đủ: Tên dịch vụ, Cổng ngoài, IP đích nội bộ.")
		return
	}

	rule := nat.PortForwardRule{
		Name:          name,
		Protocol:      proto,
		DstPort:       dstPort,
		ToAddresses:   toIP,
		ToPorts:       toPort,
		WANInterface:  wan,
		EnableHairpin: hairpin,
	}

	ts.btnAddPF.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnAddPF.Enable() })
		cfg := nat.NewConfigurator(c, ts.state.Log)
		if err := cfg.AddPortForward(rule); err != nil {
			ts.state.Log(fmt.Sprintf("❌ Lỗi mở port: %v", err))
		} else {
			fyne.Do(func() {
				ts.entryPFName.SetText("")
				ts.entryPFDstPort.SetText("")
				ts.entryPFToIP.SetText("")
				ts.entryPFToPort.SetText("")
				ts.RefreshPF()
			})
		}
	}()
}

func (ts *TabSecurity) RefreshPF() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		return
	}

	ts.btnRefreshPF.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnRefreshPF.Enable() })
		cfg := nat.NewConfigurator(c, ts.state.Log)
		rules, err := cfg.ListPortForwards()
		fyne.Do(func() {
			if err == nil {
				ts.pfData = rules
				ts.pfList.Refresh()
				ts.state.Log(fmt.Sprintf("✓ Đã tải %d luật mở cổng (Port Forwarding).", len(rules)))
			}
		})
	}()
}

func (ts *TabSecurity) autoConfigureAllHairpin() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	ts.btnAutoHairpin.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnAutoHairpin.Enable() })
		cfg := nat.NewConfigurator(c, ts.state.Log)

		// Collect known subnets from AppState VLANs
		var customSubnets []string
		for _, v := range ts.state.GetVLANs() {
			if v.CIDR != "" {
				customSubnets = append(customSubnets, v.CIDR)
			}
		}

		count, err := cfg.EnableHairpinAllSubnets(customSubnets...)
		if err != nil {
			ts.state.Log(fmt.Sprintf("⚠ Lỗi cấu hình Auto Hairpin: %v", err))
		} else {
			ts.state.Log(fmt.Sprintf("✓ Đã kích hoạt Auto Hairpin NAT cho %d dải mạng nội bộ!", count))
		}
	}()
}

func (ts *TabSecurity) confirmDeletePF(rule nat.PortForwardRule) {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	msg := fmt.Sprintf("Bạn có chắc chắn muốn đóng cổng '%s' (%s %s ➔ %s)?", rule.Name, rule.Protocol, rule.DstPort, rule.ToAddresses)
	dialog.ShowConfirm("Xác Nhận Xoá Mở Port", msg, func(ok bool) {
		if !ok {
			return
		}
		go func() {
			cfg := nat.NewConfigurator(c, ts.state.Log)
			if err := cfg.DeletePortForward(rule.ID); err == nil {
				ts.RefreshPF()
			}
		}()
	}, ts.window)
}
