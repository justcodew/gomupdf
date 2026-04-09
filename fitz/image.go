package fitz

import (
	"fmt"
)

// ImageInfo represents information about an image in a PDF.
type ImageInfo struct {
	Number     int       // Image number in the page
	Bbox       Rect      // Bounding box on the page
	Matrix     Matrix    // Transformation matrix
	Width      int       // Image width in pixels
	Height     int       // Image height in pixels
	Colorspace string    // Colorspace name (e.g., "RGB", "Gray", "CMYK")
	N          int       // Number of color components
	Xres       int       // X resolution (dpi)
	Yres       int       // Y resolution (dpi)
	BPC        int       // Bits per component
	Size       int       // Size in bytes
	HasMask    bool      // Whether image has a mask
	Digest     []byte    // MD5 digest (if hashes=true)
}

// Image represents an extracted image with its pixel data.
type Image struct {
	Width    int
	Height   int
	N        int       // Color components
	BPC      int       // Bits per component
	Colorspace string
	Samples  []byte    // Raw pixel data
}

// GetImages returns a list of image information on the page.
func (p *Page) GetImages() ([]ImageInfo, error) {
	// TODO: implement using CGO
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	return nil, fmt.Errorf("image extraction not yet implemented")
}

// GetImageXObjects returns a list of XObject (form) images on the page.
func (p *Page) GetImageXObjects() ([]ImageInfo, error) {
	// TODO: implement using CGO
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	return nil, fmt.Errorf("image xobject extraction not yet implemented")
}
