package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type MainWindow struct {
	App          fyne.App
	Window       fyne.Window
	State        *AppState
	ConnBar      *ConnectionBar
	LogPanel     *LogPanel
	Split        *container.Split
	TabDashboard *TabDashboard
	TabNetwork   *TabNetwork
	TabSecurity  *TabSecurity
	TabIPSec     *TabIPSec
	TabBackup    *TabBackup
	TabClear     *TabClear
}

func NewMainWindow(a fyne.App, w fyne.Window) *MainWindow {
	if w == nil {
		w = a.NewWindow("Mikrotik RouterOS 7 — Automation Tool")
	}

	logPanel := NewLogPanel(w)
	state := NewAppState(logPanel)
	connBar := NewConnectionBar(state, a)

	tabDash := NewTabDashboard(state, w)
	tabNet := NewTabNetwork(state, w)
	tabSec := NewTabSecurity(state)
	tabIPSec := NewTabIPSec(state)
	tabBk := NewTabBackup(state, w)
	tabClr := NewTabClear(state, w)

	mw := &MainWindow{
		App:          a,
		Window:       w,
		State:        state,
		ConnBar:      connBar,
		LogPanel:     logPanel,
		TabDashboard: tabDash,
		TabNetwork:   tabNet,
		TabSecurity:  tabSec,
		TabIPSec:     tabIPSec,
		TabBackup:    tabBk,
		TabClear:     tabClr,
	}

	logPanel.Log("Chào mừng đến với Mikrotik RouterOS 7 Automation Tool (Desktop Edition)!")
	logPanel.Log("Chọn phương thức kết nối (REST API hoặc SSH) rồi nhấn 'Kết nối'.")

	return mw
}

// ShowMainView renders the desktop interface with a wide connection bar, tabs, and resizable spacious log panel.
func (mw *MainWindow) ShowMainView() {
	mw.Window.Resize(fyne.NewSize(1020, 850))
	mw.Window.CenterOnScreen()

	connBarContent := mw.ConnBar.BuildDesktopContainer()

	appTabs := container.NewAppTabs(
		container.NewTabItem("📊 Dashboard", mw.TabDashboard.BuildDesktopContainer()),
		container.NewTabItem("🌐 Cấu Hình Mạng", mw.TabNetwork.BuildDesktopContainer()),
		container.NewTabItem("🔒 Bảo Mật", mw.TabSecurity.BuildDesktopContainer()),
		container.NewTabItem("🔐 IPSec VPN", mw.TabIPSec.BuildDesktopContainer()),
		container.NewTabItem("💾 Sao Lưu & Phục Hồi", mw.TabBackup.BuildDesktopContainer()),
		container.NewTabItem("🗑️ Xoá Cấu Hình", mw.TabClear.BuildDesktopContainer()),
	)
	appTabs.SetTabLocation(container.TabLocationTop)

	// Resizable vertical split: 65% Tabs, 35% Logs
	mw.Split = container.NewVSplit(appTabs, mw.LogPanel.Container())
	mw.Split.SetOffset(0.65)

	mw.LogPanel.SetSplitController(func(mode int) {
		switch mode {
		case 0: // Normal / Default
			mw.Split.SetOffset(0.65)
		case 1: // Maximize logs
			mw.Split.SetOffset(0.20)
		}
	})

	content := container.NewBorder(
		connBarContent,
		nil,
		nil,
		nil,
		mw.Split,
	)

	mw.Window.SetContent(content)
}

func (mw *MainWindow) ShowAndRun() {
	mw.ShowMainView()
	mw.Window.ShowAndRun()
}
