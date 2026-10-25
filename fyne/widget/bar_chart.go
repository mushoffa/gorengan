package widget

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	fynewidget "fyne.io/fyne/v2/widget"
)

var (
	_ desktop.Hoverable   = (*BarChart)(nil)
	_ fyne.Scrollable     = (*BarChart)(nil)
	_ fyne.Draggable      = (*BarChart)(nil)
	_ fyne.DoubleTappable = (*BarChart)(nil)
)

// BarChart is a custom Fyne widget:
//   - a title label in the header (e.g. "Histogram")
//   - hovering shows "position: value" beside the label, plus a cursor line
//   - mouse wheel zooms around the cursor, drag pans, double-click resets
//   - all colors come from Theme
type BarChart struct {
	fynewidget.BaseWidget

	Label         string
	X, Y          []float64 // bar positions and heights (same length)
	XMax, YMax    float64   // data ranges, both start at 0
	YStep         float64   // y tick spacing (0 = automatic)
	MinSpan       float64   // smallest visible x range when zoomed (default 5)
	MinBarWidth   float32   // bars never get thinner than this many pixels
	ShowTicks     bool      // axes, grid and tick labels (off = plain look, no axes at all)
	ShowXAxis     bool      // with ShowTicks on: false hides the x axis (labels, ticks, vertical grid, axis line); default true
	ShowYAxis     bool      // with ShowTicks on: false hides the y axis (labels, horizontal grid, axis line); default true
	Fill          bool      // true = solid bars (default); false = outlined bars
	ShowFillCheck bool      // show the "Fill" check box in the header (default true)
	Theme         ChartTheme
	UseAppTheme   bool                          // draw with the app theme from Settings().SetTheme instead of Theme
	HeaderHeight  float32                       // height of the label row above the plot (default 30)
	LabelPadding  Insets                        // label row: Left = before the title, Right = after the Fill check, Top/Bottom shift it vertically
	PlotPadding   Insets                        // space between the widget edge and the plot area; Top = gap below the label row
	Format        func(x int, y float64) string // hover text; default "x: y"

	// zoom state
	viewSet  bool
	vLo, vHi float64 // visible x range as of the last layout

	// geometry of the plot area as of the last layout
	active ChartTheme // colours of the last layout (Theme, or the app theme)

	plotLeft, plotTop, pw, ph float32

	// hover state
	hovering bool
	hoverPos fyne.Position
	lookup   map[int]float64 // rounded x -> value, rebuilt on layout

	renderer *barChartRenderer
}

// NewBarChart creates a chart with the default (light) theme and automatic axis ranges.
func NewBarChart(label string, x, y []float64) *BarChart {
	chart := &BarChart{
		Label:         label,
		MinSpan:       5,
		MinBarWidth:   2,
		Fill:          true,
		ShowFillCheck: true,
		ShowXAxis:     true,
		ShowYAxis:     true,
		HeaderHeight:  headerHeight,
		LabelPadding:  Insets{Left: padSide, Right: padSide},
		PlotPadding:   Insets{Left: padSide, Right: padSide, Bottom: padSide},
		Theme:         DefaultTheme(),
	}
	chart.setData(x, y)
	chart.ExtendBaseWidget(chart)
	return chart
}

// setData stores the data and fits the axis ranges to it. x = nil means the
// index of each value (0, 1, 2, ...). The zoom is reset.
func (chart *BarChart) setData(x, y []float64) {
	if x == nil {
		x = indices(len(y))
	}
	chart.X, chart.Y = x, y

	xs, ys := niceStep(maxOf(x)), niceStep(maxOf(y))
	chart.XMax = math.Ceil(maxOf(x)/xs) * xs
	chart.YMax = math.Ceil(maxOf(y)/ys) * ys
	chart.viewSet = false
}

// SetData replaces the data and redraws. Pass x = nil to use the index of each
// value as its x position. XMax and YMax are recalculated from the new data and
// the zoom is reset, so set any fixed range after calling SetData.
// Call it from the UI thread (in a goroutine, wrap it in fyne.Do).
func (chart *BarChart) SetData(x, y []float64) { chart.setData(x, y); chart.Refresh() }

func (chart *BarChart) SetTheme(t ChartTheme) { chart.Theme = t; chart.Refresh() }

// showX / showY tell whether the x / y axis is drawn.
func (chart *BarChart) showX() bool { return chart.ShowTicks && chart.ShowXAxis }
func (chart *BarChart) showY() bool { return chart.ShowTicks && chart.ShowYAxis }

// SetFill switches between solid bars (true) and outlined bars (false), the
// same as the Fill check box. The outline is easiest to see when zoomed in.
func (chart *BarChart) SetFill(on bool) { chart.Fill = on; chart.Refresh() }

// Show makes the chart visible (for example when its tab is selected) and
// redraws the Fill check box so it uses the current app theme.
func (chart *BarChart) Show() {
	chart.BaseWidget.Show()
	if chart.renderer != nil {
		chart.renderer.check.Refresh()
		if chart.UseAppTheme {
			chart.Refresh() // re-read the app theme (it may have changed while hidden)
		}
	}
}

// View returns the visible x range.
func (chart *BarChart) View() (lo, hi float64) {
	if !chart.viewSet {
		return 0, chart.XMax
	}
	return chart.vLo, chart.vHi
}

// SetView shows the x range [lo, hi], clamped to the data range.
func (chart *BarChart) SetView(lo, hi float64) { chart.setView(lo, hi); chart.Refresh() }

// ZoomAt scales the visible range by scale (<1 zooms in, >1 zooms out),
// keeping the data position anchor fixed on screen.
func (chart *BarChart) ZoomAt(anchor, scale float64) {
	lo, hi := chart.View()
	chart.setView(anchor-(anchor-lo)*scale, anchor+(hi-anchor)*scale)
	chart.Refresh()
}

// ZoomBy zooms around the centre of the visible range.
func (chart *BarChart) ZoomBy(scale float64) {
	lo, hi := chart.View()
	chart.ZoomAt((lo+hi)/2, scale)
}

func (chart *BarChart) ResetZoom() { chart.viewSet = false; chart.Refresh() }

func (chart *BarChart) CreateRenderer() fyne.WidgetRenderer {
	renderer := &barChartRenderer{
		chart:  chart,
		bg:     canvas.NewRectangle(color.Transparent),
		plotBg: canvas.NewRectangle(color.Transparent),
		cursor: canvas.NewLine(color.White),
		label:  canvas.NewText("", color.White),
		value:  canvas.NewText("", color.White),
		check: NewCompactCheck("Fill", func(on bool) {
			chart.Fill = on
			chart.Refresh()
		}),
		variant: fyne.CurrentApp().Settings().ThemeVariant(),
	}
	renderer.cursor.StrokeWidth = 1
	renderer.label.TextStyle = fyne.TextStyle{Bold: true}
	renderer.label.TextSize = headerTextSize
	renderer.value.TextSize = headerTextSize
	chart.renderer = renderer
	return renderer
}

func (chart *BarChart) MouseIn(e *desktop.MouseEvent)    { chart.setHover(e.Position, true) }
func (chart *BarChart) MouseMoved(e *desktop.MouseEvent) { chart.setHover(e.Position, true) }
func (chart *BarChart) MouseOut()                        { chart.setHover(fyne.Position{}, false) }

func (chart *BarChart) setHover(p fyne.Position, on bool) {
	chart.hovering, chart.hoverPos = on, p
	if chart.renderer != nil {
		chart.renderer.updateHover()
	}
}

// Scrolled zooms with the mouse wheel / touchpad, anchored at the cursorenderer.
func (chart *BarChart) Scrolled(e *fyne.ScrollEvent) {
	dy := float64(e.Scrolled.DY)
	if dy == 0 || chart.pw <= 0 {
		return
	}
	lo, hi := chart.View()
	anchor := math.Min(math.Max(chart.xAtPixel(e.Position.X), lo), hi)
	chart.hoverPos = e.Position
	chart.ZoomAt(anchor, math.Pow(0.85, dy/10)) // wheel up (DY>0) zooms in
}

// Dragged pans the view.
func (chart *BarChart) Dragged(e *fyne.DragEvent) {
	if chart.pw <= 0 {
		return
	}
	lo, hi := chart.View()
	d := -float64(e.Dragged.DX) / float64(chart.pw) * (hi - lo)
	chart.hoverPos = e.Position
	chart.setView(lo+d, hi+d)
	chart.Refresh()
}

func (chart *BarChart) DragEnd() {}

// DoubleTapped resets the zoom.
func (chart *BarChart) DoubleTapped(*fyne.PointEvent) { chart.ResetZoom() }

// styleBar colours a bar: solid when Fill is on, outline only when it is off.
// hover selects the highlight colourenderer.
func (chart *BarChart) styleBar(rect *canvas.Rectangle, hover bool) {
	col := chart.active.Bar
	if hover {
		col = chart.active.BarHover
	}
	if chart.Fill {
		rect.FillColor, rect.StrokeWidth = col, 0
		return
	}
	rect.FillColor = color.Transparent
	rect.StrokeColor, rect.StrokeWidth = col, 1
}

func (chart *BarChart) setView(lo, hi float64) {
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

func (chart *BarChart) xAtPixel(px float32) float64 {
	return chart.vLo + float64((px-chart.plotLeft)/chart.pw)*(chart.vHi-chart.vLo)
}

func (chart *BarChart) xToPixel(v float64) float32 {
	return chart.plotLeft + float32((v-chart.vLo)/(chart.vHi-chart.vLo))*chart.pw
}

func (chart *BarChart) yToPixel(v float64) float32 {
	return chart.plotTop + chart.ph - float32(v/chart.YMax)*chart.ph
}

const (
	headerHeight   float32 = 30
	headerTextSize float32 = 13
	tickTextSize   float32 = 11
	padSide        float32 = 8

	// Gap between the plot edge and its tick labels.
	tickLabelGap float32 = 6
)

// Insets is space around something, in pixels.
type Insets struct{ Top, Right, Bottom, Left float32 }

// tickLabelSize is the size of a tick label drawn at the tick text size.
func tickLabelSize(s string) (w, h float32) {
	t := canvas.NewText(s, color.White)
	t.TextSize = tickTextSize
	m := t.MinSize()
	return m.Width, m.Height
}

type anchor int

const (
	anchorLeft anchor = iota
	anchorCenter
	anchorRight
)

type barChartRenderer struct {
	chart   *BarChart
	objects []fyne.CanvasObject

	// persistent objects (kept across layouts so hovering is cheap)
	bg, plotBg   *canvas.Rectangle
	check        *CompactCheck
	variant      fyne.ThemeVariant // app theme variant the check box was last drawn with
	cursor       *canvas.Line
	label, value *canvas.Text

	bars    map[int]*canvas.Rectangle
	hovered *canvas.Rectangle
}

func (renderer *barChartRenderer) MinSize() fyne.Size           { return fyne.NewSize(240, 120) }
func (renderer *barChartRenderer) Objects() []fyne.CanvasObject { return renderer.objects }
func (renderer *barChartRenderer) Destroy()                     {}

func (renderer *barChartRenderer) Refresh() {
	renderer.Layout(renderer.chart.Size())
	canvas.Refresh(renderer.chart)
}

// Layout rebuilds the plot for the current size, data, theme and zoom.
func (renderer *barChartRenderer) Layout(size fyne.Size) {
	chart := renderer.chart
	// Pick the colours for this layout: the app theme (Settings().SetTheme) or Theme.
	chart.active = chart.Theme
	if chart.UseAppTheme {
		chart.active = ChartThemeFromApp()
	}
	t := chart.active

	// Widgets inside a hidden tab miss app theme changes; redraw on a variant change.
	if v := fyne.CurrentApp().Settings().ThemeVariant(); v != renderer.variant {
		renderer.variant = v
		renderer.check.Refresh()
	}
	if renderer.check.Checked() != chart.Fill {
		renderer.check.SetChecked(chart.Fill)
	}

	chart.vLo, chart.vHi = chart.View()

	// Plot area: PlotPadding, plus room for the tick labels when ShowTicks is on.
	hh := chart.HeaderHeight
	if hh <= 0 {
		hh = headerHeight
	}
	padL, padB := chart.PlotPadding.Left, chart.PlotPadding.Bottom
	// The tick labels get their own room, so padding of 0 never clips them.
	if chart.showY() {
		wl, _ := tickLabelSize(fmt.Sprintf("%g", chart.YMax))
		padL += wl + tickLabelGap
	}
	if chart.showX() {
		_, hl := tickLabelSize("0")
		padB += hl + tickLabelGap
	}
	chart.plotLeft, chart.plotTop = padL, hh+chart.PlotPadding.Top
	chart.pw = size.Width - padL - chart.PlotPadding.Right
	chart.ph = size.Height - chart.plotTop - padB

	// persistent objects
	renderer.bg.FillColor = t.Background
	renderer.bg.Move(fyne.NewPos(0, 0))
	renderer.bg.Resize(size)
	renderer.plotBg.FillColor = t.PlotBackground
	renderer.plotBg.Move(fyne.NewPos(chart.plotLeft, chart.plotTop))
	renderer.plotBg.Resize(fyne.NewSize(maxf(chart.pw, 0), maxf(chart.ph, 0)))
	renderer.cursor.StrokeColor = t.Cursor

	renderer.label.Text, renderer.label.Color = chart.Label, t.Label
	ls := renderer.label.MinSize()
	renderer.label.Resize(ls)
	lp := chart.LabelPadding
	cy := lp.Top + (hh-lp.Top-lp.Bottom)/2 // vertical centre of the label row
	ty := cy - ls.Height/2
	renderer.label.Move(fyne.NewPos(lp.Left, ty))
	renderer.value.Color = t.Value
	renderer.value.Move(fyne.NewPos(lp.Left+ls.Width+8, ty))

	renderer.objects = []fyne.CanvasObject{renderer.bg, renderer.plotBg}
	renderer.bars = map[int]*canvas.Rectangle{}
	renderer.hovered = nil

	chart.lookup = make(map[int]float64, len(chart.X))
	for i := 0; i < len(chart.X) && i < len(chart.Y); i++ {
		chart.lookup[int(math.Round(chart.X[i]))] = chart.Y[i]
	}

	if chart.pw > 0 && chart.ph > 0 && chart.XMax > 0 && chart.YMax > 0 && chart.vHi > chart.vLo {
		renderer.buildPlot()
	}
	if chart.ShowFillCheck {
		cs := renderer.check.MinSize()
		renderer.check.Resize(cs)
		renderer.check.Move(fyne.NewPos(size.Width-lp.Right-cs.Width, cy-cs.Height/2))
		renderer.objects = append(renderer.objects, renderer.check)
	}
	renderer.objects = append(renderer.objects, renderer.cursor, renderer.label, renderer.value)
	renderer.updateHover()
}

func (renderer *barChartRenderer) buildPlot() {
	chart := renderer.chart
	t := chart.active
	lo, hi := chart.vLo, chart.vHi
	x0, x1 := chart.plotLeft, chart.plotLeft+chart.pw
	bottom := chart.plotTop + chart.ph

	if chart.showY() {
		yStep := chart.YStep
		if yStep <= 0 {
			yStep = niceStep(chart.YMax)
		}
		for v := 0.0; v <= chart.YMax+1e-9; v += yStep {
			y := chart.yToPixel(v)
			renderer.addLine(x0, y, x1, y, t.Grid)
			renderer.addText(fmt.Sprintf("%g", v), t.Axis, x0-6, y, anchorRight, true)
		}
	}
	if chart.showX() {
		xStep := niceStep(hi - lo)
		for v := math.Ceil(lo/xStep) * xStep; v <= hi+1e-9; v += xStep {
			x := chart.xToPixel(v)
			renderer.addLine(x, chart.plotTop, x, bottom, t.Grid)
			renderer.addLine(x, bottom, x, bottom+4, t.Axis)
			renderer.addText(fmt.Sprintf("%g", v), t.Axis, x, bottom+6, anchorCenter, false)
		}
	}

	// bars (only those inside the visible range)
	barW := float32(math.Max(float64(chart.MinBarWidth), float64(chart.pw)/(hi-lo)*0.8))
	n := len(chart.X)
	if len(chart.Y) < n {
		n = len(chart.Y)
	}
	for i := 0; i < n; i++ {
		x, y := chart.X[i], chart.Y[i]
		if y <= 0 || x < lo-1 || x > hi+1 {
			continue
		}
		cx := chart.xToPixel(x)
		left, right := cx-barW/2, cx+barW/2
		left, right = maxf(left, x0), minf(right, x1)
		if right <= left {
			continue
		}
		top := chart.yToPixel(math.Min(y, chart.YMax))
		rect := canvas.NewRectangle(color.Transparent)
		chart.styleBar(rect, false)
		rect.Move(fyne.NewPos(left, top))
		rect.Resize(fyne.NewSize(right-left, bottom-top))
		renderer.objects = append(renderer.objects, rect)
		renderer.bars[int(math.Round(x))] = rect
	}

	if chart.showY() {
		renderer.addLine(x0, chart.plotTop, x0, bottom, t.Axis)
	}
	if chart.showX() {
		renderer.addLine(x0, bottom, x1, bottom, t.Axis)
	}
}

// updateHover moves the cursor line, highlights the bar and updates the
// "position: value" text. It never rebuilds the plot, so it is cheap.
func (renderer *barChartRenderer) updateHover() {
	chart := renderer.chart
	if renderer.hovered != nil {
		chart.styleBar(renderer.hovered, false)
		renderer.hovered.Refresh()
		renderer.hovered = nil
	}

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
		renderer.cursor.Hide()
		renderer.value.Text = ""
		renderer.value.Resize(renderer.value.MinSize())
		renderer.value.Refresh()
		renderer.cursor.Refresh()
		return
	}

	px := chart.xToPixel(float64(idx))
	renderer.cursor.Position1 = fyne.NewPos(px, chart.plotTop)
	renderer.cursor.Position2 = fyne.NewPos(px, chart.plotTop+chart.ph)
	renderer.cursor.Show()
	renderer.cursor.Refresh()

	y := chart.lookup[idx] // 0 where there is no bar
	if chart.Format != nil {
		renderer.value.Text = chart.Format(idx, y)
	} else {
		renderer.value.Text = fmt.Sprintf("%d: %g", idx, y)
	}
	renderer.value.Resize(renderer.value.MinSize())
	renderer.value.Refresh()

	if rect, ok := renderer.bars[idx]; ok {
		chart.styleBar(rect, true)
		rect.Refresh()
		renderer.hovered = rect
	}
}

func (renderer *barChartRenderer) addLine(x1, y1, x2, y2 float32, col color.Color) {
	l := canvas.NewLine(col)
	l.StrokeWidth = 1
	l.Position1 = fyne.NewPos(x1, y1)
	l.Position2 = fyne.NewPos(x2, y2)
	renderer.objects = append(renderer.objects, l)
}

// addText places text so that (ax, y) is its left/centre/right anchor; y is the
// top of the text, or its vertical centre when vCenter is set.
func (renderer *barChartRenderer) addText(s string, col color.Color, ax, y float32, a anchor, vCenter bool) {
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
	renderer.objects = append(renderer.objects, t)
}
