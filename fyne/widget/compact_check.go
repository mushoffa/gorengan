package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	checkIconSize = 16
	checkGap      = 4
	checkPadY     = 2
)

// CompactCheck is a small check box with a label, sized for a chart header
// (Fyne's own Check is taller). The box is filled with the accent color when checked.
type CompactCheck struct {
	widget.BaseWidget
	label    string
	checked  bool
	onChange func(checked bool)
}

// NewCompactCheck creates an unchecked box. onChange is called only when the user toggles it.
func NewCompactCheck(label string, onChange func(checked bool)) *CompactCheck {
	c := &CompactCheck{label: label, onChange: onChange}
	c.ExtendBaseWidget(c)
	return c
}

func (c *CompactCheck) Checked() bool { return c.checked }

// SetChecked changes the state without calling onChange
// (use it to show a state that was set from code).
func (c *CompactCheck) SetChecked(checked bool) {
	c.checked = checked
	c.Refresh()
}

func (c *CompactCheck) Tapped(*fyne.PointEvent) {
	c.checked = !c.checked
	c.Refresh()
	if c.onChange != nil {
		c.onChange(c.checked)
	}
}

func (c *CompactCheck) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (c *CompactCheck) CreateRenderer() fyne.WidgetRenderer {
	r := &checkRenderer{
		c:    c,
		icon: canvas.NewImageFromResource(nil),
		text: canvas.NewText(c.label, color.Black),
	}
	r.icon.FillMode = canvas.ImageFillContain
	r.apply()
	return r
}

type checkRenderer struct {
	c    *CompactCheck
	icon *canvas.Image
	text *canvas.Text
}

func (r *checkRenderer) apply() {
	if r.c.checked {
		r.icon.Resource = theme.NewPrimaryThemedResource(theme.CheckButtonCheckedIcon())
	} else {
		r.icon.Resource = theme.CheckButtonIcon()
	}
	r.text.Color = theme.Color(theme.ColorNameForeground)
	r.text.TextSize = theme.TextSize() - 1
}

func (r *checkRenderer) Refresh() {
	r.apply()
	r.Layout(r.c.Size())
	r.icon.Refresh()
	r.text.Refresh()
}

func (r *checkRenderer) Layout(size fyne.Size) {
	r.icon.Resize(fyne.NewSize(checkIconSize, checkIconSize))
	r.icon.Move(fyne.NewPos(0, (size.Height-checkIconSize)/2))
	ts := r.text.MinSize()
	r.text.Resize(ts)
	r.text.Move(fyne.NewPos(checkIconSize+checkGap, (size.Height-ts.Height)/2))
}

func (r *checkRenderer) MinSize() fyne.Size {
	ts := r.text.MinSize()
	return fyne.NewSize(checkIconSize+checkGap+ts.Width, fyne.Max(ts.Height, checkIconSize)+2*checkPadY)
}

func (r *checkRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.icon, r.text} }
func (r *checkRenderer) Destroy()                     {}
