package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	ColorPrimary     = color.NRGBA{R: 0x1a, G: 0x73, B: 0xe8, A: 0xff} // Google Blue #1a73e8
	ColorPrimaryDark = color.NRGBA{R: 0x15, G: 0x57, B: 0xb0, A: 0xff}
	ColorDestructive = color.NRGBA{R: 0xdc, G: 0x35, B: 0x45, A: 0xff} // Red #dc3545
	ColorSuccess     = color.NRGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff} // Green #10b981

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
	return theme.DefaultTheme().Size(name)
}
