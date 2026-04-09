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
	"unsafe"
)

// Document represents a MuPDF document.
type Document struct {
	Ctx *Context
	Doc *C.fz_document
	Pdf *C.pdf_document // non-nil if it's a PDF document
}

// Open opens a document from a file.
func Open(ctx *Context, filename string) (*Document, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var doc *C.fz_document
	ctx.WithLock(func() {
		doc = C.gomupdf_open_document(ctx.ctx, cfilename)
	})

	if doc == nil {
		return nil, errors.New("failed to open document")
	}

	d := &Document{
		Ctx: ctx,
		Doc:  doc,
		Pdf:  C.pdf_document_from_fz_document(ctx.ctx, doc),
	}

	// Note: No SetFinalizer - caller must explicitly call Destroy()
	return d, nil
}

// OpenStream opens a document from a byte stream.
func OpenStream(ctx *Context, data []byte, filetype string) (*Document, error) {
	var doc *C.fz_document
	cfiletype := C.CString(filetype)
	defer C.free(unsafe.Pointer(cfiletype))

	ctx.WithLock(func() {
		doc = C.gomupdf_open_document_with_stream(
			ctx.ctx,
			cfiletype,
			(*C.uchar)(unsafe.Pointer(&data[0])),
			C.size_t(len(data)),
		)
	})

	if doc == nil {
		return nil, errors.New("failed to open document from stream")
	}

	d := &Document{
		Ctx: ctx,
		Doc:  doc,
		Pdf:  C.pdf_document_from_fz_document(ctx.ctx, doc),
	}

	// Note: No SetFinalizer - caller must explicitly call Destroy()
	return d, nil
}

// NewPDF creates a new empty PDF document.
func NewPDF(ctx *Context) (*Document, error) {
	var doc *C.fz_document

	ctx.WithLock(func() {
		doc = C.gomupdf_new_pdf_document(ctx.ctx)
	})

	if doc == nil {
		return nil, errors.New("failed to create new PDF document")
	}

	d := &Document{
		Ctx: ctx,
		Doc:  doc,
		Pdf:  C.pdf_document_from_fz_document(ctx.ctx, doc),
	}

	// Note: No SetFinalizer - caller must explicitly call Destroy()
	return d, nil
}

// Destroy releases the document.
func (d *Document) Destroy() {
	if d.Doc != nil && d.Ctx != nil {
		d.Ctx.WithLock(func() {
			C.gomupdf_drop_document(d.Ctx.ctx, d.Doc)
		})
		d.Doc = nil
		d.Pdf = nil
	}
}

// PageCount returns the number of pages in the document.
func (d *Document) PageCount() int {
	if d.Doc == nil || d.Ctx == nil {
		return 0
	}
	var count C.int
	d.Ctx.WithLock(func() {
		count = C.gomupdf_page_count(d.Ctx.ctx, d.Doc)
	})
	return int(count)
}

// IsPDF returns true if the document is a PDF.
func (d *Document) IsPDF() bool {
	return d.Pdf != nil
}

// NeedsPassword returns true if the document requires a password.
func (d *Document) NeedsPassword() bool {
	if d.Doc == nil || d.Ctx == nil {
		return false
	}
	var needs C.int
	d.Ctx.WithLock(func() {
		needs = C.gomupdf_needs_password(d.Ctx.ctx, d.Doc)
	})
	return needs != 0
}

// Authenticate checks if the password is correct.
func (d *Document) Authenticate(password string) bool {
	if d.Doc == nil || d.Ctx == nil {
		return false
	}
	cpassword := C.CString(password)
	defer C.free(unsafe.Pointer(cpassword))

	var ok C.int
	d.Ctx.WithLock(func() {
		ok = C.gomupdf_authenticate_password(d.Ctx.ctx, d.Doc, cpassword)
	})
	return ok != 0
}

// Metadata returns the document metadata.
func (d *Document) Metadata() map[string]string {
	if d.Doc == nil || d.Ctx == nil {
		return nil
	}

	// Keys to try: first the short name, then the info: prefixed name
	keys := []struct {
		short string
		long  string
	}{
		{"format", ""},
		{"title", "info:Title"},
		{"author", "info:Author"},
		{"subject", "info:Subject"},
		{"keywords", "info:Keywords"},
		{"creator", "info:Creator"},
		{"producer", "info:Producer"},
		{"creationDate", "info:CreationDate"},
		{"modDate", "info:ModDate"},
	}

	metadata := make(map[string]string)

	d.Ctx.WithLock(func() {
		for _, k := range keys {
			// Try short key first
			ckey := C.CString(k.short)
			cvalue := C.gomupdf_document_metadata(d.Ctx.ctx, d.Doc, ckey)
			C.free(unsafe.Pointer(ckey))

			val := ""
			if cvalue != nil {
				val = C.GoString(cvalue)
			}

			// If empty and there's a long form, try that
			if val == "" && k.long != "" {
				ckey = C.CString(k.long)
				cvalue = C.gomupdf_document_metadata(d.Ctx.ctx, d.Doc, ckey)
				C.free(unsafe.Pointer(ckey))
				if cvalue != nil {
					val = C.GoString(cvalue)
				}
			}

			metadata[k.short] = val
		}
	})

	return metadata
}
