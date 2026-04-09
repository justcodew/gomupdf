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
)

// Page represents a page in a document.
type Page struct {
	Ctx *Context
	Page *C.fz_page
	Doc  *C.fz_document
}

// LoadPage loads a page by its number (0-indexed).
func LoadPage(ctx *Context, doc *C.fz_document, number int) (*Page, error) {
	var page *C.fz_page

	ctx.WithLock(func() {
		page = C.gomupdf_load_page(ctx.ctx, doc, C.int(number))
	})

	if page == nil {
		return nil, errors.New("failed to load page")
	}

	p := &Page{
		Ctx:  ctx,
		Page: page,
		Doc:  doc,
	}

	// Note: Don't set finalizer here - let Document handle cleanup
	// Finalizer can cause use-after-free if TextPage is still using the fz_page
	return p, nil
}

// Destroy releases the page.
func (p *Page) Destroy() {
	if p.Page != nil && p.Ctx != nil {
		p.Ctx.WithLock(func() {
			C.gomupdf_drop_page(p.Ctx.ctx, p.Page)
		})
		p.Page = nil
	}
}

// Rect returns the page bounding rectangle.
func (p *Page) Rect() (x0, y0, x1, y1 float64) {
	if p.Page == nil || p.Ctx == nil {
		return
	}

	var rect C.fz_rect
	p.Ctx.WithLock(func() {
		rect = C.gomupdf_page_rect(p.Ctx.ctx, p.Page)
	})

	return float64(rect.x0), float64(rect.y0), float64(rect.x1), float64(rect.y1)
}

// Rotation returns the page rotation in degrees (0, 90, 180, 270).
// This only works for PDF documents; returns 0 for other formats.
func (p *Page) Rotation() int {
	if p.Page == nil || p.Ctx == nil {
		return 0
	}

	var rot C.int
	p.Ctx.WithLock(func() {
		rot = C.gomupdf_page_rotation(p.Ctx.ctx, p.Page)
	})

	return int(rot)
}

// PDFPageRotation returns the rotation for a PDF page by number (0-indexed).
// This queries the /Rotate entry in the page dictionary.
func PDFPageRotation(ctx *Context, doc *C.fz_document, pageNum int) int {
	if ctx == nil || doc == nil {
		return 0
	}
	var rot C.int
	ctx.WithLock(func() {
		rot = C.gomupdf_pdf_page_rotation(ctx.ctx, doc, C.int(pageNum))
	})
	return int(rot)
}
