package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	fynewidget "fyne.io/fyne/v2/widget"
)

const menuCornerRadius float32 = 8

var (
	_ fyne.Tappable      = (*MenuButton)(nil)
	_ desktop.Hoverable  = (*MenuButton)(nil)
	_ desktop.Cursorable = (*MenuButton)(nil)
)

// MenuButton is a flat, theme-aware button for a ButtonGroup.
// Idle: muted text. Hover: highlighted. Selected: card + bold text + arrow.
type MenuButton struct {
	fynewidget.BaseWidget

	Text     string
	Icon     fyne.Resource // optional leading icon
	OnTapped func()

	selected bool
	hovered  bool
	owner    *ButtonGroup // set when added to a group
}

// NewMenuButton creates a button with a label and tap callback.
func NewMenuButton(text string, onTapped func()) *MenuButton {
	b := &MenuButton{Text: text, OnTapped: onTapped}
	b.ExtendBaseWidget(b)
	return b
}

// NewMenuButtonWithIcon creates a button with a leading icon.
func NewMenuButtonWithIcon(text string, icon fyne.Resource, onTapped func()) *MenuButton {
	b := NewMenuButton(text, onTapped)
	b.Icon = icon
	return b
}

func (b *MenuButton) SetText(text string) {
	b.Text = text
	b.Refresh()
}

func (b *MenuButton) SetIcon(icon fyne.Resource) {
	b.Icon = icon
	b.Refresh()
}

// Selected reports whether the button is the selected one in its group.
func (b *MenuButton) Selected() bool { return b.selected }

// SetSelected selects/deselects the button. Inside a group, selection is
// exclusive: selecting this button deselects the previous one.
func (b *MenuButton) SetSelected(selected bool) {
	if b.owner != nil {
		if selected {
			b.owner.Select(b)
		} else if b.owner.selected == b {
			b.owner.Select(nil)
		}
		return
	}
	if b.selected == selected {
		return
	}
	b.selected = selected
	b.Refresh()
}

// Tapped implements fyne.Tappable.
func (b *MenuButton) Tapped(*fyne.PointEvent) {
	if b.owner != nil {
		b.owner.Select(b)
	}
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// desktop.Hoverable
func (b *MenuButton) MouseIn(*desktop.MouseEvent)    { b.hovered = true; b.Refresh() }
func (b *MenuButton) MouseMoved(*desktop.MouseEvent) {}
func (b *MenuButton) MouseOut()                      { b.hovered = false; b.Refresh() }

// desktop.Cursorable
func (b *MenuButton) Cursor() desktop.Cursor { return desktop.PointerCursor }

// palette uses the owning ButtonGroup's theme when one is set, else the app's.
func (b *MenuButton) palette() MenuPalette {
	if b.owner != nil && b.owner.customTheme != nil {
		return paletteFrom(b.owner.customTheme, fyne.CurrentApp().Settings().ThemeVariant())
	}
	return currentPalette()
}

func (b *MenuButton) CreateRenderer() fyne.WidgetRenderer {
	r := &menuButtonRenderer{
		b:     b,
		bg:    canvas.NewRectangle(color.Transparent),
		label: canvas.NewText(b.Text, color.Black),
		icon:  fynewidget.NewIcon(nil),
		arrow: fynewidget.NewIcon(theme.NavigateNextIcon()),
	}
	r.bg.CornerRadius = menuCornerRadius
	r.Refresh()
	return r
}

// ---------------------------------------------------------------------------

type menuButtonRenderer struct {
	b     *MenuButton
	bg    *canvas.Rectangle
	label *canvas.Text
	icon  *fynewidget.Icon
	arrow *fynewidget.Icon
}

func menuPadX() float32 { return theme.Padding() * 3 }
func menuPadY() float32 { return theme.Padding() * 2.5 }

func (r *menuButtonRenderer) MinSize() fyne.Size {
	is := theme.IconInlineSize()
	size := theme.TextSize()

	// Measure bold too, so the button never changes width on selection.
	reg := fyne.MeasureText(r.b.Text, size, fyne.TextStyle{})
	bold := fyne.MeasureText(r.b.Text, size, fyne.TextStyle{Bold: true})
	textW := fyne.Max(reg.Width, bold.Width)
	textH := fyne.Max(reg.Height, bold.Height)

	w := menuPadX()*2 + textW + theme.Padding()*2 + is // arrow slot always reserved
	if r.b.Icon != nil {
		w += is + theme.Padding()*2
	}
	h := fyne.Max(textH, is) + menuPadY()*2
	return fyne.NewSize(w, h)
}

func (r *menuButtonRenderer) Layout(size fyne.Size) {
	is := theme.IconInlineSize()
	iconSize := fyne.NewSize(is, is)

	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)

	x := menuPadX()
	if r.b.Icon != nil {
		r.icon.Resize(iconSize)
		r.icon.Move(fyne.NewPos(x, (size.Height-is)/2))
		x += is + theme.Padding()*2
	}

	ls := r.label.MinSize()
	r.label.Resize(ls)
	r.label.Move(fyne.NewPos(x, (size.Height-ls.Height)/2))

	r.arrow.Resize(iconSize)
	r.arrow.Move(fyne.NewPos(size.Width-menuPadX()-is, (size.Height-is)/2))
}

func (r *menuButtonRenderer) Refresh() {
	b := r.b
	p := b.palette()

	switch {
	case b.selected:
		r.bg.FillColor = p.SelectedBackground
		r.bg.StrokeColor = p.Divider
		r.bg.StrokeWidth = 1
	case b.hovered:
		r.bg.FillColor = p.Hover
		r.bg.StrokeWidth = 0
	default:
		r.bg.FillColor = color.Transparent
		r.bg.StrokeWidth = 0
	}

	r.label.Text = b.Text
	r.label.TextSize = theme.TextSize()
	switch {
	case b.selected:
		r.label.Color = p.SelectedText
		r.label.TextStyle = fyne.TextStyle{Bold: true}
	case b.hovered:
		r.label.Color = p.Text
		r.label.TextStyle = fyne.TextStyle{}
	default:
		r.label.Color = p.TextMuted
		r.label.TextStyle = fyne.TextStyle{}
	}

	if b.Icon != nil {
		r.icon.SetResource(b.Icon)
		r.icon.Show()
	} else {
		r.icon.Hide()
	}

	if b.selected {
		r.arrow.Show()
	} else {
		r.arrow.Hide()
	}

	r.Layout(b.Size())
	r.bg.Refresh()
	r.label.Refresh()
}

func (r *menuButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon, r.label, r.arrow}
}

func (r *menuButtonRenderer) Destroy() {}