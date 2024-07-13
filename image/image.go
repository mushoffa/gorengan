package image

import (
	"bytes"
	"encoding/base64"
	img "image"
	"image/png"
)

// Image is a data struct which embeds Go standard image library.
type Image struct {
	img.Image
}

// New returns a new [Image] struct.
func New(img img.Image) *Image {
	return &Image{img}
}

// Base64 encodes an image that has been decoded previously through [New] function.
// The string returned is a base64 format of the given image.
func (p *Image) Base64() (string, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(bytes)

	return encoded, nil
}

// Bytes returns an encoded data of the given image in bytes format.
func (p *Image) Bytes() ([]byte, error) {
	return encode(p)
}

// Monochrome converts registered image from [New] function and returns [Monochrome] type image.
func (p *Image) Monochrome(threshold uint8) (*Monochrome) {
	bounds := p.Bounds()

	monochrome := NewMonochrome(bounds, threshold)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixel := p.At(x,y)
			color := MonochromeModel.Convert(pixel)
			monochrome.Set(x, y, color)
		}
	}

	return monochrome
}

// encode returns a bytes format of the given image.
func encode(img img.Image) ([]byte, error) {
	buffer := new(bytes.Buffer)

	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}