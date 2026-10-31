package widget

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	fynewidget "fyne.io/fyne/v2/widget"
)

const (
	maxZoom  = 32.0
	zoomStep = 1.25 // zoom factor for one wheel notch or one ZoomIn/ZoomOut call

	// scrollUnit is the Fyne scroll delta that counts as one wheel notch.
	// Smooth-scroll devices send smaller deltas and get proportionally less zoom.
	// If wheel zoom feels too fast or slow on your platform, tune this value.
	scrollUnit     = 10.0
	maxScrollSteps = 3.0 // cap on notches applied by a single scroll event

	dividerGrab = 12 // how close (Fyne units) a drag must start to the divider to move it

	maxSupersample = 4 // max n for the n×n box filter used when downscaling

	defaultStageRadius = 12.0 // Fyne units

	overlayMargin = 8 // gap between the built-in info bar / buttons and the view edge

	defaultMinW, defaultMinH     = 320, 240 // normal minimum size, Fyne units
	defaultThumbW, defaultThumbH = 160, 120 // default thumbnail size, Fyne units

	previewMaxW, previewMaxH = 960, 720 // largest default size of the preview window
)

const (
	dragNone = iota
	dragPan
	dragDivider
)

// ImageViewer is a zoomable, pannable image widget with an optional before/after
// comparison divider and an optional rounded "stage" vignette behind the image.
// It carries its own overlay: an info bar (zoom, pixel position, gray/RGB value)
// at the bottom center and a vertical strip of zoom in / zoom out / fit buttons
// at the left edge, centered vertically.
// SetOverlay hides either part. SetDisableZoom turns zoom off entirely and
// hides both parts, since they only serve zooming.
//
// Thumbnail mode (SetEnableThumbnail) shows the viewer at thumbnail size with
// zoom disabled and the overlay hidden. Tapping it opens a new window with a
// fully featured viewer of the same image.
//
// Threading: like every Fyne widget, all methods must be called on the UI
// thread. From a goroutine (e.g. an OCR pipeline), wrap calls in fyne.Do.
type ImageViewer struct {
	fynewidget.BaseWidget

	img     image.Image // shown everywhere, or only right of the divider when comparing
	before  image.Image // shown left of the divider when comparing
	compare bool
	split   float64 // divider position, 0..1 of the widget width

	zoom   float64 // 1 = fit to window
	cx, cy float64 // image coordinate shown at the center of the view

	dragMode int
	lastPct  float64
	raster   *canvas.Raster

	downscaleFilter bool // box-filter when the image is shown smaller than 1:1

	// zoomDisabled locks the view at "fit" (zoom 1): wheel, buttons, double-tap
	// and ZoomIn/ZoomOut/Reset do nothing, and the info bar and zoom buttons
	// are hidden whatever SetOverlay asked for.
	zoomDisabled bool
	showInfo     bool // what SetOverlay asked for; only honored while zoom is enabled
	showControls bool

	// Thumbnail mode: small, zoom locked, overlay hidden, tap opens a preview
	// window with a normal viewer.
	thumbnail     bool
	thumbSize     fyne.Size
	thumbTitle    string
	previewWin    fyne.Window // open preview window, nil if none
	parentWin     fyne.Window // window the thumbnail lives in, see SetThumbnailParent
	previewViewer *ImageViewer

	// Stage: a radial gradient clipped to a rounded rectangle. Empty color
	// names (or fully transparent colors) mean no stage.
	stageCenter fyne.ThemeColorName
	stageEdge   fyne.ThemeColorName
	stageRadius float64 // Fyne units
	stageKey    stageKey
	stageImg    *image.RGBA // cached gradient for stageKey

	// built-in overlay
	info      *Surface // bottom-center info bar
	infoLabel *fynewidget.Label
	controls  *Surface // left-center zoom buttons

	zoomInBtn  *fynewidget.Button
	zoomOutBtn *fynewidget.Button
	fitBtn     *fynewidget.Button

	zoomText string
	posText  string
	valText  string
	infoMinW float32 // widest the info bar has been; keeps it from jittering

	// OnZoom is an optional hook called when the zoom changes (percent of real
	// pixel size). The built-in info bar already shows it.
	OnZoom func(percent float64)
	// OnCursor is an optional hook called while the mouse moves over the view.
	// The built-in info bar already shows it.
	OnCursor func(x, y int, value string, inside bool)
}

// NewImageViewer returns an empty ImageViewer.
func NewImageViewer() *ImageViewer {
	v := &ImageViewer{
		zoom:            1,
		split:           0.5,
		downscaleFilter: true,
		stageRadius:     defaultStageRadius,
		showInfo:        true,
		showControls:    true,
		thumbSize:       fyne.NewSize(defaultThumbW, defaultThumbH),
		thumbTitle:      "Image",
	}
	v.raster = canvas.NewRaster(v.draw)
	v.raster.ScaleMode = canvas.ImageScalePixels
	v.raster.SetMinSize(fyne.NewSize(defaultMinW, defaultMinH))
	v.buildOverlay()
	v.ExtendBaseWidget(v)
	return v
}

func (v *ImageViewer) CreateRenderer() fyne.WidgetRenderer {
	return &viewerRenderer{v: v}
}

// viewerRenderer stacks the raster and the two overlay surfaces.
type viewerRenderer struct{ v *ImageViewer }

func (r *viewerRenderer) Layout(size fyne.Size) {
	r.v.raster.Move(fyne.NewPos(0, 0))
	r.v.raster.Resize(size)
	r.v.layoutOverlay(size)
}

func (r *viewerRenderer) MinSize() fyne.Size { return r.v.raster.MinSize() }

func (r *viewerRenderer) Refresh() {
	r.v.raster.Refresh()
	r.v.info.Refresh() // re-reads theme colors
	r.v.controls.Refresh()
}

func (r *viewerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.v.raster, r.v.info, r.v.controls}
}

func (r *viewerRenderer) Destroy() {}

func (v *ImageViewer) buildOverlay() {
	v.infoLabel = fynewidget.NewLabel("")
	v.infoLabel.TextStyle = fyne.TextStyle{Monospace: true}
	v.info = NewSurface(v.infoLabel, theme.ColorNameOverlayBackground).
		SetRadius(8).
		SetInset(2).
		SetBorder(theme.ColorNameInputBorder, 1)

	flat := func(icon fyne.Resource, tapped func()) *fynewidget.Button {
		b := fynewidget.NewButtonWithIcon("", icon, tapped)
		b.Importance = fynewidget.LowImportance
		return b
	}
	v.zoomInBtn = flat(theme.ZoomInIcon(), v.ZoomIn)
	v.zoomOutBtn = flat(theme.ZoomOutIcon(), v.ZoomOut)
	v.fitBtn = flat(theme.ZoomFitIcon(), v.Reset)
	v.controls = NewSurface(container.NewVBox(v.zoomInBtn, v.zoomOutBtn, v.fitBtn), theme.ColorNameOverlayBackground).
		SetRadius(8).
		SetInset(2).
		SetBorder(theme.ColorNameInputBorder, 1)

	v.updateInfo()
	v.updateControls()
}

// SetOverlay shows or hides the info bar and the zoom buttons independently.
// While zoom is disabled both stay hidden; the choice is remembered and takes
// effect again once zoom is re-enabled.
func (v *ImageViewer) SetOverlay(info, controls bool) {
	v.showInfo, v.showControls = info, controls
	v.applyOverlayVisibility()
	v.Refresh()
}

// SetEnableZoom enables zoom (the default). The zoom buttons and the info bar
// come back as configured by SetOverlay.
func (v *ImageViewer) SetEnableZoom() { v.setZoomDisabled(false) }

// SetDisableZoom disables zoom. The image stays fitted to the window (any
// current zoom is reset), wheel/button/double-tap zoom is ignored, and the
// zoom buttons and the info bar are hidden. Panning is unaffected but has no
// visible effect at fit size.
func (v *ImageViewer) SetDisableZoom() { v.setZoomDisabled(true) }

func (v *ImageViewer) setZoomDisabled(disabled bool) {
	if v.zoomDisabled == disabled {
		return
	}
	v.zoomDisabled = disabled
	if disabled && v.img != nil {
		v.zoom = 1
		v.centerOnImage()
	}
	v.applyOverlayVisibility()
	v.refreshAndNotify()
}

// ZoomDisabled reports whether zoom is disabled, either explicitly or because
// thumbnail mode is on.
func (v *ImageViewer) ZoomDisabled() bool { return v.zoomOff() }

// SetEnableThumbnail turns thumbnail mode on. The viewer shrinks to the
// thumbnail size (see SetThumbnailSize), zoom is disabled, the zoom buttons
// and info bar are hidden, and tapping the viewer opens the image in a new
// window with zoom, zoom buttons and info bar enabled.
func (v *ImageViewer) SetEnableThumbnail() { v.setThumbnail(true) }

// SetDisableThumbnail turns thumbnail mode off and restores the normal size
// and, unless zoom was disabled separately, zoom and the overlay.
func (v *ImageViewer) SetDisableThumbnail() { v.setThumbnail(false) }

// ThumbnailEnabled reports whether thumbnail mode is on.
func (v *ImageViewer) ThumbnailEnabled() bool { return v.thumbnail }

// SetThumbnailSize sets the size of the viewer in thumbnail mode, in Fyne
// units. The default is 160x120.
func (v *ImageViewer) SetThumbnailSize(size fyne.Size) {
	v.thumbSize = size
	if v.thumbnail {
		v.raster.SetMinSize(size)
		v.Refresh()
	}
}

// SetThumbnailTitle sets the title of the preview window opened by tapping a
// thumbnail. The default is "Image". If the window is already open, its title
// changes right away.
func (v *ImageViewer) SetThumbnailTitle(title string) {
	v.thumbTitle = title
	if v.previewWin != nil {
		v.previewWin.SetTitle(title)
	}
}

// SetThumbnailParent tells the viewer which window it is shown in, so the
// preview window closes automatically when that window closes.
//
// It does this by calling parent.SetOnClosed, which replaces any handler
// already set on parent, and is itself replaced if you call SetOnClosed on
// parent afterwards. In that case call CloseThumbnailWindow from your own
// handler instead. If parent is the app's master window (SetMaster), closing
// it already closes every window and none of this is needed.
func (v *ImageViewer) SetThumbnailParent(parent fyne.Window) {
	v.parentWin = parent
	if parent != nil {
		parent.SetOnClosed(v.CloseThumbnailWindow)
	}
}

// CloseThumbnailWindow closes the preview window if it is open.
func (v *ImageViewer) CloseThumbnailWindow() {
	win := v.previewWin
	if win == nil {
		return
	}
	v.previewWin, v.previewViewer = nil, nil
	win.Close()
}

func (v *ImageViewer) setThumbnail(on bool) {
	if v.thumbnail == on {
		return
	}
	v.thumbnail = on
	if on {
		v.raster.SetMinSize(v.thumbSize)
		if v.img != nil {
			v.zoom = 1
			v.centerOnImage()
		}
	} else {
		v.raster.SetMinSize(fyne.NewSize(defaultMinW, defaultMinH))
	}
	v.applyOverlayVisibility()
	v.refreshAndNotify()
}

// Tapped opens the preview window when the viewer is a thumbnail.
func (v *ImageViewer) Tapped(*fyne.PointEvent) {
	if v.thumbnail {
		v.openPreview()
	}
}

// openPreview shows the image in a new window with a regular viewer. If the
// window is already open it is brought to the front instead. While it is
// open it follows SetImage, SetBefore, SetCompare and Clear.
func (v *ImageViewer) openPreview() {
	if v.img == nil {
		return
	}
	if v.previewWin != nil {
		v.previewWin.RequestFocus()
		return
	}
	app := fyne.CurrentApp()
	if app == nil {
		return
	}

	pv := NewImageViewer()
	pv.SetStage(v.stageCenter, v.stageEdge)
	pv.SetStageRadius(v.stageRadius)
	pv.SetDownscaleFilter(v.downscaleFilter)
	pv.SetBefore(v.before)
	pv.SetCompare(v.compare)
	pv.split = v.split
	pv.SetImage(v.img)

	win := app.NewWindow(v.thumbTitle)
	win.SetContent(pv)
	win.Resize(previewWindowSize(v.img.Bounds()))
	win.SetOnClosed(func() {
		v.previewWin, v.previewViewer = nil, nil
	})
	v.previewWin, v.previewViewer = win, pv
	win.Show()
}

// previewWindowSize fits the image into previewMaxW x previewMaxH without
// upscaling it, but never goes below the viewer's normal minimum size.
func previewWindowSize(b image.Rectangle) fyne.Size {
	w, h := float64(b.Dx()), float64(b.Dy())
	if w <= 0 || h <= 0 {
		return fyne.NewSize(previewMaxW, previewMaxH)
	}
	k := math.Min(1, math.Min(previewMaxW/w, previewMaxH/h))
	return fyne.NewSize(
		float32(math.Max(w*k, defaultMinW)),
		float32(math.Max(h*k, defaultMinH)),
	)
}

// zoomOff reports whether zoom is currently locked, by SetDisableZoom or by
// thumbnail mode.
func (v *ImageViewer) zoomOff() bool { return v.zoomDisabled || v.thumbnail }

func (v *ImageViewer) applyOverlayVisibility() {
	if v.showInfo && !v.zoomOff() {
		v.info.Show()
	} else {
		v.info.Hide()
	}
	if v.showControls && !v.zoomOff() {
		v.controls.Show()
	} else {
		v.controls.Hide()
	}
}

// layoutOverlay centers the info bar along the bottom edge and the button
// strip along the left edge.
func (v *ImageViewer) layoutOverlay(size fyne.Size) {
	m := float32(overlayMargin)

	is := v.info.MinSize()
	v.infoMinW = fyne.Max(v.infoMinW, is.Width) // never shrink, so the bar does not jitter
	is.Width = v.infoMinW
	v.info.Resize(is)
	v.info.Move(fyne.NewPos((size.Width-is.Width)/2, size.Height-is.Height-m))

	cs := v.controls.MinSize()
	v.controls.Resize(cs)
	v.controls.Move(fyne.NewPos(m, (size.Height-cs.Height)/2))
}

// updateInfo rebuilds the info bar text from the three parts.
func (v *ImageViewer) updateInfo() {
	zoom, pos := v.zoomText, v.posText
	if zoom == "" {
		zoom = "-"
	}
	if pos == "" {
		pos = "(-, -)"
	}
	text := "Zoom " + zoom + "   " + pos
	if v.valText != "" {
		text += "   " + v.valText
	}
	v.infoLabel.SetText(text)
	v.layoutOverlay(v.Size())
}

// updateControls enables only the buttons that can do something right now.
func (v *ImageViewer) updateControls() {
	set := func(b *fynewidget.Button, on bool) {
		if on {
			b.Enable()
		} else {
			b.Disable()
		}
	}
	has := v.img != nil
	set(v.zoomInBtn, has && v.zoom < maxZoom)
	set(v.zoomOutBtn, has && v.zoom > 1)
	set(v.fitBtn, has)
}

// SetImage shows img. Zoom and pan are kept if the image bounds are unchanged.
func (v *ImageViewer) SetImage(img image.Image) {
	if img == nil {
		v.Clear()
		return
	}
	if v.img == nil || v.img.Bounds() != img.Bounds() {
		v.img = img
		v.centerOnImage()
		v.zoom = 1
	}
	v.img = img
	v.refreshAndNotify()
	if v.previewViewer != nil {
		v.previewViewer.SetImage(img)
	}
}

// SetBefore sets the image shown left of the divider in compare mode.
// It is sampled in the same image coordinates as the main image; areas it
// does not cover are left transparent.
func (v *ImageViewer) SetBefore(img image.Image) {
	v.before = img
	v.Refresh()
	if v.previewViewer != nil {
		v.previewViewer.SetBefore(img)
	}
}

// SetCompare turns the before/after divider on or off.
func (v *ImageViewer) SetCompare(on bool) {
	v.compare = on
	v.Refresh()
	if v.previewViewer != nil {
		v.previewViewer.SetCompare(on)
	}
}

// SetStage enables the rounded stage behind the image using two theme colors:
// center is the middle of the gradient, edge the corners. Pass empty names to
// disable it. Presets whose colors are both fully transparent also show none.
func (v *ImageViewer) SetStage(center, edge fyne.ThemeColorName) {
	v.stageCenter, v.stageEdge = center, edge
	v.Refresh()
}

// SetStageRadius sets the stage's corner radius in Fyne units.
func (v *ImageViewer) SetStageRadius(r float64) {
	v.stageRadius = math.Max(r, 0)
	v.Refresh()
}

// SetDownscaleFilter turns the box filter on or off. When on (the default),
// thin strokes and dots survive when a large image is shown smaller than 1:1,
// at some extra drawing cost.
func (v *ImageViewer) SetDownscaleFilter(on bool) {
	v.downscaleFilter = on
	v.Refresh()
}

// Clear removes the images and resets zoom and pan. The divider position is kept.
func (v *ImageViewer) Clear() {
	v.img, v.before = nil, nil
	v.zoom = 1
	v.cx, v.cy = 0, 0
	v.Refresh()
	v.lastPct = -1
	v.zoomText, v.posText, v.valText = "", "", ""
	v.updateInfo()
	v.updateControls()
	if v.OnZoom != nil {
		v.OnZoom(0)
	}
	if v.previewViewer != nil {
		v.previewViewer.Clear()
	}
}

func (v *ImageViewer) ZoomIn()  { v.zoomAt(v.zoom*zoomStep, v.centerPos()) }
func (v *ImageViewer) ZoomOut() { v.zoomAt(v.zoom/zoomStep, v.centerPos()) }

// Reset fits the image to the window again.
func (v *ImageViewer) Reset() {
	if v.img == nil || v.zoomOff() {
		return
	}
	v.zoom = 1
	v.centerOnImage()
	v.refreshAndNotify()
}

// Percent is the current zoom as screen pixels per image pixel * 100.
func (v *ImageViewer) Percent() float64 {
	return v.currentTransform().scale * 100
}

// viewTransform maps between device pixels of the view and image coordinates.
// It is computed in one place so drawing, panning, zooming and the cursor
// readout can never disagree.
type viewTransform struct {
	w, h      float64 // viewport size, device pixels
	fit       float64 // image px -> device px at zoom 1
	scale     float64 // fit * zoom: image px -> device px now
	left, top float64 // image coordinates of the viewport's top-left corner
}

// toImage converts a viewport device-pixel position to image coordinates.
func (t viewTransform) toImage(px, py float64) (float64, float64) {
	return t.left + px/t.scale, t.top + py/t.scale
}

func (v *ImageViewer) transform(w, h float64) viewTransform {
	t := viewTransform{w: w, h: h}
	if v.img == nil || w <= 0 || h <= 0 {
		return t
	}
	b := v.img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return t
	}
	t.fit = math.Min(w/float64(b.Dx()), h/float64(b.Dy()))
	t.scale = t.fit * v.zoom
	t.left = v.cx - w/(2*t.scale)
	t.top = v.cy - h/(2*t.scale)
	return t
}

// currentTransform is the transform for the widget's present size.
func (v *ImageViewer) currentTransform() viewTransform {
	cs := v.canvasScale()
	s := v.Size()
	return v.transform(math.Round(float64(s.Width)*cs), math.Round(float64(s.Height)*cs))
}

// ---------------- interaction ----------------

func (v *ImageViewer) Scrolled(ev *fyne.ScrollEvent) {
	if v.img == nil || v.zoomOff() || ev.Scrolled.DY == 0 {
		return
	}
	steps := clampFloat(float64(ev.Scrolled.DY)/scrollUnit, -maxScrollSteps, maxScrollSteps)
	v.zoomAt(v.zoom*math.Pow(zoomStep, steps), ev.Position)
}

func (v *ImageViewer) Dragged(ev *fyne.DragEvent) {
	if v.img == nil {
		return
	}
	if v.dragMode == dragNone {
		v.dragMode = dragPan
		if v.compare && v.before != nil &&
			math.Abs(float64(ev.Position.X)-v.split*float64(v.Size().Width)) <= dividerGrab {
			v.dragMode = dragDivider
		}
	}

	switch v.dragMode {
	case dragDivider:
		if w := float64(v.Size().Width); w > 0 {
			v.split = clampFloat(float64(ev.Position.X)/w, 0, 1)
			v.Refresh()
		}
	case dragPan:
		t := v.currentTransform()
		if t.scale <= 0 {
			return
		}
		cs := v.canvasScale()
		v.cx -= float64(ev.Dragged.DX) * cs / t.scale
		v.cy -= float64(ev.Dragged.DY) * cs / t.scale
		v.clampCenter()
		v.Refresh()
	}
}

func (v *ImageViewer) DragEnd() { v.dragMode = dragNone }

func (v *ImageViewer) DoubleTapped(*fyne.PointEvent) { v.Reset() }

func (v *ImageViewer) Cursor() desktop.Cursor {
	if v.thumbnail {
		return desktop.PointerCursor
	}
	return desktop.CrosshairCursor
}

func (v *ImageViewer) MouseIn(ev *desktop.MouseEvent)    { v.report(ev.Position) }
func (v *ImageViewer) MouseMoved(ev *desktop.MouseEvent) { v.report(ev.Position) }
func (v *ImageViewer) MouseOut()                         { v.setCursorInfo(0, 0, "", false) }

// Resize also refreshes the zoom readout, because "fit" depends on the size.
func (v *ImageViewer) Resize(size fyne.Size) {
	v.BaseWidget.Resize(size)
	v.notifyZoom()
}

func (v *ImageViewer) refreshAndNotify() {
	v.Refresh()
	v.notifyZoom()
	v.updateControls()
}

func (v *ImageViewer) notifyZoom() {
	if v.img == nil {
		return
	}
	p := v.Percent()
	if math.Abs(p-v.lastPct) < 0.05 {
		return
	}
	v.lastPct = p
	v.zoomText = fmt.Sprintf("%.0f%%", p)
	v.updateInfo()
	if v.OnZoom != nil {
		v.OnZoom(p)
	}
}

// showsBefore reports whether widget x (Fyne units) is on the "before" side.
func (v *ImageViewer) showsBefore(x float32) bool {
	return v.compare && v.before != nil && float64(x) < v.split*float64(v.Size().Width)
}

func (v *ImageViewer) report(pos fyne.Position) {
	x, y, value, inside := v.probe(pos)
	v.setCursorInfo(x, y, value, inside)
}

// probe finds the pixel under pos and describes it. In compare mode it
// describes what is actually on screen, so left of the divider that is the
// "before" image.
func (v *ImageViewer) probe(pos fyne.Position) (x, y int, value string, inside bool) {
	if v.img == nil {
		return
	}
	x, y, ok := v.toImage(pos)
	if !ok {
		return 0, 0, "", false
	}
	src := v.img
	if v.showsBefore(pos.X) {
		src = v.before
	}
	if !image.Pt(x, y).In(src.Bounds()) {
		return 0, 0, "", false
	}
	return x, y, describePixel(src.At(x, y)), true
}

// setCursorInfo updates the info bar and calls the optional OnCursor hook.
func (v *ImageViewer) setCursorInfo(x, y int, value string, inside bool) {
	if inside {
		v.posText = fmt.Sprintf("(%d, %d)", x, y)
		v.valText = value
	} else {
		v.posText, v.valText = "", ""
	}
	v.updateInfo()
	if v.OnCursor != nil {
		v.OnCursor(x, y, value, inside)
	}
}

func describePixel(c color.Color) string {
	switch g := c.(type) {
	case color.Gray:
		return fmt.Sprintf("Gray %d", g.Y)
	case color.Gray16:
		return fmt.Sprintf("Gray %d", g.Y>>8)
	}
	n := toNRGBA(c)
	return fmt.Sprintf("R %d  G %d  B %d", n.R, n.G, n.B)
}

func (v *ImageViewer) canvasScale() float64 {
	if c := fyne.CurrentApp().Driver().CanvasForObject(v); c != nil {
		return float64(c.Scale())
	}
	return 1
}

func (v *ImageViewer) centerPos() fyne.Position {
	s := v.Size()
	return fyne.NewPos(s.Width/2, s.Height/2)
}

func (v *ImageViewer) centerOnImage() {
	b := v.img.Bounds()
	v.cx = float64(b.Min.X) + float64(b.Dx())/2
	v.cy = float64(b.Min.Y) + float64(b.Dy())/2
}

func (v *ImageViewer) clampCenter() {
	b := v.img.Bounds()
	if v.zoom <= 1 {
		v.centerOnImage()
		return
	}
	v.cx = clampFloat(v.cx, float64(b.Min.X), float64(b.Max.X))
	v.cy = clampFloat(v.cy, float64(b.Min.Y), float64(b.Max.Y))
}

// zoomAt zooms so that the image point under pos stays under pos.
func (v *ImageViewer) zoomAt(newZoom float64, pos fyne.Position) {
	if v.img == nil || v.zoomOff() {
		return
	}
	newZoom = clampFloat(newZoom, 1, maxZoom)
	if newZoom == v.zoom {
		return
	}
	t := v.currentTransform()
	if t.fit <= 0 {
		return
	}
	cs := v.canvasScale()
	px, py := float64(pos.X)*cs, float64(pos.Y)*cs
	ix, iy := t.toImage(px, py) // image point under the cursor

	s2 := t.fit * newZoom
	v.cx = ix + (t.w/2-px)/s2
	v.cy = iy + (t.h/2-py)/s2
	v.zoom = newZoom
	v.clampCenter()
	v.refreshAndNotify()
}

// toImage maps a position inside the widget to image pixel coordinates.
func (v *ImageViewer) toImage(pos fyne.Position) (int, int, bool) {
	t := v.currentTransform()
	if t.scale <= 0 {
		return 0, 0, false
	}
	cs := v.canvasScale()
	fx, fy := t.toImage(float64(pos.X)*cs, float64(pos.Y)*cs)
	x, y := int(math.Floor(fx)), int(math.Floor(fy))
	if !image.Pt(x, y).In(v.img.Bounds()) {
		return 0, 0, false
	}
	return x, y, true
}

// draw renders the visible part of the image at the given pixel size.
func (v *ImageViewer) draw(w, h int) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 {
		return out
	}

	center, edge, hasStage := v.stageColors()
	if hasStage {
		v.paintStage(out, center, edge)
	}

	if t := v.transform(float64(w), float64(h)); t.scale > 0 {
		comparing := v.compare && v.before != nil
		v.paintImage(out, t, comparing)
		if comparing {
			paintDivider(out, int(v.split*float64(w)), v.canvasScale())
		}
	}

	if hasStage {
		maskCorners(out, v.stageRadius*v.canvasScale())
	}
	return out
}

// pixelReader returns a premultiplied 8-bit sampler for img. Coordinates
// outside img's bounds read as transparent. Common image types avoid
// per-pixel color conversions.
func pixelReader(img image.Image) func(x, y int) (r, g, b, a uint32) {
	bounds := img.Bounds()
	switch m := img.(type) {
	case *image.Gray:
		return func(x, y int) (uint32, uint32, uint32, uint32) {
			if !image.Pt(x, y).In(bounds) {
				return 0, 0, 0, 0
			}
			g := uint32(m.Pix[m.PixOffset(x, y)])
			return g, g, g, 255
		}
	case *image.NRGBA:
		return func(x, y int) (uint32, uint32, uint32, uint32) {
			if !image.Pt(x, y).In(bounds) {
				return 0, 0, 0, 0
			}
			i := m.PixOffset(x, y)
			a := uint32(m.Pix[i+3])
			return uint32(m.Pix[i]) * a / 255, uint32(m.Pix[i+1]) * a / 255, uint32(m.Pix[i+2]) * a / 255, a
		}
	case *image.RGBA:
		return func(x, y int) (uint32, uint32, uint32, uint32) {
			if !image.Pt(x, y).In(bounds) {
				return 0, 0, 0, 0
			}
			i := m.PixOffset(x, y)
			return uint32(m.Pix[i]), uint32(m.Pix[i+1]), uint32(m.Pix[i+2]), uint32(m.Pix[i+3])
		}
	}
	return func(x, y int) (uint32, uint32, uint32, uint32) {
		if !image.Pt(x, y).In(bounds) {
			return 0, 0, 0, 0
		}
		r, g, b, a := img.At(x, y).RGBA() // 16-bit, premultiplied
		return r >> 8, g >> 8, b >> 8, a >> 8
	}
}

// paintImage draws the image (and, left of the divider, the "before" image)
// over whatever is already in out. Rows are rendered in parallel.
func (v *ImageViewer) paintImage(out *image.RGBA, t viewTransform, comparing bool) {
	w, h := out.Bounds().Dx(), out.Bounds().Dy()

	after := pixelReader(v.img)
	before := after
	splitX := 0 // pixels left of splitX use `before`
	bounds := v.img.Bounds()
	if comparing {
		before = pixelReader(v.before)
		splitX = int(v.split * float64(w))
		bounds = bounds.Union(v.before.Bounds())
	}

	// n×n box filter when the image is shown smaller than 1:1.
	n := 1
	if v.downscaleFilter && t.scale < 1 {
		n = int(math.Min(math.Ceil(1/t.scale), maxSupersample))
	}
	nn := uint32(n * n)

	// Source column for every (pixel, subsample); shared by all rows.
	xs := make([]int, w*n)
	for x := 0; x < w; x++ {
		for i := 0; i < n; i++ {
			xs[x*n+i] = int(math.Floor(t.left + (float64(x)+(float64(i)+0.5)/float64(n))/t.scale))
		}
	}

	renderRows := func(y0, y1 int) {
		ys := make([]int, n)
		for y := y0; y < y1; y++ {
			for j := 0; j < n; j++ {
				ys[j] = int(math.Floor(t.top + (float64(y)+(float64(j)+0.5)/float64(n))/t.scale))
			}
			if ys[n-1] < bounds.Min.Y || ys[0] >= bounds.Max.Y {
				continue // row is entirely outside the image
			}
			row := out.PixOffset(0, y)
			for x := 0; x < w; x++ {
				src := after
				if x < splitX {
					src = before
				}
				var sr, sg, sb, sa uint32
				for j := 0; j < n; j++ {
					for i := 0; i < n; i++ {
						r, g, b, a := src(xs[x*n+i], ys[j])
						sr, sg, sb, sa = sr+r, sg+g, sb+b, sa+a
					}
				}
				if nn > 1 {
					sr, sg, sb, sa = sr/nn, sg/nn, sb/nn, sa/nn
				}
				if sa == 0 {
					continue
				}
				o := row + x*4
				if sa >= 255 {
					out.Pix[o], out.Pix[o+1], out.Pix[o+2], out.Pix[o+3] = uint8(sr), uint8(sg), uint8(sb), 255
					continue
				}
				// "over" compositing onto what is already there (the stage)
				inv := 255 - sa
				out.Pix[o] = uint8(sr + uint32(out.Pix[o])*inv/255)
				out.Pix[o+1] = uint8(sg + uint32(out.Pix[o+1])*inv/255)
				out.Pix[o+2] = uint8(sb + uint32(out.Pix[o+2])*inv/255)
				out.Pix[o+3] = uint8(sa + uint32(out.Pix[o+3])*inv/255)
			}
		}
	}

	workers := runtime.GOMAXPROCS(0)
	chunk := (h + workers - 1) / workers
	var wg sync.WaitGroup
	for y0 := 0; y0 < h; y0 += chunk {
		y1 := y0 + chunk
		if y1 > h {
			y1 = h
		}
		wg.Add(1)
		go func(y0, y1 int) {
			defer wg.Done()
			renderRows(y0, y1)
		}(y0, y1)
	}
	wg.Wait()
}

// paintDivider draws the vertical compare line, thicker on HiDPI screens.
func paintDivider(out *image.RGBA, x int, canvasScale float64) {
	c := toNRGBA(theme.Color(theme.ColorNamePrimary))
	half := int(math.Max(1, math.Round(canvasScale)))
	w, h := out.Bounds().Dx(), out.Bounds().Dy()
	for y := 0; y < h; y++ {
		for dx := -half; dx <= half; dx++ {
			px := x + dx
			if px < 0 || px >= w {
				continue
			}
			i := out.PixOffset(px, y)
			out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = c.R, c.G, c.B, 255
		}
	}
}

type stageKey struct {
	w, h         int
	center, edge color.NRGBA
}

// stageColors resolves the stage theme colors. ok is false when there is no stage.
func (v *ImageViewer) stageColors() (center, edge color.NRGBA, ok bool) {
	if v.stageCenter == "" && v.stageEdge == "" {
		return
	}
	if v.stageCenter != "" {
		center = toNRGBA(theme.Color(v.stageCenter))
	}
	if v.stageEdge != "" {
		edge = toNRGBA(theme.Color(v.stageEdge))
	}
	return center, edge, center.A != 0 || edge.A != 0
}

// paintStage copies the cached radial gradient into out, rebuilding the cache
// only when the size or the colors change. The gradient is a plain rectangle;
// the rounded corners are applied last by maskCorners so they also clip the image.
func (v *ImageViewer) paintStage(out *image.RGBA, center, edge color.NRGBA) {
	w, h := out.Bounds().Dx(), out.Bounds().Dy()
	key := stageKey{w, h, center, edge}
	if v.stageImg == nil || v.stageKey != key {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		halfW, halfH := float64(w)/2, float64(h)/2
		for y := 0; y < h; y++ {
			dy := (float64(y) + 0.5 - halfH) / halfH
			for x := 0; x < w; x++ {
				dx := (float64(x) + 0.5 - halfW) / halfW
				c := mixNRGBA(center, edge, (dx*dx+dy*dy)/2) // 0 in the middle .. 1 in the corners
				a := float64(c.A) / 255
				i := img.PixOffset(x, y)
				img.Pix[i] = uint8(float64(c.R) * a) // RGBA wants premultiplied alpha
				img.Pix[i+1] = uint8(float64(c.G) * a)
				img.Pix[i+2] = uint8(float64(c.B) * a)
				img.Pix[i+3] = c.A
			}
		}
		v.stageImg, v.stageKey = img, key
	}
	copy(out.Pix, v.stageImg.Pix)
}

// maskCorners rounds the four corners of out with a one-pixel anti-aliased
// edge. Only the corner squares are visited; the rest of the image is untouched.
func maskCorners(out *image.RGBA, radius float64) {
	w, h := out.Bounds().Dx(), out.Bounds().Dy()
	radius = math.Min(radius, float64(min(w, h))/2)
	r := int(math.Ceil(radius))
	if r <= 0 {
		return
	}
	for dy := 0; dy < r; dy++ {
		for dx := 0; dx < r; dx++ {
			qx := radius - (float64(dx) + 0.5)
			qy := radius - (float64(dy) + 0.5)
			if qx <= 0 || qy <= 0 {
				continue
			}
			cov := clampFloat(0.5-(math.Hypot(qx, qy)-radius), 0, 1)
			if cov >= 1 {
				continue
			}
			scalePixel(out, dx, dy, cov)
			scalePixel(out, w-1-dx, dy, cov)
			scalePixel(out, dx, h-1-dy, cov)
			scalePixel(out, w-1-dx, h-1-dy, cov)
		}
	}
}

// scalePixel multiplies a premultiplied pixel by coverage (0..1).
func scalePixel(img *image.RGBA, x, y int, cov float64) {
	i := img.PixOffset(x, y)
	for k := 0; k < 4; k++ {
		img.Pix[i+k] = uint8(float64(img.Pix[i+k])*cov + 0.5)
	}
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func toNRGBA(c color.Color) color.NRGBA { return color.NRGBAModel.Convert(c).(color.NRGBA) }

// mixNRGBA blends a towards b by t (0..1), alpha included.
func mixNRGBA(a, b color.NRGBA, t float64) color.NRGBA {
	mix := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
	return color.NRGBA{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B), A: mix(a.A, b.A)}
}