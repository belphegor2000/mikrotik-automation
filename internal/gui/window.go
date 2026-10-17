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
	LogPanel     *LogPanel
	Split        *container.Split
	appTabs      *container.AppTabs
	TabDashboard *TabDashboard
	TabNetwork   *TabNetwork
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

	i18n.Subscribe(mw.UpdateLanguage)
	mw.UpdateLanguage()

	logPanel.Log("Chào mừng đến với Mikrotik RouterOS 7 Automation Tool (Desktop Edition)!")
	logPanel.Log("Chọn phương thức kết nối (REST API hoặc SSH) rồi nhấn 'Kết nối'.")

	return mw
}

// ShowMainView renders the desktop interface with a wide connection bar, tabs, and resizable spacious log panel.
func (mw *MainWindow) ShowMainView() {
	connBarContent := mw.ConnBar.BuildDesktopContainer()

	mw.appTabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("tab_dashboard"), mw.TabDashboard.BuildDesktopContainer()),
		container.NewTabItem(i18n.T("tab_network"), mw.TabNetwork.BuildDesktopContainer()),
		container.NewTabItem(i18n.T("tab_security"), mw.TabSecurity.BuildDesktopContainer()),
		container.NewTabItem(i18n.T("tab_ipsec"), mw.TabIPSec.BuildDesktopContainer()),
		container.NewTabItem(i18n.T("tab_backup"), mw.TabBackup.BuildDesktopContainer()),
		container.NewTabItem(i18n.T("tab_clear"), mw.TabClear.BuildDesktopContainer()),
	)
	mw.appTabs.SetTabLocation(container.TabLocationTop)

	// Resizable vertical split: 65% Tabs, 35% Logs
	mw.Split = container.NewVSplit(mw.appTabs, mw.LogPanel.Container())
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
	mw.Window.Resize(fyne.NewSize(1000, 750))
	mw.Window.CenterOnScreen()

	// Ensure Windows OS message pump finishes WM_SIZE and snaps window to center
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
		if mw.appTabs != nil && len(mw.appTabs.Items) >= 6 {
			mw.appTabs.Items[0].Text = i18n.T("tab_dashboard")
			mw.appTabs.Items[1].Text = i18n.T("tab_network")
			mw.appTabs.Items[2].Text = i18n.T("tab_security")
			mw.appTabs.Items[3].Text = i18n.T("tab_ipsec")
			mw.appTabs.Items[4].Text = i18n.T("tab_backup")
			mw.appTabs.Items[5].Text = i18n.T("tab_clear")
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

