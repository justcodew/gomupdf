package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Pixmap represents a raster image.
type Pixmap struct {
	ctx    *cgo_bindings.Context
	pixmap *cgo_bindings.Pixmap
}

// Width returns the pixmap width.
func (p *Pixmap) Width() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Width()
}

// Height returns the pixmap height.
func (p *Pixmap) Height() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Height()
}

// Stride returns the pixmap stride (bytes per row).
func (p *Pixmap) Stride() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Stride()
}

// N returns the number of color components (e.g., 4 for RGBA).
func (p *Pixmap) N() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.N()
}

// Samples returns the raw pixel data.
func (p *Pixmap) Samples() []byte {
	if p.pixmap == nil {
		return nil
	}
	return p.pixmap.Samples()
}

// Save saves the pixmap as an image file.
func (p *Pixmap) Save(filename string) error {
	if p.pixmap == nil {
		return fmt.Errorf("pixmap is nil")
	}

	switch ext := getExt(filename); ext {
	case ".png":
		return p.pixmap.SavePNG(filename)
	case ".jpg", ".jpeg":
		return p.pixmap.SaveJPEG(filename, 90)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
}

// Close releases the pixmap resources.
func (p *Pixmap) Close() {
	if p.pixmap != nil {
		p.pixmap.Destroy()
		p.pixmap = nil
	}
}

// PNG returns the pixmap as PNG bytes.
func (p *Pixmap) PNG() ([]byte, error) {
	if p.pixmap == nil {
		return nil, fmt.Errorf("pixmap is nil")
	}
	return p.pixmap.PNGBytes()
}

// JPEG returns the pixmap as JPEG bytes.
func (p *Pixmap) JPEG(quality int) ([]byte, error) {
	if p.pixmap == nil {
		return nil, fmt.Errorf("pixmap is nil")
	}
	return p.pixmap.JPEGBytes(quality)
}

// Pixel returns the pixel value at (x, y).
func (p *Pixmap) Pixel(x, y int) int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Pixel(x, y)
}

// SetPixel sets the pixel value at (x, y).
func (p *Pixmap) SetPixel(x, y, val int) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.SetPixel(x, y, val)
}

// ClearWith clears the pixmap with the given value.
func (p *Pixmap) ClearWith(value int) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.ClearWith(value)
}

// Invert inverts the pixmap colors.
func (p *Pixmap) Invert() {
	if p.pixmap == nil {
		return
	}
	p.pixmap.Invert()
}

// Gamma applies gamma correction.
func (p *Pixmap) Gamma(gamma float64) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.Gamma(gamma)
}

// Tint applies tinting.
func (p *Pixmap) Tint(black, white int) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.Tint(black, white)
}

// String returns a string representation of the pixmap.
func (p *Pixmap) String() string {
	if p.pixmap == nil {
		return "Pixmap(<nil>)"
	}
	return fmt.Sprintf("Pixmap(%dx%d, %d channels)", p.Width(), p.Height(), p.N())
}

func getExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
		if filename[i] == '/' || filename[i] == '\\' {
			break
		}
	}
	return ""
}
