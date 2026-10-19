package gui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"mikrotik-automation/internal/i18n"
)

type MainWindow struct {
	App          fyne.App
	Window       fyne.Window
	State        *AppState
	ConnBar      *ConnectionBar
	HeaderBar    *HeaderBar
	Sidebar      *Sidebar
	LogPanel     *LogPanel
	LogTray      *LogTray
	ContentHost  *fyne.Container
	Split        *container.Split
	appTabs      *container.AppTabs
	tabViews     []fyne.CanvasObject
	TabDashboard *TabDashboard
	TabNetwork   *TabNetwork
	TabQueue     *TabQueue
	TabSecurity  *TabSecurity
	TabIPSec     *TabIPSec
	TabBackup    *TabBackup
	TabClear     *TabClear
}

func NewMainWindow(a fyne.App, w fyne.Window) *MainWindow {
	savedLang := a.Preferences().StringWithFallback("app_language", i18n.LangVI)
	i18n.SetLang(savedLang)

	if w == nil {
		w = a.NewWindow(i18n.T("app_title"))
	} else {
		w.SetTitle(i18n.T("app_title"))
	}

	logPanel := NewLogPanel(w)
	logTray := NewLogTray(logPanel, w)

	state := NewAppState(logPanel)
	connBar := NewConnectionBar(state, a)

	tabDash := NewTabDashboard(state, w)
	tabNet := NewTabNetwork(state, w)
	tabNet.GetConnectedHost = func() string {
		return connBar.GetConnectedHost()
	}
	tabNet.OnHostChanged = func(newHost string, onDone func(bool)) {
		connBar.ReconnectToHost(newHost, onDone)
	}
	tabQueue := NewTabQueue(state, w)
	tabSec := NewTabSecurity(state, w)
	tabIPSec := NewTabIPSec(state)
	tabBk := NewTabBackup(state, w)
	tabClr := NewTabClear(state, w)

	headerBar := NewHeaderBar(connBar, a, w)

	mw := &MainWindow{
		App:          a,
		Window:       w,
		State:        state,
		ConnBar:      connBar,
		HeaderBar:    headerBar,
		LogPanel:     logPanel,
		LogTray:      logTray,
		TabDashboard: tabDash,
		TabNetwork:   tabNet,
		TabQueue:     tabQueue,
		TabSecurity:  tabSec,
		TabIPSec:     tabIPSec,
		TabBackup:    tabBk,
		TabClear:     tabClr,
	}

	mw.Sidebar = NewSidebar(state, a, func(idx int) {
		mw.SwitchView(idx)
	})

	i18n.Subscribe(mw.UpdateLanguage)
	mw.UpdateLanguage()

	logPanel.Log("Chào mừng đến với Mikrotik RouterOS 7 Automation Tool (Enterprise Edition)!")
	logPanel.Log("Nhấn '🔌 Kết Nối / Đổi Router' trên thanh Header để bắt đầu kết nối.")

	return mw
}

// SwitchView switches the active content view in the ContentHost.
func (mw *MainWindow) SwitchView(index int) {
	if mw.ContentHost == nil || len(mw.tabViews) == 0 {
		return
	}
	if index >= 0 && index < len(mw.tabViews) {
		mw.ContentHost.Objects = []fyne.CanvasObject{mw.tabViews[index]}
		mw.ContentHost.Refresh()
		if mw.appTabs != nil && index < len(mw.appTabs.Items) {
			mw.appTabs.SelectIndex(index)
		}
	}
}

// ShowMainView renders the modern Enterprise Shell layout.
func (mw *MainWindow) ShowMainView() {
	mw.tabViews = []fyne.CanvasObject{
		mw.TabDashboard.BuildDesktopContainer(),
		mw.TabNetwork.BuildDesktopContainer(),
		mw.TabQueue.BuildDesktopContainer(),
		mw.TabSecurity.BuildDesktopContainer(),
		mw.TabIPSec.BuildDesktopContainer(),
		mw.TabBackup.BuildDesktopContainer(),
		mw.TabClear.BuildDesktopContainer(),
	}

	mw.ContentHost = container.NewStack(mw.tabViews[0])

	// Keep appTabs instantiated for testing and backward-compatibility
	mw.appTabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("tab_dashboard"), mw.tabViews[0]),
		container.NewTabItem(i18n.T("tab_network"), mw.tabViews[1]),
		container.NewTabItem(i18n.T("tab_queue"), mw.tabViews[2]),
		container.NewTabItem(i18n.T("tab_security"), mw.tabViews[3]),
		container.NewTabItem(i18n.T("tab_ipsec"), mw.tabViews[4]),
		container.NewTabItem(i18n.T("tab_backup"), mw.tabViews[5]),
		container.NewTabItem(i18n.T("tab_clear"), mw.tabViews[6]),
	)

	// Work area with Sidebar on the left and ContentHost filling the rest
	workArea := container.NewBorder(
		nil, nil,
		mw.Sidebar.Container(),
		nil,
		mw.ContentHost,
	)

	// Resizable vertical split: WorkArea on top (82%), LogTray at bottom (18%)
	mw.Split = container.NewVSplit(workArea, mw.LogTray.Container())
	mw.Split.SetOffset(0.82)

	// When log tray expands or collapses, adjust split offset smoothly
	mw.LogTray.OnHeightChg = func(isExpanded bool) {
		if isExpanded {
			mw.Split.SetOffset(0.68)
		} else {
			mw.Split.SetOffset(0.92)
		}
	}

	mw.LogPanel.SetSplitController(func(mode int) {
		switch mode {
		case 0:
			mw.Split.SetOffset(0.82)
		case 1:
			mw.Split.SetOffset(0.25)
		}
	})

	shellContent := container.NewBorder(
		mw.HeaderBar.Container(),
		nil,
		nil,
		nil,
		mw.Split,
	)

	mw.Window.SetContent(shellContent)
	mw.Window.Resize(fyne.NewSize(1080, 760))
	mw.Window.CenterOnScreen()

	time.AfterFunc(80*time.Millisecond, func() {
		fyne.Do(mw.Window.CenterOnScreen)
	})
}

func (mw *MainWindow) ShowAndRun() {
	mw.ShowMainView()
	mw.Window.ShowAndRun()
}

func (mw *MainWindow) UpdateLanguage() {
	fyne.Do(func() {
		if mw.Window != nil {
			mw.Window.SetTitle(i18n.T("app_title"))
		}
		if mw.HeaderBar != nil {
			mw.HeaderBar.UpdateLanguage()
		}
		if mw.Sidebar != nil {
			mw.Sidebar.UpdateLanguage()
		}
		if mw.LogTray != nil {
			mw.LogTray.UpdateLanguage()
		}
		if mw.appTabs != nil && len(mw.appTabs.Items) >= 7 {
			mw.appTabs.Items[0].Text = i18n.T("tab_dashboard")
			mw.appTabs.Items[1].Text = i18n.T("tab_network")
			mw.appTabs.Items[2].Text = i18n.T("tab_queue")
			mw.appTabs.Items[3].Text = i18n.T("tab_security")
			mw.appTabs.Items[4].Text = i18n.T("tab_ipsec")
			mw.appTabs.Items[5].Text = i18n.T("tab_backup")
			mw.appTabs.Items[6].Text = i18n.T("tab_clear")
			mw.appTabs.Refresh()
		}
		if mw.ConnBar != nil {
			mw.ConnBar.UpdateLanguage()
		}
		if mw.LogPanel != nil {
			mw.LogPanel.UpdateLanguage()
		}
		if mw.TabDashboard != nil {
			mw.TabDashboard.UpdateLanguage()
		}
		if mw.TabNetwork != nil {
			mw.TabNetwork.UpdateLanguage()
		}
		if mw.TabQueue != nil {
			mw.TabQueue.UpdateLanguage()
		}
		if mw.TabSecurity != nil {
			mw.TabSecurity.UpdateLanguage()
		}
		if mw.TabIPSec != nil {
			mw.TabIPSec.UpdateLanguage()
		}
		if mw.TabBackup != nil {
			mw.TabBackup.UpdateLanguage()
		}
		if mw.TabClear != nil {
			mw.TabClear.UpdateLanguage()
		}
	})
}
