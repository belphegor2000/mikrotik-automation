package gui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/api"
)

type ConnectionBar struct {
	state           *AppState
	app             fyne.App
	entryHost       *widget.Entry
	entryUser       *widget.Entry
	entryPass       *widget.Entry
	entryPort       *widget.Entry
	radioProto      *widget.RadioGroup
	btnConnect      *widget.Button
	btnDisconn      *widget.Button
	lblStatus       *widget.Label
	lblMobileStatus *widget.Label
	btnTheme        *widget.Button
	btnMobileTheme  *widget.Button
	container       *fyne.Container
}

func NewConnectionBar(state *AppState, a fyne.App) *ConnectionBar {
	cb := &ConnectionBar{state: state, app: a}

	cb.entryHost = widget.NewEntry()
	cb.entryHost.SetText("192.168.88.1")
	cb.entryHost.SetPlaceHolder("IP Router")

	cb.entryUser = widget.NewEntry()
	cb.entryUser.SetText("admin")
	cb.entryUser.SetPlaceHolder("Tài khoản")

	cb.entryPass = widget.NewPasswordEntry()
	cb.entryPass.SetPlaceHolder("Mật khẩu")

	cb.entryPort = widget.NewEntry()
	cb.entryPort.SetText("80")
	cb.entryPort.SetPlaceHolder("Cổng")

	cb.radioProto = widget.NewRadioGroup([]string{"REST API", "SSH"}, func(selected string) {
		if selected == "SSH" {
			cb.entryPort.SetText("22")
		} else {
			cb.entryPort.SetText("80")
		}
	})
	cb.radioProto.SetSelected("REST API")
	cb.radioProto.Horizontal = true

	cb.lblStatus = widget.NewLabel("⚪ Chưa kết nối")
	cb.lblMobileStatus = widget.NewLabel("⚪ Chưa kết nối")

	cb.btnConnect = widget.NewButton("🔌 Kết nối", func() {
		cb.connect()
	})

	cb.btnDisconn = widget.NewButton("❌ Ngắt kết nối", func() {
		cb.disconnect()
	})
	cb.btnDisconn.Disable()

	themeText := "☀️ Sáng"
	if CurrentThemeVariant == theme.VariantLight {
		themeText = "🌙 Tối"
	}
	cb.btnTheme = widget.NewButton(themeText, func() {
		cb.toggleTheme()
	})
	cb.btnMobileTheme = widget.NewButton(themeText, func() {
		cb.toggleTheme()
	})

	cb.container = container.NewVBox()
	return cb
}

func (cb *ConnectionBar) updateStatusText(text string) {
	if cb.lblStatus != nil {
		cb.lblStatus.SetText(text)
	}
	if cb.lblMobileStatus != nil {
		cb.lblMobileStatus.SetText(text)
	}
}

func (cb *ConnectionBar) toggleTheme() {
	if cb.app == nil {
		return
	}
	if CurrentThemeVariant == theme.VariantDark {
		cb.app.Settings().SetTheme(NewCustomAppTheme(theme.VariantLight))
		cb.btnTheme.SetText("🌙 Tối")
		cb.btnMobileTheme.SetText("🌙 Tối")
		cb.state.Log("Đã chuyển sang giao diện Sáng (Light mode).")
	} else {
		cb.app.Settings().SetTheme(NewCustomAppTheme(theme.VariantDark))
		cb.btnTheme.SetText("☀️ Sáng")
		cb.btnMobileTheme.SetText("☀️ Sáng")
		cb.state.Log("Đã chuyển sang giao diện Tối (Dark mode).")
	}
}

func (cb *ConnectionBar) BuildDesktopContainer(onToggleMobile func()) fyne.CanvasObject {
	rowInputs := container.NewGridWithColumns(6,
		container.NewVBox(widget.NewLabel("Phương thức:"), cb.radioProto),
		container.NewVBox(widget.NewLabel("Địa chỉ IP:"), cb.entryHost),
		container.NewVBox(widget.NewLabel("Cổng:"), cb.entryPort),
		container.NewVBox(widget.NewLabel("Tài khoản:"), cb.entryUser),
		container.NewVBox(widget.NewLabel("Mật khẩu:"), cb.entryPass),
		container.NewVBox(widget.NewLabel("Thao tác:"), container.NewGridWithColumns(2, cb.btnConnect, cb.btnDisconn)),
	)

	var rightControls fyne.CanvasObject
	if onToggleMobile != nil {
		btnMobile := widget.NewButton("📱 Chế độ Mobile", onToggleMobile)
		rightControls = container.NewHBox(btnMobile, cb.btnTheme)
	} else {
		rightControls = cb.btnTheme
	}

	statusCard := container.NewBorder(nil, nil, cb.lblStatus, rightControls)
	cb.container = container.NewVBox(rowInputs, statusCard)
	return cb.container
}

func (cb *ConnectionBar) BuildMobileHeader(w fyne.Window, onToggleDesktop func()) fyne.CanvasObject {
	btnConfig := widget.NewButtonWithIcon("⚙️ Cấu hình", theme.SettingsIcon(), func() {
		cb.ShowMobileSettingsDialog(w)
	})

	var rightControls fyne.CanvasObject
	if !fyne.CurrentDevice().IsMobile() && onToggleDesktop != nil {
		btnDesktop := widget.NewButton("💻 Desktop", onToggleDesktop)
		rightControls = container.NewHBox(btnConfig, cb.btnMobileTheme, btnDesktop)
	} else {
		rightControls = container.NewHBox(btnConfig, cb.btnMobileTheme)
	}

	headerCard := container.NewBorder(
		nil, nil,
		cb.lblMobileStatus,
		rightControls,
	)

	return container.NewPadded(headerCard)
}

func (cb *ConnectionBar) ShowMobileSettingsDialog(w fyne.Window) {
	formContent := container.NewVBox(
		widget.NewLabelWithStyle("⚙️ THIẾT LẬP KẾT NỐI ROUTEROS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Phương thức kết nối:"),
		cb.radioProto,
		widget.NewLabel("Địa chỉ IP Router:"),
		cb.entryHost,
		widget.NewLabel("Cổng (Port):"),
		cb.entryPort,
		widget.NewLabel("Tài khoản:"),
		cb.entryUser,
		widget.NewLabel("Mật khẩu:"),
		cb.entryPass,
		container.NewGridWithColumns(2, cb.btnConnect, cb.btnDisconn),
		widget.NewSeparator(),
		cb.lblStatus,
	)

	d := dialog.NewCustom("Kết Nối MikroTik", "Đóng", container.NewVScroll(container.NewPadded(formContent)), w)
	d.Resize(fyne.NewSize(360, 480))
	d.Show()
}

func (cb *ConnectionBar) Container() fyne.CanvasObject {
	if cb.container == nil || len(cb.container.Objects) == 0 {
		return cb.BuildDesktopContainer(nil)
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
	proto := cb.radioProto.Selected

	cb.btnConnect.Disable()
	cb.btnConnect.SetText("⏳ Đang kết nối...")
	cb.updateStatusText("🟡 Đang kết nối...")

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
				cb.state.Log(fmt.Sprintf("✗ Kết nối %s thất bại: %v", proto, err))
				cb.updateStatusText("🔴 Kết nối thất bại")
				cb.btnConnect.Enable()
				cb.btnConnect.SetText("🔌 Kết nối")
			} else {
				cb.state.SetClient(client, proto)
				cb.state.Log(fmt.Sprintf("✓ Đã kết nối đến router: %s (%s qua %s:%d)", ident, proto, host, port))
				statusStr := fmt.Sprintf("🟢 %s (%s)", ident, proto)
				cb.updateStatusText(statusStr)
				cb.btnConnect.SetText("✓ Đã kết nối")
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
	cb.updateStatusText("⚪ Đã ngắt kết nối")
	cb.btnConnect.Enable()
	cb.btnConnect.SetText("🔌 Kết nối")
	cb.btnDisconn.Disable()
	cb.state.Log("Đã ngắt kết nối router.")
}
