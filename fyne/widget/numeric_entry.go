package widget

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// numericEntry is a single line entry that only accepts numbers.
//
// Set AllowNegative to false to accept positive numbers only, or true to
// accept both positive and negative numbers. Decimals use '.' as the separator
// and can be switched off with AllowDecimal.
//
// Usage:
//
//	positiveOnly := newNumericEntry(false)
//	signed := newNumericEntry(true)
//
//	signed.OnChanged = func(s string) {
//		if v, err := signed.Value(); err == nil {
//			fmt.Println("value:", v)
//		}
//	}
type numericEntry struct {
	widget.Entry

	// AllowNegative lets the user type a leading '-'.
	// When false only positive numbers (and zero) are accepted.
	AllowNegative bool

	// AllowDecimal lets the user type one '.' decimal separator.
	// When false only whole numbers are accepted. newNumericEntry sets it to true.
	AllowDecimal bool
}

// newNumericEntry creates a numeric entry. If allowNegative is false the entry
// accepts positive numbers only, otherwise both positive and negative numbers.
func newNumericEntry(allowNegative bool) *numericEntry {
	e := &numericEntry{AllowNegative: allowNegative, AllowDecimal: true}
	e.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	e.ExtendBaseWidget(e)
	return e
}

// Value returns the entered number. It returns an error when the entry is
// empty or holds an incomplete number such as "-" or ".".
func (e *numericEntry) Value() (float64, error) {
	return strconv.ParseFloat(e.Text, 64)
}

// Bind connects a float data source to this entry. The current value is shown,
// changes in the data update the entry, and valid numbers typed by the user are
// written back to the data. Empty or incomplete text such as "-" or "." leaves
// the data at its last value.
func (e *numericEntry) Bind(data binding.Float) {
	e.Entry.Bind(floatText{data})
	e.Validator = nil // Entry.Bind installs a validator, which would bring the status icon back
}

// Unbind disconnects any data source. The text stays at the last value.
func (e *numericEntry) Unbind() {
	e.Entry.Unbind()
}

// floatText presents a binding.Float as a binding.String, formatting without
// trailing zeros ("3.5", not "3.500000") and parsing text back to a float.
type floatText struct {
	binding.Float
}

func (f floatText) Get() (string, error) {
	v, err := f.Float.Get()
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(v, 'f', -1, 64), nil
}

func (f floatText) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	return f.Float.Set(v)
}

// Keyboard asks mobile devices to show the numeric keyboard.
func (e *numericEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}

// SetText sets the text, ignoring the call if the text is not a valid number
// for the current settings.
func (e *numericEntry) SetText(text string) {
	if !e.isValidInput(text) {
		return
	}
	e.Entry.SetText(text)
}

// TypedRune filters typed characters so only valid numeric input reaches the entry.
func (e *numericEntry) TypedRune(r rune) {
	if !e.runeAllowed(r) {
		return
	}

	// Without a selection we can check the resulting text up front and
	// leave the entry (and its undo history) untouched when it is invalid.
	if e.SelectedText() == "" {
		runes := []rune(e.Text)
		pos := min(e.CursorColumn, len(runes))
		candidate := string(runes[:pos]) + string(r) + string(runes[pos:])
		if !e.isValidInput(candidate) {
			return
		}
		e.Entry.TypedRune(r)
		return
	}

	// With a selection the typed rune replaces it, so check afterwards.
	e.applyOrRevert(func() { e.Entry.TypedRune(r) })
}

// TypedShortcut lets everything through except paste, which is only accepted
// when the resulting text is still a valid number.
func (e *numericEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if _, ok := shortcut.(*fyne.ShortcutPaste); ok {
		e.applyOrRevert(func() { e.Entry.TypedShortcut(shortcut) })
		return
	}
	e.Entry.TypedShortcut(shortcut)
}

// applyOrRevert runs a change and rolls it back if it leaves invalid text in the entry.
// OnChanged is held back while this happens so listeners never see invalid text.
func (e *numericEntry) applyOrRevert(change func()) {
	prevText := e.Text
	prevRow, prevCol := e.CursorRow, e.CursorColumn

	cb := e.OnChanged
	e.OnChanged = nil

	change()
	if !e.isValidInput(e.Text) {
		e.Entry.SetText(prevText)
		e.CursorRow, e.CursorColumn = prevRow, prevCol
		e.Refresh()
	}

	e.OnChanged = cb
	if cb != nil && e.Text != prevText {
		cb(e.Text)
	}
}

// runeAllowed reports whether a typed character can ever be part of the number.
func (e *numericEntry) runeAllowed(r rune) bool {
	switch {
	case r >= '0' && r <= '9':
		return true
	case r == '-':
		return e.AllowNegative
	case r == '.':
		return e.AllowDecimal
	}
	return false
}

// isValidInput reports whether s is an acceptable number while typing.
// Incomplete forms such as "", "-", "." and "12." are allowed so the user can keep typing.
func (e *numericEntry) isValidInput(s string) bool {
	seenDot := false
	for i, r := range []rune(s) {
		switch {
		case r >= '0' && r <= '9':
		case r == '-' && i == 0 && e.AllowNegative:
		case r == '.' && e.AllowDecimal && !seenDot:
			seenDot = true
		default:
			return false
		}
	}
	return true
}

// NumericEntry is a numericEntry with a label shown on top of it.
//
// Usage:
//
//	age := NewNumericEntry("Age", false)
//	temp := NewNumericEntry("Temperature", true)
//
//	temp.OnChanged = func(v float64) { ... }
//	v, err := temp.Value()
type NumericEntry struct {
	widget.BaseWidget

	// Label is the text shown above the entry.
	Label *widget.Label
	// Entry is the underlying numeric entry.
	Entry *numericEntry

	// OnChanged is called with the new number whenever the user changes the value.
	// It is not called while the text is empty or incomplete (such as "-" or ".")
	OnChanged func(float64)

	// Spacing is the vertical gap between the label text and the entry.
	// It defaults to 0: the label text sits flush against the top-left of the
	// widget and the entry starts directly below it. Use a positive value to
	// add space. Call Refresh after changing it.
	Spacing float32

	// ThemeOverride is an optional theme for the label and entry. It is nil by
	// default, which means the widget follows the app theme. Set it right after
	// creating the widget (before it is shown), for example to NewCustomTheme().
	ThemeOverride fyne.Theme
}

// activeTheme returns the theme the label and entry are drawn with.
func (l *NumericEntry) activeTheme() fyne.Theme {
	if l.ThemeOverride != nil {
		return l.ThemeOverride
	}
	return l.Theme()
}

// NewNumericEntry creates a numeric entry with the given label on top.
// If allowNegative is false the entry accepts positive numbers only.
func NewNumericEntry(label string, allowNegative bool) *NumericEntry {
	l := &NumericEntry{
		Label: widget.NewLabel(label),
		Entry: newNumericEntry(allowNegative),
	}
	l.Entry.OnChanged = func(text string) {
		if l.OnChanged == nil {
			return
		}
		if v, err := strconv.ParseFloat(text, 64); err == nil {
			l.OnChanged(v)
		}
	}
	l.ExtendBaseWidget(l)
	return l
}

// NewNumericEntryWithData creates a numeric entry with the given label on top,
// connected to the data source.
//
//	price := binding.NewFloat()
//	entry := NewNumericEntryWithData("Price", false, price)
func NewNumericEntryWithData(label string, allowNegative bool, data binding.Float) *NumericEntry {
	l := NewNumericEntry(label, allowNegative)
	l.Bind(data)
	return l
}

// Bind connects a float data source to the entry. The current value is shown,
// changes in the data update the entry, and valid numbers typed by the user are
// written back to the data.
func (l *NumericEntry) Bind(data binding.Float) {
	l.Entry.Bind(data)
}

// Unbind disconnects any data source from the entry.
func (l *NumericEntry) Unbind() {
	l.Entry.Unbind()
}

// Value returns the entered number. It returns an error when the entry is
// empty or holds an incomplete number such as "-" or ".".
func (l *NumericEntry) Value() (float64, error) {
	return l.Entry.Value()
}

// SetLabel changes the label text.
func (l *NumericEntry) SetLabel(text string) {
	l.Label.SetText(text)
}

// CreateRenderer stacks the label above the entry.
func (l *NumericEntry) CreateRenderer() fyne.WidgetRenderer {
	content := fyne.CanvasObject(container.New(&labeledLayout{owner: l}, l.Label, l.Entry))
	if l.ThemeOverride != nil {
		content = container.NewThemeOverride(content, l.ThemeOverride)
	}
	return widget.NewSimpleRenderer(content)
}

// labeledLayout places a label (first object) above an entry (second object).
// A widget.Label draws its text inside the theme's inner padding, so the label is
// shifted left by that amount to line its text up with the entry's left border.
type labeledLayout struct {
	owner *NumericEntry
}

func (g *labeledLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	label, entry := objs[0], objs[1]
	pad := g.owner.activeTheme().Size(theme.SizeNameInnerPadding)
	labelHeight := label.MinSize().Height

	// Shift the label up and left by its inner padding so the text itself sits
	// flush with the top-left corner, then place the entry right under the text.
	label.Move(fyne.NewPos(-pad, -pad))
	label.Resize(fyne.NewSize(size.Width+pad, labelHeight))

	entry.Move(fyne.NewPos(0, labelHeight-2*pad+g.owner.Spacing))
	entry.Resize(fyne.NewSize(size.Width, entry.MinSize().Height))
}

func (g *labeledLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	labelMin, entryMin := objs[0].MinSize(), objs[1].MinSize()
	pad := g.owner.activeTheme().Size(theme.SizeNameInnerPadding)
	width := fyne.Max(entryMin.Width, labelMin.Width-pad)
	height := labelMin.Height - 2*pad + g.owner.Spacing + entryMin.Height
	return fyne.NewSize(width, height)
}
