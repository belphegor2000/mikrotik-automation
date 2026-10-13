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
	TabDashboard *TabDashboard
	TabNetwork   *TabNetwork
	TabSecurity  *TabSecurity
	TabIPSec     *TabIPSec
	TabBackup    *TabBackup
	TabClear     *TabClear
	IsMobile     bool
}

func NewMainWindow(a fyne.App, w fyne.Window) *MainWindow {
	if w == nil {
		w = a.NewWindow("Mikrotik RouterOS 7 — Automation Tool")
	}

	logPanel := NewLogPanel()
	state := NewAppState(logPanel)
	connBar := NewConnectionBar(state, a)

	tabDash := NewTabDashboard(state, w)
	tabNet := NewTabNetwork(state, w)
	tabSec := NewTabSecurity(state)
	tabIPSec := NewTabIPSec(state)
	tabBk := NewTabBackup(state, w)
	tabClr := NewTabClear(state, w)

	isMobile := fyne.CurrentDevice().IsMobile()

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
		IsMobile:     isMobile,
	}

	logPanel.Log("Chào mừng đến với Mikrotik RouterOS 7 Automation Tool (Golang Native Edition)!")
	logPanel.Log("Hỗ trợ đa nền tảng: Windows Desktop & Samsung Galaxy A56 Android.")
	logPanel.Log("Chọn phương thức kết nối (REST API hoặc SSH) rồi nhấn 'Kết nối'.")

	return mw
}

// ShowMainView switches the current window content to the main application interface.
func (mw *MainWindow) ShowMainView() {
	if mw.IsMobile {
		mw.SetMobileLayout()
	} else {
		mw.SetDesktopLayout()
	}
}

// SetDesktopLayout renders the full desktop layout with wide connection bar and bottom log panel.
func (mw *MainWindow) SetDesktopLayout() {
	mw.IsMobile = false
	if !fyne.CurrentDevice().IsMobile() {
		mw.Window.Resize(fyne.NewSize(980, 800))
		mw.Window.CenterOnScreen()
	}

	connBarContent := mw.ConnBar.BuildDesktopContainer(func() {
		mw.SetMobileLayout()
	})

	appTabs := container.NewAppTabs(
		container.NewTabItem("📊 Dashboard", mw.TabDashboard.Container()),
		container.NewTabItem("🌐 Cấu Hình Mạng", mw.TabNetwork.Container()),
		container.NewTabItem("🔒 Bảo Mật", mw.TabSecurity.Container()),
		container.NewTabItem("🔐 IPSec VPN", mw.TabIPSec.Container()),
		container.NewTabItem("💾 Sao Lưu & Phục Hồi", mw.TabBackup.Container()),
		container.NewTabItem("🗑️ Xoá Cấu Hình", mw.TabClear.Container()),
	)

	content := container.NewBorder(
		connBarContent,
		mw.LogPanel.Container(),
		nil,
		nil,
		appTabs,
	)

	mw.Window.SetContent(content)
}

// SetMobileLayout renders the mobile-optimized layout designed for Samsung Galaxy A56 (19.5:9 portrait ratio).
func (mw *MainWindow) SetMobileLayout() {
	mw.IsMobile = true
	if !fyne.CurrentDevice().IsMobile() {
		// Simulation / Preview on PC: 412x846 mimics Samsung Galaxy A56 (1080x2340 FHD+ @ 2.625 density)
		mw.Window.Resize(fyne.NewSize(412, 846))
		mw.Window.CenterOnScreen()
	}

	mobileHeader := mw.ConnBar.BuildMobileHeader(mw.Window, func() {
		mw.SetDesktopLayout()
	})

	// Wrap each tab content in container.NewVScroll for smooth mobile scrolling
	mobileTabs := container.NewAppTabs(
		container.NewTabItem("📊 Dashboard", container.NewVScroll(mw.TabDashboard.Container())),
		container.NewTabItem("🌐 Mạng", container.NewVScroll(mw.TabNetwork.Container())),
		container.NewTabItem("🔒 Bảo Mật", container.NewVScroll(mw.TabSecurity.Container())),
		container.NewTabItem("🔐 VPN", container.NewVScroll(mw.TabIPSec.Container())),
		container.NewTabItem("💾 Sao Lưu", container.NewVScroll(mw.TabBackup.Container())),
		container.NewTabItem("🗑️ Xoá", container.NewVScroll(mw.TabClear.Container())),
		container.NewTabItem("📜 Nhật Ký", mw.LogPanel.Container()),
	)

	content := container.NewBorder(
		mobileHeader,
		nil,
		nil,
		nil,
		mobileTabs,
	)

	mw.Window.SetContent(content)
}

func (mw *MainWindow) ShowAndRun() {
	mw.ShowMainView()
	mw.Window.ShowAndRun()
}
