package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/modules/queue"
)

type queueRowLayout struct{}

func (l *queueRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 5 {
		return
	}
	pad := float32(8)
	wAct := float32(80)
	wUp := float32(110)
	wDown := float32(110)

	fixed := wAct + wUp + wDown + 4*pad
	avail := size.Width - fixed
	if avail < 200 {
		avail = 200
	}
	wName := avail * 0.45
	wTarget := avail * 0.55

	h := size.Height
	x := float32(0)

	// 0: Name
	objects[0].Move(fyne.NewPos(x, 0))
	objects[0].Resize(fyne.NewSize(wName, h))
	x += wName + pad

	// 1: Target
	objects[1].Move(fyne.NewPos(x, 0))
	objects[1].Resize(fyne.NewSize(wTarget, h))
	x += wTarget + pad

	// 2: Upload
	objects[2].Move(fyne.NewPos(x, 0))
	objects[2].Resize(fyne.NewSize(wUp, h))
	x += wUp + pad

	// 3: Download
	objects[3].Move(fyne.NewPos(x, 0))
	objects[3].Resize(fyne.NewSize(wDown, h))
	x += wDown + pad

	// 4: Action (Delete)
	objects[4].Move(fyne.NewPos(x, 0))
	objects[4].Resize(fyne.NewSize(wAct, h))
}

func (l *queueRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	wName := float32(140)
	wTarget := float32(160)
	wUp := float32(110)
	wDown := float32(110)
	wAct := float32(80)
	pad := float32(8)

	minW := wName + wTarget + wUp + wDown + wAct + 4*pad
	maxH := float32(36)
	for _, o := range objects {
		if o != nil && o.MinSize().Height > maxH {
			maxH = o.MinSize().Height
		}
	}
	return fyne.NewSize(minW, maxH)
}

type TabQueue struct {
	state   *AppState
	window  fyne.Window
	subTabs *container.AppTabs

	// Sub-tab 1: Simple Queue fields
	lblTitle         *widget.Label
	lblFormTitle     *widget.Label
	lblQueueName     *widget.Label
	lblQueueTarget   *widget.Label
	lblQueueUpload   *widget.Label
	lblQueueDownload *widget.Label
	entryName        *widget.Entry
	entryTarget      *widget.Entry
	btnFromVLAN      *widget.Button
	selectUpload     *widget.SelectEntry
	selectDownload   *widget.SelectEntry
	btnAdd           *widget.Button
	lblListTitle     *widget.Label
	btnRefresh       *widget.Button
	headerLabels     []*widget.Label
	queueList        *widget.List
	queueData        []queue.QueueItem

	// Sub-tab 2: Smart CAKE QoS fields
	lblCakeTitle       *widget.Label
	lblCakeDesc        *widget.Label
	lblCakeStatus      *widget.Label
	btnRefreshCake     *widget.Button
	radioCakeMode      *widget.RadioGroup
	singleCakeBox      *fyne.Container
	dualCakeBox        *fyne.Container
	lblCakeDown        *widget.Label
	entryCakeDown      *widget.Entry
	lblCakeUp          *widget.Label
	entryCakeUp        *widget.Entry
	lblCakeWAN1Down    *widget.Label
	entryCakeWAN1Down  *widget.Entry
	lblCakeWAN1Up      *widget.Label
	entryCakeWAN1Up    *widget.Entry
	lblCakeWAN1Iface   *widget.Label
	entryCakeWAN1Iface *widget.Entry
	lblCakeWAN2Down    *widget.Label
	entryCakeWAN2Down  *widget.Entry
	lblCakeWAN2Up      *widget.Label
	entryCakeWAN2Up    *widget.Entry
	lblCakeWAN2Iface   *widget.Label
	entryCakeWAN2Iface *widget.Entry
	lblCakeDualNote    *widget.Label
	lblCakeCalc        *widget.Label
	btnApplyCake       *widget.Button
	btnRemoveCake      *widget.Button
}

var ratePresets = []string{
	"Unlimited",
	"1M",
	"2M",
	"5M",
	"10M",
	"15M",
	"20M",
	"30M",
	"50M",
	"100M",
	"200M",
}

func NewTabQueue(state *AppState, w fyne.Window) *TabQueue {
	tq := &TabQueue{state: state, window: w}

	// 1. Simple Queue elements
	tq.lblTitle = widget.NewLabelWithStyle(i18n.T("queue_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tq.lblFormTitle = widget.NewLabelWithStyle(i18n.T("queue_form_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	tq.lblQueueName = widget.NewLabel(i18n.T("queue_name_label"))
	tq.lblQueueTarget = widget.NewLabel(i18n.T("queue_target_label"))
	tq.lblQueueUpload = widget.NewLabel(i18n.T("queue_upload_label"))
	tq.lblQueueDownload = widget.NewLabel(i18n.T("queue_download_label"))

	tq.entryName = widget.NewEntry()
	tq.entryName.SetPlaceHolder("vd: NV-Ketoan, VLAN-Khach...")

	tq.entryTarget = widget.NewEntry()
	tq.entryTarget.SetPlaceHolder("vd: 192.168.10.50/32 hoặc 192.168.10.0/24")

	tq.btnFromVLAN = widget.NewButton(i18n.T("queue_from_vlan_btn"), func() {
		tq.pickFromVLANs()
	})

	tq.selectUpload = widget.NewSelectEntry(ratePresets)
	tq.selectUpload.SetText("10M")
	tq.selectUpload.SetPlaceHolder("vd: 5M, 10M...")

	tq.selectDownload = widget.NewSelectEntry(ratePresets)
	tq.selectDownload.SetText("20M")
	tq.selectDownload.SetPlaceHolder("vd: 10M, 20M...")

	tq.btnAdd = widget.NewButton(i18n.T("queue_add_btn"), func() {
		tq.addQueue()
	})
	tq.btnAdd.Importance = widget.HighImportance

	tq.lblListTitle = widget.NewLabelWithStyle(i18n.T("queue_list_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tq.btnRefresh = widget.NewButton(i18n.T("queue_refresh_btn"), func() {
		tq.Refresh()
	})

	tq.headerLabels = []*widget.Label{
		widget.NewLabelWithStyle(i18n.T("queue_col_name"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("queue_col_target"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("queue_col_upload"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("queue_col_download"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("queue_col_action"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}

	tq.queueData = make([]queue.QueueItem, 0)
	tq.queueList = widget.NewList(
		func() int {
			return len(tq.queueData)
		},
		func() fyne.CanvasObject {
			btnDel := widget.NewButton(i18n.T("queue_delete_btn"), nil)
			btnDel.Importance = widget.DangerImportance
			return container.New(&queueRowLayout{},
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				btnDel,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(tq.queueData) {
				return
			}
			item := tq.queueData[id]
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(item.Name)
			box.Objects[1].(*widget.Label).SetText(item.Target)
			box.Objects[2].(*widget.Label).SetText(item.UploadLimit)
			box.Objects[3].(*widget.Label).SetText(item.DownloadLimit)

			btnDel := box.Objects[4].(*widget.Button)
			btnDel.SetText(i18n.T("queue_delete_btn"))
			btnDel.OnTapped = func() {
				tq.confirmDelete(item)
			}
		},
	)

	// 2. Smart CAKE QoS elements
	tq.lblCakeTitle = widget.NewLabelWithStyle(i18n.T("cake_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tq.lblCakeDesc = widget.NewLabel(i18n.T("cake_desc"))
	tq.lblCakeDesc.Wrapping = fyne.TextWrapWord

	tq.lblCakeStatus = widget.NewLabelWithStyle(i18n.T("cake_status_inactive"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tq.btnRefreshCake = widget.NewButton("🔄 "+i18n.T("nat_refresh_btn"), func() {
		tq.RefreshCakeStatus()
	})

	// Single WAN inputs
	tq.lblCakeDown = widget.NewLabel(i18n.T("cake_down_label"))
	tq.entryCakeDown = widget.NewEntry()
	tq.entryCakeDown.SetText("150")
	tq.entryCakeDown.SetPlaceHolder("vd: 150")
	tq.entryCakeDown.OnChanged = func(_ string) {
		tq.updateCakeCalc()
	}

	tq.lblCakeUp = widget.NewLabel(i18n.T("cake_up_label"))
	tq.entryCakeUp = widget.NewEntry()
	tq.entryCakeUp.SetText("150")
	tq.entryCakeUp.SetPlaceHolder("vd: 150")
	tq.entryCakeUp.OnChanged = func(_ string) {
		tq.updateCakeCalc()
	}

	singleRateInputs := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, tq.lblCakeDown, nil, tq.entryCakeDown),
		container.NewBorder(nil, nil, tq.lblCakeUp, nil, tq.entryCakeUp),
	)
	tq.singleCakeBox = container.NewVBox(singleRateInputs)

	// Dual WAN inputs
	tq.lblCakeWAN1Down = widget.NewLabel(i18n.T("cake_wan1_down"))
	tq.entryCakeWAN1Down = widget.NewEntry()
	tq.entryCakeWAN1Down.SetText("150")
	tq.entryCakeWAN1Down.OnChanged = func(_ string) { tq.updateCakeCalc() }

	tq.lblCakeWAN1Up = widget.NewLabel(i18n.T("cake_wan1_up"))
	tq.entryCakeWAN1Up = widget.NewEntry()
	tq.entryCakeWAN1Up.SetText("150")
	tq.entryCakeWAN1Up.OnChanged = func(_ string) { tq.updateCakeCalc() }

	tq.lblCakeWAN1Iface = widget.NewLabel(i18n.T("cake_wan1_iface"))
	tq.entryCakeWAN1Iface = widget.NewEntry()
	tq.entryCakeWAN1Iface.SetText("pppoe-wan1")
	tq.entryCakeWAN1Iface.OnChanged = func(_ string) { tq.updateCakeCalc() }

	tq.lblCakeWAN2Down = widget.NewLabel(i18n.T("cake_wan2_down"))
	tq.entryCakeWAN2Down = widget.NewEntry()
	tq.entryCakeWAN2Down.SetText("150")
	tq.entryCakeWAN2Down.OnChanged = func(_ string) { tq.updateCakeCalc() }

	tq.lblCakeWAN2Up = widget.NewLabel(i18n.T("cake_wan2_up"))
	tq.entryCakeWAN2Up = widget.NewEntry()
	tq.entryCakeWAN2Up.SetText("150")
	tq.entryCakeWAN2Up.OnChanged = func(_ string) { tq.updateCakeCalc() }

	tq.lblCakeWAN2Iface = widget.NewLabel(i18n.T("cake_wan2_iface"))
	tq.entryCakeWAN2Iface = widget.NewEntry()
	tq.entryCakeWAN2Iface.SetText("pppoe-wan2")
	tq.entryCakeWAN2Iface.OnChanged = func(_ string) { tq.updateCakeCalc() }

	tq.lblCakeDualNote = widget.NewLabel(i18n.T("cake_dual_note"))
	tq.lblCakeDualNote.Wrapping = fyne.TextWrapWord

	wan1Box := container.NewVBox(
		widget.NewLabelWithStyle("WAN 1", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, tq.lblCakeWAN1Down, nil, tq.entryCakeWAN1Down),
		container.NewBorder(nil, nil, tq.lblCakeWAN1Up, nil, tq.entryCakeWAN1Up),
		container.NewBorder(nil, nil, tq.lblCakeWAN1Iface, nil, tq.entryCakeWAN1Iface),
	)
	wan2Box := container.NewVBox(
		widget.NewLabelWithStyle("WAN 2", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, tq.lblCakeWAN2Down, nil, tq.entryCakeWAN2Down),
		container.NewBorder(nil, nil, tq.lblCakeWAN2Up, nil, tq.entryCakeWAN2Up),
		container.NewBorder(nil, nil, tq.lblCakeWAN2Iface, nil, tq.entryCakeWAN2Iface),
	)

	tq.dualCakeBox = container.NewVBox(
		container.NewGridWithColumns(2, wan1Box, wan2Box),
		tq.lblCakeDualNote,
	)
	tq.dualCakeBox.Hide()

	// Mode radio
	tq.radioCakeMode = widget.NewRadioGroup([]string{
		i18n.T("cake_mode_single"),
		i18n.T("cake_mode_dual"),
	}, func(sel string) {
		if strings.Contains(sel, "2") || strings.Contains(sel, "Dual") {
			tq.singleCakeBox.Hide()
			tq.dualCakeBox.Show()
		} else {
			tq.singleCakeBox.Show()
			tq.dualCakeBox.Hide()
		}
		tq.updateCakeCalc()
	})
	tq.radioCakeMode.Horizontal = true
	tq.radioCakeMode.SetSelected(i18n.T("cake_mode_single"))

	tq.lblCakeCalc = widget.NewLabel("")
	tq.lblCakeCalc.Wrapping = fyne.TextWrapWord
	tq.updateCakeCalc()

	tq.btnApplyCake = widget.NewButton(i18n.T("cake_btn_apply"), func() {
		tq.applyCakeQoS()
	})
	tq.btnApplyCake.Importance = widget.HighImportance

	tq.btnRemoveCake = widget.NewButton(i18n.T("cake_btn_remove"), func() {
		tq.confirmRemoveCake()
	})
	tq.btnRemoveCake.Importance = widget.DangerImportance

	// Hook to refresh on connect
	tq.state.OnConnect = append(tq.state.OnConnect, func() {
		tq.Refresh()
		tq.RefreshCakeStatus()
	})

	return tq
}

func (tq *TabQueue) buildSimpleQueueView() fyne.CanvasObject {
	nameWrap := container.NewGridWrap(fyne.NewSize(180, 36), tq.entryName)
	nameSection := container.NewHBox(
		tq.lblQueueName,
		nameWrap,
		widget.NewLabel("   "),
		tq.lblQueueTarget,
		widget.NewLabel(" "),
	)
	btnWrap := container.NewHBox(widget.NewLabel(" "), tq.btnFromVLAN)
	row1 := container.NewBorder(nil, nil, nameSection, btnWrap, tq.entryTarget)

	uploadWrap := container.NewGridWrap(fyne.NewSize(110, 36), tq.selectUpload)
	downloadWrap := container.NewGridWrap(fyne.NewSize(110, 36), tq.selectDownload)
	btnAddWrap := container.NewGridWrap(fyne.NewSize(190, 36), tq.btnAdd)

	row2 := container.NewHBox(
		tq.lblQueueUpload,
		uploadWrap,
		widget.NewLabel("   "),
		tq.lblQueueDownload,
		downloadWrap,
		widget.NewLabel("   "),
		btnAddWrap,
	)

	formBox := container.NewVBox(
		tq.lblFormTitle,
		row1,
		row2,
	)

	listHeader := container.NewBorder(nil, nil, tq.lblListTitle, tq.btnRefresh)
	columnHeaders := container.New(&queueRowLayout{},
		tq.headerLabels[0],
		tq.headerLabels[1],
		tq.headerLabels[2],
		tq.headerLabels[3],
		tq.headerLabels[4],
	)

	tableCard := container.NewBorder(
		container.NewVBox(listHeader, widget.NewSeparator(), columnHeaders),
		nil, nil, nil,
		tq.queueList,
	)

	return container.NewBorder(
		container.NewVBox(
			tq.lblTitle,
			formBox,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		tableCard,
	)
}

func (tq *TabQueue) buildCakeQoSView() fyne.CanvasObject {
	statusBar := container.NewBorder(nil, nil, tq.lblCakeStatus, tq.btnRefreshCake)

	modeRow := container.NewHBox(
		widget.NewLabelWithStyle(i18n.T("cake_mode_label"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tq.radioCakeMode,
	)

	btnRow := container.NewGridWithColumns(2, tq.btnApplyCake, tq.btnRemoveCake)

	featureInfo := widget.NewLabel(
		"🌟 Ưu điểm của Smart CAKE QoS:\n" +
			" • Thuật toán CAKE (Common Applications Kept Enhanced) thế hệ mới nhất trong RouterOS 7.\n" +
			" • diffserv4: Tự động phân loại 4 mức ưu tiên (Bulk, Best Effort, Video, Voice/Gaming/DNS).\n" +
			" • triple-isolate: Cách ly công bằng giữa từng Host nội bộ, ngăn chặn 1 máy download làm lag cả mạng.\n" +
			" • Tự động tính toán trần băng thông 95% ngăn ngừa ISP bufferbloat, giảm độ trễ Ping về mức thấp nhất.\n" +
			" • Hỗ trợ hoàn hảo cả 1 WAN và 2 WAN (Dual ISP cân bằng tải, thiết lập queue riêng cho từng đường).",
	)
	featureInfo.Wrapping = fyne.TextWrapWord

	cakeContent := container.NewVBox(
		tq.lblCakeTitle,
		tq.lblCakeDesc,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Trạng thái hiện tại:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		statusBar,
		widget.NewSeparator(),
		modeRow,
		tq.singleCakeBox,
		tq.dualCakeBox,
		tq.lblCakeCalc,
		container.NewPadded(btnRow),
		widget.NewSeparator(),
		featureInfo,
	)

	return container.NewVScroll(container.NewPadded(cakeContent))
}

func (tq *TabQueue) BuildDesktopContainer() fyne.CanvasObject {
	viewSimple := tq.buildSimpleQueueView()
	viewCake := tq.buildCakeQoSView()

	tq.subTabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("queue_sub_simple"), viewSimple),
		container.NewTabItem(i18n.T("queue_sub_cake"), viewCake),
	)

	return tq.subTabs
}

func (tq *TabQueue) Container() fyne.CanvasObject {
	return tq.BuildDesktopContainer()
}

func (tq *TabQueue) UpdateLanguage() {
	if tq.subTabs != nil && len(tq.subTabs.Items) >= 2 {
		tq.subTabs.Items[0].Text = i18n.T("queue_sub_simple")
		tq.subTabs.Items[1].Text = i18n.T("queue_sub_cake")
		tq.subTabs.Refresh()
	}

	if tq.lblTitle != nil {
		tq.lblTitle.SetText(i18n.T("queue_title"))
	}
	if tq.lblFormTitle != nil {
		tq.lblFormTitle.SetText(i18n.T("queue_form_title"))
	}
	if tq.lblQueueName != nil {
		tq.lblQueueName.SetText(i18n.T("queue_name_label"))
	}
	if tq.lblQueueTarget != nil {
		tq.lblQueueTarget.SetText(i18n.T("queue_target_label"))
	}
	if tq.lblQueueUpload != nil {
		tq.lblQueueUpload.SetText(i18n.T("queue_upload_label"))
	}
	if tq.lblQueueDownload != nil {
		tq.lblQueueDownload.SetText(i18n.T("queue_download_label"))
	}
	if tq.btnFromVLAN != nil {
		tq.btnFromVLAN.SetText(i18n.T("queue_from_vlan_btn"))
	}
	if tq.btnAdd != nil {
		tq.btnAdd.SetText(i18n.T("queue_add_btn"))
	}
	if tq.lblListTitle != nil {
		tq.lblListTitle.SetText(i18n.T("queue_list_title"))
	}
	if tq.btnRefresh != nil {
		tq.btnRefresh.SetText(i18n.T("queue_refresh_btn"))
	}
	if len(tq.headerLabels) >= 5 {
		tq.headerLabels[0].SetText(i18n.T("queue_col_name"))
		tq.headerLabels[1].SetText(i18n.T("queue_col_target"))
		tq.headerLabels[2].SetText(i18n.T("queue_col_upload"))
		tq.headerLabels[3].SetText(i18n.T("queue_col_download"))
		tq.headerLabels[4].SetText(i18n.T("queue_col_action"))
	}
	if tq.queueList != nil {
		tq.queueList.Refresh()
	}

	if tq.lblCakeTitle != nil {
		tq.lblCakeTitle.SetText(i18n.T("cake_title"))
	}
	if tq.lblCakeDesc != nil {
		tq.lblCakeDesc.SetText(i18n.T("cake_desc"))
	}
	if tq.btnRefreshCake != nil {
		tq.btnRefreshCake.SetText("🔄 " + i18n.T("nat_refresh_btn"))
	}
	if tq.radioCakeMode != nil {
		isDual := strings.Contains(tq.radioCakeMode.Selected, "2") || strings.Contains(tq.radioCakeMode.Selected, "Dual")
		tq.radioCakeMode.Options = []string{i18n.T("cake_mode_single"), i18n.T("cake_mode_dual")}
		if isDual {
			tq.radioCakeMode.SetSelected(i18n.T("cake_mode_dual"))
		} else {
			tq.radioCakeMode.SetSelected(i18n.T("cake_mode_single"))
		}
		tq.radioCakeMode.Refresh()
	}
	if tq.lblCakeDown != nil {
		tq.lblCakeDown.SetText(i18n.T("cake_down_label"))
	}
	if tq.lblCakeUp != nil {
		tq.lblCakeUp.SetText(i18n.T("cake_up_label"))
	}
	if tq.lblCakeWAN1Down != nil {
		tq.lblCakeWAN1Down.SetText(i18n.T("cake_wan1_down"))
	}
	if tq.lblCakeWAN1Up != nil {
		tq.lblCakeWAN1Up.SetText(i18n.T("cake_wan1_up"))
	}
	if tq.lblCakeWAN1Iface != nil {
		tq.lblCakeWAN1Iface.SetText(i18n.T("cake_wan1_iface"))
	}
	if tq.lblCakeWAN2Down != nil {
		tq.lblCakeWAN2Down.SetText(i18n.T("cake_wan2_down"))
	}
	if tq.lblCakeWAN2Up != nil {
		tq.lblCakeWAN2Up.SetText(i18n.T("cake_wan2_up"))
	}
	if tq.lblCakeWAN2Iface != nil {
		tq.lblCakeWAN2Iface.SetText(i18n.T("cake_wan2_iface"))
	}
	if tq.lblCakeDualNote != nil {
		tq.lblCakeDualNote.SetText(i18n.T("cake_dual_note"))
	}
	if tq.btnApplyCake != nil {
		tq.btnApplyCake.SetText(i18n.T("cake_btn_apply"))
	}
	if tq.btnRemoveCake != nil {
		tq.btnRemoveCake.SetText(i18n.T("cake_btn_remove"))
	}
	tq.updateCakeCalc()
	tq.RefreshCakeStatus()
}

func (tq *TabQueue) updateCakeCalc() {
	if tq.lblCakeCalc == nil {
		return
	}
	isDual := tq.radioCakeMode != nil && (strings.Contains(tq.radioCakeMode.Selected, "2") || strings.Contains(tq.radioCakeMode.Selected, "Dual"))
	if isDual {
		if tq.entryCakeWAN1Down == nil || tq.entryCakeWAN1Up == nil || tq.entryCakeWAN2Down == nil || tq.entryCakeWAN2Up == nil {
			return
		}
		w1D, err1 := strconv.ParseFloat(strings.TrimSpace(tq.entryCakeWAN1Down.Text), 64)
		w1U, err2 := strconv.ParseFloat(strings.TrimSpace(tq.entryCakeWAN1Up.Text), 64)
		w2D, err3 := strconv.ParseFloat(strings.TrimSpace(tq.entryCakeWAN2Down.Text), 64)
		w2U, err4 := strconv.ParseFloat(strings.TrimSpace(tq.entryCakeWAN2Up.Text), 64)
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil && w1D > 0 && w1U > 0 && w2D > 0 && w2U > 0 {
			w1D95 := w1D * 0.95
			w1U95 := w1U * 0.95
			w2D95 := w2D * 0.95
			w2U95 := w2U * 0.95
			tq.lblCakeCalc.SetText(fmt.Sprintf(
				"📊 Giới hạn 95%% chống Bufferbloat Dual WAN:\n"+
					" • WAN 1 (%s): Down ~ %.1f Mbps (%d Kbps) | Up ~ %.1f Mbps (%d Kbps)\n"+
					" • WAN 2 (%s): Down ~ %.1f Mbps (%d Kbps) | Up ~ %.1f Mbps (%d Kbps)\n"+
					"💡 Mỗi đường ISP sẽ được RouterOS gắn 1 Queue CAKE độc lập với headroom 5%% riêng biệt, triệt tiêu 100%% hiện tượng nghẽn modem.",
				strings.TrimSpace(tq.entryCakeWAN1Iface.Text), w1D95, int(w1D95*1000), w1U95, int(w1U95*1000),
				strings.TrimSpace(tq.entryCakeWAN2Iface.Text), w2D95, int(w2D95*1000), w2U95, int(w2U95*1000),
			))
		} else {
			tq.lblCakeCalc.SetText("Nhập tốc độ gói cước Down/Up cho cả 2 đường WAN để tự động tính giới hạn 95% chống Bufferbloat.")
		}
		return
	}

	if tq.entryCakeDown == nil || tq.entryCakeUp == nil {
		return
	}
	downVal, err1 := strconv.ParseFloat(strings.TrimSpace(tq.entryCakeDown.Text), 64)
	upVal, err2 := strconv.ParseFloat(strings.TrimSpace(tq.entryCakeUp.Text), 64)
	if err1 == nil && err2 == nil && downVal > 0 && upVal > 0 {
		d95 := downVal * 0.95
		u95 := upVal * 0.95
		tq.lblCakeCalc.SetText(fmt.Sprintf("📊 Giới hạn 95%% chống Bufferbloat: Down ~ %.1f Mbps (%d Kbps) | Up ~ %.1f Mbps (%d Kbps)\n💡 Tự động giữ 5%% headroom để đường truyền ISP không bị nghẽn buffer. Đảm bảo Ping khi chơi game luôn ổn định ngay cả khi đang tải nặng.", d95, int(d95*1000), u95, int(u95*1000)))
	} else {
		tq.lblCakeCalc.SetText("Nhập tốc độ gói cước Down/Up để tự động tính toán giới hạn 95% chống Bufferbloat.")
	}
}

func (tq *TabQueue) applyCakeQoS() {
	c := tq.state.GetClient()
	if !c.IsConnected() {
		tq.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	isDual := tq.radioCakeMode != nil && (strings.Contains(tq.radioCakeMode.Selected, "2") || strings.Contains(tq.radioCakeMode.Selected, "Dual"))
	if isDual {
		w1DStr := strings.TrimSpace(tq.entryCakeWAN1Down.Text)
		w1UStr := strings.TrimSpace(tq.entryCakeWAN1Up.Text)
		w2DStr := strings.TrimSpace(tq.entryCakeWAN2Down.Text)
		w2UStr := strings.TrimSpace(tq.entryCakeWAN2Up.Text)
		w1Iface := strings.TrimSpace(tq.entryCakeWAN1Iface.Text)
		w2Iface := strings.TrimSpace(tq.entryCakeWAN2Iface.Text)

		w1D, err1 := strconv.ParseFloat(w1DStr, 64)
		w1U, err2 := strconv.ParseFloat(w1UStr, 64)
		w2D, err3 := strconv.ParseFloat(w2DStr, 64)
		w2U, err4 := strconv.ParseFloat(w2UStr, 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || w1D <= 0 || w1U <= 0 || w2D <= 0 || w2U <= 0 {
			dialog.ShowError(fmt.Errorf("Vui lòng nhập đầy đủ tốc độ gói cước (> 0 Mbps) cho cả WAN 1 và WAN 2"), tq.window)
			return
		}
		if w1Iface == "" {
			w1Iface = "pppoe-wan1"
		}
		if w2Iface == "" {
			w2Iface = "pppoe-wan2"
		}

		tq.btnApplyCake.Disable()
		go func() {
			defer fyne.Do(func() { tq.btnApplyCake.Enable() })
			cfg := queue.NewConfigurator(c, tq.state.Log)
			err := cfg.ConfigureCakeQoSDual(w1D, w1U, w2D, w2U, w1Iface, w2Iface)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Lỗi cấu hình Dual CAKE QoS: %v", err), tq.window)
				})
				return
			}
			fyne.Do(func() {
				tq.RefreshCakeStatus()
				tq.Refresh()
				dialog.ShowInformation("Thành công", "Đã kích hoạt Smart CAKE QoS cho Dual WAN thành công! Cả 2 đường ISP đã được tối ưu chống giật lag Bufferbloat.", tq.window)
			})
		}()
		return
	}

	downStr := strings.TrimSpace(tq.entryCakeDown.Text)
	upStr := strings.TrimSpace(tq.entryCakeUp.Text)

	downVal, err1 := strconv.ParseFloat(downStr, 64)
	upVal, err2 := strconv.ParseFloat(upStr, 64)

	if err1 != nil || err2 != nil || downVal <= 0 || upVal <= 0 {
		dialog.ShowError(fmt.Errorf("Vui lòng nhập tốc độ gói cước hợp lệ (> 0 Mbps)"), tq.window)
		return
	}

	tq.btnApplyCake.Disable()
	go func() {
		defer fyne.Do(func() { tq.btnApplyCake.Enable() })

		cfg := queue.NewConfigurator(c, tq.state.Log)
		err := cfg.ConfigureCakeQoS(downVal, upVal)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Lỗi cấu hình CAKE QoS: %v", err), tq.window)
			})
			return
		}

		fyne.Do(func() {
			tq.RefreshCakeStatus()
			tq.Refresh()
			dialog.ShowInformation("Thành công", "Đã kích hoạt Smart CAKE QoS thành công! Router đã được tối ưu chống giật lag Bufferbloat.", tq.window)
		})
	}()
}

func (tq *TabQueue) confirmRemoveCake() {
	c := tq.state.GetClient()
	if !c.IsConnected() {
		tq.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	dialog.ShowConfirm("Xác Nhận Gỡ Bỏ CAKE QoS", "Bạn có chắc chắn muốn gỡ bỏ Smart CAKE QoS?", func(ok bool) {
		if !ok {
			return
		}
		tq.btnRemoveCake.Disable()
		go func() {
			defer fyne.Do(func() { tq.btnRemoveCake.Enable() })
			cfg := queue.NewConfigurator(c, tq.state.Log)
			if err := cfg.RemoveCakeQoS(); err != nil {
				tq.state.Log(fmt.Sprintf("❌ Lỗi gỡ bỏ CAKE QoS: %v", err))
			} else {
				fyne.Do(func() {
					tq.RefreshCakeStatus()
					tq.Refresh()
				})
			}
		}()
	}, tq.window)
}

func (tq *TabQueue) RefreshCakeStatus() {
	c := tq.state.GetClient()
	if !c.IsConnected() {
		if tq.lblCakeStatus != nil {
			tq.lblCakeStatus.SetText(i18n.T("cake_status_inactive"))
		}
		return
	}

	if tq.btnRefreshCake != nil {
		tq.btnRefreshCake.Disable()
	}
	go func() {
		defer fyne.Do(func() {
			if tq.btnRefreshCake != nil {
				tq.btnRefreshCake.Enable()
			}
		})
		cfg := queue.NewConfigurator(c, tq.state.Log)
		active, maxLimit, err := cfg.GetCakeStatus()
		fyne.Do(func() {
			if tq.lblCakeStatus == nil {
				return
			}
			if err == nil && active {
				tq.lblCakeStatus.SetText(fmt.Sprintf(i18n.T("cake_status_active"), maxLimit))
			} else {
				tq.lblCakeStatus.SetText(i18n.T("cake_status_inactive"))
			}
		})
	}()
}

func (tq *TabQueue) pickFromVLANs() {
	vlans := tq.state.GetVLANs()
	if len(vlans) == 0 {
		tq.state.Log("⚠ Chưa có danh sách VLAN trong ứng dụng. Vui lòng cấu hình VLAN tại Tab Cấu Hình Mạng trước.")
		return
	}

	options := make([]string, 0, len(vlans))
	vlanMap := make(map[string]struct{ name, cidr string })
	for _, v := range vlans {
		opt := fmt.Sprintf("VLAN %d - %s (%s)", v.VLANID, v.Name, v.CIDR)
		options = append(options, opt)
		vlanMap[opt] = struct{ name, cidr string }{
			name: fmt.Sprintf("vlan%d-%s", v.VLANID, v.Name),
			cidr: v.CIDR,
		}
	}

	selectPopup := widget.NewSelect(options, func(selected string) {
		if val, found := vlanMap[selected]; found {
			tq.entryName.SetText(val.name)
			tq.entryTarget.SetText(val.cidr)
			tq.state.Log(fmt.Sprintf("📋 Đã chọn nhanh Target từ VLAN: %s -> %s", val.name, val.cidr))
		}
	})
	if len(options) > 0 {
		selectPopup.SetSelected(options[0])
	}
	dialog.ShowCustom("Chọn VLAN Để Giới Hạn Băng Thông", "Đóng", selectPopup, tq.window)
}

func (tq *TabQueue) addQueue() {
	c := tq.state.GetClient()
	if !c.IsConnected() {
		tq.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	name := strings.TrimSpace(tq.entryName.Text)
	target := strings.TrimSpace(tq.entryTarget.Text)
	up := strings.TrimSpace(tq.selectUpload.Text)
	down := strings.TrimSpace(tq.selectDownload.Text)

	if name == "" {
		tq.state.Log("✗ Vui lòng nhập Tên Queue!")
		return
	}
	if target == "" {
		tq.state.Log("✗ Vui lòng nhập địa chỉ Target (IP hoặc Subnet)!")
		return
	}

	tq.btnAdd.Disable()
	go func() {
		defer fyne.Do(func() { tq.btnAdd.Enable() })

		cfg := queue.NewConfigurator(c, tq.state.Log)
		err := cfg.AddQueue(queue.QueueItem{
			Name:          name,
			Target:        target,
			UploadLimit:   up,
			DownloadLimit: down,
		})
		if err != nil {
			tq.state.Log(fmt.Sprintf("✗ Lỗi: %v", err))
			return
		}

		fyne.Do(func() {
			tq.entryName.SetText("")
			tq.entryTarget.SetText("")
		})
		tq.Refresh()
	}()
}

func (tq *TabQueue) Refresh() {
	c := tq.state.GetClient()
	if !c.IsConnected() {
		return
	}

	tq.btnRefresh.Disable()
	go func() {
		defer fyne.Do(func() { tq.btnRefresh.Enable() })

		cfg := queue.NewConfigurator(c, tq.state.Log)
		items, err := cfg.ListQueues()
		fyne.Do(func() {
			if err == nil {
				tq.queueData = items
				tq.queueList.Refresh()
				tq.state.Log(fmt.Sprintf("✓ Đã tải %d Simple Queues từ Router.", len(items)))
			}
		})
	}()
}

func (tq *TabQueue) confirmDelete(item queue.QueueItem) {
	c := tq.state.GetClient()
	if !c.IsConnected() {
		tq.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	msg := fmt.Sprintf(i18n.T("queue_delete_confirm_msg"), item.Name)
	dialog.ShowConfirm(i18n.T("queue_delete_confirm_title"), msg, func(ok bool) {
		if !ok {
			return
		}
		go func() {
			cfg := queue.NewConfigurator(c, tq.state.Log)
			targetID := item.ID
			if targetID == "" {
				targetID = item.Name
			}
			err := cfg.DeleteQueue(targetID)
			if err != nil {
				tq.state.Log(fmt.Sprintf("✗ Lỗi xoá queue: %v", err))
			} else {
				tq.Refresh()
			}
		}()
	}, tq.window)
}
