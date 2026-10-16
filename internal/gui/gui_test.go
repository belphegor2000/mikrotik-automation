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
