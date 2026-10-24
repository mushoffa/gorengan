package widget

import (
	"image/color"
)

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
