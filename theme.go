package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

var (
	orange                     = &color.NRGBA{R: 198, G: 123, B: 0, A: 255}
	orangeTransparent          = &color.NRGBA{R: 198, G: 123, B: 0, A: 180}
	orangeTransparentSelection = &color.NRGBA{R: 198, G: 123, B: 0, A: 100}
	colorDarkHover             = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f}
	colorDarkDisabled          = color.NRGBA{R: 80, G: 80, B: 81, A: 255}
	colorDarkInputBackground   = color.NRGBA{R: 0x25, G: 0x25, B: 0x28, A: 0xff}
)

type AppTheme struct{}

var _ fyne.Theme = (*AppTheme)(nil)

func (m AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return orange
	case theme.ColorNameScrollBar:
		return orangeTransparent
	case theme.ColorNameHover, theme.ColorNameFocus:
		return colorDarkHover
	case theme.ColorNameSelection:
		return orangeTransparentSelection
	case theme.ColorNameDisabled:
		return colorDarkDisabled
	case theme.ColorNameInputBackground:
		return colorDarkInputBackground
	}

	variant = theme.VariantDark

	return theme.DefaultTheme().Color(name, variant)
}
func (m AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	return theme.DefaultTheme().Font(style)
	//return Resources.ResourceGoNotoKurrentRegularTtf
}

func (m AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
