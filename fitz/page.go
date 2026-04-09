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
	if p.doc.doc != nil {
		rot := cgo_bindings.PDFPageRotation(p.ctx, p.doc.doc.Doc, p.index)
		if rot != 0 {
			return rot
		}
	}
	return p.page.Rotation()
}

// SetRotation sets the page rotation.
func (p *Page) SetRotation(rotation int) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.SetPageRotation(p.ctx, p.doc.doc.Doc, p.index, rotation)
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
func (p *Page) Annots() ([]*Annot, error) {
	if p.page == nil || p.doc == nil {
		return nil, fmt.Errorf("page is nil")
	}

	var annots []*Annot
	a := cgo_bindings.FirstAnnot(p.ctx, p.doc.doc.Doc, p.page.Page)
	for a != nil {
		annots = append(annots, &Annot{ctx: p.ctx, annot: a})
		a = a.Next()
	}
	return annots, nil
}

// AddAnnot creates and adds a new annotation to the page.
func (p *Page) AddAnnot(annotType int) (*Annot, error) {
	if p.page == nil || p.doc == nil {
		return nil, fmt.Errorf("page is nil")
	}
	annot, err := cgo_bindings.CreateAnnot(p.ctx, p.doc.doc.Doc, p.page.Page, annotType)
	if err != nil {
		return nil, err
	}
	return &Annot{ctx: p.ctx, annot: annot}, nil
}

// AddHighlightAnnot adds a highlight annotation with the given quads.
func (p *Page) AddHighlightAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotHighlight))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		if err := annot.annot.SetQuadPoints(cgoQuads); err != nil {
			annot.Delete()
			return nil, err
		}
	}
	annot.Update()
	return annot, nil
}

// AddStrikeoutAnnot adds a strikeout annotation.
func (p *Page) AddStrikeoutAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotStrikeOut))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		annot.annot.SetQuadPoints(cgoQuads)
	}
	annot.Update()
	return annot, nil
}

// AddUnderlineAnnot adds an underline annotation.
func (p *Page) AddUnderlineAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotUnderline))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		annot.annot.SetQuadPoints(cgoQuads)
	}
	annot.Update()
	return annot, nil
}

// AddSquigglyAnnot adds a squiggly annotation.
func (p *Page) AddSquigglyAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotSquiggly))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		annot.annot.SetQuadPoints(cgoQuads)
	}
	annot.Update()
	return annot, nil
}

// AddTextAnnot adds a text (sticky note) annotation.
func (p *Page) AddTextAnnot(point Point, text string) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotText))
	if err != nil {
		return nil, err
	}
	annot.annot.SetContents(text)
	annot.Update()
	return annot, nil
}

// AddFreeTextAnnot adds a free text annotation.
func (p *Page) AddFreeTextAnnot(rect Rect, text string) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotFreeText))
	if err != nil {
		return nil, err
	}
	annot.annot.SetRect(rect.X0, rect.Y0, rect.X1, rect.Y1)
	annot.annot.SetContents(text)
	annot.Update()
	return annot, nil
}

// AddRedactAnnot adds a redaction annotation.
func (p *Page) AddRedactAnnot(rect Rect, text string) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotRedact))
	if err != nil {
		return nil, err
	}
	annot.annot.SetRect(rect.X0, rect.Y0, rect.X1, rect.Y1)
	if text != "" {
		annot.annot.SetContents(text)
	}
	annot.Update()
	return annot, nil
}

// ApplyRedactions applies all redaction annotations on this page.
func (p *Page) ApplyRedactions() error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.ApplyRedactions(p.ctx, p.doc.doc.Doc, p.page.Page)
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

// AddLink adds a link to the page pointing to a URI.
func (p *Page) AddLink(rect Rect, uri string) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.CreateLink(p.ctx, p.doc.doc.Doc, p.page.Page,
		rect.X0, rect.Y0, rect.X1, rect.Y1, uri, -1)
}

// SearchFor searches for text on this page, returning bounding rectangles of matches.
func (p *Page) SearchFor(text string, maxHits int) ([]Rect, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	if maxHits <= 0 {
		maxHits = 100
	}

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

// CropBox returns the page crop box.
func (p *Page) CropBox() Rect {
	if p.page == nil || p.doc == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := cgo_bindings.PageBox(p.ctx, p.doc.doc.Doc, p.index, "CropBox")
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// SetCropBox sets the page crop box.
func (p *Page) SetCropBox(r Rect) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.SetPageBox(p.ctx, p.doc.doc.Doc, p.index, "CropBox", r.X0, r.Y0, r.X1, r.Y1)
}

// MediaBox returns the page media box.
func (p *Page) MediaBox() Rect {
	if p.page == nil || p.doc == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := cgo_bindings.PageBox(p.ctx, p.doc.doc.Doc, p.index, "MediaBox")
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// SetMediaBox sets the page media box.
func (p *Page) SetMediaBox(r Rect) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.SetPageBox(p.ctx, p.doc.doc.Doc, p.index, "MediaBox", r.X0, r.Y0, r.X1, r.Y1)
}

// Widgets returns all form widgets on this page.
func (p *Page) Widgets() ([]*Widget, error) {
	if p.page == nil || p.doc == nil {
		return nil, fmt.Errorf("page is nil")
	}
	var widgets []*Widget
	w := cgo_bindings.FirstWidget(p.ctx, p.doc.doc.Doc, p.page.Page)
	for w != nil {
		widgets = append(widgets, &Widget{ctx: p.ctx, widget: w})
		w = w.Next()
	}
	return widgets, nil
}

// String returns a string representation of the page.
func (p *Page) String() string {
	if p.page == nil {
		return "Page(<closed>)"
	}
	return fmt.Sprintf("Page(%v)", p.Rect())
}
