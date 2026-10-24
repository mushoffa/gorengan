package widget

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	fynewidget "fyne.io/fyne/v2/widget"
)

var (
	_ desktop.Hoverable   = (*LineChart)(nil)
	_ fyne.Scrollable     = (*LineChart)(nil)
	_ fyne.Draggable      = (*LineChart)(nil)
	_ fyne.DoubleTappable = (*LineChart)(nil)
)

// LineChart is the line version of BarChart. It has the same features:
//   - title label in the header (e.g. "HPP")
//   - hovering shows "position: value" beside the label, with a cursor line
//     and a marker dot on the curve
//   - mouse wheel zooms around the cursor, drag pans, double-click resets
//   - a "Fill" checkbox in the header (top right) fills the area under the line
//   - colours come from Theme (Bar = line, Fill = area, BarHover = marker)
//
// X must be sorted in ascending order.
type LineChart struct {
	fynewidget.BaseWidget

	Label         string
	X, Y          []float64              // same length; X sorted ascending
	XMax          float64                // x range is 0..XMax
	YMin, YMax    float64                // y range; YMin is negative when the data has negatives
	YStep         float64                // y tick spacing (0 = automatic), e.g. 1 for ticks -1, 0, 1
	YFormat       func(v float64) string // y tick label text (nil = plain number)
	FillToZero    bool                   // fill towards the y=0 line instead of the bottom edge
	MinSpan       float64                // smallest visible x range when zoomed (default 5)
	LineWidth     float32                // default 1.5
	Fill          bool                   // fill the area under the line (the checkbox toggles this)
	ShowFillCheck bool                   // show the "Fill" checkbox in the header (default true)
	ShowTicks     bool
	Theme         ChartTheme
	Format        func(x int, y float64) string // hover text; default "x: y"

	viewSet  bool
	vLo, vHi float64

	plotLeft, plotTop, pw, ph float32

	hovering bool
	hoverPos fyne.Position
	lookup   map[int]float64

	renderer *lineChartRenderer
}

// NewLineChart creates a line chart. Pass x = nil to use the index of each
// value as its x position (0, 1, 2, ...), which suits histogram data.
func NewLineChart(label string, x, y []float64) *LineChart {
	if x == nil {
		x = make([]float64, len(y))
		for i := range x {
			x[i] = float64(i)
		}
	}
	chart := &LineChart{
		Label:         label,
		X:             x,
		Y:             y,
		MinSpan:       5,
		LineWidth:     1.5,
		ShowFillCheck: true,
		Theme:         DefaultTheme(),
	}
	minY, maxY := math.Min(0, minOf(y)), maxOf(y)
	ys := niceStep(maxY - minY)
	chart.XMax = maxOf(x)
	chart.YMin = math.Floor(minY/ys) * ys
	chart.YMax = math.Ceil(maxY/ys) * ys
	if chart.YMax <= chart.YMin {
		chart.YMax = chart.YMin + 1
	}
	chart.ExtendBaseWidget(chart)
	return chart
}

func (chart *LineChart) SetData(x, y []float64) { chart.X, chart.Y = x, y; chart.Refresh() }

func (chart *LineChart) SetTheme(t ChartTheme) { chart.Theme = t; chart.Refresh() }

// Show makes the chart visible (for example when its tab is selected) and
// redraws the Fill check box so it uses the current app theme.
func (chart *LineChart) Show() {
	chart.BaseWidget.Show()
	if chart.renderer != nil {
		chart.renderer.check.Refresh()
	}
}

// SetFill turns the area under the line on or off (same as the checkbox).
func (chart *LineChart) SetFill(on bool) { chart.Fill = on; chart.Refresh() }

func (chart *LineChart) View() (lo, hi float64) {
	if !chart.viewSet {
		return 0, chart.XMax
	}
	return chart.vLo, chart.vHi
}

func (chart *LineChart) SetView(lo, hi float64) { chart.setView(lo, hi); chart.Refresh() }

// ZoomAt scales the visible range (<1 zooms in, >1 zooms out) keeping the data
// position anchor fixed on screen.
func (chart *LineChart) ZoomAt(anchor, scale float64) {
	lo, hi := chart.View()
	chart.setView(anchor-(anchor-lo)*scale, anchor+(hi-anchor)*scale)
	chart.Refresh()
}

func (chart *LineChart) ZoomBy(scale float64) {
	lo, hi := chart.View()
	chart.ZoomAt((lo+hi)/2, scale)
}

func (chart *LineChart) ResetZoom() { chart.viewSet = false; chart.Refresh() }

func (chart *LineChart) CreateRenderer() fyne.WidgetRenderer {
	renderer := &lineChartRenderer{
		chart:  chart,
		bg:     canvas.NewRectangle(color.Transparent),
		plotBg: canvas.NewRectangle(color.Transparent),
		cursor: canvas.NewLine(color.White),
		marker: canvas.NewCircle(color.White),
		check: NewCompactCheck("Fill", func(on bool) {
			chart.Fill = on
			chart.Refresh()
		}),
		variant: fyne.CurrentApp().Settings().ThemeVariant(),
		label:   canvas.NewText("", color.White),
		value:   canvas.NewText("", color.White),
	}
	renderer.cursor.StrokeWidth = 1
	renderer.label.TextStyle = fyne.TextStyle{Bold: true}
	renderer.label.TextSize = headerTextSize
	renderer.value.TextSize = headerTextSize
	chart.renderer = renderer
	return renderer
}

func (chart *LineChart) MouseIn(e *desktop.MouseEvent)    { chart.setHover(e.Position, true) }
func (chart *LineChart) MouseMoved(e *desktop.MouseEvent) { chart.setHover(e.Position, true) }
func (chart *LineChart) MouseOut()                        { chart.setHover(fyne.Position{}, false) }

func (chart *LineChart) setHover(p fyne.Position, on bool) {
	chart.hovering, chart.hoverPos = on, p
	if chart.renderer != nil {
		chart.renderer.updateHover()
	}
}

func (chart *LineChart) Scrolled(e *fyne.ScrollEvent) {
	dy := float64(e.Scrolled.DY)
	if dy == 0 || chart.pw <= 0 {
		return
	}
	lo, hi := chart.View()
	anchor := math.Min(math.Max(chart.xAtPixel(e.Position.X), lo), hi)
	chart.hoverPos = e.Position
	chart.ZoomAt(anchor, math.Pow(0.85, dy/10)) // wheel up (DY>0) zooms in
}

func (chart *LineChart) Dragged(e *fyne.DragEvent) {
	if chart.pw <= 0 {
		return
	}
	lo, hi := chart.View()
	d := -float64(e.Dragged.DX) / float64(chart.pw) * (hi - lo)
	chart.hoverPos = e.Position
	chart.setView(lo+d, hi+d)
	chart.Refresh()
}

func (chart *LineChart) DragEnd() {}

// DoubleTapped resets the zoom.
func (chart *LineChart) DoubleTapped(*fyne.PointEvent) { chart.ResetZoom() }

func (chart *LineChart) setView(lo, hi float64) {
	if chart.XMax <= 0 {
		return
	}
	minSpan := chart.MinSpan
	if minSpan <= 0 {
		minSpan = 5
	}
	span := math.Min(math.Max(hi-lo, minSpan), chart.XMax)
	mid := (lo + hi) / 2
	lo, hi = mid-span/2, mid+span/2
	if lo < 0 {
		lo, hi = 0, span
	}
	if hi > chart.XMax {
		lo, hi = chart.XMax-span, chart.XMax
	}
	chart.vLo, chart.vHi, chart.viewSet = lo, hi, true
}

func (chart *LineChart) xAtPixel(px float32) float64 {
	return chart.vLo + float64((px-chart.plotLeft)/chart.pw)*(chart.vHi-chart.vLo)
}

func (chart *LineChart) xToPixel(v float64) float32 {
	return chart.plotLeft + float32((v-chart.vLo)/(chart.vHi-chart.vLo))*chart.pw
}

// clampY limits a value to the visible y range [YMin, YMax]. The line and the
// fill both interpolate between clamped values so they always match.
func (chart *LineChart) clampY(v float64) float64 {
	return math.Min(math.Max(v, chart.YMin), chart.YMax)
}

func (chart *LineChart) yToPixel(v float64) float32 {
	v = chart.clampY(v)
	return chart.plotTop + chart.ph - float32((v-chart.YMin)/(chart.YMax-chart.YMin))*chart.ph
}

type lineChartRenderer struct {
	chart   *LineChart
	objects []fyne.CanvasObject

	bg, plotBg   *canvas.Rectangle
	check        *CompactCheck
	variant      fyne.ThemeVariant // app theme variant the check box was last drawn with
	cursor       *canvas.Line
	marker       *canvas.Circle
	label, value *canvas.Text
}

const markerSize float32 = 7

func (r *lineChartRenderer) MinSize() fyne.Size           { return fyne.NewSize(240, 120) }
func (r *lineChartRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *lineChartRenderer) Destroy()                     {}

func (r *lineChartRenderer) Refresh() {
	r.Layout(r.chart.Size())
	canvas.Refresh(r.chart)
}

func (r *lineChartRenderer) Layout(size fyne.Size) {
	chart := r.chart
	t := chart.Theme

	// Widgets inside a hidden tab miss app theme changes; redraw on a variant change.
	if v := fyne.CurrentApp().Settings().ThemeVariant(); v != r.variant {
		r.variant = v
		r.check.Refresh()
	}
	if r.check.Checked() != chart.Fill {
		r.check.SetChecked(chart.Fill)
	}

	chart.vLo, chart.vHi = chart.View()

	padL, padB := padSide, padSide
	if chart.ShowTicks {
		padL, padB = 40, 22
		for _, v := range chart.yTicks() {
			tx := canvas.NewText(chart.formatY(v), color.White)
			tx.TextSize = tickTextSize
			if w := tx.MinSize().Width + 12; w > padL {
				padL = w
			}
		}
	}
	chart.plotLeft, chart.plotTop = padL, headerHeight
	chart.pw = size.Width - padL - padSide
	chart.ph = size.Height - headerHeight - padB

	r.bg.FillColor = t.Background
	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)
	r.plotBg.FillColor = t.PlotBackground
	r.plotBg.Move(fyne.NewPos(chart.plotLeft, chart.plotTop))
	r.plotBg.Resize(fyne.NewSize(maxf(chart.pw, 0), maxf(chart.ph, 0)))
	r.cursor.StrokeColor = t.Cursor
	r.marker.FillColor = t.BarHover
	r.marker.Resize(fyne.NewSize(markerSize, markerSize))

	r.label.Text, r.label.Color = chart.Label, t.Label
	ls := r.label.MinSize()
	r.label.Resize(ls)
	ty := (headerHeight - ls.Height) / 2
	r.label.Move(fyne.NewPos(padSide, ty))
	r.value.Color = t.Value
	r.value.Move(fyne.NewPos(padSide+ls.Width+8, ty))

	r.objects = []fyne.CanvasObject{r.bg, r.plotBg}

	chart.lookup = make(map[int]float64, len(chart.X))
	for i := 0; i < len(chart.X) && i < len(chart.Y); i++ {
		chart.lookup[int(math.Round(chart.X[i]))] = chart.Y[i]
	}

	if chart.pw > 0 && chart.ph > 0 && chart.XMax > 0 && chart.YMax > chart.YMin && chart.vHi > chart.vLo {
		r.buildPlot()
	}
	if chart.ShowFillCheck {
		cs := r.check.MinSize()
		r.check.Resize(cs)
		r.check.Move(fyne.NewPos(size.Width-padSide-cs.Width, (headerHeight-cs.Height)/2))
		r.objects = append(r.objects, r.check)
	}
	r.objects = append(r.objects, r.cursor, r.marker, r.label, r.value)
	r.updateHover()
}

func (r *lineChartRenderer) buildPlot() {
	chart := r.chart
	t := chart.Theme
	lo, hi := chart.vLo, chart.vHi
	x0, x1 := chart.plotLeft, chart.plotLeft+chart.pw
	bottom := chart.plotTop + chart.ph

	if chart.ShowTicks {
		for _, v := range chart.yTicks() {
			y := chart.yToPixel(v)
			col := t.Grid
			if v == 0 && chart.YMin < 0 {
				col = t.Axis // emphasise the zero line
			}
			r.addLine(x0, y, x1, y, col, 1)
			r.addText(chart.formatY(v), t.Axis, x0-6, y, anchorRight, true)
		}
		xStep := niceStep(hi - lo)
		for v := math.Ceil(lo/xStep) * xStep; v <= hi+1e-9; v += xStep {
			x := chart.xToPixel(v)
			r.addLine(x, chart.plotTop, x, bottom, t.Grid, 1)
			r.addLine(x, bottom, x, bottom+4, t.Axis, 1)
			r.addText(fmt.Sprintf("%g", tidy(v)), t.Axis, x, bottom+6, anchorCenter, false)
		}
	}

	if chart.Fill {
		r.addFill()
	}

	// curve: one segment per pair of neighbouring points, clipped to the view
	lw := chart.LineWidth
	if lw <= 0 {
		lw = 1.5
	}
	n := len(chart.X)
	if len(chart.Y) < n {
		n = len(chart.Y)
	}
	for i := 0; i+1 < n; i++ {
		ax, ay, bx, by := chart.X[i], chart.clampY(chart.Y[i]), chart.X[i+1], chart.clampY(chart.Y[i+1])
		if bx <= ax || (bx < lo) || (ax > hi) {
			continue
		}
		sx, sy, ex, ey := ax, ay, bx, by
		if sx < lo {
			sy = ay + (by-ay)*(lo-ax)/(bx-ax)
			sx = lo
		}
		if ex > hi {
			ey = ay + (by-ay)*(hi-ax)/(bx-ax)
			ex = hi
		}
		r.addLine(chart.xToPixel(sx), chart.yToPixel(sy), chart.xToPixel(ex), chart.yToPixel(ey), t.Bar, lw)
	}

	if chart.ShowTicks {
		r.addLine(x0, chart.plotTop, x0, bottom, t.Axis, 1)
		r.addLine(x0, bottom, x1, bottom, t.Axis, 1)
	}
}

// updateHover moves the cursor line and marker and updates the hover text.
// It never rebuilds the plot, so it is cheap.
func (r *lineChartRenderer) updateHover() {
	chart := r.chart

	p := chart.hoverPos
	inside := chart.hovering && chart.pw > 0 &&
		p.X >= chart.plotLeft && p.X <= chart.plotLeft+chart.pw &&
		p.Y >= chart.plotTop && p.Y <= chart.plotTop+chart.ph
	idx := 0
	if inside {
		idx = int(math.Round(chart.xAtPixel(p.X)))
		inside = idx >= 0 && float64(idx) <= chart.XMax
	}

	if !inside {
		r.cursor.Hide()
		r.marker.Hide()
		r.value.Text = ""
		r.value.Resize(r.value.MinSize())
		r.value.Refresh()
		r.cursor.Refresh()
		r.marker.Refresh()
		return
	}

	px := chart.xToPixel(float64(idx))
	r.cursor.Position1 = fyne.NewPos(px, chart.plotTop)
	r.cursor.Position2 = fyne.NewPos(px, chart.plotTop+chart.ph)
	r.cursor.Show()
	r.cursor.Refresh()

	y, ok := chart.lookup[idx]
	switch {
	case !ok:
		r.value.Text = fmt.Sprintf("%d: -", idx)
		r.marker.Hide()
	case chart.Format != nil:
		r.value.Text = chart.Format(idx, y)
	default:
		r.value.Text = fmt.Sprintf("%d: %g", idx, y)
	}
	if ok {
		r.marker.Move(fyne.NewPos(px-markerSize/2, chart.yToPixel(y)-markerSize/2))
		r.marker.Show()
	}
	r.marker.Refresh()
	r.value.Resize(r.value.MinSize())
	r.value.Refresh()
}

func (r *lineChartRenderer) addLine(x1, y1, x2, y2 float32, col color.Color, width float32) {
	l := canvas.NewLine(col)
	l.StrokeWidth = width
	l.Position1 = fyne.NewPos(x1, y1)
	l.Position2 = fyne.NewPos(x2, y2)
	r.objects = append(r.objects, l)
}

// addText places text so that (ax, y) is its left/centre/right anchor; y is the
// top of the text, or its vertical centre when vCenter is set.
func (r *lineChartRenderer) addText(s string, col color.Color, ax, y float32, a anchor, vCenter bool) {
	t := canvas.NewText(s, col)
	t.TextSize = tickTextSize
	sz := t.MinSize()
	t.Resize(sz)
	switch a {
	case anchorCenter:
		ax -= sz.Width / 2
	case anchorRight:
		ax -= sz.Width
	}
	if vCenter {
		y -= sz.Height / 2
	}
	t.Move(fyne.NewPos(ax, y))
	r.objects = append(r.objects, t)
}

// addFill paints the area under the curve as one thin column per pixel.
func (r *lineChartRenderer) addFill() {
	chart := r.chart
	n := len(chart.X)
	if len(chart.Y) < n {
		n = len(chart.Y)
	}
	cols := int(chart.pw)
	if n < 2 || cols < 1 {
		return
	}
	colW := chart.pw / float32(cols)
	base := chart.yToPixel(chart.fillBase())
	span := chart.vHi - chart.vLo
	xs := chart.X[:n]

	for i := 0; i < cols; i++ {
		x := chart.vLo + (float64(i)+0.5)/float64(cols)*span
		k := sort.SearchFloat64s(xs, x) // first index with X[k] >= x
		var y float64
		switch {
		case k == 0:
			if xs[0] != x {
				continue // left of the data
			}
			y = chart.clampY(chart.Y[0])
		case k == n:
			continue // right of the data
		default:
			x0, x1 := xs[k-1], xs[k]
			ya, yb := chart.clampY(chart.Y[k-1]), chart.clampY(chart.Y[k])
			y = ya + (yb-ya)*(x-x0)/(x1-x0)
		}
		top := chart.yToPixel(y)
		t0, t1 := minf(top, base), maxf(top, base)
		if t1-t0 < 0.5 {
			continue
		}
		rect := canvas.NewRectangle(chart.Theme.Fill)
		rect.Move(fyne.NewPos(chart.plotLeft+float32(i)*colW, t0))
		rect.Resize(fyne.NewSize(colW, t1-t0))
		r.objects = append(r.objects, rect)
	}
}

// yTicks returns the y values that get a grid line and a label.
func (chart *LineChart) yTicks() []float64 {
	step := chart.YStep
	if step <= 0 {
		step = niceStep(chart.YMax - chart.YMin)
	}
	first := math.Ceil(chart.YMin/step-1e-9) * step
	var ticks []float64
	for i := 0; i < 1000; i++ {
		v := tidy(first + float64(i)*step)
		if v > chart.YMax+1e-9 {
			break
		}
		ticks = append(ticks, v)
	}
	return ticks
}

// formatY returns the label for a y tick (YFormat, or a plain number).
func (chart *LineChart) formatY(v float64) string {
	if chart.YFormat != nil {
		return chart.YFormat(v)
	}
	return fmt.Sprintf("%g", tidy(v))
}

// fillBase is the y value the fill extends to: the bottom edge, or the zero
// line when FillToZero is set.
func (chart *LineChart) fillBase() float64 {
	if chart.FillToZero {
		return chart.clampY(0)
	}
	return chart.YMin
}
