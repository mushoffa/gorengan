package chart

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	gwidget "github.com/mushoffa/gorengan/fyne/widget"
)

// HistogramChannel is the data of one selectable channel.
type HistogramChannel struct {
	Name string
	Bins [256]int // number of pixels per level 0..255
}

// Histogram is a Photoshop-style histogram panel:
//
//	Channel: [ picker  v ]                      [x] Fill
//	┌────────────────────┐   dark plot (hover shows a line)
//	└────────────────────┘
//	 black-to-white bar with two range markers
//	 Mean   Std Dev   Median   Peak      Pixels   Level   Count   Percentile
//
// Level, Count and Percentile follow the mouse over the plot; with no mouse
// over it they describe the range selected with the two markers (nothing
// while the markers are at both ends). Percentile is the share of pixels at or
// below the level, or "from-to" for a range.
//
// Like LineChart, the "Fill" check box (top right) toggles the Fill field, and
// SetFill changes it from code.
type Histogram struct {
	// Fill fills the area under the histogram (default true); false draws only
	// the outline. The "Fill" check box toggles it; use SetFill to change it
	// from code so the check box and the plot follow.
	Fill bool

	channels []HistogramChannel
	selected int
	suspend  bool // true while the picker is changed from code
	lo, hi   int  // selected range
	hover    int  // level under the mouse, -1 for none

	picker *widget.Select
	plot   *histogramPlot
	bar    *levelBar

	fillCheck *gwidget.CompactCheck

	mean, stdDev, median, peak       *caption
	pixels, level, count, percentile *caption

	obj fyne.CanvasObject

	// OnFillChange is called when the user toggles the Fill check box
	// (not when SetFill is called from code).
	OnFillChange func(fill bool)
}

func NewHistogram() *Histogram {
	h := &Histogram{lo: 0, hi: 255, hover: -1, Fill: true}

	h.plot = newHistogramPlot()
	h.plot.OnHover = func(level int) {
		h.hover = level
		h.showReadout()
	}

	h.bar = newLevelBar()
	h.bar.OnChange = func(lo, hi int) {
		h.lo, h.hi = lo, hi
		h.plot.SetRange(lo, hi)
		h.showReadout()
	}

	h.picker = widget.NewSelect(nil, func(name string) {
		if h.suspend {
			return
		}
		for i, c := range h.channels {
			if c.Name == name {
				h.showChannel(i)
				return
			}
		}
	})
	h.picker.PlaceHolder = "-"
	h.picker.Disable()

	// two columns of four rows, values lined up within a column
	leftKey := measureWidth("Mean:", "Std Dev:", "Median:", "Peak:")
	rightKey := measureWidth("Pixels:", "Level:", "Count:", "Percentile:")
	meanRow, mean := newStat("Mean:", leftKey)
	stdRow, stdDev := newStat("Std Dev:", leftKey)
	medianRow, median := newStat("Median:", leftKey)
	peakRow, peak := newStat("Peak:", leftKey)
	pixelsRow, pixels := newStat("Pixels:", rightKey)
	levelRow, level := newStat("Level:", rightKey)
	countRow, count := newStat("Count:", rightKey)
	percentRow, percentile := newStat("Percentile:", rightKey)
	h.mean, h.stdDev, h.median, h.peak = mean, stdDev, median, peak
	h.pixels, h.level, h.count, h.percentile = pixels, level, count, percentile

	tight := layout.NewCustomPaddedVBoxLayout(0)
	stats := container.NewGridWithColumns(2,
		container.New(tight, meanRow, stdRow, medianRow, peakRow),
		container.New(tight, pixelsRow, levelRow, countRow, percentRow),
	)

	// The check box calls back only when the user toggles it.
	h.fillCheck = gwidget.NewCompactCheck("Fill", func(on bool) {
		h.Fill = on
		h.plot.SetFilled(on)
		if h.OnFillChange != nil {
			h.OnFillChange(on)
		}
	})
	h.fillCheck.SetChecked(h.Fill)
	h.plot.SetFilled(h.Fill)

	channelRow := container.NewBorder(nil, nil, newCaption("Channel:", theme.ColorNameForeground), h.fillCheck, h.picker)

	h.obj = container.NewVBox(channelRow, h.plot, h.bar, stats)
	h.clear()
	return h
}

// Object returns the widget to put on screen.
func (h *Histogram) Object() fyne.CanvasObject { return h.obj }

// SetFill turns the filled area on or off (same as the "Fill" check box, but
// it does not call OnFillChange).
func (h *Histogram) SetFill(on bool) {
	h.Fill = on
	h.fillCheck.SetChecked(on)
	h.plot.SetFilled(on)
}

// SetBins shows a single channel called "Luminosity".
func (h *Histogram) SetBins(bins [256]int) {
	h.SetChannels([]HistogramChannel{{Name: "Luminosity", Bins: bins}})
}

// SetChannels replaces the data. The channel the user picked stays selected
// when it still exists; an empty list clears the panel.
func (h *Histogram) SetChannels(channels []HistogramChannel) {
	h.channels = channels
	if len(channels) == 0 {
		h.suspend = true
		h.picker.Options = nil
		h.picker.ClearSelected()
		h.suspend = false
		h.picker.Disable()
		h.clear()
		return
	}

	names := make([]string, len(channels))
	for i, c := range channels {
		names[i] = c.Name
	}
	sel := h.selected
	if sel < 0 || sel >= len(channels) {
		sel = 0
	}

	h.suspend = true
	h.picker.Options = names
	h.picker.Enable()
	h.picker.SetSelectedIndex(sel)
	h.suspend = false

	h.showChannel(sel)
}

func (h *Histogram) showChannel(i int) {
	h.selected = i
	bins := h.channels[i].Bins
	s := computeStats(bins)

	h.plot.SetBins(bins)
	if s.Pixels == 0 {
		h.clear()
		return
	}
	h.mean.SetText(fmt.Sprintf("%.2f", s.Mean))
	h.stdDev.SetText(fmt.Sprintf("%.2f", s.StdDev))
	h.median.SetText(fmt.Sprint(s.Median))
	h.peak.SetText(fmt.Sprint(s.Peak))
	h.pixels.SetText(fmt.Sprint(s.Pixels))
	h.showReadout()
}

// showReadout fills Level / Count / Percentile from the mouse position, or
// from the marker range when the mouse is elsewhere.
func (h *Histogram) showReadout() {
	if len(h.channels) == 0 {
		h.setReadout("-", "-", "-")
		return
	}
	bins := h.channels[h.selected].Bins
	total := countIn(bins, 0, 255)
	if total == 0 {
		h.setReadout("-", "-", "-")
		return
	}

	switch {
	case h.hover >= 0:
		h.setReadout(
			fmt.Sprint(h.hover),
			fmt.Sprint(bins[h.hover]),
			fmt.Sprintf("%.2f%%", percent(countIn(bins, 0, h.hover), total)),
		)
	case h.lo > 0 || h.hi < 255:
		below := 0.0
		if h.lo > 0 {
			below = percent(countIn(bins, 0, h.lo-1), total)
		}
		h.setReadout(
			fmt.Sprintf("%d-%d", h.lo, h.hi),
			fmt.Sprint(countIn(bins, h.lo, h.hi)),
			fmt.Sprintf("%.1f-%.1f%%", below, percent(countIn(bins, 0, h.hi), total)),
		)
	default:
		h.setReadout("-", "-", "-")
	}
}

func (h *Histogram) setReadout(level, count, percentile string) {
	h.level.SetText(level)
	h.count.SetText(count)
	h.percentile.SetText(percentile)
}

// clear shows an empty plot and "-" for every number.
func (h *Histogram) clear() {
	h.plot.SetBins([256]int{})
	for _, c := range []*caption{h.mean, h.stdDev, h.median, h.peak, h.pixels} {
		c.SetText("-")
	}
	h.setReadout("-", "-", "-")
}