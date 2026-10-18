package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/modules/dashboard"
)

type TabDashboard struct {
	state            *AppState
	window           fyne.Window
	container        *fyne.Container
	lblTitle         *widget.Label
	lblIdentityTitle *widget.Label
	lblIdentity      *widget.Label
	lblBoardTitle    *widget.Label
	lblBoard         *widget.Label
	lblVersionTitle  *widget.Label
	lblVersion       *widget.Label
	lblUptimeTitle   *widget.Label
	lblUptime        *widget.Label
	lblIfaceTitle    *widget.Label
	progCPU          *widget.ProgressBar
	lblCPU           *widget.Label
	progRAM          *widget.ProgressBar
	lblRAM           *widget.Label
	progHDD          *widget.ProgressBar
	lblHDD           *widget.Label
	ifaceList        *widget.List
	ifaceHeaders     []*widget.Label
	ifaceData        []dashboard.InterfaceItem
	btnRefresh       *widget.Button
	btnReboot        *widget.Button
	cardHardware     *fyne.Container
}

func NewTabDashboard(state *AppState, w fyne.Window) *TabDashboard {
	td := &TabDashboard{state: state, window: w}

	td.lblTitle = widget.NewLabelWithStyle(i18n.T("dash_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblIdentityTitle = widget.NewLabelWithStyle(i18n.T("dash_model"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblIdentity = widget.NewLabelWithStyle("--", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	td.lblBoardTitle = widget.NewLabelWithStyle(i18n.T("dash_card_dev"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblBoard = widget.NewLabelWithStyle("--", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	td.lblVersionTitle = widget.NewLabelWithStyle(i18n.T("dash_version"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblVersion = widget.NewLabel("--")

	td.lblUptimeTitle = widget.NewLabelWithStyle(i18n.T("dash_uptime"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	td.lblUptime = widget.NewLabelWithStyle("--", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	td.lblIfaceTitle = widget.NewLabelWithStyle(i18n.T("dash_iface_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	td.cardHardware = container.NewGridWithColumns(2,
		container.NewVBox(td.lblIdentityTitle, td.lblIdentity),
		container.NewVBox(td.lblBoardTitle, td.lblBoard),
		container.NewVBox(td.lblVersionTitle, td.lblVersion),
		container.NewVBox(td.lblUptimeTitle, td.lblUptime),
	)

	td.progCPU = widget.NewProgressBar()
	td.lblCPU = widget.NewLabel("CPU: 0%")

	td.progRAM = widget.NewProgressBar()
	td.lblRAM = widget.NewLabel("RAM: 0 MB / 0 MB (0%)")

	td.progHDD = widget.NewProgressBar()
	td.lblHDD = widget.NewLabel(fmt.Sprintf("%s 0 MB / 0 MB (0%%)", i18n.T("dash_storage")))

	td.ifaceHeaders = []*widget.Label{
		widget.NewLabelWithStyle(i18n.T("dash_col_name"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("dash_col_type"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("dash_col_status"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("dash_col_mac"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(i18n.T("dash_col_comment"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}

	// Interfaces list (Desktop 5-column table)
	td.ifaceData = make([]dashboard.InterfaceItem, 0)
	td.ifaceList = widget.NewList(
		func() int {
			return len(td.ifaceData)
		},
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(5,
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
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

			stText := i18n.T("dash_link_down")
			if item.Disabled {
				stText = "⚪ Disabled"
			} else if item.Running {
				stText = i18n.T("dash_link_up")
			}
			box.Objects[2].(*widget.Label).SetText(stText)
			box.Objects[3].(*widget.Label).SetText(item.MAC)
			box.Objects[4].(*widget.Label).SetText(item.Comment)
		},
	)

	td.btnRefresh = widget.NewButton(i18n.T("dash_refresh"), func() {
		td.Refresh()
	})
	td.btnRefresh.Importance = widget.HighImportance

	td.btnReboot = widget.NewButton(i18n.T("dash_reboot"), func() {
		td.confirmReboot()
	})
	td.btnReboot.Importance = widget.DangerImportance

	// Register on-connect hook
	td.state.OnConnect = append(td.state.OnConnect, func() {
		td.Refresh()
	})

	return td
}

func (td *TabDashboard) BuildDesktopContainer() fyne.CanvasObject {
	cardResource := container.NewGridWithColumns(3,
		container.NewVBox(td.lblCPU, td.progCPU),
		container.NewVBox(td.lblRAM, td.progRAM),
		container.NewVBox(td.lblHDD, td.progHDD),
	)
	btnBar := container.NewBorder(nil, nil, td.btnRefresh, td.btnReboot)

	var headerObjects []fyne.CanvasObject
	for _, h := range td.ifaceHeaders {
		headerObjects = append(headerObjects, h)
	}
	headerRow := container.NewGridWithColumns(5, headerObjects...)
	tableContainer := container.NewBorder(headerRow, nil, nil, nil, td.ifaceList)

	return container.NewBorder(
		container.NewVBox(
			td.lblTitle,
			td.cardHardware,
			widget.NewSeparator(),
			cardResource,
			widget.NewSeparator(),
			td.lblIfaceTitle,
		),
		btnBar,
		nil,
		nil,
		tableContainer,
	)
}

func (td *TabDashboard) Container() fyne.CanvasObject {
	return td.BuildDesktopContainer()
}

func (td *TabDashboard) UpdateLanguage() {
	if td.lblTitle != nil {
		td.lblTitle.SetText(i18n.T("dash_title"))
	}
	if td.lblIdentityTitle != nil {
		td.lblIdentityTitle.SetText(i18n.T("dash_model"))
	}
	if td.lblBoardTitle != nil {
		td.lblBoardTitle.SetText(i18n.T("dash_card_dev"))
	}
	if td.lblVersionTitle != nil {
		td.lblVersionTitle.SetText(i18n.T("dash_version"))
	}
	if td.lblUptimeTitle != nil {
		td.lblUptimeTitle.SetText(i18n.T("dash_uptime"))
	}
	if td.lblIfaceTitle != nil {
		td.lblIfaceTitle.SetText(i18n.T("dash_iface_title"))
	}
	if td.btnRefresh != nil {
		td.btnRefresh.SetText(i18n.T("dash_refresh"))
	}
	if td.btnReboot != nil {
		td.btnReboot.SetText(i18n.T("dash_reboot"))
	}
	if len(td.ifaceHeaders) >= 5 {
		td.ifaceHeaders[0].SetText(i18n.T("dash_col_name"))
		td.ifaceHeaders[1].SetText(i18n.T("dash_col_type"))
		td.ifaceHeaders[2].SetText(i18n.T("dash_col_status"))
		td.ifaceHeaders[3].SetText(i18n.T("dash_col_mac"))
		td.ifaceHeaders[4].SetText(i18n.T("dash_col_comment"))
	}
	if td.ifaceList != nil {
		td.ifaceList.Refresh()
	}
}

func (td *TabDashboard) Refresh() {
	c := td.state.GetClient()
	if !c.IsConnected() {
		td.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	td.btnRefresh.Disable()
	td.btnRefresh.SetText(i18n.T("dash_loading"))

	go func() {
		svc := dashboard.NewService(c, td.state.Log)
		info, err := svc.GetSystemInfo()
		ifaces, err2 := svc.GetInterfaces()

		fyne.Do(func() {
			td.btnRefresh.Enable()
			td.btnRefresh.SetText(i18n.T("dash_refresh"))
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
			td.lblHDD.SetText(fmt.Sprintf("%s %.1fMB / %.1fMB (%.1f%%)", i18n.T("dash_storage"), info.UsedHDDMB, info.TotalHDDMB, info.HDDPct))

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

	dialog.ShowConfirm(i18n.T("dash_reboot_confirm_title"), i18n.T("dash_reboot_confirm_msg"), func(ok bool) {
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
