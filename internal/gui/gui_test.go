package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
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

