package image

import (
	"bytes"
	"encoding/base64"
	gimg "image"
	"image/png"
)

// Image is a data struct which embeds Go standard image library.
type Image struct {
	gimg.Image
}

// New returns a new [Image] struct.
func New(img gimg.Image) *Image {
	return &Image{img}
}

// Base64 encodes an image that has been decoded previously through [New] function.
// The string returned is a base64 format of the given image.
func (i *Image) Base64() (string, error) {
	bytes, err := i.Bytes()
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(bytes)

	return encoded, nil
}

// Bytes returns an encoded data of the given image in bytes format.
func (i *Image) Bytes() ([]byte, error) {
	return encode(i)
}


// encode returns a bytes format of the given image.
func encode(img gimg.Image) ([]byte, error) {
	buffer := new(bytes.Buffer)

	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}