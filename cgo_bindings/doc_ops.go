package cgo

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/mupdf/lib -lmupdf -lmupdfcpp
#cgo CFLAGS: -I/opt/homebrew/opt/mupdf/include

#include "bindings.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// SaveOptions controls how a PDF document is saved.
type SaveOptions struct {
	Garbage          int  // 0=off, 1=gc, 2=renumber, 3=deduplicate
	Clean            bool // Clean content streams
	Compress         int  // 0=off, 1=zlib, 2=brotli
	CompressImages   bool // Compress image streams
	CompressFonts    bool // Compress font streams
	Decompress       bool // Decompress all streams
	Linear           bool // Linearize for web
	ASCII            bool // ASCII hex encode
	Incremental      bool // Write only changed objects
	Pretty           bool // Pretty print dicts
	Sanitize         bool // Sanitize content streams
	Appearance       bool // Recreate appearance streams
	PreserveMetadata bool // Keep metadata unchanged
}

// SaveDocument saves a PDF document to a file.
func (d *Document) SaveDocument(filename string, opts *SaveOptions) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var cOpts SaveOptions
	if opts != nil {
		cOpts = *opts
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_save_document(
			d.Ctx.ctx, d.Doc, cfilename,
			C.int(cOpts.Garbage), boolToInt(cOpts.Clean), C.int(cOpts.Compress),
			boolToInt(cOpts.CompressImages), boolToInt(cOpts.CompressFonts),
			boolToInt(cOpts.Decompress), boolToInt(cOpts.Linear), boolToInt(cOpts.ASCII),
			boolToInt(cOpts.Incremental), boolToInt(cOpts.Pretty),
			boolToInt(cOpts.Sanitize), boolToInt(cOpts.Appearance), boolToInt(cOpts.PreserveMetadata),
		)
	})

	if rc != 0 {
		return fmt.Errorf("failed to save document: %s", GetLastError())
	}
	return nil
}

// WriteDocument writes a PDF document to bytes.
func (d *Document) WriteDocument(opts *SaveOptions) ([]byte, error) {
	if d.Doc == nil || d.Ctx == nil {
		return nil, errors.New("document is nil")
	}

	var cOpts SaveOptions
	if opts != nil {
		cOpts = *opts
	}

	var outData *C.uchar
	var outLen C.size_t

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_write_document(
			d.Ctx.ctx, d.Doc,
			&outData, &outLen,
			C.int(cOpts.Garbage), boolToInt(cOpts.Clean), C.int(cOpts.Compress),
			boolToInt(cOpts.CompressImages), boolToInt(cOpts.CompressFonts),
			boolToInt(cOpts.Decompress), boolToInt(cOpts.Linear), boolToInt(cOpts.ASCII),
			boolToInt(cOpts.Incremental), boolToInt(cOpts.Pretty),
			boolToInt(cOpts.Sanitize), boolToInt(cOpts.Appearance), boolToInt(cOpts.PreserveMetadata),
		)
	})

	if rc != 0 || outData == nil {
		return nil, fmt.Errorf("failed to write document: %s", GetLastError())
	}
	defer C.gomupdf_free(unsafe.Pointer(outData))

	return C.GoBytes(unsafe.Pointer(outData), C.int(outLen)), nil
}

// InsertPage creates and inserts a new blank page.
func (d *Document) InsertPage(at int, x0, y0, x1, y1 float64, rotation int) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_insert_page(
			d.Ctx.ctx, d.Doc, C.int(at),
			C.float(x0), C.float(y0), C.float(x1), C.float(y1), C.int(rotation),
		)
	})

	if rc != 0 {
		return fmt.Errorf("failed to insert page: %s", GetLastError())
	}
	return nil
}

// DeletePage removes a page by number.
func (d *Document) DeletePage(number int) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_delete_page(d.Ctx.ctx, d.Doc, C.int(number))
	})

	if rc != 0 {
		return fmt.Errorf("failed to delete page: %s", GetLastError())
	}
	return nil
}

// DeletePageRange removes pages from start to end (exclusive).
func (d *Document) DeletePageRange(start, end int) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_delete_page_range(d.Ctx.ctx, d.Doc, C.int(start), C.int(end))
	})

	if rc != 0 {
		return fmt.Errorf("failed to delete page range: %s", GetLastError())
	}
	return nil
}

// SetMetadata sets a metadata key-value pair on the document.
func (d *Document) SetMetadata(key, value string) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_metadata(d.Ctx.ctx, d.Doc, ckey, cvalue)
	})

	if rc != 0 {
		return fmt.Errorf("failed to set metadata: %s", GetLastError())
	}
	return nil
}

// Permissions returns the document permission flags.
func (d *Document) Permissions() int {
	if d.Doc == nil || d.Ctx == nil {
		return 0
	}
	var perm C.int
	d.Ctx.WithLock(func() {
		perm = C.gomupdf_pdf_permissions(d.Ctx.ctx, d.Doc)
	})
	return int(perm)
}

// OutlineEntry represents a single entry in the document's table of contents.
type OutlineEntry struct {
	Title  string
	Page   int
	Level  int
	URI    string
	IsOpen bool
}

// GetOutline loads and returns the document's table of contents.
func (d *Document) GetOutline() ([]OutlineEntry, error) {
	if d.Doc == nil || d.Ctx == nil {
		return nil, errors.New("document is nil")
	}

	var count C.int
	d.Ctx.WithLock(func() {
		count = C.gomupdf_pdf_outline_count(d.Ctx.ctx, d.Doc)
	})

	if count <= 0 {
		return nil, nil
	}

	entries := make([]OutlineEntry, count)
	for i := 0; i < int(count); i++ {
		var title, uri *C.char
		var page, level, isOpen C.int

		d.Ctx.WithLock(func() {
			C.gomupdf_pdf_outline_get(d.Ctx.ctx, d.Doc, C.int(i),
				&title, &page, &level, &uri, &isOpen)
		})

		entries[i] = OutlineEntry{
			Title:  C.GoString(title),
			Page:   int(page),
			Level:  int(level),
			URI:    C.GoString(uri),
			IsOpen: isOpen != 0,
		}
	}

	return entries, nil
}

// LinkInfo represents a hyperlink on a page.
type LinkInfo struct {
	Rect    Rect
	URI     string
	Page    int // resolved page number (-1 if external)
	Next    *LinkInfo
}

// Rect represents a rectangle in PDF coordinates.
type Rect struct {
	X0, Y0, X1, Y1 float64
}

// LoadLinks loads all links from a page.
func LoadLinks(ctx *Context, page *Page) ([]LinkInfo, error) {
	if ctx == nil || page == nil {
		return nil, errors.New("nil context or page")
	}

	var clink *C.fz_link
	ctx.WithLock(func() {
		clink = C.gomupdf_page_load_links(ctx.ctx, page.Page)
	})
	if clink == nil {
		return nil, nil
	}
	defer C.gomupdf_drop_link(ctx.ctx, clink)

	var links []LinkInfo
	for l := clink; l != nil; l = C.gomupdf_link_next(l) {
		r := C.gomupdf_link_rect(l)
		uri := C.gomupdf_link_uri(l)

		link := LinkInfo{
			Rect: Rect{X0: float64(r.x0), Y0: float64(r.y0), X1: float64(r.x1), Y1: float64(r.y1)},
			URI:  C.GoString(uri),
			Page: -1,
		}
		links = append(links, link)
	}

	return links, nil
}

// SearchText searches for text in a structured text page, returning matching quads.
func SearchText(ctx *Context, stextPage *C.fz_stext_page, needle string, maxHits int) []Rect {
	if ctx == nil || stextPage == nil || needle == "" || maxHits <= 0 {
		return nil
	}

	cneedle := C.CString(needle)
	defer C.free(unsafe.Pointer(cneedle))

	hits := make([]C.fz_quad, maxHits)

	var count C.int
	ctx.WithLock(func() {
		count = C.gomupdf_search_text(ctx.ctx, stextPage, cneedle, C.int(maxHits), &hits[0])
	})

	if count <= 0 {
		return nil
	}

	rects := make([]Rect, count)
	for i := 0; i < int(count); i++ {
		r := C.gomupdf_quad_rect(hits[i])
		rects[i] = Rect{X0: float64(r.x0), Y0: float64(r.y0), X1: float64(r.x1), Y1: float64(r.y1)}
	}
	return rects
}

// helper
func boolToInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
