package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mikrotik-automation/internal/i18n"
)

func cleanTitleWithIcon(rawTitle, icon string) string {
	s := strings.TrimSpace(rawTitle)
	s = strings.TrimPrefix(s, icon)
	s = strings.TrimSpace(s)
	return fmt.Sprintf("%s  %s", icon, s)
}

type navDef struct {
	key  string
	icon string
}

var navItems = []navDef{
	{key: "tab_dashboard", icon: "📊"},
	{key: "tab_network", icon: "🌐"},
	{key: "tab_queue", icon: "⚡"},
	{key: "tab_security", icon: "🔒"},
	{key: "tab_ipsec", icon: "🔐"},
	{key: "tab_backup", icon: "💾"},
	{key: "tab_clear", icon: "🗑️"},
}

// Sidebar provides collapsible navigation for the application shell.
type Sidebar struct {
	app         fyne.App
	state       *AppState
	isCollapsed bool
	activeIndex int
	buttons     []*widget.Button
	btnToggle   *widget.Button
	lblFooter   *widget.Label
	itemsBox    *fyne.Container
	container   *fyne.Container
	OnSelect    func(index int)
}

func NewSidebar(state *AppState, app fyne.App, onSelect func(int)) *Sidebar {
	sb := &Sidebar{
		app:         app,
		state:       state,
		activeIndex: 0,
		OnSelect:    onSelect,
	}

	// Restore collapsed preference if saved
	if app != nil {
		sb.isCollapsed = app.Preferences().BoolWithFallback("sidebar_collapsed", false)
	}

	sb.buttons = make([]*widget.Button, len(navItems))
	for i, def := range navItems {
		idx := i
		title := cleanTitleWithIcon(i18n.T(def.key), def.icon)
		btnText := title
		if sb.isCollapsed {
			btnText = def.icon
		}
		btn := widget.NewButton(btnText, func() {
			sb.SetActive(idx)
			if sb.OnSelect != nil {
				sb.OnSelect(idx)
			}
		})
		btn.Alignment = widget.ButtonAlignLeading
		if i == 0 {
			btn.Importance = widget.HighImportance
		}
		sb.buttons[i] = btn
	}

	toggleText := "◀  Thu Gọn"
	if sb.isCollapsed {
		toggleText = "▶"
	}
	sb.btnToggle = widget.NewButton(toggleText, func() {
		sb.Toggle()
	})

	sb.lblFooter = widget.NewLabelWithStyle("RouterOS v7", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	sb.buildUI()

	// Update footer on router connect
	sb.state.OnConnect = append(sb.state.OnConnect, func() {
		c := sb.state.GetClient()
		if c != nil && c.IsConnected() {
			sb.lblFooter.SetText("🟢 RouterOS v7")
		} else {
			sb.lblFooter.SetText("⚪ Offline")
		}
	})

	return sb
}

func (sb *Sidebar) buildUI() {
	itemObjects := make([]fyne.CanvasObject, len(sb.buttons))
	for i, b := range sb.buttons {
		itemObjects[i] = b
	}
	sb.itemsBox = container.NewVBox(itemObjects...)

	footerBox := container.NewVBox(
		widget.NewSeparator(),
		sb.btnToggle,
		sb.lblFooter,
	)

	w := float32(220)
	if sb.isCollapsed {
		w = float32(64)
	}

	sb.container = container.NewBorder(
		nil,
		footerBox,
		nil,
		nil,
		container.NewVScroll(sb.itemsBox),
	)
	sb.container.Resize(fyne.NewSize(w, sb.container.MinSize().Height))
}

func (sb *Sidebar) Container() fyne.CanvasObject {
	if sb.container == nil {
		sb.buildUI()
	}
	return sb.container
}

func (sb *Sidebar) SetActive(index int) {
	if index < 0 || index >= len(sb.buttons) {
		return
	}
	sb.activeIndex = index
	for i, btn := range sb.buttons {
		if i == index {
			btn.Importance = widget.HighImportance
		} else {
			btn.Importance = widget.MediumImportance
		}
		btn.Refresh()
	}
}

func (sb *Sidebar) Toggle() {
	sb.isCollapsed = !sb.isCollapsed
	if sb.app != nil {
		sb.app.Preferences().SetBool("sidebar_collapsed", sb.isCollapsed)
	}
	sb.Refresh()
}

func (sb *Sidebar) IsCollapsed() bool {
	return sb.isCollapsed
}

func (sb *Sidebar) Refresh() {
	for i, def := range navItems {
		title := cleanTitleWithIcon(i18n.T(def.key), def.icon)
		if sb.isCollapsed {
			sb.buttons[i].SetText(def.icon)
			sb.buttons[i].Alignment = widget.ButtonAlignCenter
		} else {
			sb.buttons[i].SetText(title)
			sb.buttons[i].Alignment = widget.ButtonAlignLeading
		}
	}

	if sb.isCollapsed {
		sb.btnToggle.SetText("▶")
		sb.lblFooter.SetText("v7")
	} else {
		sb.btnToggle.SetText("◀  Thu Gọn")
		c := sb.state.GetClient()
		if c != nil && c.IsConnected() {
			sb.lblFooter.SetText("🟢 RouterOS v7")
		} else {
			sb.lblFooter.SetText("RouterOS v7")
		}
	}

	if sb.container != nil {
		w := float32(220)
		if sb.isCollapsed {
			w = float32(64)
		}
		sb.container.Resize(fyne.NewSize(w, sb.container.Size().Height))
		sb.container.Refresh()
	}
}

func (sb *Sidebar) UpdateLanguage() {
	sb.Refresh()
}
