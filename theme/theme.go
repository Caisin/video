package theme

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

//go:embed SourceHanSansK-Bold.ttf
var font []byte

//go:embed caisin.png
var icon []byte
var _ fyne.Theme = (*CaisinTheme)(nil)
var dft = theme.DefaultTheme()

type CaisinTheme struct {
	fyne.Theme
}

func (t CaisinTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return dft.Color(n, v)
}

func (t CaisinTheme) Font(fyne.TextStyle) fyne.Resource {
	return fyne.NewStaticResource("SourceHanSansK-Bold.ttf", font)
}

func GetLogo() fyne.Resource {
	return fyne.NewStaticResource("logo.png", icon)
}

func (t CaisinTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if name == "logo" {
		return fyne.NewStaticResource("logo.png", icon)
	}
	return dft.Icon(name)
}

func (t CaisinTheme) Size(s fyne.ThemeSizeName) float32 {
	return dft.Size(s)

}
