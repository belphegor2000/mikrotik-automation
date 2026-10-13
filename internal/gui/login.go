package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/auth"
)

// NewLoginView creates a CanvasObject representing the login screen.
func NewLoginView(a fyne.App, onLoginSuccess func()) fyne.CanvasObject {
	authMgr, err := auth.LoadOrCreateAuthManager()
	if err != nil {
		authMgr, _ = auth.LoadOrCreateAuthManager()
	}

	header := widget.NewLabelWithStyle("🔐 MIKROTIK ROUTEROS 7", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subHeader := widget.NewLabelWithStyle("Automation & Network Controller", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	entryUser := widget.NewEntry()
	entryUser.SetText("admin")
	entryUser.SetPlaceHolder("Tài khoản quản trị")

	entryPass := widget.NewPasswordEntry()
	entryPass.SetPlaceHolder("Mật khẩu")

	lblError := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	doLogin := func() {
		u := entryUser.Text
		p := entryPass.Text

		if u == "" || p == "" {
			lblError.SetText("⚠️ Vui lòng nhập đầy đủ tài khoản và mật khẩu!")
			return
		}

		if authMgr.Authenticate(u, p) {
			if onLoginSuccess != nil {
				onLoginSuccess()
			}
		} else {
			lblError.SetText("❌ Tài khoản hoặc mật khẩu không chính xác!")
			entryPass.SetText("")
		}
	}

	btnLogin := widget.NewButton("✓ ĐĂNG NHẬP HỆ THỐNG", doLogin)
	btnLogin.Importance = widget.HighImportance

	entryPass.OnSubmitted = func(_ string) {
		doLogin()
	}

	lblHint := widget.NewLabelWithStyle("💡 Mặc định: admin / Loveisdie1@", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Tài khoản:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		entryUser,
		widget.NewLabelWithStyle("Mật khẩu:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		entryPass,
		lblError,
		container.NewPadded(btnLogin),
		lblHint,
	)

	card := container.NewVBox(
		header,
		subHeader,
		widget.NewSeparator(),
		form,
	)

	box := container.NewGridWrap(fyne.NewSize(350, 340), card)
	return container.NewVScroll(container.NewCenter(box))
}

// ShowLoginWindow displays a modal-like login window (retained for backward compatibility).
func ShowLoginWindow(a fyne.App, onLoginSuccess func()) {
	w := a.NewWindow("Đăng Nhập — Mikrotik Automation Tool")
	w.Resize(fyne.NewSize(380, 320))
	w.CenterOnScreen()

	loginSuccess := false
	loginView := NewLoginView(a, func() {
		loginSuccess = true
		w.Close()
		if onLoginSuccess != nil {
			onLoginSuccess()
		}
	})

	w.SetOnClosed(func() {
		if !loginSuccess {
			a.Quit()
		}
	})

	w.SetContent(loginView)
	w.Show()
}

