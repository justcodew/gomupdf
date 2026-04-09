package pdf

// GetOCRCharRotatePos rotates character bounding box based on page rotation
func GetOCRCharRotatePos(x1, y1, x2, y2, pageWidth, pageHeight, rotation int) []int {
	pos := []int{x1, y1, x2, y2}

	switch rotation {
	case 90:
		pos = []int{pageWidth - y1, x1, pageWidth - y2, x2}
	case 270:
		pos = []int{pageHeight - y1, x1, pageHeight - y2, x2}
	case 180:
		pos = []int{pageWidth - x1, pageHeight - y1, pageWidth - x2, pageHeight - y2}
	}

	return pos
}

// GetOCRPolyRotatePos rotates polygon coordinates based on page rotation
func GetOCRPolyRotatePos(x, y, width, height, pageWidth, pageHeight, rotation int) []float64 {
	poly := []float64{
		float64(x), float64(y),
		float64(x + width), float64(y),
		float64(x + width), float64(y + height),
		float64(x), float64(y + height),
	}

	switch rotation {
	case 90:
		poly = []float64{
			float64(pageWidth - y), float64(x),
			float64(pageWidth - y), float64(x + width),
			float64(pageWidth - y - height), float64(x + width),
			float64(pageWidth - y - height), float64(x),
		}
	case 270:
		poly = []float64{
			float64(y), float64(pageHeight - x),
			float64(y), float64(pageHeight - (x + width)),
			float64(y + height), float64(pageHeight - (x + width)),
			float64(y + height), float64(pageHeight - x),
		}
	case 180:
		poly = []float64{
			float64(pageWidth - x), float64(pageHeight - y),
			float64(pageWidth - (x + width)), float64(pageHeight - y),
			float64(pageWidth - (x + width)), float64(pageHeight - (y + height)),
			float64(pageWidth - x), float64(pageHeight - (y + height)),
		}
	}

	return poly
}

// GetImgRotatePos rotates image position based on page rotation
func GetImgRotatePos(x, y, width, height, pageWidth, pageHeight, rotation int) ImagePosition {
	pos := ImagePosition{X: x, Y: y, Width: width, Height: height}

	switch rotation {
	case 90:
		pos = ImagePosition{X: pageHeight - y - height, Y: x, Width: height, Height: width}
	case 270:
		pos = ImagePosition{X: pageHeight - y, Y: x, Width: height, Height: width}
	case 180:
		pos = ImagePosition{X: pageWidth - x, Y: pageHeight - y, Width: width, Height: height}
	}

	return pos
}
