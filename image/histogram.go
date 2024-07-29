package image

// Histogram returns array of data from registered [Monochrome] image type
// vertical = false, horizontal axis profile
// vertical = true, vertical axis profile
func (p *Monochrome) Histogram(vertical bool) []uint16 {
	bounds := p.Bounds()

	minRow := bounds.Min.Y
	maxRow := bounds.Max.Y
	minColumn := bounds.Min.X
	maxColumn := bounds.Max.X

	if vertical {
		minRow, minColumn = minColumn, minRow
		maxRow, maxColumn = maxColumn, maxRow
	}

	histogram := make([]uint16, maxRow)

	for row := minRow; row < maxRow; row++ {
		for column := minColumn; column < maxColumn; column++ {
			c := column
			r := row

			if vertical {
				c, r = r, c
			}

			pixel := p.At(c, r).(MonochromeColor).Y
			// 8-bit grayscale color
			// 0 = black
			// 255 = white
			if pixel == 0 {
				histogram[row]++
			}
		}
	}

	return histogram
}
