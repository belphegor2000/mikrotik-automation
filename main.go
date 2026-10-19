package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"mikrotik-automation/internal/gui"
)

//go:embed icon.png
var appIconBytes []byte

func main() {
	a := app.NewWithID("com.mikrotik.automation")
	if len(appIconBytes) > 0 {
		iconRes := fyne.NewStaticResource("icon.png", appIconBytes)
		a.SetIcon(iconRes)
	}
	a.Settings().SetTheme(gui.NewCustomAppTheme(gui.CurrentThemeVariant))

	w := a.NewWindow("Mikrotik RouterOS 7 — Automation Tool")
	if len(appIconBytes) > 0 {
		w.SetIcon(a.Icon())
	}
	mw := gui.NewMainWindow(a, w)

	loginView := gui.NewLoginView(a, func() {
		mw.ShowMainView()
	})

	w.SetContent(loginView)
	w.Resize(fyne.NewSize(420, 360))
	w.CenterOnScreen()

	w.ShowAndRun()
}
