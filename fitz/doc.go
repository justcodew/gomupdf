package fitz

import (
	"fmt"
	"io"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Document represents a PDF or other document format.
type Document struct {
	ctx *cgo_bindings.Context
	doc *cgo_bindings.Document
}

// Open opens a document from a file using a new context.
func Open(filename string) (*Document, error) {
	ctx := cgo_bindings.NewContext()
	doc, err := cgo_bindings.Open(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// OpenWithContext opens a document from a file using the provided context.
func OpenWithContext(ctx *cgo_bindings.Context, filename string) (*Document, error) {
	doc, err := cgo_bindings.Open(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// OpenStream opens a document from a stream.
func OpenStream(r io.Reader, filetype string) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}
	ctx := cgo_bindings.New()
	doc, err := cgo_bindings.OpenStream(ctx, data, filetype)
	if err != nil {
		return nil, fmt.Errorf("failed to open document from stream: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// OpenStreamWithContext opens a document from a byte slice using the provided context.
func OpenStreamWithContext(ctx *cgo_bindings.Context, data []byte, filetype string) (*Document, error) {
	doc, err := cgo_bindings.OpenStream(ctx, data, filetype)
	if err != nil {
		return nil, fmt.Errorf("failed to open document from stream: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// NewPDF creates a new empty PDF document.
func NewPDF() (*Document, error) {
	ctx := cgo_bindings.New()
	doc, err := cgo_bindings.NewPDF(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// Close closes the document.
func (d *Document) Close() error {
	if d.doc != nil {
		d.doc.Destroy()
		d.doc = nil
	}
	return nil
}

// PageCount returns the number of pages in the document.
func (d *Document) PageCount() int {
	if d.doc == nil {
		return 0
	}
	return d.doc.PageCount()
}

// IsPDF returns true if the document is a PDF.
func (d *Document) IsPDF() bool {
	if d.doc == nil {
		return false
	}
	return d.doc.IsPDF()
}

// NeedsPassword returns true if the document requires a password.
func (d *Document) NeedsPassword() bool {
	if d.doc == nil {
		return false
	}
	return d.doc.NeedsPassword()
}

// Authenticate checks if the password is correct.
func (d *Document) Authenticate(password string) bool {
	if d.doc == nil {
		return false
	}
	return d.doc.Authenticate(password)
}

// Metadata returns the document metadata.
func (d *Document) Metadata() map[string]string {
	if d.doc == nil {
		return nil
	}
	return d.doc.Metadata()
}

// Page returns a Page for the given page number (0-indexed).
func (d *Document) Page(number int) (*Page, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	if number < 0 || number >= d.doc.PageCount() {
		return nil, fmt.Errorf("page number out of range: %d", number)
	}

	page, err := cgo_bindings.LoadPage(d.ctx, d.doc.Doc, number)
	if err != nil {
		return nil, fmt.Errorf("failed to load page: %w", err)
	}

	return &Page{
		ctx:   d.ctx,
		page:  page,
		doc:   d,
		index: number,
	}, nil
}

// SaveOptions contains options for saving a document.
type SaveOptions = cgo_bindings.SaveOptions

// Save saves the document to a file.
func (d *Document) Save(filename string, opts *SaveOptions) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.SaveDocument(filename, opts)
}

// SaveToBytes writes the document to a byte slice.
func (d *Document) SaveToBytes(opts *SaveOptions) ([]byte, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	return d.doc.WriteDocument(opts)
}

// NewPage creates and inserts a new blank page.
func (d *Document) NewPage(at int, width, height float64, rotation int) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.InsertPage(at, 0, 0, width, height, rotation)
}

// DeletePage removes a page by number.
func (d *Document) DeletePage(number int) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.DeletePage(number)
}

// SetMetadata sets a metadata field.
func (d *Document) SetMetadata(key, value string) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.SetMetadata(key, value)
}

// Permissions returns the document permission flags.
func (d *Document) Permissions() int {
	if d.doc == nil {
		return 0
	}
	return d.doc.Permissions()
}

// OutlineEntry represents a TOC entry.
type OutlineEntry = cgo_bindings.OutlineEntry

// GetOutline returns the document's table of contents.
func (d *Document) GetOutline() ([]OutlineEntry, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	return d.doc.GetOutline()
}

// String returns a string representation of the document.
func (d *Document) String() string {
	if d.doc == nil {
		return "Document(<closed>)"
	}
	return fmt.Sprintf("Document(%q, %d pages)", d.Metadata()["format"], d.PageCount())
}

// GetDoc returns the underlying cgo document.
func (d *Document) GetDoc() *cgo_bindings.Document {
	return d.doc
}

// XRefLength returns the XRef table length.
func (d *Document) XRefLength() int {
	if d.doc == nil {
		return 0
	}
	return cgo_bindings.XRefLength(d.ctx, d.doc.Doc)
}

// XRefGetKey returns the value of a key in an XRef object.
func (d *Document) XRefGetKey(xref int, key string) string {
	if d.doc == nil {
		return ""
	}
	return cgo_bindings.XRefGetKey(d.ctx, d.doc.Doc, xref, key)
}

// XRefIsStream returns whether an XRef entry is a stream.
func (d *Document) XRefIsStream(xref int) bool {
	if d.doc == nil {
		return false
	}
	return cgo_bindings.XRefIsStream(d.ctx, d.doc.Doc, xref)
}

// EmbeddedFileCount returns the number of embedded files.
func (d *Document) EmbeddedFileCount() int {
	if d.doc == nil {
		return 0
	}
	return cgo_bindings.EmbeddedFileCount(d.ctx, d.doc.Doc)
}

// EmbeddedFileName returns the name of an embedded file.
func (d *Document) EmbeddedFileName(idx int) string {
	if d.doc == nil {
		return ""
	}
	return cgo_bindings.EmbeddedFileName(d.ctx, d.doc.Doc, idx)
}

// EmbeddedFileGet returns the data of an embedded file.
func (d *Document) EmbeddedFileGet(idx int) ([]byte, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	return cgo_bindings.EmbeddedFileGet(d.ctx, d.doc.Doc, idx)
}

// AddEmbeddedFile adds an embedded file to the document.
func (d *Document) AddEmbeddedFile(filename, mimetype string, data []byte) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return cgo_bindings.AddEmbeddedFile(d.ctx, d.doc.Doc, filename, mimetype, data)
}
