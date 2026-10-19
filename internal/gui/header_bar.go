package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
)

// HeaderBar renders a modern, slim top header bar (48px) with brand logo,
// real-time device chip status, language/theme toggles, and connection modal button.
type HeaderBar struct {
	app        fyne.App
	window     fyne.Window
	connBar    *ConnectionBar
	container  *fyne.Container
	lblBrand   *widget.Label
	lblDevice  *widget.Label
	btnLang    *widget.Button
	btnTheme   *widget.Button
	btnConnect *widget.Button
	btnDisconn *widget.Button
}

func NewHeaderBar(connBar *ConnectionBar, app fyne.App, w fyne.Window) *HeaderBar {
	hb := &HeaderBar{
		app:     app,
		window:  w,
		connBar: connBar,
	}

	hb.lblBrand = widget.NewLabelWithStyle("⚡ MIKROTIK AUTOMATION", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	hb.lblDevice = widget.NewLabelWithStyle(fmtChip(i18n.T("conn_not")), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	hb.btnLang = widget.NewButton(i18n.T("lang_toggle_btn"), func() {
		hb.connBar.toggleLanguage()
		hb.UpdateLanguage()
	})

	themeText := i18n.T("theme_light")
	if CurrentThemeVariant == theme.VariantLight {
		themeText = i18n.T("theme_dark")
	}
	hb.btnTheme = widget.NewButton(themeText, func() {
		hb.connBar.toggleTheme()
		hb.UpdateLanguage()
	})

	hb.btnConnect = widget.NewButton(i18n.T("hdr_btn_manage"), func() {
		hb.connBar.ShowDialog(hb.window)
	})
	hb.btnConnect.Importance = widget.HighImportance

	hb.btnDisconn = widget.NewButton("❌", func() {
		hb.connBar.disconnect()
	})
	hb.btnDisconn.Importance = widget.DangerImportance
	hb.btnDisconn.Disable()

	// Bind status change listener from ConnectionBar
	hb.connBar.OnStatusChanged = func(status string, isConnected bool) {
		fyne.Do(func() {
			if isConnected {
				hb.lblDevice.SetText(fmtChip(status))
				hb.btnDisconn.Enable()
				hb.btnConnect.SetText(i18n.T("hdr_btn_change"))
			} else {
				hb.lblDevice.SetText(fmtChip(i18n.T("conn_not")))
				hb.btnDisconn.Disable()
				hb.btnConnect.SetText(i18n.T("hdr_btn_connect"))
			}
		})
	}

	rightCluster := container.NewHBox(
		hb.btnLang,
		hb.btnTheme,
		hb.btnConnect,
		hb.btnDisconn,
	)

	centerCluster := container.NewCenter(hb.lblDevice)

	topBar := container.NewBorder(
		nil, nil,
		hb.lblBrand,
		rightCluster,
		centerCluster,
	)

	hb.container = container.NewVBox(
		container.NewPadded(topBar),
		widget.NewSeparator(),
	)

	return hb
}

func fmtChip(status string) string {
	s := strings.TrimSpace(status)
	if !strings.HasPrefix(s, "[") {
		return "[ " + s + " ]"
	}
	return s
}

func (hb *HeaderBar) Container() fyne.CanvasObject {
	return hb.container
}

func (hb *HeaderBar) UpdateLanguage() {
	if hb.btnLang != nil {
		hb.btnLang.SetText(i18n.T("lang_toggle_btn"))
	}
	if hb.btnTheme != nil {
		if CurrentThemeVariant == theme.VariantLight {
			hb.btnTheme.SetText(i18n.T("theme_dark"))
		} else {
			hb.btnTheme.SetText(i18n.T("theme_light"))
		}
	}
	c := hb.connBar.state.GetClient()
	if c == nil || !c.IsConnected() {
		if hb.lblDevice != nil {
			hb.lblDevice.SetText(fmtChip(i18n.T("conn_not")))
		}
		if hb.btnConnect != nil {
			hb.btnConnect.SetText(i18n.T("hdr_btn_connect"))
		}
	} else {
		if hb.btnConnect != nil {
			hb.btnConnect.SetText(i18n.T("hdr_btn_change"))
		}
	}
}
