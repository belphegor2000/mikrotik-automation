package gui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/api"
)

type ConnectionBar struct {
	state      *AppState
	app        fyne.App
	entryHost  *widget.Entry
	entryUser  *widget.Entry
	entryPass  *widget.Entry
	entryPort  *widget.Entry
	radioProto *widget.RadioGroup
	btnConnect *widget.Button
	btnDisconn *widget.Button
	lblStatus  *widget.Label
	btnTheme   *widget.Button
	container  *fyne.Container
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

	cb.btnConnect = widget.NewButton("🔌 Kết nối", func() {
		cb.connect()
	})
	cb.btnConnect.Importance = widget.HighImportance

	cb.btnDisconn = widget.NewButton("❌ Ngắt kết nối", func() {
		cb.disconnect()
	})
	cb.btnDisconn.Importance = widget.DangerImportance
	cb.btnDisconn.Disable()

	themeText := "☀️ Sáng"
	if CurrentThemeVariant == theme.VariantLight {
		themeText = "🌙 Tối"
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

func (cb *ConnectionBar) toggleTheme() {
	if cb.app == nil {
		return
	}
	if CurrentThemeVariant == theme.VariantDark {
		cb.app.Settings().SetTheme(NewCustomAppTheme(theme.VariantLight))
		cb.btnTheme.SetText("🌙 Tối")
		cb.state.Log("Đã chuyển sang giao diện Sáng (Light mode).")
	} else {
		cb.app.Settings().SetTheme(NewCustomAppTheme(theme.VariantDark))
		cb.btnTheme.SetText("☀️ Sáng")
		cb.state.Log("Đã chuyển sang giao diện Tối (Dark mode).")
	}
}

func (cb *ConnectionBar) BuildDesktopContainer() fyne.CanvasObject {
	rowInputs := container.NewGridWithColumns(6,
		container.NewVBox(widget.NewLabelWithStyle("Phương thức:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), cb.radioProto),
		container.NewVBox(widget.NewLabelWithStyle("Địa chỉ IP:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), cb.entryHost),
		container.NewVBox(widget.NewLabelWithStyle("Cổng:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), cb.entryPort),
		container.NewVBox(widget.NewLabelWithStyle("Tài khoản:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), cb.entryUser),
		container.NewVBox(widget.NewLabelWithStyle("Mật khẩu:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), cb.entryPass),
		container.NewVBox(widget.NewLabelWithStyle("Thao tác:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), container.NewGridWithColumns(2, cb.btnConnect, cb.btnDisconn)),
	)

	statusCard := container.NewBorder(nil, nil, cb.lblStatus, cb.btnTheme)
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
