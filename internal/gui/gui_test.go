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

func TestMainWindowLayoutSwitching(t *testing.T) {
	a := test.NewApp()
	w := test.NewWindow(nil)

	mw := NewMainWindow(a, w)
	if mw == nil {
		t.Fatal("Expected mw to be non-nil")
	}

	// Test Desktop Layout
	mw.SetDesktopLayout()
	if mw.IsMobile {
		t.Errorf("Expected IsMobile to be false in desktop layout")
	}

	// Test Mobile Layout (Samsung Galaxy A56 portrait simulation)
	mw.SetMobileLayout()
	if !mw.IsMobile {
		t.Errorf("Expected IsMobile to be true in mobile layout")
	}

	// Switch back to Desktop
	mw.SetDesktopLayout()
	if mw.IsMobile {
		t.Errorf("Expected IsMobile to be false after switching back")
	}
}
