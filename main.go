package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"mikrotik-automation/internal/gui"
)

func main() {
	a := app.NewWithID("com.mikrotik.automation")
	a.Settings().SetTheme(gui.NewCustomAppTheme(gui.CurrentThemeVariant))

	w := a.NewWindow("Mikrotik RouterOS 7 — Automation Tool")
	mw := gui.NewMainWindow(a, w)

	loginView := gui.NewLoginView(a, func() {
		mw.ShowMainView()
	})

	w.SetContent(loginView)
	w.Resize(fyne.NewSize(420, 360))
	w.CenterOnScreen()

	w.ShowAndRun()
}
