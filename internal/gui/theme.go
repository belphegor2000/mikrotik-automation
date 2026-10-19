package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CurrentThemeVariant tracks the active UI theme variant (Dark/Light).
var CurrentThemeVariant fyne.ThemeVariant = theme.VariantDark

// EnterpriseTheme implements fyne.Theme with custom enterprise-grade styling.
type EnterpriseTheme struct {
	isDark bool
}

// NewEnterpriseTheme creates a new EnterpriseTheme instance.
func NewEnterpriseTheme(isDark bool) *EnterpriseTheme {
	return &EnterpriseTheme{isDark: isDark}
}

// NewCustomAppTheme provides backward compatibility with main.go.
func NewCustomAppTheme(variant fyne.ThemeVariant) fyne.Theme {
	CurrentThemeVariant = variant
	return NewEnterpriseTheme(variant == theme.VariantDark)
}

func (t *EnterpriseTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Respect internal theme dark/light setting if explicit, otherwise use variant
	dark := t.isDark
	if dark {
		switch name {
		case theme.ColorNameBackground:
			return color.RGBA{R: 11, G: 17, B: 32, A: 255} // #0B1120 Deep Slate
		case theme.ColorNameMenuBackground, theme.ColorNameHeaderBackground, theme.ColorNameOverlayBackground:
			return color.RGBA{R: 30, G: 41, B: 59, A: 255} // #1E293B Surface Card
		case theme.ColorNameInputBackground:
			return color.RGBA{R: 15, G: 23, B: 42, A: 255} // #0F172A Input field
		case theme.ColorNameButton:
			return color.RGBA{R: 30, G: 41, B: 59, A: 255} // #1E293B
		case theme.ColorNamePrimary:
			return color.RGBA{R: 56, G: 189, B: 248, A: 255} // #38BDF8 Cyber Cyan
		case theme.ColorNameForeground:
			return color.RGBA{R: 241, G: 245, B: 249, A: 255} // #F1F5F9 Crisp text
		case theme.ColorNamePlaceHolder:
			return color.RGBA{R: 100, G: 116, B: 139, A: 255} // #64748B Slate muted
		case theme.ColorNameHover:
			return color.RGBA{R: 255, G: 255, B: 255, A: 25} // Semi-transparent subtle white highlight
		case theme.ColorNameFocus:
			return color.RGBA{R: 56, G: 189, B: 248, A: 255} // #38BDF8
		case theme.ColorNameSeparator:
			return color.RGBA{R: 51, G: 65, B: 85, A: 255} // #334155
		case theme.ColorNameDisabled:
			return color.RGBA{R: 71, G: 85, B: 105, A: 255} // #475569
		case theme.ColorNameDisabledButton:
			return color.RGBA{R: 20, G: 30, B: 45, A: 255}
		case theme.ColorNameSelection:
			return color.RGBA{R: 3, G: 105, B: 161, A: 160} // #0369A1
		case theme.ColorNameError:
			return color.RGBA{R: 244, G: 63, B: 94, A: 255} // #F43F5E Neon Rose
		case theme.ColorNameSuccess:
			return color.RGBA{R: 16, G: 185, B: 129, A: 255} // #10B981 Emerald
		case theme.ColorNameWarning:
			return color.RGBA{R: 245, G: 158, B: 11, A: 255} // #F59E0B Amber
		case theme.ColorNameForegroundOnError:
			return color.RGBA{R: 255, G: 255, B: 255, A: 255}
		case theme.ColorNameShadow:
			return color.RGBA{R: 0, G: 0, B: 0, A: 120}
		}
	} else {
		// Light Mode
		switch name {
		case theme.ColorNameBackground:
			return color.RGBA{R: 248, G: 250, B: 252, A: 255} // #F8FAFC Pure Slate
		case theme.ColorNameMenuBackground, theme.ColorNameHeaderBackground, theme.ColorNameOverlayBackground:
			return color.RGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF Pure White
		case theme.ColorNameInputBackground:
			return color.RGBA{R: 255, G: 255, B: 255, A: 255}
		case theme.ColorNameButton:
			return color.RGBA{R: 241, G: 245, B: 249, A: 255} // #F1F5F9
		case theme.ColorNamePrimary:
			return color.RGBA{R: 2, G: 132, B: 199, A: 255} // #0284C7 Cobalt Blue
		case theme.ColorNameForeground:
			return color.RGBA{R: 15, G: 23, B: 42, A: 255} // #0F172A Slate Dark
		case theme.ColorNamePlaceHolder:
			return color.RGBA{R: 148, G: 163, B: 184, A: 255} // #94A3B8
		case theme.ColorNameHover:
			return color.RGBA{R: 15, G: 23, B: 42, A: 20} // Semi-transparent subtle dark tint
		case theme.ColorNameFocus:
			return color.RGBA{R: 2, G: 132, B: 199, A: 255}
		case theme.ColorNameSeparator:
			return color.RGBA{R: 226, G: 232, B: 240, A: 255}
		case theme.ColorNameDisabled:
			return color.RGBA{R: 203, G: 213, B: 225, A: 255}
		case theme.ColorNameDisabledButton:
			return color.RGBA{R: 241, G: 245, B: 249, A: 255}
		case theme.ColorNameSelection:
			return color.RGBA{R: 186, G: 230, B: 253, A: 160}
		case theme.ColorNameError:
			return color.RGBA{R: 225, G: 29, B: 72, A: 255} // #E11D48
		case theme.ColorNameSuccess:
			return color.RGBA{R: 5, G: 150, B: 105, A: 255} // #059669
		case theme.ColorNameWarning:
			return color.RGBA{R: 217, G: 119, B: 6, A: 255} // #D97706
		case theme.ColorNameForegroundOnError:
			return color.RGBA{R: 255, G: 255, B: 255, A: 255}
		case theme.ColorNameShadow:
			return color.RGBA{R: 0, G: 0, B: 0, A: 40}
		}
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (t *EnterpriseTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *EnterpriseTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *EnterpriseTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputBorder:
		return 1.0
	case theme.SizeNameInputRadius:
		return 8.0
	case theme.SizeNameSelectionRadius:
		return 8.0
	case theme.SizeNamePadding:
		return 6.0
	case theme.SizeNameText:
		return 13.0
	case theme.SizeNameHeadingText:
		return 18.0
	case theme.SizeNameSubHeadingText:
		return 15.0
	case theme.SizeNameCaptionText:
		return 11.0
	default:
		return theme.DefaultTheme().Size(name)
	}
}
