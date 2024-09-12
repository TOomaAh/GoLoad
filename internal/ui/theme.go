package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 33, G: 33, B: 33, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0, G: 153, B: 204, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0, G: 178, B: 238, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 179, G: 179, B: 179, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0, G: 128, B: 170, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 102, G: 102, B: 102, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 153, B: 204, A: 255}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0, G: 153, B: 204, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 3
	default:
		return theme.DefaultTheme().Size(name)
	}
}
