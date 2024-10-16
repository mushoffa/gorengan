package image

import (
	img "image"
	"image/color"
)

// Monochrome color is either a black or white.
// 0 = White
// 255 = Black
const (
	Black uint8 = 0
	White uint8 = 255
)

// MonochromeModel for the color of black and white type.
var MonochromeModel color.Model = color.ModelFunc(monochromeModel)

// MonochromeColor represents a 1-bit color.
type MonochromeColor struct {
	Y uint8
}

// RGBA returns the RGBA values of [MonochromeColor] type.
func (c MonochromeColor) RGBA() (r, g, b, a uint32) {
	if c.Y == Black {
		return 0, 0, 0, 0xffff
	}
	return 0xffff, 0xffff, 0xffff, 0xffff
}

func monochromeModel(c color.Color) color.Color {
	if _, ok := c.(MonochromeColor); ok {
		return c
	}
	r, g, b, _ := c.RGBA()

	r = r >> 8
	g = g >> 8
	b = b >> 8

	y := 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b)

	return MonochromeColor{uint8(y)}
}

// Monochrome is an in-memory image whose At method returns [color.Monochrome] values.
type Monochrome struct {
	// Pix holds the image's pixels, as gray values. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*1].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect img.Rectangle
	// Threshold is the value which determines black / white pixel
	Threshold uint8
	// Inverse is parameter value to change base contrast of monochrome color black-white / white-black
	inverse bool
}

// NewMonochrome returns a new [Monochrome] image with the given bounds.
func NewMonochrome(r img.Rectangle, threshold uint8) *Monochrome {
	width, height := r.Dx(), r.Dy()

	pixBuffer := 1 * width * height
	pix := make([]uint8, pixBuffer)
	stride := 1 * width

	return &Monochrome{
		Pix:       pix,
		Stride:    stride,
		Rect:      r,
		Threshold: threshold,
		inverse: false,
	}
}

// getColor returns monochrome color and invert if inverse param is true
func (p *Monochrome) getColor(c uint8) uint8 {
	if p.inverse {
		return ^c
	}

	return c
}

// Inverse set [inverse] parameter to true
func (p *Monochrome) Inverse() {
	p.inverse = true
}

// ColorModel returns the [MonochromeModel] color type.
func (p *Monochrome) ColorModel() color.Model {
	return MonochromeModel
}

// Bounds returns rectangle boundary of the image.
func (p *Monochrome) Bounds() img.Rectangle {
	return p.Rect
}

// At returns pixel color of the image at (x,y) coordinate point.
func (p *Monochrome) At(x, y int) color.Color {
	if !(img.Point{x, y}.In(p.Rect)) {
		return MonochromeColor{Y: uint8(White)}
	}
	i := p.PixOffset(x, y)
	//pixel := Pixel(p.Pix[i]&(1<<uint(x%8)) != 0)
	pixel := p.Pix[i] & (1 << uint(x%8))
	color := Black
	if pixel != 0 {
		color = White
	}

	return MonochromeColor{Y: color}
}

// Set determines black or white pixel color at (x,y) coordinate point
// by comparing given [MonochromeColor] model value to defined threshold value.
func (p *Monochrome) Set(x, y int, c color.Color) {
	if !(img.Point{x, y}.In(p.Rect)) {
		return
	}

	i := p.PixOffset(x, y)
	_y := MonochromeModel.Convert(c).(MonochromeColor).Y
	p.Pix[i] = p.getColor(Black)

	if _y > p.Threshold {
		p.Pix[i] = p.getColor(White)
	}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *Monochrome) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*1
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *Monochrome) SubImage(r img.Rectangle) img.Image {
	r = r.Intersect(p.Rect)
	// If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be inside
	// either r1 or r2 if the intersection is empty. Without explicitly checking for
	// this, the Pix[i:] expression below can panic.
	if r.Empty() {
		return &Monochrome{}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &Monochrome{
		Pix:    p.Pix[i:],
		Stride: p.Stride,
		Rect:   r,
	}
}
