package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"whispering-tiger-ui/Resources"
)

var (
	orange            = &color.NRGBA{R: 198, G: 123, B: 0, A: 255}
	orangeTransparent = &color.NRGBA{R: 198, G: 123, B: 0, A: 180}
)

type AppTheme struct{}

var _ fyne.Theme = (*AppTheme)(nil)

func (m AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return orange
	case theme.ColorNameScrollBar:
		return orangeTransparent
	}

	variant = theme.VariantDark

	return theme.DefaultTheme().Color(name, variant)
}
func (m AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return Resources.ResourceGoNotoTtf
}

func (m AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
