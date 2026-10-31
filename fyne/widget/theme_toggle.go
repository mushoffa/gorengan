package widget

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	fwidget "fyne.io/fyne/v2/widget"
)

// ThemeMode is the state of a ThemeToggle.
type ThemeMode int

const (
	ModeLight ThemeMode = iota // default
	ModeDark
)

// ---------------------------------------------------------------------------
// Default icons
// ---------------------------------------------------------------------------

var defaultLightIcon = &fyne.StaticResource{
	StaticName: "sun.svg",
	StaticContent: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M7.5 12a4.5 4.5 0 1 0 9 0a4.5 4.5 0 1 0-9 0zM19.60 13.00L22.60 13.00L22.60 11.00L19.60 11.00zM16.67 18.08L18.79 20.20L20.20 18.79L18.08 16.67zM11.00 19.60L11.00 22.60L13.00 22.60L13.00 19.60zM5.92 16.67L3.80 18.79L5.21 20.20L7.33 18.08zM4.40 11.00L1.40 11.00L1.40 13.00L4.40 13.00zM7.33 5.92L5.21 3.80L3.80 5.21L5.92 7.33zM13.00 4.40L13.00 1.40L11.00 1.40L11.00 4.40zM18.08 7.33L20.20 5.21L18.79 3.80L16.67 5.92z"/></svg>`),
}

var defaultDarkIcon = &fyne.StaticResource{
	StaticName: "moon.svg",
	StaticContent: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>`),
}

// ---------------------------------------------------------------------------
// Colors
// ---------------------------------------------------------------------------

// ToggleColors is the palette used while the toggle is in one mode. Any nil
// field falls back to the built-in default for that mode, so a partial
// palette is fine.
type ToggleColors struct {
	Track        color.Color // rounded background behind both options
	TrackBorder  color.Color // outline of the track
	Thumb        color.Color // sliding highlight behind the selected option
	ThumbBorder  color.Color // outline of the thumb
	ActiveText   color.Color // label and icon of the selected option
	InactiveText color.Color // label and icon of the other option
}

// DefaultToggleColors returns the built-in palette for a mode.
func DefaultToggleColors(m ThemeMode) ToggleColors {
	if m == ModeDark {
		return ToggleColors{
			Track:        color.NRGBA{R: 0x17, G: 0x18, B: 0x1c, A: 0xff},
			TrackBorder:  color.NRGBA{R: 0x2a, G: 0x2b, B: 0x31, A: 0xff},
			Thumb:        color.NRGBA{R: 0x2b, G: 0x2c, B: 0x32, A: 0xff},
			ThumbBorder:  color.NRGBA{R: 0x3a, G: 0x3b, B: 0x42, A: 0xff},
			ActiveText:   color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff},
			InactiveText: color.NRGBA{R: 0x8e, G: 0x8e, B: 0x98, A: 0xff},
		}
	}
	return ToggleColors{
		Track:        color.NRGBA{R: 0xec, G: 0xec, B: 0xf0, A: 0xff},
		TrackBorder:  color.NRGBA{R: 0xec, G: 0xec, B: 0xf0, A: 0x00}, // none
		Thumb:        color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		ThumbBorder:  color.NRGBA{R: 0xe2, G: 0xe2, B: 0xe8, A: 0xff},
		ActiveText:   color.NRGBA{R: 0x11, G: 0x11, B: 0x14, A: 0xff},
		InactiveText: color.NRGBA{R: 0x6e, G: 0x6e, B: 0x78, A: 0xff},
	}
}

// mergeColors fills the nil fields of custom from base.
func mergeColors(custom *ToggleColors, base ToggleColors) ToggleColors {
	if custom == nil {
		return base
	}
	pick := func(c, d color.Color) color.Color {
		if c != nil {
			return c
		}
		return d
	}
	return ToggleColors{
		Track:        pick(custom.Track, base.Track),
		TrackBorder:  pick(custom.TrackBorder, base.TrackBorder),
		Thumb:        pick(custom.Thumb, base.Thumb),
		ThumbBorder:  pick(custom.ThumbBorder, base.ThumbBorder),
		ActiveText:   pick(custom.ActiveText, base.ActiveText),
		InactiveText: pick(custom.InactiveText, base.InactiveText),
	}
}

// lerpColor blends a into b; p is 0..1.
func lerpColor(a, b color.Color, p float32) color.Color {
	x := color.NRGBAModel.Convert(a).(color.NRGBA)
	y := color.NRGBAModel.Convert(b).(color.NRGBA)
	mix := func(u, v uint8) uint8 {
		return uint8(float32(u) + (float32(v)-float32(u))*p + 0.5)
	}
	return color.NRGBA{R: mix(x.R, y.R), G: mix(x.G, y.G), B: mix(x.B, y.B), A: mix(x.A, y.A)}
}

// ---------------------------------------------------------------------------
// Icon tinting (so icons follow the active / inactive text color)
// ---------------------------------------------------------------------------

var svgShapeTag = regexp.MustCompile(`<(path|circle|rect|ellipse|polygon|polyline|line)\b([^>]*)>`)

// tintIcon recolors an SVG resource. "currentColor" is replaced, and shapes
// without an explicit fill get one. Non-SVG resources (PNG, ...) are returned
// unchanged.
func tintIcon(res fyne.Resource, c color.Color) fyne.Resource {
	if res == nil || !strings.HasSuffix(strings.ToLower(res.Name()), ".svg") {
		return res
	}
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	hex := fmt.Sprintf("#%02x%02x%02x", n.R, n.G, n.B)

	s := strings.ReplaceAll(string(res.Content()), "currentColor", hex)
	s = svgShapeTag.ReplaceAllStringFunc(s, func(tag string) string {
		if strings.Contains(tag, "fill=") {
			return tag
		}
		m := svgShapeTag.FindStringSubmatch(tag)
		return "<" + m[1] + ` fill="` + hex + `"` + m[2] + ">"
	})

	// The name is part of Fyne's SVG cache key, so it must change with the color.
	name := strings.TrimSuffix(res.Name(), ".svg") + "-" + hex[1:] + ".svg"
	return fyne.NewStaticResource(name, []byte(s))
}

// ---------------------------------------------------------------------------
// ThemeToggle widget
// ---------------------------------------------------------------------------

// ThemeToggle is a two-option "Light | Dark" segmented control. It has its own
// light and dark look and does not depend on the application theme.
type ThemeToggle struct {
	fwidget.BaseWidget

	// OnChanged is optional. It is called after the app theme has been
	// switched, when the user taps the other option.
	OnChanged func(ThemeMode)

	settings    fyne.Settings
	lightTheme  fyne.Theme
	darkTheme   fyne.Theme
	mode        ThemeMode
	progress    float32 // 0 = light, 1 = dark; drives thumb position and color blend
	anim        *fyne.Animation
	lightIcon   fyne.Resource
	darkIcon    fyne.Resource
	lightLabel  string
	darkLabel   string
	lightColors *ToggleColors
	darkColors  *ToggleColors
}

// NewThemeToggle creates a toggle in light mode. It switches the application
// theme itself through settings (pass a.Settings()): Fyne's light theme for
// Light, the dark theme for Dark. It does not change the current theme until
// the mode changes.
func NewThemeToggle(settings fyne.Settings) *ThemeToggle {
	t := &ThemeToggle{
		settings:   settings,
		lightTheme: theme.LightTheme(),
		darkTheme:  theme.DarkTheme(),
		mode:       ModeLight,
		lightLabel: "Light",
		darkLabel:  "Dark",
	}
	t.ExtendBaseWidget(t)
	return t
}

// Mode returns the current mode.
func (t *ThemeToggle) Mode() ThemeMode { return t.mode }

// SetMode changes the mode programmatically and switches the app theme to
// match. It does not call OnChanged. Before the widget is first laid out the
// change is instant, afterwards it is animated.
func (t *ThemeToggle) SetMode(m ThemeMode) {
	if m == t.mode {
		return
	}
	t.mode = m
	t.applyAppTheme()
	t.animateTo(float32(m))
	t.Refresh()
}

// SetLightTheme sets the app theme applied in light mode (nil = theme.LightTheme()).
func (t *ThemeToggle) SetLightTheme(th fyne.Theme) {
	if th == nil {
		th = theme.LightTheme()
	}
	t.lightTheme = th
	if t.mode == ModeLight {
		t.applyAppTheme()
	}
}

// SetDarkTheme sets the app theme applied in dark mode (nil = theme.DarkTheme()).
func (t *ThemeToggle) SetDarkTheme(th fyne.Theme) {
	if th == nil {
		th = theme.DarkTheme()
	}
	t.darkTheme = th
	if t.mode == ModeDark {
		t.applyAppTheme()
	}
}

func (t *ThemeToggle) applyAppTheme() {
	if t.settings == nil {
		return
	}
	if t.mode == ModeDark {
		t.settings.SetTheme(t.darkTheme)
	} else {
		t.settings.SetTheme(t.lightTheme)
	}
}

// SetLightColors customises the palette used in light mode (nil = default).
func (t *ThemeToggle) SetLightColors(c *ToggleColors) {
	t.lightColors = c
	t.Refresh()
}

// SetDarkColors customises the palette used in dark mode (nil = default).
func (t *ThemeToggle) SetDarkColors(c *ToggleColors) {
	t.darkColors = c
	t.Refresh()
}

// SetLightIcon sets a custom icon for the Light option. Pass nil for the
// default sun. SVG icons are recolored to match the palette.
func (t *ThemeToggle) SetLightIcon(res fyne.Resource) {
	t.lightIcon = res
	t.Refresh()
}

// SetDarkIcon sets a custom icon for the Dark option. Pass nil for the
// default moon. SVG icons are recolored to match the palette.
func (t *ThemeToggle) SetDarkIcon(res fyne.Resource) {
	t.darkIcon = res
	t.Refresh()
}

// SetLabels changes the option texts (defaults: "Light" and "Dark").
func (t *ThemeToggle) SetLabels(light, dark string) {
	t.lightLabel, t.darkLabel = light, dark
	t.Refresh()
}

func (t *ThemeToggle) resolvedLightIcon() fyne.Resource {
	if t.lightIcon != nil {
		return t.lightIcon
	}
	return defaultLightIcon
}

func (t *ThemeToggle) resolvedDarkIcon() fyne.Resource {
	if t.darkIcon != nil {
		return t.darkIcon
	}
	return defaultDarkIcon
}

func (t *ThemeToggle) palettes() (light, dark ToggleColors) {
	return mergeColors(t.lightColors, DefaultToggleColors(ModeLight)),
		mergeColors(t.darkColors, DefaultToggleColors(ModeDark))
}

func (t *ThemeToggle) animateTo(target float32) {
	if t.anim != nil {
		t.anim.Stop()
		t.anim = nil
	}
	if t.Size().IsZero() { // not laid out yet: no animation
		t.progress = target
		return
	}
	start := t.progress
	t.anim = fyne.NewAnimation(160*time.Millisecond, func(p float32) {
		t.progress = start + (target-start)*p
		t.Refresh()
	})
	t.anim.Curve = fyne.AnimationEaseInOut
	t.anim.Start()
}

// Tapped implements fyne.Tappable: the left half selects Light, the right half Dark.
func (t *ThemeToggle) Tapped(ev *fyne.PointEvent) {
	next := ModeLight
	if ev.Position.X >= t.Size().Width/2 {
		next = ModeDark
	}
	if next == t.mode {
		return
	}
	t.SetMode(next) // switches the app theme
	if t.OnChanged != nil {
		t.OnChanged(next)
	}
}

// Cursor implements desktop.Cursorable.
func (t *ThemeToggle) Cursor() desktop.Cursor { return desktop.PointerCursor }

// CreateRenderer implements fyne.Widget.
func (t *ThemeToggle) CreateRenderer() fyne.WidgetRenderer {
	r := &themeToggleRenderer{
		t:        t,
		track:    canvas.NewRectangle(color.Transparent),
		thumb:    canvas.NewRectangle(color.Transparent),
		lightImg: canvas.NewImageFromResource(nil),
		darkImg:  canvas.NewImageFromResource(nil),
		lightLbl: canvas.NewText("", color.Black),
		darkLbl:  canvas.NewText("", color.Black),
	}
	r.track.StrokeWidth = 1
	r.thumb.StrokeWidth = 1
	r.lightImg.FillMode = canvas.ImageFillContain
	r.darkImg.FillMode = canvas.ImageFillContain
	r.Refresh()
	return r
}

// ---------------------------------------------------------------------------
// Renderer
// ---------------------------------------------------------------------------

const (
	toggleIconSize = float32(18)
	toggleGap      = float32(8)
	toggleHPad     = float32(18) // horizontal padding inside each option
	toggleVPad     = float32(10) // vertical padding inside each option
	toggleInset    = float32(4)  // gap between track edge and thumb
)

type themeToggleRenderer struct {
	t                 *ThemeToggle
	track, thumb      *canvas.Rectangle
	lightImg, darkImg *canvas.Image
	lightLbl, darkLbl *canvas.Text
}

func (r *themeToggleRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.track, r.thumb, r.lightImg, r.darkImg, r.lightLbl, r.darkLbl}
}

func (r *themeToggleRenderer) Destroy() {
	if r.t.anim != nil {
		r.t.anim.Stop()
	}
}

func (r *themeToggleRenderer) MinSize() fyne.Size {
	ts := theme.TextSize()
	bold := fyne.TextStyle{Bold: true}
	w := fyne.Max(
		fyne.MeasureText(r.t.lightLabel, ts, bold).Width,
		fyne.MeasureText(r.t.darkLabel, ts, bold).Width,
	)
	textH := fyne.MeasureText("Mg", ts, bold).Height

	half := toggleHPad*2 + toggleIconSize + toggleGap + w
	h := fyne.Max(toggleIconSize, textH) + toggleVPad*2 + toggleInset*2
	return fyne.NewSize(half*2+toggleInset*2, h)
}

func (r *themeToggleRenderer) Layout(size fyne.Size) {
	r.track.Resize(size)
	r.track.CornerRadius = size.Height / 2

	innerH := size.Height - toggleInset*2
	half := (size.Width - toggleInset*2) / 2

	r.thumb.Resize(fyne.NewSize(half, innerH))
	r.thumb.CornerRadius = innerH / 2
	r.thumb.Move(fyne.NewPos(toggleInset+r.t.progress*half, toggleInset))

	place := func(i int, img *canvas.Image, lbl *canvas.Text) {
		ls := lbl.MinSize()
		total := toggleIconSize + toggleGap + ls.Width
		x := toggleInset + float32(i)*half + (half-total)/2

		img.Resize(fyne.NewSize(toggleIconSize, toggleIconSize))
		img.Move(fyne.NewPos(x, (size.Height-toggleIconSize)/2))

		lbl.Resize(ls)
		lbl.Move(fyne.NewPos(x+toggleIconSize+toggleGap, (size.Height-ls.Height)/2))
	}
	place(0, r.lightImg, r.lightLbl)
	place(1, r.darkImg, r.darkLbl)
}

func (r *themeToggleRenderer) Refresh() {
	t := r.t
	pl, pd := t.palettes()
	p := t.progress

	// Surfaces blend from the light palette to the dark one as the thumb slides.
	r.track.FillColor = lerpColor(pl.Track, pd.Track, p)
	r.track.StrokeColor = lerpColor(pl.TrackBorder, pd.TrackBorder, p)
	r.thumb.FillColor = lerpColor(pl.Thumb, pd.Thumb, p)
	r.thumb.StrokeColor = lerpColor(pl.ThumbBorder, pd.ThumbBorder, p)

	// The Light option is "active" in the light palette and "inactive" in the dark one.
	lc := lerpColor(pl.ActiveText, pd.InactiveText, p)
	dc := lerpColor(pl.InactiveText, pd.ActiveText, p)

	style := func(active bool) fyne.TextStyle { return fyne.TextStyle{Bold: active} }

	r.lightLbl.Text, r.lightLbl.Color, r.lightLbl.TextStyle = t.lightLabel, lc, style(t.mode == ModeLight)
	r.darkLbl.Text, r.darkLbl.Color, r.darkLbl.TextStyle = t.darkLabel, dc, style(t.mode == ModeDark)
	r.lightLbl.TextSize = theme.TextSize()
	r.darkLbl.TextSize = theme.TextSize()

	setIcon := func(img *canvas.Image, res fyne.Resource, c color.Color) {
		tinted := tintIcon(res, c)
		if img.Resource == nil || img.Resource.Name() != tinted.Name() {
			img.Resource = tinted
			img.Refresh()
		}
	}
	setIcon(r.lightImg, t.resolvedLightIcon(), lc)
	setIcon(r.darkImg, t.resolvedDarkIcon(), dc)

	r.Layout(t.Size())
	r.track.Refresh()
	r.thumb.Refresh()
	r.lightLbl.Refresh()
	r.darkLbl.Refresh()
}