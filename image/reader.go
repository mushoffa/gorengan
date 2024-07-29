package image

import (
	img "image"
	"os"
)

// ReadFile reads image file format from given path
// and decode it to Go standard image format
func ReadFile(path string) (img.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_img, _, err := img.Decode(file)
	if err != nil {
		return nil, err
	}

	return _img, nil
}
