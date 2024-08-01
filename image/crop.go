package image

import (
	img "image"
)

type CropRectangle interface {
	SubImage(r img.Rectangle) img.Image
}

func (p *Image) Crop(minX, minY, maxX, maxY int) img.Image {
	r := img.Rect(minX, minY, maxX, maxY)
	cropped := p.Data.(CropRectangle).SubImage(r)

	return cropped
}

func (p *Monochrome) Crop(minX, minY, maxX, maxY int) img.Image {
	cropped := p.SubImage(img.Rect(minX, minY, maxX, maxY))

	return cropped
}
