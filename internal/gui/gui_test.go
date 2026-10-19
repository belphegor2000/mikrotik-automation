package gui

import (
	"sync"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"mikrotik-automation/internal/modules/vlan"
)

func TestLoginViewCreation(t *testing.T) {
	a := test.NewApp()
	called := false
	view := NewLoginView(a, func() {
		called = true
	})
	if view == nil {
		t.Fatal("Expected login view to be non-nil")
	}
	_ = called
}

func TestMainWindowDesktopLayout(t *testing.T) {
	a := test.NewApp()
	w := test.NewWindow(nil)

	mw := NewMainWindow(a, w)
	if mw == nil {
		t.Fatal("Expected mw to be non-nil")
	}

	mw.ShowMainView()
	if mw.Split == nil {
		t.Fatal("Expected Split to be non-nil in desktop view")
	}
	if mw.Split.Offset < 0.5 {
		t.Errorf("Expected Split offset to prioritize tabs by default, got %f", mw.Split.Offset)
	}

	// Verify log logging and clearing
	mw.LogPanel.Log("Test log line 1")
	mw.LogPanel.Log("Test log line 2")
	if len(mw.LogPanel.logLines) != 4 { // 2 init logs + 2 test logs
		t.Errorf("Expected 4 log lines, got %d", len(mw.LogPanel.logLines))
	}

	mw.LogPanel.Clear()
	if len(mw.LogPanel.logLines) != 0 {
		t.Errorf("Expected 0 log lines after clear, got %d", len(mw.LogPanel.logLines))
	}
}

func TestTabNetworkVLANAndBridgeRows(t *testing.T) {
	_ = test.NewApp()
	w := test.NewWindow(nil)
	lp := NewLogPanel(w)
	state := NewAppState(lp)
	tn := NewTabNetwork(state, w)

	// Verify Bridge default IP
	if tn.entryBridgeIP.Text != "192.168.88.1/24" {
		t.Errorf("Expected default bridge IP 192.168.88.1/24, got %s", tn.entryBridgeIP.Text)
	}

	// Verify 5 initial VLAN rows
	if len(tn.vlanRows) != 5 {
		t.Fatalf("Expected 5 initial VLAN rows, got %d", len(tn.vlanRows))
	}

	// Add 6th row
	tn.addVLANRow()
	if len(tn.vlanRows) != 6 {
		t.Fatalf("Expected 6 VLAN rows after add, got %d", len(tn.vlanRows))
	}

	// Delete 6th row -> should reduce to 5
	tn.removeOrClearRow(tn.vlanRows[5])
	if len(tn.vlanRows) != 5 {
		t.Fatalf("Expected 5 VLAN rows after removing 6th row, got %d", len(tn.vlanRows))
	}

	// Fill row 0 and row 1 with valid VLANs
	tn.vlanRows[0].entryID.SetText("10")
	tn.vlanRows[0].entryName.SetText("users")
	tn.vlanRows[0].entryCIDR.SetText("192.168.10.0/24")

	tn.vlanRows[1].entryID.SetText("20")
	tn.vlanRows[1].entryName.SetText("servers")
	tn.vlanRows[1].entryCIDR.SetText("192.168.20.0/24")

	// Verify auto-calculated gateway on row 0
	if tn.vlanRows[0].lblGateway.Text != "192.168.10.1" {
		t.Errorf("Expected gateway 192.168.10.1, got %s", tn.vlanRows[0].lblGateway.Text)
	}

	// Sync from UI: rows 2, 3, 4 are empty and should be ignored
	vlans, err := tn.syncVLANsFromUI()
	if err != nil {
		t.Fatalf("Unexpected error syncing VLANs: %v", err)
	}
	if len(vlans) != 2 {
		t.Fatalf("Expected 2 synced VLANs, got %d", len(vlans))
	}
	if vlans[0].VLANID != 10 || vlans[1].VLANID != 20 {
		t.Errorf("Unexpected synced VLAN IDs: %v", vlans)
	}

	// Clear all rows -> still 5 rows, but texts empty
	tn.clearAllVLANRows()
	if len(tn.vlanRows) != 5 {
		t.Fatalf("Expected 5 VLAN rows after clear all, got %d", len(tn.vlanRows))
	}
	if tn.vlanRows[0].entryID.Text != "" {
		t.Errorf("Expected row 0 ID to be empty after clear, got %s", tn.vlanRows[0].entryID.Text)
	}
}

func TestLanguageSwitchingInGUI(t *testing.T) {
	a := test.NewApp()
	a.Preferences().SetString("app_language", "vi")
	w := test.NewWindow(nil)

	mw := NewMainWindow(a, w)
	mw.ShowMainView()

	// Initial check: Vietnamese
	if mw.ConnBar.btnLang.Text != "🌐 English" {
		t.Errorf("Expected btnLang to show '🌐 English' in VI mode, got '%s'", mw.ConnBar.btnLang.Text)
	}
	if mw.appTabs.Items[1].Text != "🌐 Cấu Hình Mạng" {
		t.Errorf("Expected tab 1 to be '🌐 Cấu Hình Mạng', got '%s'", mw.appTabs.Items[1].Text)
	}
	if mw.appTabs.Items[2].Text != "⚡ Băng Thông" {
		t.Errorf("Expected tab 2 to be '⚡ Băng Thông', got '%s'", mw.appTabs.Items[2].Text)
	}
	if mw.TabQueue.btnAdd.Text != "➕ Thêm Giới Hạn Băng Thông" {
		t.Errorf("Expected queue_add_btn in VI, got '%s'", mw.TabQueue.btnAdd.Text)
	}
	if mw.TabSecurity.btnDisable.Text != "🚫 Tắt Dịch Vụ Không Dùng" {
		t.Errorf("Expected sec_srv_btn in VI, got '%s'", mw.TabSecurity.btnDisable.Text)
	}
	if mw.HeaderBar.lblDevice.Text != "[ ⚪ Chưa kết nối ]" {
		t.Errorf("Expected HeaderBar lblDevice in VI, got '%s'", mw.HeaderBar.lblDevice.Text)
	}

	// Toggle language to English
	mw.ConnBar.toggleLanguage()

	if mw.ConnBar.btnLang.Text != "🌐 Tiếng Việt" {
		t.Errorf("Expected btnLang to show '🌐 Tiếng Việt' in EN mode, got '%s'", mw.ConnBar.btnLang.Text)
	}
	if mw.appTabs.Items[1].Text != "🌐 Network Setup" {
		t.Errorf("Expected tab 1 to be '🌐 Network Setup', got '%s'", mw.appTabs.Items[1].Text)
	}
	if mw.appTabs.Items[2].Text != "⚡ Bandwidth" {
		t.Errorf("Expected tab 2 to be '⚡ Bandwidth', got '%s'", mw.appTabs.Items[2].Text)
	}
	if mw.TabQueue.btnAdd.Text != "➕ Add Bandwidth Queue" {
		t.Errorf("Expected queue_add_btn in EN, got '%s'", mw.TabQueue.btnAdd.Text)
	}
	if mw.TabSecurity.btnDisable.Text != "🚫 Disable Insecure Services" {
		t.Errorf("Expected sec_srv_btn in EN, got '%s'", mw.TabSecurity.btnDisable.Text)
	}
	if mw.TabClear.btnClear.Text != "🗑️ Run 15-Step Automated Reset" {
		t.Errorf("Expected clr_btn in EN, got '%s'", mw.TabClear.btnClear.Text)
	}
	if mw.HeaderBar.lblDevice.Text != "[ ⚪ Disconnected ]" {
		t.Errorf("Expected HeaderBar lblDevice in EN, got '%s'", mw.HeaderBar.lblDevice.Text)
	}
	if pref := a.Preferences().String("app_language"); pref != "en" {
		t.Errorf("Expected preference 'app_language' to be 'en', got '%s'", pref)
	}

	// Toggle language back to Vietnamese
	mw.ConnBar.toggleLanguage()

	if mw.ConnBar.btnLang.Text != "🌐 English" {
		t.Errorf("Expected btnLang to show '🌐 English' after toggle back, got '%s'", mw.ConnBar.btnLang.Text)
	}
	if mw.appTabs.Items[1].Text != "🌐 Cấu Hình Mạng" {
		t.Errorf("Expected tab 1 to revert to '🌐 Cấu Hình Mạng', got '%s'", mw.appTabs.Items[1].Text)
	}
	if mw.appTabs.Items[2].Text != "⚡ Băng Thông" {
		t.Errorf("Expected tab 2 to revert to '⚡ Băng Thông', got '%s'", mw.appTabs.Items[2].Text)
	}
	if mw.TabQueue.btnAdd.Text != "➕ Thêm Giới Hạn Băng Thông" {
		t.Errorf("Expected queue_add_btn in VI after toggle back, got '%s'", mw.TabQueue.btnAdd.Text)
	}
	if mw.HeaderBar.lblDevice.Text != "[ ⚪ Chưa kết nối ]" {
		t.Errorf("Expected HeaderBar lblDevice in VI after toggle back, got '%s'", mw.HeaderBar.lblDevice.Text)
	}
	if pref := a.Preferences().String("app_language"); pref != "vi" {
		t.Errorf("Expected preference 'app_language' to be 'vi', got '%s'", pref)
	}
}

func TestAppStateVLANConcurrency(t *testing.T) {
	state := NewAppState(nil)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			entries := state.GetVLANs()
			_ = len(entries)
		}(i)
		go func(id int) {
			defer wg.Done()
			state.SetVLANs([]vlan.VLANEntry{
				{VLANID: id, Name: "test", CIDR: "192.168.1.0/24"},
			})
		}(i)
	}
	wg.Wait()
}

func TestTabSecuritySubTabsAndNAT(t *testing.T) {
	_ = test.NewApp()
	w := test.NewWindow(nil)
	lp := NewLogPanel(w)
	state := NewAppState(lp)
	ts := NewTabSecurity(state, w)

	container := ts.Container()
	if container == nil {
		t.Fatal("Expected TabSecurity container to be non-nil")
	}

	if ts.subTabs == nil || len(ts.subTabs.Items) < 2 {
		t.Fatalf("Expected at least 2 sub-tabs in TabSecurity, got %d", len(ts.subTabs.Items))
	}

	if ts.btnAddPF == nil || ts.btnAutoHairpin == nil {
		t.Fatal("Expected NAT buttons to be non-nil")
	}

	// Verify default inputs
	if ts.selectPFProto.Selected != "TCP" {
		t.Errorf("Expected default protocol TCP, got %s", ts.selectPFProto.Selected)
	}
}

func TestTabQueueSubTabsAndCakeQoS(t *testing.T) {
	_ = test.NewApp()
	w := test.NewWindow(nil)
	lp := NewLogPanel(w)
	state := NewAppState(lp)
	tq := NewTabQueue(state, w)

	c := tq.Container()
	if c == nil {
		t.Fatal("Expected TabQueue container to be non-nil")
	}

	if tq.subTabs == nil || len(tq.subTabs.Items) < 2 {
		t.Fatalf("Expected 2 sub-tabs in TabQueue, got %d", len(tq.subTabs.Items))
	}

	// Verify CAKE QoS fields
	if tq.entryCakeDown == nil || tq.entryCakeUp == nil || tq.btnApplyCake == nil {
		t.Fatal("Expected CAKE QoS widgets to be non-nil")
	}

	// Test 95% calculation
	tq.entryCakeDown.SetText("200")
	tq.entryCakeUp.SetText("100")
	tq.updateCakeCalc()

	// 200 * 0.95 = 190, 100 * 0.95 = 95
	calcText := tq.lblCakeCalc.Text
	if calcText == "" || len(calcText) < 10 {
		t.Errorf("Expected calc label to have text, got empty")
	}
}

func TestTabNetworkPBRSubTab(t *testing.T) {
	_ = test.NewApp()
	w := test.NewWindow(nil)
	lp := NewLogPanel(w)
	state := NewAppState(lp)
	tn := NewTabNetwork(state, w)
	c := tn.Container()
	if c == nil {
		t.Fatal("Expected TabNetwork container to be non-nil")
	}

	if tn.subTabs == nil || len(tn.subTabs.Items) < 4 {
		t.Fatalf("Expected 4 sub-tabs in TabNetwork, got %d", len(tn.subTabs.Items))
	}

	if tn.btnAddPBR == nil || tn.radioPBRType == nil || tn.selectPBRWAN == nil {
		t.Fatal("Expected PBR controls to be initialized")
	}

	// Verify recursive routing checkbox exists
	if tn.chkRecursive == nil {
		t.Fatal("Expected chkRecursive to be initialized")
	}
}

func TestEnterpriseThemeColors(t *testing.T) {
	darkTheme := NewEnterpriseTheme(true)
	lightTheme := NewEnterpriseTheme(false)

	darkBg := darkTheme.Color(theme.ColorNameBackground, theme.VariantDark)
	r, g, b, _ := darkBg.RGBA()
	// Check dark background is deep slate (low values)
	if r > 0x2000 || g > 0x3000 || b > 0x4000 {
		t.Errorf("Expected dark background to be deep slate, got rgba(%d, %d, %d)", r, g, b)
	}

	lightBg := lightTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	lr, lg, lb, _ := lightBg.RGBA()
	// Check light background is near pure white/slate (high values)
	if lr < 0xF000 || lg < 0xF000 || lb < 0xF000 {
		t.Errorf("Expected light background to be near white, got rgba(%d, %d, %d)", lr, lg, lb)
	}

	// Verify Corner Radius
	radius := darkTheme.Size(theme.SizeNameInputRadius)
	if radius != 8.0 {
		t.Errorf("Expected input radius 8.0, got %f", radius)
	}
}

func TestSidebarToggleAndCollapse(t *testing.T) {
	a := test.NewApp()
	lp := NewLogPanel(test.NewWindow(nil))
	state := NewAppState(lp)

	selectedIdx := -1
	sb := NewSidebar(state, a, func(idx int) {
		selectedIdx = idx
	})

	if sb.Container() == nil {
		t.Fatal("Expected Sidebar container to be non-nil")
	}

	if sb.IsCollapsed() {
		t.Errorf("Expected sidebar to start expanded by default")
	}

	// Test click
	sb.SetActive(2)
	if sb.activeIndex != 2 {
		t.Errorf("Expected activeIndex 2, got %d", sb.activeIndex)
	}

	// Test toggle
	sb.Toggle()
	if !sb.IsCollapsed() {
		t.Errorf("Expected sidebar to be collapsed after Toggle()")
	}

	// Test toggle back
	sb.Toggle()
	if sb.IsCollapsed() {
		t.Errorf("Expected sidebar to be expanded after second Toggle()")
	}
	_ = selectedIdx
}

func TestLogTrayCollapseExpandAndAutoExpand(t *testing.T) {
	w := test.NewWindow(nil)
	lp := NewLogPanel(w)
	tray := NewLogTray(lp, w)

	if tray.Container() == nil {
		t.Fatal("Expected LogTray container to be non-nil")
	}

	if tray.IsExpanded() {
		t.Errorf("Expected LogTray to be collapsed by default")
	}

	// Test manual toggle
	tray.Toggle()
	if !tray.IsExpanded() {
		t.Errorf("Expected LogTray to be expanded after Toggle()")
	}

	tray.SetExpanded(false)
	if tray.IsExpanded() {
		t.Errorf("Expected LogTray to be collapsed after SetExpanded(false)")
	}

	// Test smart auto-expand on deployment step
	tray.Log("▶ [BƯỚC 1/3] CẤU HÌNH PPPoE WAN...")
	if !tray.IsExpanded() {
		t.Errorf("Expected LogTray to auto-expand on [BƯỚC log line")
	}
}

func TestHeaderBarDeviceChip(t *testing.T) {
	a := test.NewApp()
	w := test.NewWindow(nil)
	lp := NewLogPanel(w)
	state := NewAppState(lp)
	cb := NewConnectionBar(state, a)
	hb := NewHeaderBar(cb, a, w)

	if hb.Container() == nil {
		t.Fatal("Expected HeaderBar container to be non-nil")
	}

	if hb.lblDevice.Text == "" {
		t.Errorf("Expected device chip label to have text")
	}

	// Simulate connection status change
	cb.updateStatusText("192.168.88.1 (REST API)")
	if hb.lblDevice.Text == "" {
		t.Errorf("Expected device chip label to reflect connected status")
	}
}

