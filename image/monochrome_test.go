package image

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonochrome_Success(t *testing.T) {
	img, err := ReadFile("./test/input_color_1.png")
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	bounds := img.Bounds()
	monochrome := NewMonochrome(bounds, 200)	

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixel := img.At(x,y)
			monochromeColor := MonochromeModel.Convert(pixel)
			monochrome.Set(x, y, monochromeColor)
		}
	}

	elapsed := time.Since(start)
	t.Log("Monochrome took: ", elapsed)

	err = WriteFile(monochrome, "./test", "output_monochrome_1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
}

func TestMonochromeLazyConcurrent_Success(t *testing.T) {
	img, err := ReadFile("./test/input_color_1.png")
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	monochrome := New(img).Monochrome(200)
	elapsed := time.Since(start)
	t.Log("MonochromeConcurrent: ", elapsed)

	err = WriteFile(monochrome, "./test", "output_monochrome_1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
}

func BenchmarkMonochrome(b *testing.B) {
	img, err := ReadFile("./test/input_color_1.png")
	if err != nil {
		b.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		bounds := img.Bounds()
		monochrome := NewMonochrome(bounds, 200)	

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				pixel := img.At(x,y)
				monochromeColor := MonochromeModel.Convert(pixel)
				monochrome.Set(x, y, monochromeColor)
			}
		}
	}
}

func BenchmarkMonochromeConcurrent(b *testing.B) {
	img, err := ReadFile("./test/input_color_1.png")
	if err != nil {
		b.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		New(img).Monochrome(200)
	}
}