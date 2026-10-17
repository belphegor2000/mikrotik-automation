package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
	"mikrotik-automation/internal/modules/cleaner"
)

type TabClear struct {
	state     *AppState
	window    fyne.Window
	container *fyne.Container
	lblTitle  *widget.Label
	lblDesc   *widget.Label
	btnClear  *widget.Button
}

func NewTabClear(state *AppState, w fyne.Window) *TabClear {
	tc := &TabClear{state: state, window: w}

	tc.lblTitle = widget.NewLabelWithStyle(i18n.T("clr_title"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tc.lblDesc = widget.NewLabel(i18n.T("clr_warn_desc"))
	tc.lblDesc.Wrapping = fyne.TextWrapWord

	tc.btnClear = widget.NewButton(i18n.T("clr_btn"), func() {
		tc.confirmClear()
	})
	tc.btnClear.Importance = widget.DangerImportance

	return tc
}

func (tc *TabClear) BuildDesktopContainer() fyne.CanvasObject {
	card := container.NewVBox(
		tc.lblTitle,
		tc.lblDesc,
		widget.NewSeparator(),
		container.NewPadded(tc.btnClear),
	)
	return card
}

func (tc *TabClear) Container() fyne.CanvasObject {
	return tc.BuildDesktopContainer()
}

func (tc *TabClear) UpdateLanguage() {
	if tc.lblTitle != nil {
		tc.lblTitle.SetText(i18n.T("clr_title"))
	}
	if tc.lblDesc != nil {
		tc.lblDesc.SetText(i18n.T("clr_warn_desc"))
	}
	if tc.btnClear != nil {
		tc.btnClear.SetText(i18n.T("clr_btn"))
	}
}

func (tc *TabClear) confirmClear() {
	c := tc.state.GetClient()
	if !c.IsConnected() {
		tc.state.Log("✗ Chưa kết nối đến router!")
		return
	}

	dialog.ShowConfirm(i18n.T("clr_confirm_title"), i18n.T("clr_confirm_msg"), func(ok bool) {
		if !ok {
			return
		}
		tc.btnClear.Disable()
		tc.btnClear.SetText(i18n.T("clr_btn_ing"))
		go func() {
			defer fyne.Do(func() {
				tc.btnClear.Enable()
				tc.btnClear.SetText(i18n.T("clr_btn"))
			})
			cln := cleaner.NewCleaner(c, tc.state.Log)
			cln.CleanAll()
		}()
	}, tc.window)
}
