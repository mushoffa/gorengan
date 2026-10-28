package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	fynewidget "fyne.io/fyne/v2/widget"
)

// SidePanel is the custom container widget.
//
//	┌───────────────┐
//	│    Label      │  optional – no space is used when the text is ""
//	│ ┌───────────┐ │
//	│ │  Header   │ │  pinned to the top
//	│ └───────────┘ │
//	│ ┌───────────┐ │
//	│ │           │ │
//	│ │   Body    │ │  fills remaining space, scrolls when content overflows
//	│ │           │ │
//	│ └───────────┘ │
//	│ ┌───────────┐ │
//	│ │  Footer   │ │  pinned to the bottom
//	│ └───────────┘ │
//	└───────────────┘
//
// Every color, radius and padding comes from the active Fyne theme and is
// re-read on each Refresh, so light / dark / custom themes work automatically.
type SidePanel struct {
	fynewidget.BaseWidget

	label  *fynewidget.Label
	header fyne.CanvasObject // nil => header section is not shown
	footer fyne.CanvasObject // nil => footer section is not shown

	bodyBox       *fyne.Container   // VBox: items keep their min height and stack from the top
	scroll        *container.Scroll // wraps bodyBox, shows scrollbar only when needed
	bodyMinHeight float32
}

// NewSidePanel creates a SidePanel. Pass "" as label to hide the label completely and
// nil as header/footer to omit those sections.
func NewSidePanel(label string, header, footer fyne.CanvasObject, body ...fyne.CanvasObject) *SidePanel {
	s := &SidePanel{
		header:        header,
		footer:        footer,
		bodyMinHeight: 120,
	}

	s.label = fynewidget.NewLabelWithStyle(label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	s.label.Truncation = fyne.TextTruncateEllipsis

	// VBox gives every child only its MinSize height, so one item does not
	// stretch over the whole body; extra space stays empty below the items.
	s.bodyBox = container.NewVBox(body...)
	s.scroll = container.NewVScroll(container.NewPadded(s.bodyBox))

	s.ExtendBaseWidget(s)
	return s
}

// SetLabel changes the label. An empty string hides it (and its padding).
func (s *SidePanel) SetLabel(text string) {
	s.label.SetText(text)
	s.Refresh()
}

// SetHeader replaces the header content (nil removes the header section).
func (s *SidePanel) SetHeader(o fyne.CanvasObject) {
	s.header = o
	s.Refresh()
}

// SetFooter replaces the footer content (nil removes the footer section).
func (s *SidePanel) SetFooter(o fyne.CanvasObject) {
	s.footer = o
	s.Refresh()
}

// SetBody replaces all body items.
func (s *SidePanel) SetBody(items ...fyne.CanvasObject) {
	s.bodyBox.Objects = items
	s.bodyBox.Refresh()
	s.Refresh()
}

// AddBody appends items to the body.
func (s *SidePanel) AddBody(items ...fyne.CanvasObject) {
	for _, it := range items {
		s.bodyBox.Add(it)
	}
	s.Refresh()
}

// ClearBody removes all body items.
func (s *SidePanel) ClearBody() {
	s.bodyBox.Objects = nil
	s.bodyBox.Refresh()
	s.Refresh()
}

// SetBodyMinHeight sets the smallest height the body area may shrink to
// (default 120). The body scrolls when its content is taller than the area.
func (s *SidePanel) SetBodyMinHeight(h float32) {
	s.bodyMinHeight = h
	s.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (s *SidePanel) CreateRenderer() fyne.WidgetRenderer {
	r := &renderer{
		s:        s,
		frame:    canvas.NewRectangle(color.Transparent),
		headerBg: canvas.NewRectangle(color.Transparent),
		bodyBg:   canvas.NewRectangle(color.Transparent),
		footerBg: canvas.NewRectangle(color.Transparent),
	}
	r.applyTheme()
	return r
}

type renderer struct {
	s *SidePanel

	frame    *canvas.Rectangle // outer border + background
	headerBg *canvas.Rectangle
	bodyBg   *canvas.Rectangle
	footerBg *canvas.Rectangle
}

func (r *renderer) hasLabel() bool { return r.s.label.Text != "" }

// applyTheme (re)reads every colour/size from the current theme.
func (r *renderer) applyTheme() {
	radius := theme.Size(theme.SizeNameInputRadius)

	r.frame.FillColor = theme.Color(theme.ColorNameBackground)
	r.frame.StrokeColor = theme.Color(theme.ColorNameSeparator)
	r.frame.StrokeWidth = 1
	r.frame.CornerRadius = radius

	for _, box := range []*canvas.Rectangle{r.headerBg, r.bodyBg, r.footerBg} {
		box.FillColor = theme.Color(theme.ColorNameInputBackground)
		box.StrokeColor = theme.Color(theme.ColorNameInputBorder)
		box.StrokeWidth = 1
		box.CornerRadius = radius
	}
}

// Objects must return EVERY child (including user content) so Fyne can walk
// the tree and propagate theme changes to them.
func (r *renderer) Objects() []fyne.CanvasObject {
	objs := []fyne.CanvasObject{r.frame}
	if r.hasLabel() {
		objs = append(objs, r.s.label)
	}
	if r.s.header != nil {
		objs = append(objs, r.headerBg, r.s.header)
	}
	objs = append(objs, r.bodyBg, r.s.scroll)
	if r.s.footer != nil {
		objs = append(objs, r.footerBg, r.s.footer)
	}
	return objs
}

func (r *renderer) MinSize() fyne.Size {
	p := theme.Padding()
	s := r.s

	var width float32
	height := 2 * p
	rows := 0
	add := func(w, h float32) {
		if w > width {
			width = w
		}
		height += h
		rows++
	}

	if r.hasLabel() {
		m := s.label.MinSize()
		add(m.Width, m.Height)
	}
	if s.header != nil {
		m := s.header.MinSize()
		add(m.Width+2*p, m.Height+2*p)
	}

	// body: content width (+ padding + scrollbar + border), fixed minimum height
	bodyContent := s.scroll.Content.MinSize()
	add(bodyContent.Width+theme.Size(theme.SizeNameScrollBar)+2, s.bodyMinHeight)

	if s.footer != nil {
		m := s.footer.MinSize()
		add(m.Width+2*p, m.Height+2*p)
	}

	height += float32(rows-1) * p // gaps between rows (rows >= 1: body always exists)
	return fyne.NewSize(width+2*p, height)
}

func (r *renderer) Layout(size fyne.Size) {
	p := theme.Padding()
	s := r.s

	r.frame.Move(fyne.NewPos(0, 0))
	r.frame.Resize(size)

	x := p
	w := size.Width - 2*p
	y := p
	bottom := size.Height - p

	// Label – only takes space when it has text.
	if r.hasLabel() {
		h := s.label.MinSize().Height
		s.label.Move(fyne.NewPos(x, y))
		s.label.Resize(fyne.NewSize(w, h))
		y += h + p
	}

	// Header – pinned to top, min height only.
	if s.header != nil {
		h := s.header.MinSize().Height + 2*p
		placeBox(r.headerBg, s.header, x, y, w, h, p)
		y += h + p
	}

	// Footer – pinned to bottom, min height only.
	if s.footer != nil {
		h := s.footer.MinSize().Height + 2*p
		bottom -= h
		placeBox(r.footerBg, s.footer, x, bottom, w, h, p)
		bottom -= p
	}

	// Body – everything that is left.
	bh := bottom - y
	if bh < 0 {
		bh = 0
	}
	r.bodyBg.Move(fyne.NewPos(x, y))
	r.bodyBg.Resize(fyne.NewSize(w, bh))

	const inset = 1 // keep the scroll area inside the border stroke
	sw, sh := w-2*inset, bh-2*inset
	if sw < 0 {
		sw = 0
	}
	if sh < 0 {
		sh = 0
	}
	s.scroll.Move(fyne.NewPos(x+inset, y+inset))
	s.scroll.Resize(fyne.NewSize(sw, sh))
}

func (r *renderer) Refresh() {
	r.applyTheme()
	r.Layout(r.s.Size())

	r.frame.Refresh()
	r.headerBg.Refresh()
	r.bodyBg.Refresh()
	r.footerBg.Refresh()
	r.s.label.Refresh()
	r.s.scroll.Refresh()
}

func (r *renderer) Destroy() {}

// placeBox positions a background rectangle and its content (inset by pad).
func placeBox(bg *canvas.Rectangle, content fyne.CanvasObject, x, y, w, h, pad float32) {
	bg.Move(fyne.NewPos(x, y))
	bg.Resize(fyne.NewSize(w, h))
	content.Move(fyne.NewPos(x+pad, y+pad))
	content.Resize(fyne.NewSize(w-2*pad, h-2*pad))
}
