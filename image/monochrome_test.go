package image

import (
	"testing"

	"github.com/mushoffa/gorengan/image"

	"github.com/stretchr/testify/assert"
)

func TestMonochrome1_Success(t *testing.T) {
	img, err := image.ReadFile("./test/input_monochrome.png")
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()
	monochrome := image.NewMonochrome(bounds, 160)	

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixel := img.At(x,y)
			monochromeColor := image.MonochromeModel.Convert(pixel)
			monochrome.Set(x, y, monochromeColor)
		}
	}

	err = image.WriteFile(monochrome, "./test", "output_monochrome")
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
}

func TestMonochrome2_Success(t *testing.T) {
	img, err := image.ReadFile("./test/input_monochrome_2.png")
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()
	monochrome := image.NewMonochrome(bounds, 160)	

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixel := img.At(x,y)
			monochromeColor := image.MonochromeModel.Convert(pixel)
			monochrome.Set(x, y, monochromeColor)
		}
	}

	err = image.WriteFile(monochrome, "./test", "output_monochrome_2")
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
}