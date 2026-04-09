package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// ImageInfo represents information about an image in a PDF.
type ImageInfo struct {
	Number     int
	Bbox       Rect
	Matrix     Matrix
	Width      int
	Height     int
	Colorspace string
	N          int
	Xres       int
	Yres       int
	BPC        int
	Size       int
	HasMask    bool
	Digest     []byte
}

// Image represents an extracted image with its pixel data.
type Image struct {
	Width      int
	Height     int
	N          int
	BPC        int
	Colorspace string
	Samples    []byte
}

// GetImages returns a list of image information on the page.
func (p *Page) GetImages() ([]ImageInfo, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	// Use structured text page to find image blocks
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, err
	}
	defer stextPage.Destroy()

	blockCount := stextPage.BlockCount()
	var images []ImageInfo

	for i := 0; i < blockCount; i++ {
		block := stextPage.GetBlock(i)
		if block == nil {
			continue
		}
		if stextPage.BlockType(block) != 1 { // Skip non-image blocks
			continue
		}

		bx0, by0, bx1, by1 := stextPage.BlockBbox(block)
		images = append(images, ImageInfo{
			Number: i,
			Bbox:   Rect{X0: bx0, Y0: by0, X1: bx1, Y1: by1},
		})
	}

	return images, nil
}

// ExtractImage extracts image pixel data from a text block.
func (p *Page) ExtractImage(blockIdx int) (*Image, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, err
	}
	defer stextPage.Destroy()

	block := stextPage.GetBlock(blockIdx)
	if block == nil {
		return nil, fmt.Errorf("block %d not found", blockIdx)
	}
	if stextPage.BlockType(block) != 1 {
		return nil, fmt.Errorf("block %d is not an image block", blockIdx)
	}

	// Get the image from the block and render to pixmap
	img := cgo_bindings.BlockGetImage(p.ctx, block)
	if img == nil {
		return nil, fmt.Errorf("failed to get image from block")
	}

	w := cgo_bindings.ImageWidth(p.ctx, img)
	h := cgo_bindings.ImageHeight(p.ctx, img)
	n := cgo_bindings.ImageN(p.ctx, img)
	bpc := cgo_bindings.ImageBPC(p.ctx, img)
	csName := cgo_bindings.ImageColorspaceName(p.ctx, img)

	pix, err := cgo_bindings.NewPixmapFromImage(p.ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to extract image pixels: %w", err)
	}
	defer pix.Destroy()

	return &Image{
		Width:      w,
		Height:     h,
		N:          n,
		BPC:        bpc,
		Colorspace: csName,
		Samples:    pix.Samples(),
	}, nil
}

// GetImageXObjects returns a list of XObject images on the page.
func (p *Page) GetImageXObjects() ([]ImageInfo, error) {
	// TODO: implement XObject enumeration from page resources
	return nil, fmt.Errorf("not yet implemented")
}
