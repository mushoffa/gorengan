package widget

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	fynewidget "fyne.io/fyne/v2/widget"
)

// ---------------------------------------------------------------------------
// Self-contained styling: ButtonGroup does not use MenuPalette/currentPalette.

const groupCornerRadius float32 = 8

func groupPadX() float32 { return theme.Padding() * 3 }
func groupPadY() float32 { return theme.Padding() * 2.5 }

// groupColors are the only colours ButtonGroup needs.
type groupColors struct {
	accent     color.Color // header title
	hover      color.Color // header hover background
	divider    color.Color // vertical rail
	background color.Color // group background (painted only with a custom theme)
}

func groupColorsFrom(th fyne.Theme, v fyne.ThemeVariant) groupColors {
	return groupColors{
		accent:     th.Color(theme.ColorNamePrimary, v),
		hover:      th.Color(theme.ColorNameHover, v),
		divider:    th.Color(theme.ColorNameSeparator, v),
		background: th.Color(theme.ColorNameBackground, v),
	}
}

// ButtonGroup is a collapsible group: a clickable header (icon, title, chevron)
// above a list of buttons with a vertical rail on the left.
//
// Items can be any fyne.CanvasObject: *MenuButton (selectable, exclusive)
// or native Fyne widgets such as *widget.Button.
type ButtonGroup struct {
	fynewidget.BaseWidget

	// OnToggled is called after the group is expanded/collapsed.
	OnToggled func(expanded bool)
	// OnSelected is called when the selected MenuButton changes (nil = none).
	OnSelected func(*MenuButton)
	// AnimationDuration of expand/collapse. Set to 0 for instant.
	AnimationDuration time.Duration

	header      *buttonGroupHeader
	items       *fyne.Container
	rail        *canvas.Rectangle
	body        *fyne.Container
	clip        *container.Scroll // clips the body while it animates
	bg          *canvas.Rectangle // painted only when a custom theme is set
	content     *fyne.Container   // header + clip
	override    *container.ThemeOverride
	customTheme fyne.Theme // nil = follow the app theme
	selected    *MenuButton
	expanded bool
	progress float32 // 0 = collapsed, 1 = fully expanded
	anim     *fyne.Animation
}

// NewButtonGroup creates an expanded group. Items passed here are added in order.
func NewButtonGroup(title string, icon fyne.Resource, items ...fyne.CanvasObject) *ButtonGroup {
	g := &ButtonGroup{
		AnimationDuration: 180 * time.Millisecond,
		expanded:          true,
		progress:          1,
		items:             container.NewVBox(),
		rail:              canvas.NewRectangle(color.Transparent),
		bg:                canvas.NewRectangle(color.Transparent),
	}
	g.bg.CornerRadius = groupCornerRadius
	g.header = newButtonGroupHeader(title, icon, g.Toggle, g.colors)
	g.body = container.New(&groupBodyLayout{rail: g.rail, items: g.items}, g.rail, g.items)
	g.clip = container.NewScroll(g.body)
	g.clip.Direction = container.ScrollNone
	g.content = container.New(&groupLayout{g: g}, g.header, g.clip)
	// Fyne's theme override makes icons and native widgets inside the group
	// follow the group's theme. The proxy falls back to the app theme.
	g.override = container.NewThemeOverride(g.content, &groupThemeProxy{g: g})
	g.ExtendBaseWidget(g)

	g.appendItems(items...)
	return g
}

func (g *ButtonGroup) appendItems(items ...fyne.CanvasObject) {
	for _, it := range items {
		if it == nil {
			continue
		}
		if mb, ok := it.(*MenuButton); ok {
			mb.owner = g
		}
		g.items.Objects = append(g.items.Objects, it)
	}
}

// Add appends any widgets (MenuButton or native Fyne widgets).
func (g *ButtonGroup) Add(items ...fyne.CanvasObject) {
	g.appendItems(items...)
	g.items.Refresh()
	g.Refresh()
}

// AddButton creates a MenuButton, appends it and returns it.
func (g *ButtonGroup) AddButton(text string, onTapped func()) *MenuButton {
	b := NewMenuButton(text, onTapped)
	g.Add(b)
	return b
}

// Remove removes an item from the group.
func (g *ButtonGroup) Remove(item fyne.CanvasObject) {
	for i, o := range g.items.Objects {
		if o == item {
			g.items.Objects = append(g.items.Objects[:i], g.items.Objects[i+1:]...)
			break
		}
	}
	if mb, ok := item.(*MenuButton); ok {
		if g.selected == mb {
			g.Select(nil)
		}
		mb.owner = nil
	}
	g.items.Refresh()
	g.Refresh()
}

// SetTheme sets a theme used only by this group: its header, rail, background,
// MenuButtons, and (through Fyne's theme override) icons and native widgets.
// Pass e.g. NewCustomTheme(...), NewDarkTheme(), or any fyne.Theme.
// Pass nil to go back to following the app theme.
func (g *ButtonGroup) SetTheme(th fyne.Theme) {
	g.customTheme = th
	g.override.Refresh()
	g.Refresh()
	for _, o := range g.items.Objects {
		o.Refresh()
	}
}

// Theme returns the theme set with SetTheme, or nil when following the app.
func (g *ButtonGroup) Theme() fyne.Theme { return g.customTheme }

func (g *ButtonGroup) effectiveTheme() fyne.Theme {
	if g.customTheme != nil {
		return g.customTheme
	}
	return fyne.CurrentApp().Settings().Theme()
}

func (g *ButtonGroup) colors() groupColors {
	return groupColorsFrom(g.effectiveTheme(), fyne.CurrentApp().Settings().ThemeVariant())
}

// SetTitle changes the header text.
func (g *ButtonGroup) SetTitle(title string) {
	g.header.title = title
	g.Refresh()
}

// Select makes b the selected button (nil clears the selection).
func (g *ButtonGroup) Select(b *MenuButton) {
	if b != nil && b.owner != g {
		return
	}
	if g.selected == b {
		return
	}
	prev := g.selected
	g.selected = b
	if prev != nil {
		prev.selected = false
		prev.Refresh()
	}
	if b != nil {
		b.selected = true
		b.Refresh()
	}
	if g.OnSelected != nil {
		g.OnSelected(b)
	}
}

// Selected returns the selected button or nil.
func (g *ButtonGroup) Selected() *MenuButton { return g.selected }

// Expanded reports whether the group is open.
func (g *ButtonGroup) Expanded() bool { return g.expanded }

func (g *ButtonGroup) Expand()   { g.SetExpanded(true) }
func (g *ButtonGroup) Collapse() { g.SetExpanded(false) }
func (g *ButtonGroup) Toggle()   { g.SetExpanded(!g.expanded) }

// SetExpanded expands or collapses the group.
func (g *ButtonGroup) SetExpanded(expanded bool) {
	if g.expanded == expanded {
		return
	}
	g.expanded = expanded
	g.header.expanded = expanded
	g.animate(expanded)
	if g.OnToggled != nil {
		g.OnToggled(expanded)
	}
}

func (g *ButtonGroup) animate(open bool) {
	if g.anim != nil {
		g.anim.Stop()
		g.anim = nil
	}
	from, to := g.progress, float32(0)
	if open {
		to = 1
	}
	if g.AnimationDuration <= 0 {
		g.progress = to
		g.Refresh()
		return
	}
	g.anim = fyne.NewAnimation(g.AnimationDuration, func(f float32) {
		g.progress = from + (to-from)*f
		g.Refresh()
	})
	g.anim.Curve = fyne.AnimationEaseInOut
	g.anim.Start()
}

func (g *ButtonGroup) CreateRenderer() fyne.WidgetRenderer {
	r := &buttonGroupRenderer{g: g}
	r.Refresh()
	return r
}

// ---------------------------------------------------------------------------

type buttonGroupRenderer struct{ g *ButtonGroup }

func (r *buttonGroupRenderer) MinSize() fyne.Size { return r.g.override.MinSize() }

func (r *buttonGroupRenderer) Layout(size fyne.Size) {
	g := r.g
	g.bg.Move(fyne.NewPos(0, 0))
	g.bg.Resize(size)
	g.override.Move(fyne.NewPos(0, 0))
	g.override.Resize(size)
}

func (r *buttonGroupRenderer) Refresh() {
	g := r.g
	c := g.colors()
	g.rail.FillColor = c.divider
	g.rail.Refresh()

	if g.customTheme != nil {
		g.bg.FillColor = c.background
	} else {
		g.bg.FillColor = color.Transparent
	}
	g.bg.Refresh()

	want := g.progress > 0
	if g.clip.Visible() != want {
		if want {
			g.clip.Show()
		} else {
			g.clip.Hide()
		}
	}

	r.Layout(g.Size())
	g.header.Refresh()
	g.clip.Refresh()
}

func (r *buttonGroupRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.g.bg, r.g.override}
}

func (r *buttonGroupRenderer) Destroy() {}

// groupLayout stacks the header and the (animated, clipped) body.
type groupLayout struct{ g *ButtonGroup }

func (l *groupLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	hm := l.g.header.MinSize()
	bm := l.g.body.MinSize()
	return fyne.NewSize(fyne.Max(hm.Width, bm.Width), hm.Height+bm.Height*l.g.progress)
}

func (l *groupLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	g := l.g
	hm := g.header.MinSize()
	g.header.Move(fyne.NewPos(0, 0))
	g.header.Resize(fyne.NewSize(size.Width, hm.Height))

	bh := g.body.MinSize().Height * g.progress
	g.clip.Move(fyne.NewPos(0, hm.Height))
	g.clip.Resize(fyne.NewSize(size.Width, bh))
}

// groupThemeProxy is the theme given to Fyne's override: it always resolves to
// the group's custom theme, or to the app theme when none is set.
type groupThemeProxy struct{ g *ButtonGroup }

func (p *groupThemeProxy) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return p.g.effectiveTheme().Color(n, v)
}
func (p *groupThemeProxy) Font(s fyne.TextStyle) fyne.Resource {
	return p.g.effectiveTheme().Font(s)
}
func (p *groupThemeProxy) Icon(n fyne.ThemeIconName) fyne.Resource {
	return p.g.effectiveTheme().Icon(n)
}
func (p *groupThemeProxy) Size(n fyne.ThemeSizeName) float32 {
	return p.g.effectiveTheme().Size(n)
}

// ---------------------------------------------------------------------------
// Body layout: vertical rail on the left, items indented on the right.

type groupBodyLayout struct {
	rail  *canvas.Rectangle
	items *fyne.Container
}

func groupRailX() float32   { return groupPadX() + theme.IconInlineSize()/2 }
func groupItemsX() float32  { return groupRailX() + theme.Padding()*3 }
func groupBodyPad() float32 { return theme.Padding() }

func (l *groupBodyLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	m := l.items.MinSize()
	return fyne.NewSize(groupItemsX()+m.Width, m.Height+groupBodyPad()*2)
}

func (l *groupBodyLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	l.rail.Move(fyne.NewPos(groupRailX(), 0))
	l.rail.Resize(fyne.NewSize(1, size.Height))

	x := groupItemsX()
	l.items.Move(fyne.NewPos(x, groupBodyPad()))
	l.items.Resize(fyne.NewSize(size.Width-x, size.Height-groupBodyPad()*2))
}

// ---------------------------------------------------------------------------
// Header

var (
	_ fyne.Tappable      = (*buttonGroupHeader)(nil)
	_ desktop.Hoverable  = (*buttonGroupHeader)(nil)
	_ desktop.Cursorable = (*buttonGroupHeader)(nil)
)

type buttonGroupHeader struct {
	fynewidget.BaseWidget

	title    string
	icon     fyne.Resource
	expanded bool
	hovered  bool
	onTapped func()
	colors   func() groupColors
}

func newButtonGroupHeader(title string, icon fyne.Resource, onTapped func(), colors func() groupColors) *buttonGroupHeader {
	h := &buttonGroupHeader{title: title, icon: icon, expanded: true, onTapped: onTapped, colors: colors}
	h.ExtendBaseWidget(h)
	return h
}

func (h *buttonGroupHeader) Tapped(*fyne.PointEvent) {
	if h.onTapped != nil {
		h.onTapped()
	}
}

func (h *buttonGroupHeader) MouseIn(*desktop.MouseEvent)    { h.hovered = true; h.Refresh() }
func (h *buttonGroupHeader) MouseMoved(*desktop.MouseEvent) {}
func (h *buttonGroupHeader) MouseOut()                      { h.hovered = false; h.Refresh() }
func (h *buttonGroupHeader) Cursor() desktop.Cursor         { return desktop.PointerCursor }

func (h *buttonGroupHeader) CreateRenderer() fyne.WidgetRenderer {
	r := &buttonGroupHeaderRenderer{
		h:       h,
		bg:      canvas.NewRectangle(color.Transparent),
		label:   canvas.NewText(h.title, color.Black),
		icon:    fynewidget.NewIcon(nil),
		chevron: fynewidget.NewIcon(theme.MenuDropUpIcon()),
	}
	r.bg.CornerRadius = groupCornerRadius
	r.Refresh()
	return r
}

type buttonGroupHeaderRenderer struct {
	h       *buttonGroupHeader
	bg      *canvas.Rectangle
	label   *canvas.Text
	icon    *fynewidget.Icon
	chevron *fynewidget.Icon
}

func (r *buttonGroupHeaderRenderer) MinSize() fyne.Size {
	is := theme.IconInlineSize()
	ts := fyne.MeasureText(r.h.title, theme.TextSize(), fyne.TextStyle{Bold: true})

	w := groupPadX()*2 + ts.Width + theme.Padding()*2 + is
	if r.h.icon != nil {
		w += is + theme.Padding()*2
	}
	return fyne.NewSize(w, fyne.Max(ts.Height, is)+groupPadY()*2)
}

func (r *buttonGroupHeaderRenderer) Layout(size fyne.Size) {
	is := theme.IconInlineSize()
	iconSize := fyne.NewSize(is, is)

	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)

	x := groupPadX()
	if r.h.icon != nil {
		r.icon.Resize(iconSize)
		r.icon.Move(fyne.NewPos(x, (size.Height-is)/2))
		x += is + theme.Padding()*2
	}

	ls := r.label.MinSize()
	r.label.Resize(ls)
	r.label.Move(fyne.NewPos(x, (size.Height-ls.Height)/2))

	r.chevron.Resize(iconSize)
	r.chevron.Move(fyne.NewPos(size.Width-groupPadX()-is, (size.Height-is)/2))
}

func (r *buttonGroupHeaderRenderer) Refresh() {
	h := r.h
	c := h.colors()

	if h.hovered {
		r.bg.FillColor = c.hover
	} else {
		r.bg.FillColor = color.Transparent
	}

	r.label.Text = h.title
	r.label.TextSize = theme.TextSize()
	r.label.TextStyle = fyne.TextStyle{Bold: true}
	r.label.Color = c.accent

	if h.icon != nil {
		r.icon.SetResource(theme.NewPrimaryThemedResource(h.icon)) // follows theme accent
		r.icon.Show()
	} else {
		r.icon.Hide()
	}

	if h.expanded {
		r.chevron.SetResource(theme.MenuDropUpIcon())
	} else {
		r.chevron.SetResource(theme.MenuDropDownIcon())
	}

	r.Layout(h.Size())
	r.bg.Refresh()
	r.label.Refresh()
}

func (r *buttonGroupHeaderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon, r.label, r.chevron}
}

func (r *buttonGroupHeaderRenderer) Destroy() {}