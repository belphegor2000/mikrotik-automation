package gui

import (
	"fmt"
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
	state          *AppState
	window         fyne.Window
	container      *fyne.Container
	lblTitle       *widget.Label
	lblFormTitle   *widget.Label
	entryName      *widget.Entry
	entryTarget    *widget.Entry
	btnFromVLAN    *widget.Button
	selectUpload   *widget.SelectEntry
	selectDownload *widget.SelectEntry
	btnAdd         *widget.Button
	lblListTitle   *widget.Label
	btnRefresh     *widget.Button
	headerLabels   []*widget.Label
	queueList      *widget.List
	queueData      []queue.QueueItem
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

	tq.lblTitle = widget.NewLabelWithStyle(i18n.T("queue_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tq.lblFormTitle = widget.NewLabelWithStyle(i18n.T("queue_form_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

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

	// Hook to refresh on connect
	tq.state.OnConnect = append(tq.state.OnConnect, func() {
		tq.Refresh()
	})

	return tq
}

func (tq *TabQueue) BuildDesktopContainer() fyne.CanvasObject {
	targetRow := container.NewBorder(nil, nil, nil, tq.btnFromVLAN, tq.entryTarget)
	rateRow := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("queue_upload_label")), nil, tq.selectUpload),
		container.NewBorder(nil, nil, widget.NewLabel(i18n.T("queue_download_label")), nil, tq.selectDownload),
	)

	formBox := container.NewVBox(
		tq.lblFormTitle,
		container.NewGridWithColumns(2,
			container.NewBorder(nil, nil, widget.NewLabel(i18n.T("queue_name_label")), nil, tq.entryName),
			container.NewBorder(nil, nil, widget.NewLabel(i18n.T("queue_target_label")), nil, targetRow),
		),
		rateRow,
		container.NewPadded(tq.btnAdd),
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

func (tq *TabQueue) Container() fyne.CanvasObject {
	return tq.BuildDesktopContainer()
}

func (tq *TabQueue) UpdateLanguage() {
	if tq.lblTitle != nil {
		tq.lblTitle.SetText(i18n.T("queue_title"))
	}
	if tq.lblFormTitle != nil {
		tq.lblFormTitle.SetText(i18n.T("queue_form_title"))
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
}

func (tq *TabQueue) pickFromVLANs() {
	if len(tq.state.VLANs) == 0 {
		tq.state.Log("⚠ Chưa có danh sách VLAN trong ứng dụng. Vui lòng cấu hình VLAN tại Tab Cấu Hình Mạng trước.")
		return
	}

	options := make([]string, 0, len(tq.state.VLANs))
	vlanMap := make(map[string]struct{ name, cidr string })
	for _, v := range tq.state.VLANs {
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
