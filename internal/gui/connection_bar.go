package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/api"
	"mikrotik-automation/internal/i18n"
)

type ConnectionBar struct {
	state          *AppState
	app            fyne.App
	lblProtoTitle  *widget.Label
	lblHostTitle   *widget.Label
	lblPortTitle   *widget.Label
	lblUserTitle   *widget.Label
	lblPassTitle   *widget.Label
	entryHost      *widget.Entry
	entryUser      *widget.Entry
	entryPass      *widget.Entry
	entryPort      *widget.Entry
	selectProto    *widget.Select
	btnConnect     *widget.Button
	btnDisconn     *widget.Button
	lblStatus      *widget.Label
	btnLang        *widget.Button
	btnTheme       *widget.Button
	container      *fyne.Container
}

func NewConnectionBar(state *AppState, a fyne.App) *ConnectionBar {
	cb := &ConnectionBar{state: state, app: a}

	cb.lblProtoTitle = widget.NewLabelWithStyle(i18n.T("conn_proto"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblHostTitle = widget.NewLabelWithStyle(i18n.T("conn_ip"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblPortTitle = widget.NewLabelWithStyle(i18n.T("conn_port"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblUserTitle = widget.NewLabelWithStyle(i18n.T("conn_user"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblPassTitle = widget.NewLabelWithStyle(i18n.T("conn_pass"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	cb.entryHost = widget.NewEntry()
	cb.entryHost.SetText("192.168.88.1")
	cb.entryHost.SetPlaceHolder("192.168.88.1")

	cb.entryUser = widget.NewEntry()
	cb.entryUser.SetText("admin")
	cb.entryUser.SetPlaceHolder("admin")

	cb.entryPass = widget.NewPasswordEntry()
	cb.entryPass.SetPlaceHolder("Password")

	cb.entryPort = widget.NewEntry()
	cb.entryPort.SetText("80")
	cb.entryPort.SetPlaceHolder("80")

	cb.selectProto = widget.NewSelect([]string{"REST API", "SSH"}, func(selected string) {
		if selected == "SSH" {
			cb.entryPort.SetText("22")
		} else {
			cb.entryPort.SetText("80")
		}
	})
	cb.selectProto.SetSelected("REST API")

	cb.lblStatus = widget.NewLabel(i18n.T("conn_not"))

	cb.btnConnect = widget.NewButton(i18n.T("conn_btn"), func() {
		cb.connect()
	})
	cb.btnConnect.Importance = widget.HighImportance

	cb.btnDisconn = widget.NewButton(i18n.T("conn_disconn"), func() {
		cb.disconnect()
	})
	cb.btnDisconn.Importance = widget.DangerImportance
	cb.btnDisconn.Disable()

	cb.btnLang = widget.NewButton(i18n.T("lang_toggle_btn"), func() {
		cb.toggleLanguage()
	})

	themeText := i18n.T("theme_light")
	if CurrentThemeVariant == theme.VariantLight {
		themeText = i18n.T("theme_dark")
	}
	cb.btnTheme = widget.NewButton(themeText, func() {
		cb.toggleTheme()
	})

	cb.container = container.NewVBox()
	return cb
}

func (cb *ConnectionBar) updateStatusText(text string) {
	if cb.lblStatus != nil {
		cb.lblStatus.SetText(text)
	}
}

func (cb *ConnectionBar) toggleLanguage() {
	newLang := i18n.ToggleLang()
	if cb.app != nil {
		cb.app.Preferences().SetString("app_language", newLang)
	}
	cb.UpdateLanguage()
	if newLang == i18n.LangEN {
		cb.state.Log("Language switched to English.")
	} else {
		cb.state.Log("Đã chuyển ngôn ngữ sang Tiếng Việt.")
	}
}

func (cb *ConnectionBar) UpdateLanguage() {
	if cb.lblProtoTitle != nil {
		cb.lblProtoTitle.SetText(i18n.T("conn_proto"))
	}
	if cb.lblHostTitle != nil {
		cb.lblHostTitle.SetText(i18n.T("conn_ip"))
	}
	if cb.lblPortTitle != nil {
		cb.lblPortTitle.SetText(i18n.T("conn_port"))
	}
	if cb.lblUserTitle != nil {
		cb.lblUserTitle.SetText(i18n.T("conn_user"))
	}
	if cb.lblPassTitle != nil {
		cb.lblPassTitle.SetText(i18n.T("conn_pass"))
	}
	if cb.btnLang != nil {
		cb.btnLang.SetText(i18n.T("lang_toggle_btn"))
	}
	if cb.btnTheme != nil {
		if CurrentThemeVariant == theme.VariantLight {
			cb.btnTheme.SetText(i18n.T("theme_dark"))
		} else {
			cb.btnTheme.SetText(i18n.T("theme_light"))
		}
	}
	c := cb.state.GetClient()
	if c == nil || !c.IsConnected() {
		cb.updateStatusText(i18n.T("conn_not"))
		cb.btnConnect.SetText(i18n.T("conn_btn"))
	} else {
		cb.btnConnect.SetText(i18n.T("conn_btn_done"))
	}
	if cb.btnDisconn != nil {
		cb.btnDisconn.SetText(i18n.T("conn_disconn"))
	}
}

func (cb *ConnectionBar) toggleTheme() {
	if cb.app == nil {
		return
	}
	if CurrentThemeVariant == theme.VariantDark {
		cb.app.Settings().SetTheme(NewCustomAppTheme(theme.VariantLight))
		cb.btnTheme.SetText(i18n.T("theme_dark"))
		cb.state.Log("Đã chuyển sang giao diện Sáng (Light mode).")
	} else {
		cb.app.Settings().SetTheme(NewCustomAppTheme(theme.VariantDark))
		cb.btnTheme.SetText(i18n.T("theme_light"))
		cb.state.Log("Đã chuyển sang giao diện Tối (Dark mode).")
	}
}

func (cb *ConnectionBar) BuildDesktopContainer() fyne.CanvasObject {
	rowInputs := container.NewGridWithColumns(5,
		container.NewVBox(cb.lblProtoTitle, cb.selectProto),
		container.NewVBox(cb.lblHostTitle, cb.entryHost),
		container.NewVBox(cb.lblPortTitle, cb.entryPort),
		container.NewVBox(cb.lblUserTitle, cb.entryUser),
		container.NewVBox(cb.lblPassTitle, cb.entryPass),
	)

	actionControls := container.NewHBox(cb.btnConnect, cb.btnDisconn, cb.btnLang, cb.btnTheme)
	statusCard := container.NewBorder(nil, nil, cb.lblStatus, actionControls)
	cb.container = container.NewVBox(container.NewPadded(rowInputs), widget.NewSeparator(), statusCard)
	return cb.container
}

func (cb *ConnectionBar) Container() fyne.CanvasObject {
	if cb.container == nil || len(cb.container.Objects) == 0 {
		return cb.BuildDesktopContainer()
	}
	return cb.container
}

func (cb *ConnectionBar) connect() {
	host := cb.entryHost.Text
	user := cb.entryUser.Text
	pass := cb.entryPass.Text
	port, _ := strconv.Atoi(cb.entryPort.Text)
	if port <= 0 {
		port = 80
	}
	proto := cb.selectProto.Selected

	cb.btnConnect.Disable()
	cb.btnConnect.SetText(i18n.T("conn_btn_ing"))
	cb.updateStatusText(i18n.T("conn_trying"))

	go func() {
		var client api.MikrotikClient
		if proto == "SSH" {
			client = api.NewSSHClient()
		} else {
			client = api.NewRESTClient()
		}

		ident, err := client.Connect(host, user, pass, port, false)
		fyne.Do(func() {
			if err != nil {
				cb.state.Log(fmt.Sprintf("✗ %s (%s): %v", i18n.T("conn_failed"), proto, err))
				cb.updateStatusText(i18n.T("conn_failed"))
				cb.btnConnect.Enable()
				cb.btnConnect.SetText(i18n.T("conn_btn"))
			} else {
				cb.state.SetClient(client, proto)
				cb.state.Log(fmt.Sprintf("✓ %s: %s (%s qua %s:%d)", i18n.T("conn_btn_done"), ident, proto, host, port))
				statusStr := fmt.Sprintf("🟢 %s (%s)", ident, proto)
				cb.updateStatusText(statusStr)
				cb.btnConnect.SetText(i18n.T("conn_btn_done"))
				cb.btnDisconn.Enable()

				// Trigger on-connect hooks
				for _, hook := range cb.state.OnConnect {
					hook()
				}
			}
		})
	}()
}

func (cb *ConnectionBar) disconnect() {
	c := cb.state.GetClient()
	if c != nil {
		_ = c.Disconnect()
	}
	cb.updateStatusText(i18n.T("conn_off"))
	cb.btnConnect.Enable()
	cb.btnConnect.SetText(i18n.T("conn_btn"))
	cb.btnDisconn.Disable()
	cb.state.Log(i18n.T("conn_off"))
}

func (cb *ConnectionBar) GetConnectedHost() string {
	if cb.entryHost != nil {
		return strings.TrimSpace(cb.entryHost.Text)
	}
	return ""
}

// ReconnectToHost updates entryHost, closes current connection, and retries
// connecting to newHost with multiple attempts.
func (cb *ConnectionBar) ReconnectToHost(newHost string, onDone func(success bool)) {
	newHost = strings.TrimSpace(newHost)
	if newHost == "" {
		if onDone != nil {
			onDone(false)
		}
		return
	}

	fyne.Do(func() {
		cb.entryHost.SetText(newHost)
		cb.updateStatusText(fmt.Sprintf("🟡 Đang chuyển sang %s...", newHost))
		cb.btnConnect.Disable()
		cb.btnConnect.SetText("⏳ Đang kết nối lại...")
		cb.btnDisconn.Disable()
	})

	c := cb.state.GetClient()
	if c != nil {
		_ = c.Disconnect()
	}

	go func() {
		user := cb.entryUser.Text
		pass := cb.entryPass.Text
		port, _ := strconv.Atoi(cb.entryPort.Text)
		if port <= 0 {
			port = 80
		}
		proto := cb.selectProto.Selected

		maxRetries := 6
		retryInterval := 2500 * time.Millisecond

		// Initial small pause to let router finish assigning IP
		time.Sleep(1000 * time.Millisecond)

		for attempt := 1; attempt <= maxRetries; attempt++ {
			cb.state.Log(fmt.Sprintf("  🔄 [Lần %d/%d] Đang thử kết nối lại tới Router tại IP mới: %s...", attempt, maxRetries, newHost))

			var client api.MikrotikClient
			if proto == "SSH" {
				client = api.NewSSHClient()
			} else {
				client = api.NewRESTClient()
			}

			ident, err := client.Connect(newHost, user, pass, port, false)
			if err == nil {
				fyne.Do(func() {
					cb.state.SetClient(client, proto)
					cb.state.Log(fmt.Sprintf("✓ Đã tự động kết nối lại thành công đến Router: %s (%s qua %s:%d)!", ident, proto, newHost, port))
					statusStr := fmt.Sprintf("🟢 %s (%s)", ident, proto)
					cb.updateStatusText(statusStr)
					cb.btnConnect.Enable()
					cb.btnConnect.SetText("✓ Đã kết nối")
					cb.btnDisconn.Enable()

					for _, hook := range cb.state.OnConnect {
						hook()
					}
					if onDone != nil {
						onDone(true)
					}
				})
				return
			}

			if attempt < maxRetries {
				time.Sleep(retryInterval)
			}
		}

		fyne.Do(func() {
			cb.state.Log(fmt.Sprintf("⚠ Không thể tự động kết nối lại tới %s sau %d lần thử. Vui lòng kiểm tra lại kết nối mạng của máy tính rồi bấm '🔌 Kết nối'.", newHost, maxRetries))
			cb.updateStatusText("🔴 Kết nối lại thất bại")
			cb.btnConnect.Enable()
			cb.btnConnect.SetText("🔌 Kết nối")
			cb.btnDisconn.Disable()
			if onDone != nil {
				onDone(false)
			}
		})
	}()
}

