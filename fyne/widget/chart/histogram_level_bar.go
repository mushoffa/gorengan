package chart

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	markerNone = iota
	markerLow
	markerHigh
)

// levelBar is the black-to-white bar under a histogram with two triangle
// markers. Dragging (or tapping near) a marker moves it; together they select
// a range of levels.
type levelBar struct {
	widget.BaseWidget
	raster *canvas.Raster
	lo, hi int
	active int // the marker being dragged

	OnChange func(lo, hi int)
}

func newLevelBar() *levelBar {
	b := &levelBar{lo: 0, hi: 255}
	b.raster = canvas.NewRaster(b.draw)
	b.raster.ScaleMode = canvas.ImageScalePixels
	b.raster.SetMinSize(fyne.NewSize(180, histBarHeight+histMarkerHeight))
	b.ExtendBaseWidget(b)
	return b
}

func (b *levelBar) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(b.raster) }

func (b *levelBar) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (b *levelBar) Tapped(ev *fyne.PointEvent) {
	level := b.levelAt(ev.Position.X)
	b.set(b.nearest(level), level)
}

func (b *levelBar) Dragged(ev *fyne.DragEvent) {
	level := b.levelAt(ev.Position.X)
	if b.active == markerNone {
		b.active = b.nearest(level)
	}
	b.set(b.active, level)
}

func (b *levelBar) DragEnd() { b.active = markerNone }

// levelAt maps an x position inside the widget to a level 0..255.
func (b *levelBar) levelAt(x float32) int {
	cs := scaleOf(b)
	pad := histPad * cs
	pw := float64(b.Size().Width)*cs - 2*pad
	if pw <= 1 {
		return 0
	}
	level := int(math.Round((float64(x)*cs - pad) / (pw - 1) * 255))
	if level < 0 {
		return 0
	}
	if level > 255 {
		return 255
	}
	return level
}

// nearest picks the marker closest to level (when they overlap: the one on that side).
func (b *levelBar) nearest(level int) int {
	if b.lo == b.hi {
		if level < b.lo {
			return markerLow
		}
		return markerHigh
	}
	if abs(level-b.lo) <= abs(level-b.hi) {
		return markerLow
	}
	return markerHigh
}

func (b *levelBar) set(marker, level int) {
	switch marker {
	case markerLow:
		if level > b.hi {
			level = b.hi
		}
		b.lo = level
	case markerHigh:
		if level < b.lo {
			level = b.lo
		}
		b.hi = level
	default:
		return
	}
	b.Refresh()
	if b.OnChange != nil {
		b.OnChange(b.lo, b.hi)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (b *levelBar) draw(w, h int) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	cs := scaleOf(b)
	pad := int(histPad * cs)
	pw := w - 2*pad
	barH := int(histBarHeight * cs)
	if pw < 2 || barH <= 0 || barH >= h {
		return out
	}

	// black-to-white gradient
	for x := 0; x < pw; x++ {
		v := uint8(float64(x) * 255 / float64(pw-1))
		c := color.NRGBA{R: v, G: v, B: v, A: 255}
		for y := 0; y < barH; y++ {
			setPix(out, pad+x, y, c)
		}
	}

	// the two markers: upward triangles under the bar
	fg := toNRGBA(theme.Color(theme.ColorNameForeground))
	rows := h - barH - 1
	halfMax := rows * 6 / 10
	for _, level := range []int{b.lo, b.hi} {
		cx := pad + level*(pw-1)/255
		for dy := 0; dy < rows; dy++ {
			half := 0
			if rows > 1 {
				half = dy * halfMax / (rows - 1)
			}
			for x := cx - half; x <= cx+half; x++ {
				if x >= 0 && x < w {
					setPix(out, x, barH+1+dy, fg)
				}
			}
		}
	}
	return out
}

func toNRGBA(c color.Color) color.NRGBA { return color.NRGBAModel.Convert(c).(color.NRGBA) }