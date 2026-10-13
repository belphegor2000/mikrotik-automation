package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/dashboard"
)

type TabDashboard struct {
	state       *AppState
	window      fyne.Window
	container   *fyne.Container
	lblIdentity *widget.Label
	lblBoard    *widget.Label
	lblVersion  *widget.Label
	lblUptime   *widget.Label
	progCPU     *widget.ProgressBar
	lblCPU      *widget.Label
	progRAM     *widget.ProgressBar
	lblRAM      *widget.Label
	progHDD     *widget.ProgressBar
	lblHDD      *widget.Label
	ifaceList   *widget.List
	ifaceData   []dashboard.InterfaceItem
	btnRefresh  *widget.Button
	btnReboot   *widget.Button
}

func NewTabDashboard(state *AppState, w fyne.Window) *TabDashboard {
	td := &TabDashboard{state: state, window: w}

	td.lblIdentity = widget.NewLabelWithStyle("--", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblBoard = widget.NewLabelWithStyle("--", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblVersion = widget.NewLabel("--")
	td.lblUptime = widget.NewLabelWithStyle("--", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	cardHardware := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Tên Router:"), td.lblIdentity),
		container.NewVBox(widget.NewLabel("Thiết Bị / Model:"), td.lblBoard),
		container.NewVBox(widget.NewLabel("Phiên Bản OS:"), td.lblVersion),
		container.NewVBox(widget.NewLabel("Thời Gian Hoạt Động:"), td.lblUptime),
	)

	td.progCPU = widget.NewProgressBar()
	td.lblCPU = widget.NewLabel("CPU: 0%")

	td.progRAM = widget.NewProgressBar()
	td.lblRAM = widget.NewLabel("RAM: 0 MB / 0 MB (0%)")

	td.progHDD = widget.NewProgressBar()
	td.lblHDD = widget.NewLabel("Ổ Cứng: 0 MB / 0 MB (0%)")

	cardResource := container.NewGridWithColumns(3,
		container.NewVBox(td.lblCPU, td.progCPU),
		container.NewVBox(td.lblRAM, td.progRAM),
		container.NewVBox(td.lblHDD, td.progHDD),
	)

	// Interfaces list
	td.ifaceData = make([]dashboard.InterfaceItem, 0)
	td.ifaceList = widget.NewList(
		func() int {
			return len(td.ifaceData)
		},
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(5,
				widget.NewLabel("Name"),
				widget.NewLabel("Type"),
				widget.NewLabel("Status"),
				widget.NewLabel("MAC"),
				widget.NewLabel("Comment"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(td.ifaceData) {
				return
			}
			item := td.ifaceData[id]
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(item.Name)
			box.Objects[1].(*widget.Label).SetText(item.Type)

			stText := "🔴 Link Down"
			if item.Disabled {
				stText = "⚪ Disabled"
			} else if item.Running {
				stText = "🟢 Link Up"
			}
			box.Objects[2].(*widget.Label).SetText(stText)
			box.Objects[3].(*widget.Label).SetText(item.MAC)
			box.Objects[4].(*widget.Label).SetText(item.Comment)
		},
	)

	td.btnRefresh = widget.NewButton("🔄 Làm Mới Thông Số", func() {
		td.Refresh()
	})

	td.btnReboot = widget.NewButton("⚠️ Khởi Động Lại Router (Reboot)", func() {
		td.confirmReboot()
	})

	btnBar := container.NewBorder(nil, nil, td.btnRefresh, td.btnReboot)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("📊 Giám Sát Hệ Thống & Tài Nguyên", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			cardHardware,
			cardResource,
			widget.NewLabelWithStyle("Danh Sách Cổng Mạng (Interfaces):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		btnBar,
		nil,
		nil,
		td.ifaceList,
	)

	td.container = content

	// Register on-connect hook
	td.state.OnConnect = append(td.state.OnConnect, func() {
		td.Refresh()
	})

	return td
}

func (td *TabDashboard) Container() fyne.CanvasObject {
	return td.container
}

func (td *TabDashboard) Refresh() {
	c := td.state.GetClient()
	if !c.IsConnected() {
		td.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	td.btnRefresh.Disable()
	td.btnRefresh.SetText("⏳ Đang tải...")

	go func() {
		svc := dashboard.NewService(c, td.state.Log)
		info, err := svc.GetSystemInfo()
		ifaces, err2 := svc.GetInterfaces()

		fyne.Do(func() {
			td.btnRefresh.Enable()
			td.btnRefresh.SetText("🔄 Làm Mới Thông Số")
			if err != nil {
				td.state.Log(fmt.Sprintf("✗ Lỗi lấy thông số hệ thống: %v", err))
				return
			}
			td.lblIdentity.SetText(info.Identity)
			td.lblBoard.SetText(info.BoardName)
			td.lblVersion.SetText(info.Version)
			td.lblUptime.SetText(info.Uptime)

			td.progCPU.SetValue(float64(info.CPULoad) / 100.0)
			td.lblCPU.SetText(fmt.Sprintf("CPU: %d%% (%d Cores)", info.CPULoad, info.CPUCount))

			td.progRAM.SetValue(info.MemPct / 100.0)
			td.lblRAM.SetText(fmt.Sprintf("RAM: %.1fMB / %.1fMB (%.1f%%)", info.UsedMemMB, info.TotalMemMB, info.MemPct))

			td.progHDD.SetValue(info.HDDPct / 100.0)
			td.lblHDD.SetText(fmt.Sprintf("Ổ Cứng: %.1fMB / %.1fMB (%.1f%%)", info.UsedHDDMB, info.TotalHDDMB, info.HDDPct))

			if err2 == nil {
				td.ifaceData = ifaces
				td.ifaceList.Refresh()
			}
			td.state.Log("✓ Đã cập nhật thông số Dashboard thành công.")
		})
	}()
}

func (td *TabDashboard) confirmReboot() {
	c := td.state.GetClient()
	if !c.IsConnected() {
		td.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	dialog.ShowConfirm("Khởi Động Lại", "Bạn có chắc chắn muốn khởi động lại Router không? Mạng sẽ tạm thời gián đoạn.", func(ok bool) {
		if !ok {
			return
		}
		td.btnReboot.Disable()
		go func() {
			svc := dashboard.NewService(c, td.state.Log)
			_ = svc.RebootRouter()
			fyne.Do(func() {
				td.btnReboot.Enable()
			})
		}()
	}, td.window)
}
