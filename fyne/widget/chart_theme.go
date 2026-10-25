package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Custom color names. A Fyne theme passed to Settings().SetTheme can return
// colours for these names to style the charts; any name it does not handle
// falls back to a standard Fyne colour (see ChartThemeFromApp).
const (
	ColorNameChartBackground fyne.ThemeColorName = "chartBackground"
	ColorNameChartPlot       fyne.ThemeColorName = "chartPlot"
	ColorNameChartBar        fyne.ThemeColorName = "chartBar"
	ColorNameChartFill       fyne.ThemeColorName = "chartFill"
	ColorNameChartBarHover   fyne.ThemeColorName = "chartBarHover"
	ColorNameChartCursor     fyne.ThemeColorName = "chartCursor"
	ColorNameChartLabel      fyne.ThemeColorName = "chartLabel"
	ColorNameChartValue      fyne.ThemeColorName = "chartValue"
	ColorNameChartAxis       fyne.ThemeColorName = "chartAxis"
	ColorNameChartGrid       fyne.ThemeColorName = "chartGrid"
)

// ChartThemeFromApp builds a ChartTheme from the theme set with
// fyne.CurrentApp().Settings().SetTheme, for the current light/dark variant.
//
// Each chart color is looked up by its custom name first (ColorNameChart...).
// A theme that does not define that name (Fyne returns a fully transparent
// color for unknown names) set to a standard Fyne color:
//
//	Background  <- background          Label  <- foreground
//	Plot        <- input background    Value  <- placeholder
//	Bar         <- primary             Axis   <- foreground
//	Fill        <- primary, ~35% alpha Grid   <- separator
//	BarHover    <- warning             Cursor <- foreground, ~55% alpha
func ChartThemeFromApp() ChartTheme {
	s := fyne.CurrentApp().Settings()
	th, v := s.Theme(), s.ThemeVariant()

	std := func(n fyne.ThemeColorName) color.Color { return th.Color(n, v) }
	pick := func(custom fyne.ThemeColorName, fallback color.Color) color.Color {
		if c := th.Color(custom, v); !isTransparent(c) {
			return c
		}
		return fallback
	}

	fg, primary := std(theme.ColorNameForeground), std(theme.ColorNamePrimary)
	return ChartTheme{
		Background:     pick(ColorNameChartBackground, std(theme.ColorNameBackground)),
		PlotBackground: pick(ColorNameChartPlot, std(theme.ColorNameInputBackground)),
		Bar:            pick(ColorNameChartBar, primary),
		Fill:           pick(ColorNameChartFill, withAlpha(primary, 0x59)),
		BarHover:       pick(ColorNameChartBarHover, std(theme.ColorNameWarning)),
		Cursor:         pick(ColorNameChartCursor, withAlpha(fg, 0x90)),
		Label:          pick(ColorNameChartLabel, fg),
		Value:          pick(ColorNameChartValue, std(theme.ColorNamePlaceHolder)),
		Axis:           pick(ColorNameChartAxis, fg),
		Grid:           pick(ColorNameChartGrid, std(theme.ColorNameSeparator)),
	}
}

func isTransparent(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a == 0
}

func withAlpha(c color.Color, a uint8) color.Color {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	n.A = a
	return n
}

// ChartTheme holds every color the chart uses.
type ChartTheme struct {
	Background     color.Color // whole widget, including the header
	PlotBackground color.Color // area behind the bars
	Bar            color.Color // bar colour / line colour
	Fill           color.Color // area under the line when Fill is ticked
	BarHover       color.Color // bar under the mouse cursor
	Cursor         color.Color // vertical hover line
	Label          color.Color // the title text, e.g. "HPP"
	Value          color.Color // the "345: 26" hover text
	Axis           color.Color
	Grid           color.Color
}

// DarkTheme matches the look of the reference screenshot. Select it with "base": "dark".
func DarkTheme() ChartTheme {
	return ChartTheme{
		Background:     color.NRGBA{0x2b, 0x2d, 0x30, 0xff},
		PlotBackground: color.NRGBA{0x55, 0x58, 0x5c, 0xff},
		Bar:            color.NRGBA{0xe6, 0xe6, 0xe6, 0xff},
		Fill:           color.NRGBA{0xe6, 0xe6, 0xe6, 0x59},
		BarHover:       color.NRGBA{0xff, 0xc1, 0x07, 0xff},
		Cursor:         color.NRGBA{0xff, 0xff, 0xff, 0x90},
		Label:          color.NRGBA{0xff, 0xff, 0xff, 0xff},
		Value:          color.NRGBA{0x9a, 0x9d, 0xa1, 0xff},
		Axis:           color.NRGBA{0xcc, 0xcc, 0xcc, 0xff},
		Grid:           color.NRGBA{0xff, 0xff, 0xff, 0x28},
	}
}

// LightTheme is a white background with matplotlib-blue bars. Select it with "base": "light".
func LightTheme() ChartTheme {
	return ChartTheme{
		Background:     color.NRGBA{0xff, 0xff, 0xff, 0xff},
		PlotBackground: color.NRGBA{0xff, 0xff, 0xff, 0xff},
		Bar:            color.NRGBA{0x1f, 0x77, 0xb4, 0xff},
		Fill:           color.NRGBA{0x1f, 0x77, 0xb4, 0x59},
		BarHover:       color.NRGBA{0xff, 0x7f, 0x0e, 0xff},
		Cursor:         color.NRGBA{0x00, 0x00, 0x00, 0x80},
		Label:          color.NRGBA{0x20, 0x20, 0x20, 0xff},
		Value:          color.NRGBA{0x70, 0x70, 0x70, 0xff},
		Axis:           color.NRGBA{0x00, 0x00, 0x00, 0xff},
		Grid:           color.NRGBA{0x00, 0x00, 0x00, 0x20},
	}
}

// DefaultTheme is used when no theme is set: the original light look.
func DefaultTheme() ChartTheme { return LightTheme() }
