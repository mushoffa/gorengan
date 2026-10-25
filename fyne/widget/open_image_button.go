package widget

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	fwidget "fyne.io/fyne/v2/widget"
)

// openImageExtensions are the file types the dialog offers.
var openImageExtensions = []string{".png", ".jpg", ".jpeg"}

const (
	defaultDialogRatio = 0.8 // dialog size as a fraction of the parent window
	minDialogRatio     = 0.2
)

type openImageConfig struct {
	text           string
	icon           fyne.Resource
	importance     fwidget.Importance
	ratioW, ratioH float32
	startDir       string
}

// OpenImageOption customizes NewOpenImageButton.
type OpenImageOption func(*openImageConfig)

// OpenImageText sets the button label (default "Open Image").
func OpenImageText(text string) OpenImageOption {
	return func(c *openImageConfig) { c.text = text }
}

// OpenImageIcon sets the button icon (default theme.FolderOpenIcon).
func OpenImageIcon(icon fyne.Resource) OpenImageOption {
	return func(c *openImageConfig) { c.icon = icon }
}

// OpenImageImportance sets how prominent the button looks (default
// MediumImportance). The colors come from the active Fyne theme, so
// HighImportance uses the theme's primary color, and light/dark switches and
// custom themes apply automatically.
func OpenImageImportance(importance fwidget.Importance) OpenImageOption {
	return func(c *openImageConfig) { c.importance = importance }
}

// OpenImageDialogRatio sets the dialog size as a fraction of the parent
// window's canvas: 0.8, 0.6 is 80% of the width and 60% of the height.
// Values are clamped to 0.2..1. The size is read each time the dialog opens,
// so it follows the window if the user resizes it.
func OpenImageDialogRatio(width, height float32) OpenImageOption {
	return func(c *openImageConfig) {
		c.ratioW, c.ratioH = clampRatio(width), clampRatio(height)
	}
}

// OpenImageStartDir sets the folder the dialog opens in
// (default: the process working directory).
func OpenImageStartDir(dir string) OpenImageOption {
	return func(c *openImageConfig) { c.startDir = dir }
}

// NewOpenImageButton returns a standard, theme-styled Fyne button that opens a file dialog on win,
// decodes the chosen PNG or JPEG, and passes it to callback together with the
// file's URI string (for example "file:///home/me/scan.png"; use
// storage.ParseURI to get its Name or Path). Cancelling the dialog does
// nothing. Errors are shown in an error dialog and callback is not called.
//
// The dialog and the callback run on the UI thread.
func NewOpenImageButton(win fyne.Window, callback func(uri string, img image.Image), opts ...OpenImageOption) *fwidget.Button {
	cfg := openImageConfig{
		text:       "Open Image",
		icon:       theme.FolderOpenIcon(),
		importance: fwidget.MediumImportance,
		ratioW:     defaultDialogRatio,
		ratioH:     defaultDialogRatio,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	btn := fwidget.NewButtonWithIcon(cfg.text, cfg.icon, func() {
		showOpenImageDialog(win, cfg, callback)
	})
	btn.Importance = cfg.importance
	return btn
}

func showOpenImageDialog(win fyne.Window, cfg openImageConfig, callback func(string, image.Image)) {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if r == nil {
			return // cancelled
		}
		defer r.Close()

		img, _, err := image.Decode(r)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if callback != nil {
			callback(r.URI().String(), img)
		}
	}, win)

	d.SetFilter(storage.NewExtensionFileFilter(openImageExtensions))
	if dir := startDirLister(cfg.startDir); dir != nil {
		d.SetLocation(dir)
	}

	size := win.Canvas().Size()
	d.Resize(fyne.NewSize(size.Width*cfg.ratioW, size.Height*cfg.ratioH))
	d.Show()
}

// startDirLister returns the folder to open in, or nil to let the dialog choose.
func startDirLister(dir string) fyne.ListableURI {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if dir == "" {
		return nil
	}
	lister, err := storage.ListerForURI(storage.NewFileURI(dir))
	if err != nil {
		return nil
	}
	return lister
}

func clampRatio(r float32) float32 {
	if r < minDialogRatio {
		return minDialogRatio
	}
	if r > 1 {
		return 1
	}
	return r
}
