package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	fynewidget "fyne.io/fyne/v2/widget"
)

const (
	// optionVPad is the vertical padding inside each option (above and below
	// the text). Set to 0 for the tightest rows.
	optionVPad float32 = 2

	// headerGap is the space between the header and the first row.
	headerGap float32 = 4
)

type option struct {
	fynewidget.BaseWidget
	text     string
	selected bool
	hovered  bool
	onTap    func()
}

func newOption(text string, onTap func()) *option {
	o := &option{text: text, onTap: onTap}
	o.ExtendBaseWidget(o)
	return o
}

func (o *option) setSelected(s bool) {
	if o.selected != s {
		o.selected = s
		o.Refresh()
	}
}

func (o *option) Tapped(*fyne.PointEvent)       { o.onTap() }
func (o *option) MouseIn(*desktop.MouseEvent)    { o.hovered = true; o.Refresh() }
func (o *option) MouseMoved(*desktop.MouseEvent) {}
func (o *option) MouseOut()                      { o.hovered = false; o.Refresh() }
func (o *option) Cursor() desktop.Cursor         { return desktop.PointerCursor }

func (o *option) CreateRenderer() fyne.WidgetRenderer {
	txt := canvas.NewText(o.text, theme.Color(theme.ColorNameForeground))
	txt.TextSize = theme.TextSize()

	r := &optionRenderer{
		o:    o,
		bg:   canvas.NewRectangle(color.Transparent),
		icon: fynewidget.NewIcon(theme.RadioButtonIcon()),
		text: txt,
	}
	r.bg.CornerRadius = theme.InputRadiusSize()
	r.Refresh()
	return r
}

type optionRenderer struct {
	o    *option
	bg   *canvas.Rectangle
	icon *fynewidget.Icon
	text *canvas.Text
}

func (r *optionRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)

	is := theme.IconInlineSize()
	r.icon.Resize(fyne.NewSquareSize(is))
	r.icon.Move(fyne.NewPos(theme.Padding(), (size.Height-is)/2))

	ts := r.text.MinSize()
	lx := theme.Padding()*2 + is
	r.text.Resize(ts)
	r.text.Move(fyne.NewPos(lx, (size.Height-ts.Height)/2))
}

func (r *optionRenderer) MinSize() fyne.Size {
	is := theme.IconInlineSize()
	ts := r.text.MinSize()
	return fyne.NewSize(
		theme.Padding()*2+is+ts.Width+theme.Padding(), // icon gap + text + right margin
		fyne.Max(is, ts.Height)+2*optionVPad,
	)
}

func (r *optionRenderer) Refresh() {
	if r.o.selected {
		r.icon.SetResource(theme.NewPrimaryThemedResource(theme.RadioButtonCheckedIcon()))
	} else {
		r.icon.SetResource(theme.NewThemedResource(theme.RadioButtonIcon()))
	}

	r.text.Text = r.o.text
	r.text.TextSize = theme.TextSize()
	r.text.Color = theme.Color(theme.ColorNameForeground)
	r.text.Refresh()

	if r.o.hovered {
		r.bg.FillColor = theme.Color(theme.ColorNameHover)
	} else {
		r.bg.FillColor = color.Transparent
	}
	r.bg.CornerRadius = theme.InputRadiusSize()
	r.bg.Refresh()
	r.icon.Refresh()
}

func (r *optionRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon, r.text}
}
func (r *optionRenderer) Destroy() {}

// RadioGroup is a titled, theme-aware, adaptive radio button group.
type RadioGroup struct {
	fynewidget.BaseWidget

	Title string
	// OnChanged receives the zero-based index of the selected option.
	OnChanged func(index uint)

	// RowSpacing is the vertical gap between options.
	// Negative means "use the theme default" (theme.Padding()/2).
	RowSpacing float32

	// MaxColumns limits the column count chosen by the adaptive layout.
	// 0 = automatic (one row, then two columns, then one column),
	// 1 = always a single column.
	MaxColumns int

	options  []*option
	selected uint
	cols     int // last column count chosen by the layout
}

// New creates a group. The first option is selected by default.
func NewRadioGroup(title string, labels []string, onChanged func(index uint)) *RadioGroup {
	g := &RadioGroup{
		Title:      title,
		OnChanged:  onChanged,
		cols:       1,
		RowSpacing: -1, // follow the theme
	}
	for i, l := range labels {
		idx := uint(i)
		g.options = append(g.options, newOption(l, func() { g.userSelect(idx) }))
	}
	if len(g.options) > 0 {
		g.options[0].selected = true
	}
	g.ExtendBaseWidget(g)
	return g
}

// SetRowSpacing sets the vertical gap between options and re-lays out the
// group. Pass a negative value to go back to the theme default.
func (g *RadioGroup) SetRowSpacing(spacing float32) {
	g.RowSpacing = spacing
	g.Refresh()
}

// SetMaxColumns limits how many columns the layout may use (0 = automatic,
// 1 = always a single column) and re-lays out the group.
func (g *RadioGroup) SetMaxColumns(n int) {
	g.MaxColumns = n
	g.Refresh()
}

// Selected returns the selected index.
func (g *RadioGroup) Selected() uint { return g.selected }

// SetSelected selects an option without firing the callback.
func (g *RadioGroup) SetSelected(index uint) {
	if int(index) >= len(g.options) {
		return
	}
	g.apply(index)
}

func (g *RadioGroup) userSelect(index uint) {
	if index == g.selected {
		return
	}
	g.apply(index)
	if g.OnChanged != nil {
		g.OnChanged(index)
	}
}

func (g *RadioGroup) apply(index uint) {
	g.selected = index
	for i, o := range g.options {
		o.setSelected(uint(i) == index)
	}
}

func (g *RadioGroup) CreateRenderer() fyne.WidgetRenderer {
	header := canvas.NewText(g.Title, theme.Color(theme.ColorNameForeground))
	header.TextStyle = fyne.TextStyle{Bold: true}
	header.TextSize = theme.TextSize()

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = theme.InputRadiusSize()

	r := &groupRenderer{g: g, bg: bg, header: header}
	r.objs = append(r.objs, bg, header)
	for _, o := range g.options {
		r.objs = append(r.objs, o)
	}
	return r
}

type groupRenderer struct {
	g      *RadioGroup
	bg     *canvas.Rectangle
	header *canvas.Text
	objs   []fyne.CanvasObject
}

// rowGap is the vertical gap between rows. Layout and MinSize both use it,
// so the group's height always matches its content.
func (r *groupRenderer) rowGap() float32 {
	if r.g.RowSpacing >= 0 {
		return r.g.RowSpacing
	}
	return theme.Padding() / 2
}

// columnWidths returns per-column widths and the row count for a column count.
func (r *groupRenderer) columnWidths(cols int) (widths []float32, rows int) {
	n := len(r.g.options)
	rows = (n + cols - 1) / cols
	widths = make([]float32, cols)
	for i, o := range r.g.options {
		c := i / rows // column-major fill, like the reference image
		widths[c] = fyne.Max(widths[c], o.MinSize().Width)
	}
	return
}

func (r *groupRenderer) totalWidth(cols int) float32 {
	w, _ := r.columnWidths(cols)
	var sum float32
	for _, v := range w {
		sum += v
	}
	return sum + float32(cols-1)*theme.Padding()*2
}

// pickColumns: single row if it fits, else two columns, else one column,
// never exceeding MaxColumns when it is set.
func (r *groupRenderer) pickColumns(avail float32) int {
	n := len(r.g.options)
	if n <= 1 || r.g.MaxColumns == 1 {
		return 1
	}
	for _, c := range []int{n, 2} {
		if r.g.MaxColumns > 0 && c > r.g.MaxColumns {
			continue
		}
		if r.totalWidth(c) <= avail {
			return c
		}
	}
	return 1
}

func (r *groupRenderer) Layout(size fyne.Size) {
	inset := theme.InnerPadding()
	p := theme.Padding()
	gap := r.rowGap()
	r.bg.Resize(size)

	hs := r.header.MinSize()
	r.header.Resize(hs)
	r.header.Move(fyne.NewPos(inset, inset))

	avail := size.Width - 2*inset
	cols := r.pickColumns(avail)
	widths, rows := r.columnWidths(cols)

	top := inset + hs.Height + headerGap
	x := inset
	for c := 0; c < cols; c++ {
		y := top
		for row := 0; row < rows; row++ {
			i := c*rows + row
			if i >= len(r.g.options) {
				break
			}
			o := r.g.options[i]
			h := o.MinSize().Height
			o.Move(fyne.NewPos(x, y))
			o.Resize(fyne.NewSize(widths[c], h))
			y += h + gap
		}
		x += widths[c] + p*2
	}

	// Height depends on column count; ask the parent to re-measure if it changed.
	if cols != r.g.cols {
		r.g.cols = cols
		r.g.Refresh()
	}
}

func (r *groupRenderer) MinSize() fyne.Size {
	inset := theme.InnerPadding()
	hs := r.header.MinSize()
	n := len(r.g.options)
	if n == 0 {
		return fyne.NewSize(hs.Width+2*inset, hs.Height+2*inset)
	}
	cols := r.g.cols
	if cols < 1 {
		cols = 1
	}
	rows := (n + cols - 1) / cols
	oh := r.g.options[0].MinSize().Height

	// Width: widest single column, so the parent can shrink the group freely.
	w := fyne.Max(hs.Width, r.totalWidth(1))
	h := inset + hs.Height + headerGap +
		float32(rows)*oh + float32(rows-1)*r.rowGap() +
		inset
	return fyne.NewSize(w+2*inset, h)
}

func (r *groupRenderer) Refresh() {
	r.bg.FillColor = theme.Color(theme.ColorNameInputBackground)
	r.bg.CornerRadius = theme.InputRadiusSize()
	r.bg.Refresh()

	r.header.Text = r.g.Title
	r.header.TextSize = theme.TextSize()
	r.header.Color = theme.Color(theme.ColorNameForeground)
	r.header.Refresh()

	for _, o := range r.g.options {
		o.Refresh()
	}
	r.Layout(r.g.Size()) // re-apply spacing and column changes immediately
}

func (r *groupRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *groupRenderer) Destroy()                       {}