package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	ColorPrimary     = color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff} // Tailwind Blue-600 #2563eb
	ColorPrimaryDark = color.NRGBA{R: 0x1d, G: 0x4e, B: 0xd8, A: 0xff} // Blue-700 #1d4ed8
	ColorDestructive = color.NRGBA{R: 0xdc, G: 0x26, B: 0x26, A: 0xff} // Red-600 #dc2626
	ColorSuccess     = color.NRGBA{R: 0x16, G: 0xa3, B: 0x4a, A: 0xff} // Green-600 #16a34a
	ColorWarning     = color.NRGBA{R: 0xd9, G: 0x77, B: 0x06, A: 0xff} // Amber-600 #d97706

	// CurrentThemeVariant tracks active theme (Dark by default)
	CurrentThemeVariant fyne.ThemeVariant = theme.VariantDark
)

// CustomAppTheme implements fyne.Theme with custom Blue and Red colors and dynamic Variant.
type CustomAppTheme struct {
	Variant fyne.ThemeVariant
}

func NewCustomAppTheme(v fyne.ThemeVariant) *CustomAppTheme {
	CurrentThemeVariant = v
	return &CustomAppTheme{Variant: v}
}

func (t *CustomAppTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return ColorPrimary
	case theme.ColorNameError:
		return ColorDestructive
	case theme.ColorNameSuccess:
		return ColorSuccess
	case theme.ColorNameWarning:
		return ColorWarning
	default:
		return theme.DefaultTheme().Color(name, t.Variant)
	}
}

func (t *CustomAppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *CustomAppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *CustomAppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 8.0 // Bo góc 8px hiện đại theo chuẩn Fluent/Material 3
	case theme.SizeNamePadding:
		return 6.0
	default:
		return theme.DefaultTheme().Size(name)
	}
}
