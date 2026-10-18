package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/auth"
	"mikrotik-automation/internal/i18n"
)

// NewLoginView creates a CanvasObject representing the login screen.
func NewLoginView(a fyne.App, onLoginSuccess func()) fyne.CanvasObject {
	authMgr, err := auth.LoadOrCreateAuthManager()
	if err != nil {
		authMgr, _ = auth.LoadOrCreateAuthManager()
	}

	header := widget.NewLabelWithStyle(i18n.T("login_header"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subHeader := widget.NewLabelWithStyle(i18n.T("login_sub"), fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	lblUserTitle := widget.NewLabelWithStyle(i18n.T("login_user_label"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	entryUser := widget.NewEntry()
	entryUser.SetText("admin")
	entryUser.SetPlaceHolder(i18n.T("login_user"))

	lblPassTitle := widget.NewLabelWithStyle(i18n.T("login_pass_label"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	entryPass := widget.NewPasswordEntry()
	entryPass.SetPlaceHolder(i18n.T("login_pass"))

	lblError := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	lblHint := widget.NewLabelWithStyle(i18n.T("login_hint"), fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	btnLogin := widget.NewButton(i18n.T("login_btn"), nil)
	btnLogin.Importance = widget.HighImportance

	btnLang := widget.NewButton(i18n.T("lang_toggle_btn"), nil)
	btnChangePass := widget.NewButton(i18n.T("login_chg_pass_btn"), nil)

	updateLanguage := func() {
		header.SetText(i18n.T("login_header"))
		subHeader.SetText(i18n.T("login_sub"))
		lblUserTitle.SetText(i18n.T("login_user_label"))
		entryUser.SetPlaceHolder(i18n.T("login_user"))
		lblPassTitle.SetText(i18n.T("login_pass_label"))
		entryPass.SetPlaceHolder(i18n.T("login_pass"))
		lblHint.SetText(i18n.T("login_hint"))
		btnLogin.SetText(i18n.T("login_btn"))
		btnLang.SetText(i18n.T("lang_toggle_btn"))
		btnChangePass.SetText(i18n.T("login_chg_pass_btn"))
	}
	i18n.Subscribe(updateLanguage)

	btnLang.OnTapped = func() {
		newLang := i18n.ToggleLang()
		if a != nil {
			a.Preferences().SetString("app_language", newLang)
		}
	}

	doLogin := func() {
		u := entryUser.Text
		p := entryPass.Text

		if u == "" || p == "" {
			lblError.SetText(i18n.T("login_err_empty"))
			return
		}

		if authMgr.Authenticate(u, p) {
			if onLoginSuccess != nil {
				onLoginSuccess()
			}
		} else {
			lblError.SetText(i18n.T("login_err_fail"))
			entryPass.SetText("")
		}
	}

	btnLogin.OnTapped = doLogin
	entryPass.OnSubmitted = func(_ string) {
		doLogin()
	}

	btnChangePass.OnTapped = func() {
		// Find main window if available
		var parentWin fyne.Window
		if a != nil && len(a.Driver().AllWindows()) > 0 {
			parentWin = a.Driver().AllWindows()[0]
		}

		entryOld := widget.NewPasswordEntry()
		entryOld.SetPlaceHolder(i18n.T("chg_pass_old"))
		entryNew := widget.NewPasswordEntry()
		entryNew.SetPlaceHolder(i18n.T("chg_pass_new"))
		entryConfirm := widget.NewPasswordEntry()
		entryConfirm.SetPlaceHolder(i18n.T("chg_pass_confirm"))
		lblModalErr := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

		formContent := container.NewVBox(
			widget.NewLabelWithStyle(i18n.T("chg_pass_old"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			entryOld,
			widget.NewLabelWithStyle(i18n.T("chg_pass_new"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			entryNew,
			widget.NewLabelWithStyle(i18n.T("chg_pass_confirm"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			entryConfirm,
			lblModalErr,
		)

		if parentWin != nil {
			d := dialog.NewCustomConfirm(
				i18n.T("chg_pass_title"),
				i18n.T("chg_pass_btn"),
				i18n.T("btn_cancel"),
				formContent,
				func(ok bool) {
					if !ok {
						return
					}
					u := entryUser.Text
					if u == "" {
						u = "admin"
					}
					if entryNew.Text != entryConfirm.Text {
						lblError.SetText(i18n.T("chg_pass_err_match"))
						return
					}
					err := authMgr.ChangePassword(u, entryOld.Text, entryNew.Text)
					if err != nil {
						lblError.SetText("✗ " + err.Error())
					} else {
						lblError.SetText(i18n.T("chg_pass_success"))
						entryPass.SetText("")
					}
				},
				parentWin,
			)
			d.Resize(fyne.NewSize(360, 300))
			d.Show()
		}
	}

	topBar := container.NewHBox(
		btnLang,
		container.NewCenter(),
		btnChangePass,
	)

	form := container.NewVBox(
		lblUserTitle,
		entryUser,
		lblPassTitle,
		entryPass,
		lblError,
		container.NewPadded(btnLogin),
		lblHint,
		widget.NewSeparator(),
		container.NewCenter(topBar),
	)

	card := container.NewVBox(
		header,
		subHeader,
		widget.NewSeparator(),
		form,
	)

	box := container.NewGridWrap(fyne.NewSize(380, 410), card)
	return container.NewVScroll(container.NewCenter(box))
}

// ShowLoginWindow displays a modal-like login window (retained for backward compatibility).
func ShowLoginWindow(a fyne.App, onLoginSuccess func()) {
	w := a.NewWindow(i18n.T("app_title"))
	w.Resize(fyne.NewSize(420, 440))
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
