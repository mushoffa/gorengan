package chart

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// histogramStats are the whole-image numbers of a 256-bin histogram.
type histogramStats struct {
	Pixels int64   // total number of pixels
	Mean   float64 // average level
	StdDev float64 // standard deviation of the level (population)
	Median int     // level at which half of the pixels are at or below
	Peak   int     // the level with the most pixels
}

func computeStats(bins [256]int) histogramStats {
	var s histogramStats
	var weighted float64
	for level, n := range bins {
		s.Pixels += int64(n)
		weighted += float64(level) * float64(n)
		if n > bins[s.Peak] {
			s.Peak = level
		}
	}
	if s.Pixels == 0 {
		return s
	}

	s.Mean = weighted / float64(s.Pixels)

	var squares float64
	for level, n := range bins {
		d := float64(level) - s.Mean
		squares += d * d * float64(n)
	}
	s.StdDev = math.Sqrt(squares / float64(s.Pixels))

	var cumulative int64
	for level, n := range bins {
		cumulative += int64(n)
		if cumulative*2 >= s.Pixels {
			s.Median = level
			break
		}
	}
	return s
}

// countIn is how many pixels have a level in [lo, hi].
func countIn(bins [256]int, lo, hi int) int64 {
	if lo < 0 {
		lo = 0
	}
	if hi > 255 {
		hi = 255
	}
	var n int64
	for level := lo; level <= hi; level++ {
		n += int64(bins[level])
	}
	return n
}

// percent is part of total as a percentage (0 when total is 0).
func percent(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func smallTextSize() float32 { return theme.TextSize() - 1 }

// caption is one line of small text in a theme color. It re-reads the color on
// Refresh, so it follows light/dark changes.
type caption struct {
	widget.BaseWidget
	text      *canvas.Text
	colorName fyne.ThemeColorName
}

func newCaption(text string, colorName fyne.ThemeColorName) *caption {
	c := &caption{text: canvas.NewText(text, color.Black), colorName: colorName}
	c.ExtendBaseWidget(c)
	return c
}

func (c *caption) SetText(s string) {
	c.text.Text = s
	c.Refresh()
}

func (c *caption) CreateRenderer() fyne.WidgetRenderer {
	r := &captionRenderer{c: c}
	r.apply()
	return r
}

type captionRenderer struct{ c *caption }

func (r *captionRenderer) apply() {
	r.c.text.Color = theme.Color(r.c.colorName)
	r.c.text.TextSize = smallTextSize()
}

func (r *captionRenderer) Refresh() {
	r.apply()
	r.Layout(r.c.Size())
	r.c.text.Refresh()
}

func (r *captionRenderer) Layout(size fyne.Size) {
	ts := r.c.text.MinSize()
	r.c.text.Resize(ts)
	r.c.text.Move(fyne.NewPos(0, (size.Height-ts.Height)/2))
}

func (r *captionRenderer) MinSize() fyne.Size           { return r.c.text.MinSize() }
func (r *captionRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.c.text} }
func (r *captionRenderer) Destroy()                     {}

// measureWidth is the width of the widest of the given texts at the caption size.
func measureWidth(texts ...string) float32 {
	var w float32
	for _, s := range texts {
		t := canvas.NewText(s, color.Black)
		t.TextSize = smallTextSize()
		if m := t.MinSize().Width; m > w {
			w = m
		}
	}
	return w
}

const statGap = 4

// keyValueLayout puts the value a fixed distance from the left edge, so the
// values of a column line up.
type keyValueLayout struct{ keyWidth float32 }

func (l keyValueLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	if len(objs) < 2 {
		return
	}
	objs[0].Resize(fyne.NewSize(l.keyWidth, size.Height))
	objs[0].Move(fyne.NewPos(0, 0))
	x := l.keyWidth + statGap
	objs[1].Resize(fyne.NewSize(fyne.Max(size.Width-x, 0), size.Height))
	objs[1].Move(fyne.NewPos(x, 0))
}

func (l keyValueLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	if len(objs) < 2 {
		return fyne.NewSize(0, 0)
	}
	k, v := objs[0].MinSize(), objs[1].MinSize()
	return fyne.NewSize(l.keyWidth+statGap+v.Width, fyne.Max(k.Height, v.Height))
}

// newStat returns a "Key:  value" row and the caption to set the value on.
func newStat(key string, keyWidth float32) (fyne.CanvasObject, *caption) {
	k := newCaption(key, theme.ColorNameDisabled)
	v := newCaption("-", theme.ColorNameForeground)
	return container.New(keyValueLayout{keyWidth: keyWidth}, k, v), v
}
