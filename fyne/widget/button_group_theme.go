package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MenuPalette holds every colour used by MenuButton and ButtonGroup.
// Any nil field in a custom palette falls back to the light/dark default.
type MenuPalette struct {
	Background         color.Color // window / sidebar background
	Accent             color.Color // group header text + icon
	Text               color.Color // normal text (hover)
	TextMuted          color.Color // idle button text
	SelectedBackground color.Color // selected button card
	SelectedText       color.Color // selected button text
	Hover              color.Color // hover background
	Divider            color.Color // rail line + selected card border
}

func rgb(hex uint32) color.Color {
	return color.NRGBA{R: uint8(hex >> 16), G: uint8(hex >> 8), B: uint8(hex), A: 0xFF}
}

// LightPalette is the default light palette.
func LightPalette() MenuPalette {
	return MenuPalette{
		Background:         rgb(0xF4F5F7),
		Accent:             rgb(0x1E90FF),
		Text:               rgb(0x1F2937),
		TextMuted:          rgb(0x6B7280),
		SelectedBackground: rgb(0xFFFFFF),
		SelectedText:       rgb(0x111827),
		Hover:              rgb(0xE5E7EB),
		Divider:            rgb(0xD1D5DB),
	}
}

// DarkPalette is the default dark palette.
func DarkPalette() MenuPalette {
	return MenuPalette{
		Background:         rgb(0x1F2430),
		Accent:             rgb(0x4DA3FF),
		Text:               rgb(0xE5E7EB),
		TextMuted:          rgb(0x9CA3AF),
		SelectedBackground: rgb(0x2B3140),
		SelectedText:       rgb(0xFFFFFF),
		Hover:              rgb(0x272D3A),
		Divider:            rgb(0x3A4152),
	}
}

func (p MenuPalette) merge(def MenuPalette) MenuPalette {
	pick := func(c, d color.Color) color.Color {
		if c == nil {
			return d
		}
		return c
	}
	return MenuPalette{
		Background:         pick(p.Background, def.Background),
		Accent:             pick(p.Accent, def.Accent),
		Text:               pick(p.Text, def.Text),
		TextMuted:          pick(p.TextMuted, def.TextMuted),
		SelectedBackground: pick(p.SelectedBackground, def.SelectedBackground),
		SelectedText:       pick(p.SelectedText, def.SelectedText),
		Hover:              pick(p.Hover, def.Hover),
		Divider:            pick(p.Divider, def.Divider),
	}
}

func isDark(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535
	return lum < 0.5
}

// MenuTheme is a fyne.Theme that also exposes a MenuPalette.
// The palette colours are mapped onto Fyne's standard colour names too,
// so the rest of the app (native buttons, entries...) matches.
type MenuTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
	forced  bool // false = follow the OS light/dark setting
	custom  *MenuPalette
}

var _ fyne.Theme = (*MenuTheme)(nil)

// NewLightTheme always uses the light palette.
func NewLightTheme() *MenuTheme {
	return &MenuTheme{base: theme.DefaultTheme(), variant: theme.VariantLight, forced: true}
}

// NewDarkTheme always uses the dark palette.
func NewDarkTheme() *MenuTheme {
	return &MenuTheme{base: theme.DefaultTheme(), variant: theme.VariantDark, forced: true}
}

// NewSystemTheme follows the operating system's light/dark setting.
func NewSystemTheme() *MenuTheme {
	return &MenuTheme{base: theme.DefaultTheme()}
}

// NewCustomTheme uses your own palette. Unset (nil) colours get defaults.
func NewCustomTheme(p MenuPalette) *MenuTheme {
	def := LightPalette()
	if p.Background != nil && isDark(p.Background) {
		def = DarkPalette()
	}
	p = p.merge(def)

	variant := theme.VariantLight
	if isDark(p.Background) {
		variant = theme.VariantDark
	}
	return &MenuTheme{base: theme.DefaultTheme(), variant: variant, forced: true, custom: &p}
}

// MenuPalette returns the palette for the given variant.
func (t *MenuTheme) MenuPalette(v fyne.ThemeVariant) MenuPalette {
	if t.custom != nil {
		return *t.custom
	}
	if t.forced {
		v = t.variant
	}
	if v == theme.VariantLight {
		return LightPalette()
	}
	return DarkPalette()
}

func (t *MenuTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if t.forced {
		v = t.variant
	}
	p := t.MenuPalette(v)
	switch name {
	case theme.ColorNameBackground:
		return p.Background
	case theme.ColorNameForeground:
		return p.Text
	case theme.ColorNamePrimary:
		return p.Accent
	case theme.ColorNameHover:
		return p.Hover
	case theme.ColorNameSeparator:
		return p.Divider
	}
	return t.base.Color(name, v)
}

func (t *MenuTheme) Font(s fyne.TextStyle) fyne.Resource     { return t.base.Font(s) }
func (t *MenuTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return t.base.Icon(n) }
func (t *MenuTheme) Size(n fyne.ThemeSizeName) float32       { return t.base.Size(n) }

// currentPalette returns the palette of the active theme. If the app uses a
// theme that is not a MenuTheme, the palette is derived from standard colours.
func currentPalette() MenuPalette {
	s := fyne.CurrentApp().Settings()
	return paletteFrom(s.Theme(), s.ThemeVariant())
}

// paletteFrom converts any theme into a MenuPalette.
func paletteFrom(th fyne.Theme, v fyne.ThemeVariant) MenuPalette {
	if mt, ok := th.(interface {
		MenuPalette(fyne.ThemeVariant) MenuPalette
	}); ok {
		return mt.MenuPalette(v)
	}

	fg := th.Color(theme.ColorNameForeground, v)
	return MenuPalette{
		Background:         th.Color(theme.ColorNameBackground, v),
		Accent:             th.Color(theme.ColorNamePrimary, v),
		Text:               fg,
		TextMuted:          th.Color(theme.ColorNamePlaceHolder, v),
		SelectedBackground: th.Color(theme.ColorNameInputBackground, v),
		SelectedText:       fg,
		Hover:              th.Color(theme.ColorNameHover, v),
		Divider:            th.Color(theme.ColorNameSeparator, v),
	}
}