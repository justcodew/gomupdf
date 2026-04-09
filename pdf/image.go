package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
)

// IsPureWhiteOrBlack detects if image is pure white or pure black
func IsPureWhiteOrBlack(img image.Image, whiteThreshold, blackThreshold int) bool {
	if img == nil {
		return false
	}

	bounds := img.Bounds()
	var sumR, sumG, sumB int64
	var count int64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			sumR += int64(r >> 8)
			sumG += int64(g >> 8)
			sumB += int64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return false
	}

	avgR := float64(sumR) / float64(count)
	avgG := float64(sumG) / float64(count)
	avgB := float64(sumB) / float64(count)
	avg := (avgR + avgG + avgB) / 3.0

	// Check pure white
	if avg >= float64(whiteThreshold) {
		return true
	}

	// Check pure black
	if avg <= float64(blackThreshold) {
		return true
	}

	return false
}

// IsPureWhiteOrBlackWithYCrCb uses YCrCb color space for better detection
func IsPureWhiteOrBlackWithYCrCb(img image.Image, whiteThreshold, blackThreshold int) bool {
	if img == nil {
		return false
	}

	bounds := img.Bounds()
	var sumY float64
	var count int64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Calculate Y (luminance)
			yVal := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			sumY += yVal
			count++
		}
	}

	if count == 0 {
		return false
	}

	avgY := sumY / float64(count)

	// Check pure white
	if avgY >= float64(whiteThreshold) {
		return true
	}

	// Check pure black
	if avgY <= float64(blackThreshold) {
		return true
	}

	return false
}

// ResizeImageByLongEdge resizes image with long edge not exceeding maxLength
func ResizeImageByLongEdge(img image.Image, maxLength int) image.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	maxEdge := math.Max(float64(width), float64(height))

	if maxEdge <= float64(maxLength) {
		return img
	}

	// Calculate new dimensions
	var newWidth, newHeight int
	if width > height {
		newWidth = maxLength
		newHeight = int(float64(height) * (float64(maxLength) / float64(width)))
	} else {
		newHeight = maxLength
		newWidth = int(float64(width) * (float64(maxLength) / float64(height)))
	}

	// Simple nearest neighbor scaling
	scaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	scaleX := float64(width) / float64(newWidth)
	scaleY := float64(height) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			if srcX >= width {
				srcX = width - 1
			}
			if srcY >= height {
				srcY = height - 1
			}
			scaled.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return scaled
}

// SaveImageAsPNG saves image as PNG
func SaveImageAsPNG(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// SaveImageAsJPEG saves image as JPEG
func SaveImageAsJPEG(img image.Image, path string, quality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
}

// ConvertToRGBA converts any image to RGBA
func ConvertToRGBA(img image.Image) *image.RGBA {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba
}

// RotateImage90 rotates image 90 degrees clockwise
func RotateImage90(img image.Image) image.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	rotated := image.NewRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotated.Set(height-y, x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return rotated
}

// RotateImage rotates image by angle (0, 90, 180, 270)
func RotateImage(img image.Image, rotation int) image.Image {
	if img == nil {
		return nil
	}

	switch rotation {
	case 0:
		return img
	case 90:
		return RotateImage90(img)
	case 180:
		return RotateImage180(img)
	case 270:
		return RotateImage270(img)
	default:
		return img
	}
}

// RotateImage180 rotates image 180 degrees
func RotateImage180(img image.Image) image.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rotated.Set(bounds.Dx()-x-1, bounds.Dy()-y-1, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return rotated
}

// RotateImage270 rotates image 270 degrees clockwise
func RotateImage270(img image.Image) image.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	rotated := image.NewRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotated.Set(y, width-x-1, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	return rotated
}

// ImageFromBytes creates image from PNG/JPEG bytes
func ImageFromBytes(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// CreateImageFromBuffer creates image from raw bytes (typically from PDF pixmap)
func CreateImageFromBuffer(buf []byte, width, height int) image.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// Assume RGB format (3 bytes per pixel)
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if idx+2 < len(buf) {
				r := buf[idx]
				g := buf[idx+1]
				b := buf[idx+2]
				rgba.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				idx += 3
			}
		}
	}

	return rgba
}

// ScaleImage scales image using bilinear interpolation
func ScaleImage(img image.Image, scaleX, scaleY float64) image.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	newWidth := int(float64(bounds.Dx()) * scaleX)
	newHeight := int(float64(bounds.Dy()) * scaleY)

	if newWidth <= 0 || newHeight <= 0 {
		return img
	}

	scaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest neighbor scaling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scaleX)
			srcY := int(float64(y) / scaleY)
			if srcX >= bounds.Dx() {
				srcX = bounds.Dx() - 1
			}
			if srcY >= bounds.Dy() {
				srcY = bounds.Dy() - 1
			}
			scaled.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return scaled
}
