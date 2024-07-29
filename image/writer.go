package image

import (
	"fmt"
	img "image"
	"image/png"
	"os"
)

// WriteFile writes image to local file directory of given path
// currently generates image file in .png extension format
func WriteFile(_img img.Image, path, filename string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s.png", path, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, _img)
	if err != nil {
		return err
	}

	return nil
}
