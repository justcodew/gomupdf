package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Page represents a page in a document.
type Page struct {
	ctx   *cgo_bindings.Context
	page  *cgo_bindings.Page
	doc   *Document
	index int // page number (0-indexed)
}

// Rect returns the page bounding rectangle.
func (p *Page) Rect() Rect {
	if p.page == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := p.page.Rect()
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// Rotation returns the page rotation in degrees.
func (p *Page) Rotation() int {
	if p.page == nil || p.doc == nil {
		return 0
	}
	// Try PDF-specific rotation first (reads /Rotate from page dict)
	if p.doc.doc != nil {
		rot := cgo_bindings.PDFPageRotation(p.ctx, p.doc.doc.Doc, p.index)
		if rot != 0 {
			return rot
		}
	}
	return p.page.Rotation()
}

// Number returns the page number (0-indexed).
func (p *Page) Number() int {
	return p.index
}

// Pixmap returns a pixmap rendering of the page.
func (p *Page) Pixmap(matrix Matrix, alpha bool) (*Pixmap, error) {
	if p.page == nil || p.ctx == nil {
		return nil, fmt.Errorf("page is nil")
	}

	pix, err := cgo_bindings.RenderPage(p.ctx, p.page.Page,
		matrix.A, matrix.B, matrix.C, matrix.D, matrix.E, matrix.F, alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to render page: %w", err)
	}

	return &Pixmap{
		ctx:  p.ctx,
		pixmap: pix,
	}, nil
}

// Annots returns a list of annotations on the page.
// TODO: implement
func (p *Page) Annots() ([]*Annot, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	return nil, fmt.Errorf("annotations not yet implemented")
}

// LinkInfo represents a hyperlink on a page.
type LinkInfo = cgo_bindings.LinkInfo

// GetLinks loads all links from this page.
func (p *Page) GetLinks() ([]LinkInfo, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	return cgo_bindings.LoadLinks(p.ctx, p.page)
}

// SearchFor searches for text on this page, returning bounding rectangles of matches.
func (p *Page) SearchFor(text string, maxHits int) ([]Rect, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	if maxHits <= 0 {
		maxHits = 100
	}

	// Create structured text page
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, err
	}
	defer stextPage.Destroy()

	cgoRects := cgo_bindings.SearchText(p.ctx, stextPage.CStextPage(), text, maxHits)

	rects := make([]Rect, len(cgoRects))
	for i, r := range cgoRects {
		rects[i] = Rect{X0: r.X0, Y0: r.Y0, X1: r.X1, Y1: r.Y1}
	}
	return rects, nil
}

// String returns a string representation of the page.
func (p *Page) String() string {
	if p.page == nil {
		return "Page(<closed>)"
	}
	return fmt.Sprintf("Page(%v)", p.Rect())
}
