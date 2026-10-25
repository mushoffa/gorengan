package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	fynewidget "fyne.io/fyne/v2/widget"
)

// Surface draws a rounded, theme-colored backdrop behind a single piece of
// content and pads the content with an inset. It can optionally draw a
// border and a soft drop shadow.
//
//	s := widget.NewSurface(entry, theme.ColorNameInputBackground).
//		SetRadius(8).
//		SetBorder(theme.ColorNameInputBorder, 1).
//		SetShadow(3)
//
// The setters update the widget and refresh it. If a setter changes the
// widget's MinSize (inset, border, shadow, content), the parent container may
// need its own Refresh to pick up the new size.
type Surface struct {
	fynewidget.BaseWidget

	content     fyne.CanvasObject
	color       fyne.ThemeColorName
	borderColor fyne.ThemeColorName

	radius       float32
	inset        float32
	borderWidth  float32
	shadowOffset float32
}

// NewSurface returns a Surface filled with the given theme color,
// e.g. theme.ColorNameBackground or theme.ColorNameInputBackground.
// A nil content is allowed; the surface then renders as an empty backdrop.
func NewSurface(content fyne.CanvasObject, colorName fyne.ThemeColorName) *Surface {
	s := &Surface{
		color:       colorName,
		borderColor: theme.ColorNameInputBorder,
		radius:      12,
		inset:       10,
		content:     content,
	}
	if s.content == nil {
		// Invisible placeholder so the renderer never has to nil-check.
		s.content = canvas.NewRectangle(color.Transparent)
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetContent replaces the wrapped content (nil clears it).
func (s *Surface) SetContent(c fyne.CanvasObject) *Surface {
	if c == nil {
		c = canvas.NewRectangle(color.Transparent)
	}
	s.content = c
	s.Refresh()
	return s
}

// SetColor changes the fill to another theme color name.
func (s *Surface) SetColor(name fyne.ThemeColorName) *Surface {
	s.color = name
	s.Refresh()
	return s
}

// SetRadius sets the corner radius. Values below 0 are treated as 0.
func (s *Surface) SetRadius(r float32) *Surface {
	s.radius = fyne.Max(r, 0)
	s.Refresh()
	return s
}

// SetInset sets the space between the edge and the content.
// Values below 0 are treated as 0.
func (s *Surface) SetInset(in float32) *Surface {
	s.inset = fyne.Max(in, 0)
	s.Refresh()
	return s
}

// SetBorder draws an outline using a theme color. Width 0 disables it.
func (s *Surface) SetBorder(name fyne.ThemeColorName, width float32) *Surface {
	s.borderColor = name
	s.borderWidth = fyne.Max(width, 0)
	s.Refresh()
	return s
}

// SetShadow draws a soft shadow offset downward by the given amount.
// 0 disables it. The offset is reserved inside the widget's bounds.
func (s *Surface) SetShadow(offset float32) *Surface {
	s.shadowOffset = fyne.Max(offset, 0)
	s.Refresh()
	return s
}

// CreateRenderer implements fyne.Widget.
func (s *Surface) CreateRenderer() fyne.WidgetRenderer {
	r := &surfaceRenderer{
		s:      s,
		shadow: canvas.NewRectangle(color.Transparent),
		bg:     canvas.NewRectangle(color.Transparent),
	}
	r.applyStyle()
	return r
}

type surfaceRenderer struct {
	s      *Surface
	shadow *canvas.Rectangle
	bg     *canvas.Rectangle
}

// pad is the total distance from the widget edge to the content.
func (r *surfaceRenderer) pad() float32 { return r.s.inset + r.s.borderWidth }

func (r *surfaceRenderer) applyStyle() {
	s := r.s
	r.bg.FillColor = theme.Color(s.color)
	r.bg.CornerRadius = s.radius
	r.bg.StrokeWidth = s.borderWidth
	r.bg.StrokeColor = theme.Color(s.borderColor)

	r.shadow.FillColor = color.NRGBA{A: 45}
	r.shadow.CornerRadius = s.radius
	if s.shadowOffset > 0 {
		r.shadow.Show()
	} else {
		r.shadow.Hide()
	}
}

func (r *surfaceRenderer) Layout(size fyne.Size) {
	s := r.s
	bgSize := fyne.NewSize(size.Width, fyne.Max(size.Height-s.shadowOffset, 0))

	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(bgSize)

	r.shadow.Move(fyne.NewPos(0, s.shadowOffset))
	r.shadow.Resize(bgSize)

	p := r.pad()
	s.content.Move(fyne.NewPos(p, p))
	s.content.Resize(fyne.NewSize(
		fyne.Max(bgSize.Width-2*p, 0),
		fyne.Max(bgSize.Height-2*p, 0),
	))
}

func (r *surfaceRenderer) MinSize() fyne.Size {
	m := r.s.content.MinSize()
	p := r.pad()
	return fyne.NewSize(m.Width+2*p, m.Height+2*p+r.s.shadowOffset)
}

// Refresh re-reads theme colors (so light/dark switches update the card),
// re-applies styling and re-lays out in case inset, border or shadow changed.
func (r *surfaceRenderer) Refresh() {
	r.applyStyle()
	r.Layout(r.s.Size())
	r.shadow.Refresh()
	r.bg.Refresh()
	r.s.content.Refresh()
}

func (r *surfaceRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.shadow, r.bg, r.s.content}
}

func (r *surfaceRenderer) Destroy() {}
