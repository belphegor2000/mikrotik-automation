package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/api"
	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/profile"
)


type ConnectionBar struct {
	state          *AppState
	app            fyne.App
	profileMgr     *profile.ProfileManager
	lblProfile     *widget.Label
	selectProfile  *widget.Select
	btnSaveProfile *widget.Button
	btnDelProfile  *widget.Button
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
	OnStatusChanged func(status string, isConnected bool)
	activeDialog   dialog.Dialog
}


func NewConnectionBar(state *AppState, a fyne.App) *ConnectionBar {
	cb := &ConnectionBar{state: state, app: a}

	cb.lblProtoTitle = widget.NewLabelWithStyle(i18n.T("conn_proto"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblHostTitle = widget.NewLabelWithStyle(i18n.T("conn_ip"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblPortTitle = widget.NewLabelWithStyle(i18n.T("conn_port"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblUserTitle = widget.NewLabelWithStyle(i18n.T("conn_user"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.lblPassTitle = widget.NewLabelWithStyle(i18n.T("conn_pass"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Profile / Address Book
	cb.profileMgr, _ = profile.NewProfileManager()
	cb.lblProfile = widget.NewLabelWithStyle(i18n.T("conn_profile_select"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	cb.selectProfile = widget.NewSelect([]string{i18n.T("conn_profile_prompt")}, func(selected string) {
		if selected == "" || selected == i18n.T("conn_profile_prompt") {
			return
		}
		if cb.profileMgr != nil {
			if p, ok := cb.profileMgr.FindProfile(selected); ok {
				cb.entryHost.SetText(p.Host)
				cb.entryPort.SetText(fmt.Sprintf("%d", p.Port))
				cb.entryUser.SetText(p.User)
				cb.entryPass.SetText(p.Password)
				cb.selectProto.SetSelected(p.Proto)
				cb.state.Log(fmt.Sprintf("✓ Đã nạp thông tin Router: %s (%s:%d)", p.Name, p.Host, p.Port))
			}
		}
	})
	cb.selectProfile.SetSelected(i18n.T("conn_profile_prompt"))

	cb.btnSaveProfile = widget.NewButton(i18n.T("conn_profile_save"), func() {
		cb.promptSaveProfile()
	})
	cb.btnDelProfile = widget.NewButton(i18n.T("conn_profile_delete"), func() {
		cb.deleteCurrentProfile()
	})

	cb.reloadProfileOptions()

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
	if cb.OnStatusChanged != nil {
		c := cb.state.GetClient()
		isConnected := c != nil && c.IsConnected()
		cb.OnStatusChanged(text, isConnected)
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
	if cb.lblProfile != nil {
		cb.lblProfile.SetText(i18n.T("conn_profile_select"))
	}
	if cb.btnSaveProfile != nil {
		cb.btnSaveProfile.SetText(i18n.T("conn_profile_save"))
	}
	if cb.btnDelProfile != nil {
		cb.btnDelProfile.SetText(i18n.T("conn_profile_delete"))
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

func (cb *ConnectionBar) reloadProfileOptions() {
	if cb.profileMgr == nil || cb.selectProfile == nil {
		return
	}
	profs := cb.profileMgr.ListProfiles()
	opts := []string{i18n.T("conn_profile_prompt")}
	for _, p := range profs {
		opts = append(opts, p.Name)
	}
	cb.selectProfile.Options = opts
	cb.selectProfile.Refresh()
}

func (cb *ConnectionBar) promptSaveProfile() {
	host := strings.TrimSpace(cb.entryHost.Text)
	if host == "" {
		cb.state.Log("⚠ IP Router không được để trống khi lưu profile.")
		return
	}
	var parentWin fyne.Window
	if cb.app != nil && len(cb.app.Driver().AllWindows()) > 0 {
		parentWin = cb.app.Driver().AllWindows()[0]
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("vd: Chi nhánh Quận 1, Router Kho, Văn phòng...")
	nameEntry.SetText("Router " + host)

	d := dialog.NewCustomConfirm(
		"Lưu Vào Danh Bạ Router",
		"Lưu",
		"Huỷ",
		container.NewVBox(widget.NewLabel("Nhập tên gợi nhớ cho Router này:"), nameEntry),
		func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(nameEntry.Text)
			if name == "" {
				name = "Router " + host
			}
			port, _ := strconv.Atoi(cb.entryPort.Text)
			if port <= 0 {
				if cb.selectProto.Selected == "SSH" {
					port = 22
				} else {
					port = 80
				}
			}
			prof := profile.DeviceProfile{
				Name:     name,
				Host:     host,
				Port:     port,
				User:     cb.entryUser.Text,
				Password: cb.entryPass.Text,
				Proto:    cb.selectProto.Selected,
			}
			if err := cb.profileMgr.SaveProfile(prof); err != nil {
				cb.state.Log(fmt.Sprintf("⚠ Lỗi lưu profile: %v", err))
			} else {
				cb.state.Log(fmt.Sprintf("✓ Đã lưu '%s' (%s) vào danh bạ an toàn.", name, host))
				cb.reloadProfileOptions()
				cb.selectProfile.SetSelected(name)
			}
		},
		parentWin,
	)
	if parentWin != nil {
		d.Show()
	}
}

func (cb *ConnectionBar) deleteCurrentProfile() {
	sel := cb.selectProfile.Selected
	if sel == "" || sel == i18n.T("conn_profile_prompt") {
		return
	}
	if err := cb.profileMgr.DeleteProfile(sel); err == nil {
		cb.state.Log(fmt.Sprintf("✓ Đã xoá profile '%s' khỏi danh bạ.", sel))
		cb.reloadProfileOptions()
		cb.selectProfile.SetSelected(i18n.T("conn_profile_prompt"))
	}
}

func (cb *ConnectionBar) toggleTheme() {
	if cb.app == nil {
		return
	}
	if CurrentThemeVariant == theme.VariantDark {
		CurrentThemeVariant = theme.VariantLight
		cb.app.Settings().SetTheme(NewEnterpriseTheme(false))
		cb.btnTheme.SetText(i18n.T("theme_dark"))
		cb.state.Log("Đã chuyển sang giao diện Sáng (Light mode).")
	} else {
		CurrentThemeVariant = theme.VariantDark
		cb.app.Settings().SetTheme(NewEnterpriseTheme(true))
		cb.btnTheme.SetText(i18n.T("theme_light"))
		cb.state.Log("Đã chuyển sang giao diện Tối (Dark mode).")
	}
}

// ShowDialog displays the connection modal over the given parent window.
func (cb *ConnectionBar) ShowDialog(parent fyne.Window) {
	if parent == nil {
		return
	}

	profileBar := container.NewBorder(
		nil, nil,
		cb.lblProfile,
		container.NewHBox(cb.btnSaveProfile, cb.btnDelProfile),
		cb.selectProfile,
	)

	rowInputs1 := container.NewGridWithColumns(3,
		container.NewVBox(cb.lblProtoTitle, cb.selectProto),
		container.NewVBox(cb.lblHostTitle, cb.entryHost),
		container.NewVBox(cb.lblPortTitle, cb.entryPort),
	)
	rowInputs2 := container.NewGridWithColumns(2,
		container.NewVBox(cb.lblUserTitle, cb.entryUser),
		container.NewVBox(cb.lblPassTitle, cb.entryPass),
	)

	actionBox := container.NewHBox(cb.btnConnect, cb.btnDisconn)

	content := container.NewVBox(
		widget.NewLabelWithStyle(i18n.T("conn_modal_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewPadded(profileBar),
		widget.NewSeparator(),
		rowInputs1,
		rowInputs2,
		widget.NewSeparator(),
		container.NewBorder(nil, nil, cb.lblStatus, actionBox),
	)

	cb.activeDialog = dialog.NewCustom(i18n.T("conn_modal_title"), i18n.T("conn_modal_close"), container.NewPadded(content), parent)
	cb.activeDialog.Resize(fyne.NewSize(540, 360))
	cb.activeDialog.Show()
}

func (cb *ConnectionBar) BuildDesktopContainer() fyne.CanvasObject {
	profileBar := container.NewBorder(
		nil, nil,
		cb.lblProfile,
		container.NewHBox(cb.btnSaveProfile, cb.btnDelProfile),
		cb.selectProfile,
	)

	rowInputs := container.NewGridWithColumns(5,
		container.NewVBox(cb.lblProtoTitle, cb.selectProto),
		container.NewVBox(cb.lblHostTitle, cb.entryHost),
		container.NewVBox(cb.lblPortTitle, cb.entryPort),
		container.NewVBox(cb.lblUserTitle, cb.entryUser),
		container.NewVBox(cb.lblPassTitle, cb.entryPass),
	)

	actionControls := container.NewHBox(cb.btnConnect, cb.btnDisconn, cb.btnLang, cb.btnTheme)
	statusCard := container.NewBorder(nil, nil, cb.lblStatus, actionControls)
	cb.container = container.NewVBox(
		container.NewPadded(profileBar),
		container.NewPadded(rowInputs),
		widget.NewSeparator(),
		statusCard,
	)
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

				if cb.activeDialog != nil {
					cb.activeDialog.Hide()
				}

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

