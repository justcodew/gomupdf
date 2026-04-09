package pdf

import (
	"fmt"
	"image"
	"image/color"
	"io"

	cgo "github.com/go-pymupdf/gomupdf/cgo_bindings"
	"github.com/go-pymupdf/gomupdf/fitz"
)

// getContext returns a new MuPDF context
func getContext() *cgo.Context {
	return cgo.New()
}

// PDFDocument represents a PDF document
type PDFDocument struct {
	ctx      *cgo.Context
	doc      *cgo.Document
	fitz     *fitz.Document
	filename string // original filename for reopening
}

// OpenPDF opens a PDF file
func OpenPDF(filename string) (*PDFDocument, error) {
	ctx := getContext()

	// Open with fitz for higher-level operations, using the same context
	fitzDoc, err := fitz.OpenWithContext(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF with fitz: %w", err)
	}

	// Get the underlying cgo document from fitz
	cgoDoc := fitzDoc.GetDoc()

	return &PDFDocument{
		ctx:      ctx,
		doc:      cgoDoc,
		fitz:     fitzDoc,
		filename: filename,
	}, nil
}

// OpenPDFStream opens a PDF from a stream
func OpenPDFStream(r io.Reader, filetype string) (*PDFDocument, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}

	ctx := getContext()

	// Open at cgo level
	cgoDoc, err := cgo.OpenStream(ctx, data, filetype)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF stream: %w", err)
	}

	// Also open at fitz level for high-level operations
	fitzDoc, err := fitz.OpenStreamWithContext(ctx, data, filetype)
	if err != nil {
		cgoDoc.Destroy()
		return nil, fmt.Errorf("failed to open PDF stream with fitz: %w", err)
	}

	return &PDFDocument{
		ctx:  ctx,
		doc:  cgoDoc,
		fitz: fitzDoc,
	}, nil
}

// Close closes the PDF document
func (p *PDFDocument) Close() {
	if p.doc != nil {
		p.doc.Destroy()
	}
	if p.fitz != nil {
		p.fitz.Close()
	}
}

// PageCount returns the number of pages
func (p *PDFDocument) PageCount() int {
	if p.doc != nil {
		return p.doc.PageCount()
	}
	return 0
}

// IsPDF returns true if this is a PDF document
func (p *PDFDocument) IsPDF() bool {
	if p.doc != nil {
		return p.doc.IsPDF()
	}
	return false
}

// NeedsPassword returns true if password is required
func (p *PDFDocument) NeedsPassword() bool {
	if p.doc != nil {
		return p.doc.NeedsPassword()
	}
	return false
}

// IsEncrypted returns true if document is encrypted
func (p *PDFDocument) IsEncrypted() bool {
	if p.doc != nil {
		return p.doc.NeedsPassword() // Simplified
	}
	return false
}

// Authenticate checks if password is correct
func (p *PDFDocument) Authenticate(password string) bool {
	if p.doc != nil {
		return p.doc.Authenticate(password)
	}
	return false
}

// IsScannedMode determines if PDF is a scanned document
func (p *PDFDocument) IsScannedMode() bool {
	if p.fitz == nil {
		return false
	}

	// Check first 3 pages for text
	maxPages := p.fitz.PageCount()
	if maxPages == 0 {
		return true
	}

	textCount := 0
	for i := 0; i < maxPages && i < 3; i++ {
		page, err := p.fitz.Page(i)
		if err != nil {
			continue
		}
		text, err := page.GetText(nil)
		if err == nil && text != "" {
			textCount += len(text)
		}
		page.Close()
	}

	return textCount == 0
}

// GetPage returns a page by index
func (p *PDFDocument) GetPage(index int) (*PDFPage, error) {
	if p.fitz == nil {
		return nil, fmt.Errorf("document not open")
	}

	page, err := p.fitz.Page(index)
	if err != nil {
		return nil, err
	}

	return &PDFPage{
		fitzPage: page,
		document: p,
	}, nil
}

// PDFPage represents a page in a PDF
type PDFPage struct {
	fitzPage *fitz.Page
	document *PDFDocument
}

// GetFitzPage returns the underlying fitz.Page
func (p *PDFPage) GetFitzPage() *fitz.Page {
	return p.fitzPage
}

// GetPixmap returns a pixmap of the page
func (p *PDFPage) GetPixmap(dpi float64) (*image.RGBA, int, int, error) {
	if p.fitzPage == nil {
		return nil, 0, 0, fmt.Errorf("page not available")
	}

	// Calculate zoom factor based on DPI (72 is default DPI)
	zoom := dpi / 72.0
	matrix := fitz.NewMatrixScale(zoom)

	pixmap, err := p.fitzPage.Pixmap(matrix, false)
	if err != nil {
		return nil, 0, 0, err
	}
	defer pixmap.Close()

	width := pixmap.Width()
	height := pixmap.Height()
	samples := pixmap.Samples()
	n := pixmap.N() // number of components per pixel (3 for RGB, 4 for RGBA)

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Convert samples to image
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if idx+n-1 < len(samples) {
				r, g, b, a := uint8(0), uint8(0), uint8(0), uint8(255)
				r = samples[idx]
				if n > 1 {
					g = samples[idx+1]
				}
				if n > 2 {
					b = samples[idx+2]
				}
				if n > 3 {
					a = samples[idx+3]
				}
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
				idx += n
			}
		}
	}

	return img, width, height, nil
}

// GetTextBlocks extracts text blocks with coordinates
func (p *PDFPage) GetTextBlocks() ([]TextBlock, error) {
	if p.fitzPage == nil {
		return nil, fmt.Errorf("page not available")
	}

	// Use fitz for text extraction
	fitzBlocks, err := p.fitzPage.GetTextBlocks()
	if err != nil {
		return nil, err
	}

	// Convert fitz.TextBlock to pdf.TextBlock
	blocks := make([]TextBlock, 0, len(fitzBlocks))
	for _, fb := range fitzBlocks {
		block := TextBlock{
			Bbox:   []float64{fb.Bbox.X0, fb.Bbox.Y0, fb.Bbox.X1, fb.Bbox.Y1},
			Type:   fb.Type,
			Text:   fb.Text,
			Number: fb.Number,
		}

		// Convert lines
		for _, fl := range fb.Lines {
			line := TextLine{
				Bbox: []float64{fl.Bbox.X0, fl.Bbox.Y0, fl.Bbox.X1, fl.Bbox.Y1},
				Dir:  Point{X: fl.WDir.X, Y: fl.WDir.Y},
			}

			// Convert fragments
			for _, ff := range fl.Frags {
				frag := TextSpan{
					Bbox:   []float64{ff.Bbox.X0, ff.Bbox.Y0, ff.Bbox.X1, ff.Bbox.Y1},
					Origin: Point{X: ff.Origin.X, Y: ff.Origin.Y},
					Text:   ff.Text,
					Size:   ff.Size,
					Font:   ff.Font,
					Color:  0,
					Flags:  0,
				}
				line.Spans = append(line.Spans, frag)
			}
			block.Lines = append(block.Lines, line)
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetTextDict extracts text as structured dictionary
func (p *PDFPage) GetTextDict() (map[string]interface{}, error) {
	if p.fitzPage == nil {
		return nil, fmt.Errorf("page not available")
	}

	// Get raw text dictionary
	return p.fitzPage.GetTextDict()
}

// Rotation returns the page rotation
func (p *PDFPage) Rotation() int {
	if p.fitzPage != nil {
		return p.fitzPage.Rotation()
	}
	return 0
}

// Close closes the page
func (p *PDFPage) Close() {
	if p.fitzPage != nil {
		p.fitzPage.Close()
	}
}

// TextBlock represents a block of text
type TextBlock struct {
	Bbox    []float64        `json:"bbox"`
	Type    int              `json:"type"`
	Text    string           `json:"text"`
	Number  int              `json:"number"`
	Lines   []TextLine       `json:"lines,omitempty"`
	Spans   []TextSpan       `json:"spans,omitempty"`
}

// TextLine represents a line of text
type TextLine struct {
	Bbox    []float64   `json:"bbox"`
	Dir     Point       `json:"dir"`
	Spans   []TextSpan  `json:"spans"`
}

// TextSpan represents a text span
type TextSpan struct {
	Bbox    []float64 `json:"bbox"`
	Origin  Point     `json:"origin"`
	Text    string    `json:"text"`
	Size    float64   `json:"size"`
	Font    string    `json:"font"`
	Color   int       `json:"color"`
	Flags   int       `json:"flags"`
}

// Point represents a 2D point
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
