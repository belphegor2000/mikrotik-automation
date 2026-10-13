package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/modules/security"
)

type TabSecurity struct {
	state        *AppState
	container    *fyne.Container
	checks       map[string]*widget.Check
	btnDisable   *widget.Button
	btnFW        *widget.Button
	btnBruteForce *widget.Button
}

func NewTabSecurity(state *AppState) *TabSecurity {
	ts := &TabSecurity{
		state:  state,
		checks: make(map[string]*widget.Check),
	}

	// Cột 1: Danh sách dịch vụ cần tắt
	services := security.DefaultInsecureServices
	serviceBoxes := container.NewVBox(
		widget.NewLabelWithStyle("1. Tắt Dịch Vụ Không An Toàn:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	for _, sv := range services {
		chk := widget.NewCheck(sv, nil)
		chk.SetChecked(true)
		ts.checks[sv] = chk
		serviceBoxes.Add(chk)
	}

	ts.btnDisable = widget.NewButton("⚠️ Tắt Dịch Vụ Đã Chọn", func() {
		ts.disableServices()
	})

	// Cột 2: Firewall Filter & Chống Brute Force
	descFW := "Thiết lập quy tắc tường lửa và tự động bảo vệ RouterOS:\n" +
		"• Chấp nhận kết nối Established, Related, Untracked\n" +
		"• Chặn kết nối Invalid & chặn trái phép từ WAN\n" +
		"• Tự động chống Brute Force: Quét log mỗi 5 phút\n" +
		"• Thất bại > 5 lần tự động thêm vào Blacklist (30 ngày)."

	ts.btnFW = widget.NewButton("🔒 Áp Dụng Firewall Filter", func() {
		ts.applyFirewall()
	})

	ts.btnBruteForce = widget.NewButton("🛡️ Chống Brute Force (Blacklist)", func() {
		ts.applyBruteForce()
	})

	fwBoxes := container.NewVBox(
		widget.NewLabelWithStyle("2. Firewall Filter & Chống Brute Force:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(descFW),
	)

	// Các nút chức năng nằm ngang hàng tuyệt đối trên một dòng
	buttonsRow := container.NewGridWithColumns(3, ts.btnDisable, ts.btnFW, ts.btnBruteForce)

	ts.container = container.NewVBox(
		widget.NewLabelWithStyle("🔒 Bảo Mật & Tường Lửa RouterOS 7", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, serviceBoxes, fwBoxes),
		buttonsRow,
	)
	return ts
}

func (ts *TabSecurity) Container() fyne.CanvasObject {
	return ts.container
}

func (ts *TabSecurity) disableServices() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	var selected []string
	for name, chk := range ts.checks {
		if chk.Checked {
			selected = append(selected, name)
		}
	}

	ts.btnDisable.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnDisable.Enable() })
		sec := security.NewConfigurator(c, ts.state.Log)
		_, _ = sec.DisableInsecureServices(selected)
	}()
}

func (ts *TabSecurity) applyFirewall() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	ts.btnFW.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnFW.Enable() })
		sec := security.NewConfigurator(c, ts.state.Log)
		_ = sec.ConfigureFirewallFilters(nil)
	}()
}

func (ts *TabSecurity) applyBruteForce() {
	c := ts.state.GetClient()
	if !c.IsConnected() {
		ts.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	ts.btnBruteForce.Disable()
	go func() {
		defer fyne.Do(func() { ts.btnBruteForce.Enable() })
		sec := security.NewConfigurator(c, ts.state.Log)
		_ = sec.ConfigureBruteForceProtection()
	}()
}

