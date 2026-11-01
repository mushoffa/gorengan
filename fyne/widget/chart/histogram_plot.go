package chart

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	// histPad is the side margin (Fyne units) of the plot and of the level bar.
	// Both leave it free, so the level bar's markers can sit exactly under the
	// first and last column of the plot.
	histPad          = 6.0
	histBarHeight    = 12 // the black-to-white bar, Fyne units
	histMarkerHeight = 10 // the triangles under it
)

// The plot keeps the same dark panel look in light and dark mode (like Photoshop's).
var (
	plotBackground = color.NRGBA{R: 0x46, G: 0x46, B: 0x4A, A: 0xFF}
	plotBar        = color.NRGBA{R: 0xD4, G: 0xD4, B: 0xD8, A: 0xFF}
	plotBarDim     = color.NRGBA{R: 0x86, G: 0x86, B: 0x8C, A: 0xFF} // outside the selected range
	plotCursor     = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xCC}
	plotAxis       = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x40} // the zero line of a series with negative values
)

// histogramPlot draws the bars on a dark, square-cornered panel. Levels outside
// the selected range are dimmed, and the level under the mouse gets a vertical line.
type histogramPlot struct {
	widget.BaseWidget
	raster *canvas.Raster
	bins   [256]int
	lo, hi int  // selected range (all levels by default)
	hover  int  // level under the mouse, -1 for none
	filled bool // filled bars, or just the outline of the histogram

	OnHover func(level int) // called with -1 when the mouse leaves
}

func newHistogramPlot() *histogramPlot {
	p := &histogramPlot{lo: 0, hi: 255, hover: -1, filled: true}
	p.raster = canvas.NewRaster(p.draw)
	p.raster.ScaleMode = canvas.ImageScalePixels
	p.raster.SetMinSize(fyne.NewSize(180, 90))
	p.ExtendBaseWidget(p)
	return p
}

func (p *histogramPlot) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.raster)
}

func (p *histogramPlot) SetBins(bins [256]int) {
	p.bins = bins
	p.Refresh()
}

func (p *histogramPlot) SetRange(lo, hi int) {
	p.lo, p.hi = lo, hi
	p.Refresh()
}

// SetFilled chooses filled bars (default) or only the outline.
func (p *histogramPlot) SetFilled(filled bool) {
	p.filled = filled
	p.Refresh()
}

// ---- hover (desktop.Hoverable) ----

func (p *histogramPlot) MouseIn(ev *desktop.MouseEvent)    { p.track(ev.Position) }
func (p *histogramPlot) MouseMoved(ev *desktop.MouseEvent) { p.track(ev.Position) }
func (p *histogramPlot) MouseOut() {
	p.hover = -1
	p.Refresh()
	if p.OnHover != nil {
		p.OnHover(-1)
	}
}
func (p *histogramPlot) Cursor() desktop.Cursor { return desktop.CrosshairCursor }

func (p *histogramPlot) track(pos fyne.Position) {
	cs := scaleOf(p)
	pad := histPad * cs
	pw := float64(p.Size().Width)*cs - 2*pad
	if pw <= 0 {
		return
	}
	level := int(math.Floor((float64(pos.X)*cs - pad) / pw * 256))
	if level < 0 {
		level = 0
	}
	if level > 255 {
		level = 255
	}
	p.hover = level
	p.Refresh()
	if p.OnHover != nil {
		p.OnHover(level)
	}
}

func (p *histogramPlot) draw(w, h int) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 {
		return out
	}
	pad := int(histPad * scaleOf(p))
	x0, pw := pad, w-2*pad
	if pw <= 0 {
		return out
	}

	for y := 0; y < h; y++ {
		for x := 0; x < pw; x++ {
			setPix(out, x0+x, y, plotBackground)
		}
	}

	peak := 0
	for _, v := range p.bins {
		if v > peak {
			peak = v
		}
	}
	if peak > 0 {
		lw, prevY := lineWidthOf(p), 0
		for x := 0; x < pw; x++ {
			// a pixel column covers one or more levels; use the tallest
			lo, hi := x*256/pw, (x+1)*256/pw
			if hi <= lo {
				hi = lo + 1
			}
			count := 0
			for l := lo; l < hi && l < 256; l++ {
				if p.bins[l] > count {
					count = p.bins[l]
				}
			}
			barH := int(int64(count) * int64(h) / int64(peak))

			c := plotBar
			if mid := (lo + hi - 1) / 2; mid < p.lo || mid > p.hi {
				c = plotBarDim
			}
			if p.filled {
				for y := h - barH; y < h; y++ {
					setPix(out, x0+x, y, c)
				}
				continue
			}

			// outline: join the top of this column to the top of the previous one
			y := h - barH
			if y > h-1 {
				y = h - 1
			}
			top, bottom := y, y
			if x > 0 {
				if prevY < top {
					top = prevY
				}
				if prevY > bottom {
					bottom = prevY
				}
			}
			drawVSpan(out, x0+x, top, bottom, lw, c)
			prevY = y
		}
	}

	if p.hover >= 0 {
		x := x0 + int((float64(p.hover)+0.5)/256*float64(pw))
		for y := 0; y < h; y++ {
			blendPix(out, x, y, plotCursor)
		}
	}
	return out
}

func setPix(img *image.RGBA, x, y int, c color.NRGBA) {
	i := img.PixOffset(x, y)
	img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, 255
}

// blendPix draws c over the (opaque) pixel that is already there.
func blendPix(img *image.RGBA, x, y int, c color.NRGBA) {
	if x < 0 || x >= img.Bounds().Dx() {
		return
	}
	i := img.PixOffset(x, y)
	a := uint32(c.A)
	for k, s := range [3]uint8{c.R, c.G, c.B} {
		d := uint32(img.Pix[i+k])
		img.Pix[i+k] = uint8((uint32(s)*a + d*(255-a)) / 255)
	}
}

// scaleOf is the canvas scale (device pixels per Fyne unit) of the window the object is in.
func scaleOf(o fyne.CanvasObject) float64 {
	if c := fyne.CurrentApp().Driver().CanvasForObject(o); c != nil {
		return float64(c.Scale())
	}
	return 1
}

// lineWidthOf is the thickness of outline plots in device pixels (about 1.5 Fyne units, at least 2).
func lineWidthOf(o fyne.CanvasObject) int {
	return int(math.Max(2, math.Round(1.5*scaleOf(o))))
}

// drawVSpan draws a vertical run of pixels in column x from y0 to y1, thickened
// by lineWidth, so that neighbouring runs join into one continuous line.
func drawVSpan(img *image.RGBA, x, y0, y1, lineWidth int, c color.NRGBA) {
	if x < 0 || x >= img.Bounds().Dx() {
		return
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	h := img.Bounds().Dy()
	for y := y0 - lineWidth/2; y <= y1+(lineWidth-1)/2; y++ {
		if y >= 0 && y < h {
			setPix(img, x, y, c)
		}
	}
}
