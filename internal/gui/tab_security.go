package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/modules/security"
)

type TabSecurity struct {
	state         *AppState
	container     *fyne.Container
	lblMainTitle  *widget.Label
	lblSrvTitle   *widget.Label
	lblFWTitle    *widget.Label
	lblFWDesc     *widget.Label
	checks        map[string]*widget.Check
	btnDisable    *widget.Button
	btnFW         *widget.Button
	btnBruteForce *widget.Button
	serviceBoxes  *fyne.Container
	fwBoxes       *fyne.Container
}

func NewTabSecurity(state *AppState) *TabSecurity {
	ts := &TabSecurity{
		state:  state,
		checks: make(map[string]*widget.Check),
	}

	ts.lblMainTitle = widget.NewLabelWithStyle(i18n.T("sec_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Cột 1: Danh sách dịch vụ cần tắt
	services := security.DefaultInsecureServices
	ts.lblSrvTitle = widget.NewLabelWithStyle(i18n.T("sec_srv_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	serviceBoxes := container.NewVBox(
		ts.lblSrvTitle,
	)

	for _, sv := range services {
		chk := widget.NewCheck(sv, nil)
		chk.SetChecked(true)
		ts.checks[sv] = chk
		serviceBoxes.Add(chk)
	}

	ts.btnDisable = widget.NewButton(i18n.T("sec_srv_btn"), func() {
		ts.disableServices()
	})
	ts.btnDisable.Importance = widget.WarningImportance

	// Cột 2: Firewall Filter & Chống Brute Force
	ts.btnFW = widget.NewButton(i18n.T("sec_fw_btn"), func() {
		ts.applyFirewall()
	})
	ts.btnFW.Importance = widget.HighImportance

	ts.btnBruteForce = widget.NewButton(i18n.T("sec_bf_btn"), func() {
		ts.applyBruteForce()
	})
	ts.btnBruteForce.Importance = widget.HighImportance

	ts.lblFWDesc = widget.NewLabel(i18n.T("sec_fw_desc"))
	ts.lblFWDesc.Wrapping = fyne.TextWrapWord

	ts.lblFWTitle = widget.NewLabelWithStyle(i18n.T("sec_fw_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	ts.fwBoxes = container.NewVBox(
		ts.lblFWTitle,
		ts.lblFWDesc,
	)
	ts.serviceBoxes = serviceBoxes

	return ts
}

func (ts *TabSecurity) BuildDesktopContainer() fyne.CanvasObject {
	buttonsRow := container.NewGridWithColumns(3, ts.btnDisable, ts.btnFW, ts.btnBruteForce)
	return container.NewVBox(
		ts.lblMainTitle,
		container.NewGridWithColumns(2, ts.serviceBoxes, ts.fwBoxes),
		buttonsRow,
	)
}

func (ts *TabSecurity) UpdateLanguage() {
	if ts.lblMainTitle != nil {
		ts.lblMainTitle.SetText(i18n.T("sec_title"))
	}
	if ts.lblSrvTitle != nil {
		ts.lblSrvTitle.SetText(i18n.T("sec_srv_title"))
	}
	if ts.btnDisable != nil {
		ts.btnDisable.SetText(i18n.T("sec_srv_btn"))
	}
	if ts.lblFWTitle != nil {
		ts.lblFWTitle.SetText(i18n.T("sec_fw_title"))
	}
	if ts.lblFWDesc != nil {
		ts.lblFWDesc.SetText(i18n.T("sec_fw_desc"))
	}
	if ts.btnFW != nil {
		ts.btnFW.SetText(i18n.T("sec_fw_btn"))
	}
	if ts.btnBruteForce != nil {
		ts.btnBruteForce.SetText(i18n.T("sec_bf_btn"))
	}
}

func (ts *TabSecurity) Container() fyne.CanvasObject {
	return ts.BuildDesktopContainer()
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
