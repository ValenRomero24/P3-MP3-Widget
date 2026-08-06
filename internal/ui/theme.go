package ui

import (
	_"embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Embeber fuentes desde assets
//go:embed "assets/FOT-Skip Std B.otf"
var mainFontData []byte

//go:embed assets/BMSPA___.TTF
var logoFontData []byte

var ResourceMainFont = fyne.NewStaticResource("FOT-Skip Std B.otf", mainFontData)
var ResourceLogoFont = fyne.NewStaticResource("BMSPA___", logoFontData)

type P3Theme struct{}

var _ fyne.Theme = (*P3Theme)(nil)

func (m *P3Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color{
	switch name {
	// Fondo principal de la ventana (Negro / Azul noche profundo)
	case theme.ColorNameBackground:
		return color.RGBA{R: 7, G: 9, B: 14, A: 255} // #07090E (Negro/Azul noche profundo)
	// Fondos de contenedores modales
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return color.RGBA{R: 15, G: 22, B: 36, A: 255} // #0F1624
	// Botones y tarjetas
	case theme.ColorNameButton:
		return color.RGBA{R: 20, G: 32, B: 54, A: 255} // #142036
	// Color primario / Resaltados / Foco (Azul eléctrico P3)
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameSelection:
		return color.RGBA{R: 0, G: 163, B: 255, A: 255} // #00A3FF
	// Texto Principal (Blanco cian / brillante)
	case theme.ColorNameForeground:
		return color.RGBA{R: 224, G: 244, B: 255, A: 255} // #E0F4FF
	// Sliders y barras	
	case theme.ColorNameScrollBar, theme.ColorNameInputBackground:
		return color.RGBA{R: 12, G: 20, B: 34, A: 255} 
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *P3Theme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		return ResourceLogoFont
	}

	return ResourceMainFont
}

func (m *P3Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *P3Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding: 
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameText:
		return 14
	}
	return theme.DefaultTheme().Size(name)
}