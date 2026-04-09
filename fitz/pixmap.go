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

	// Determine format from filename extension
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

// JPEG returns the pixmap as JPEG bytes.
func (p *Pixmap) JPEG() ([]byte, error) {
	// For now, save to a temp file and read back
	// TODO: implement direct JPEG encoding
	if p.pixmap == nil {
		return nil, fmt.Errorf("pixmap is nil")
	}
	return nil, fmt.Errorf("JPEG encoding not yet implemented")
}

// PNG returns the pixmap as PNG bytes.
func (p *Pixmap) PNG() ([]byte, error) {
	// For now, save to a temp file and read back
	// TODO: implement direct PNG encoding
	if p.pixmap == nil {
		return nil, fmt.Errorf("pixmap is nil")
	}
	return nil, fmt.Errorf("PNG encoding not yet implemented")
}

// String returns a string representation of the pixmap.
func (p *Pixmap) String() string {
	if p.pixmap == nil {
		return "Pixmap(<nil>)"
	}
	return fmt.Sprintf("Pixmap(%dx%d, %d channels)", p.Width(), p.Height(), p.N())
}

func getExt(filename string) string {
	// Simple extension extraction
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
