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

// DisplayList represents a cached display list for a page.
type DisplayList struct {
	ctx  *Context
	list *C.fz_display_list
}

// NewDisplayList creates a new display list.
func NewDisplayList(ctx *Context, x0, y0, x1, y1 float64) (*DisplayList, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	bounds := C.fz_rect{x0: C.float(x0), y0: C.float(y0), x1: C.float(x1), y1: C.float(y1)}
	var list *C.fz_display_list
	ctx.WithLock(func() {
		list = C.gomupdf_new_display_list(ctx.ctx, bounds)
	})
	if list == nil {
		return nil, errors.New("failed to create display list")
	}
	return &DisplayList{ctx: ctx, list: list}, nil
}

// RunPageToDisplayList runs a page into a display list.
func RunPageToDisplayList(ctx *Context, page *C.fz_page, list *DisplayList, a, b, c, d, e, f float64) error {
	if ctx == nil || page == nil || list == nil {
		return errors.New("nil arguments")
	}
	ctm := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	ctx.WithLock(func() {
		C.gomupdf_run_page_to_list(ctx.ctx, page, list.list, ctm)
	})
	return nil
}

// GetPixmap renders the display list to a pixmap.
func (dl *DisplayList) GetPixmap(a, b, c, d, e, f float64, alpha bool) (*Pixmap, error) {
	if dl.list == nil || dl.ctx == nil {
		return nil, errors.New("display list is nil")
	}
	ctm := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	var alphaInt C.int
	if alpha {
		alphaInt = 1
	}
	var pix *C.fz_pixmap
	dl.ctx.WithLock(func() {
		pix = C.gomupdf_display_list_get_pixmap(dl.ctx.ctx, dl.list, ctm, alphaInt)
	})
	if pix == nil {
		return nil, errors.New("failed to render display list")
	}
	return &Pixmap{ctx: dl.ctx, pix: pix}, nil
}

// Destroy releases the display list.
func (dl *DisplayList) Destroy() {
	if dl.list != nil && dl.ctx != nil {
		dl.ctx.WithLock(func() {
			C.gomupdf_drop_display_list(dl.ctx.ctx, dl.list)
		})
		dl.list = nil
	}
}

// PageBox returns a page box (cropbox, mediabox, etc.).
func PageBox(ctx *Context, doc *C.fz_document, pageNum int, boxType string) (x0, y0, x1, y1 float64) {
	if ctx == nil || doc == nil {
		return 0, 0, 0, 0
	}
	var r C.fz_rect
	ctx.WithLock(func() {
		switch boxType {
		case "CropBox":
			r = C.gomupdf_pdf_page_cropbox(ctx.ctx, doc, C.int(pageNum))
		case "MediaBox":
			r = C.gomupdf_pdf_page_mediabox(ctx.ctx, doc, C.int(pageNum))
		}
	})
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}

// SetPageBox sets a page box.
func SetPageBox(ctx *Context, doc *C.fz_document, pageNum int, boxType string, x0, y0, x1, y1 float64) error {
	if ctx == nil || doc == nil {
		return errors.New("nil arguments")
	}
	var rc C.int
	ctx.WithLock(func() {
		switch boxType {
		case "CropBox":
			rc = C.gomupdf_pdf_set_page_cropbox(ctx.ctx, doc, C.int(pageNum), C.float(x0), C.float(y0), C.float(x1), C.float(y1))
		case "MediaBox":
			rc = C.gomupdf_pdf_set_page_mediabox(ctx.ctx, doc, C.int(pageNum), C.float(x0), C.float(y0), C.float(x1), C.float(y1))
		}
	})
	if rc != 0 {
		return fmt.Errorf("failed to set %s: %s", boxType, GetLastError())
	}
	return nil
}

// SetPageRotation sets the page rotation.
func SetPageRotation(ctx *Context, doc *C.fz_document, pageNum, rotation int) error {
	if ctx == nil || doc == nil {
		return errors.New("nil arguments")
	}
	var rc C.int
	ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_page_rotation(ctx.ctx, doc, C.int(pageNum), C.int(rotation))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set rotation: %s", GetLastError())
	}
	return nil
}

// XRefLength returns the XRef table length.
func XRefLength(ctx *Context, doc *C.fz_document) int {
	if ctx == nil || doc == nil {
		return 0
	}
	var len C.int
	ctx.WithLock(func() {
		len = C.gomupdf_pdf_xref_length(ctx.ctx, doc)
	})
	return int(len)
}

// XRefGetKey returns the value of a key in an XRef object.
func XRefGetKey(ctx *Context, doc *C.fz_document, xref int, key string) string {
	if ctx == nil || doc == nil {
		return ""
	}
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	var s *C.char
	ctx.WithLock(func() {
		s = C.gomupdf_pdf_xref_get_key(ctx.ctx, doc, C.int(xref), ckey)
	})
	return C.GoString(s)
}

// XRefIsStream returns whether an XRef entry is a stream.
func XRefIsStream(ctx *Context, doc *C.fz_document, xref int) bool {
	if ctx == nil || doc == nil {
		return false
	}
	var result C.int
	ctx.WithLock(func() {
		result = C.gomupdf_pdf_xref_is_stream(ctx.ctx, doc, C.int(xref))
	})
	return result != 0
}

// EmbeddedFileCount returns the number of embedded files.
func EmbeddedFileCount(ctx *Context, doc *C.fz_document) int {
	if ctx == nil || doc == nil {
		return 0
	}
	var count C.int
	ctx.WithLock(func() {
		count = C.gomupdf_pdf_embedded_file_count(ctx.ctx, doc)
	})
	return int(count)
}

// EmbeddedFileName returns the name of an embedded file.
func EmbeddedFileName(ctx *Context, doc *C.fz_document, idx int) string {
	if ctx == nil || doc == nil {
		return ""
	}
	var s *C.char
	ctx.WithLock(func() {
		s = C.gomupdf_pdf_embedded_file_name(ctx.ctx, doc, C.int(idx))
	})
	return C.GoString(s)
}

// EmbeddedFileGet returns the data of an embedded file.
func EmbeddedFileGet(ctx *Context, doc *C.fz_document, idx int) ([]byte, error) {
	if ctx == nil || doc == nil {
		return nil, errors.New("nil arguments")
	}
	var outData *C.uchar
	var outLen C.size_t
	ctx.WithLock(func() {
		outData = C.gomupdf_pdf_embedded_file_get(ctx.ctx, doc, C.int(idx), &outLen)
	})
	if outData == nil || outLen == 0 {
		return nil, fmt.Errorf("failed to get embedded file %d", idx)
	}
	defer C.free(unsafe.Pointer(outData))
	return C.GoBytes(unsafe.Pointer(outData), C.int(outLen)), nil
}

// AddEmbeddedFile adds an embedded file to the document.
func AddEmbeddedFile(ctx *Context, doc *C.fz_document, filename, mimetype string, data []byte) error {
	if ctx == nil || doc == nil || filename == "" || len(data) == 0 {
		return errors.New("invalid arguments")
	}
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	var cmimetype *C.char
	if mimetype != "" {
		cmimetype = C.CString(mimetype)
		defer C.free(unsafe.Pointer(cmimetype))
	}
	var rc C.int
	ctx.WithLock(func() {
		rc = C.gomupdf_pdf_add_embedded_file(ctx.ctx, doc, cfilename, cmimetype,
			(*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	})
	if rc != 0 {
		return fmt.Errorf("failed to add embedded file: %s", GetLastError())
	}
	return nil
}

// CreateLink creates a new link on a page.
func CreateLink(ctx *Context, doc *C.fz_document, page *C.fz_page, x0, y0, x1, y1 float64, uri string, pageNum int) error {
	if ctx == nil || doc == nil || page == nil {
		return errors.New("nil arguments")
	}
	curi := C.CString(uri)
	defer C.free(unsafe.Pointer(curi))
	var rc C.int
	ctx.WithLock(func() {
		rc = C.gomupdf_pdf_create_link(ctx.ctx, doc, page,
			C.float(x0), C.float(y0), C.float(x1), C.float(y1), curi, C.int(pageNum))
	})
	if rc != 0 {
		return fmt.Errorf("failed to create link: %s", GetLastError())
	}
	return nil
}

// PageContentBegin starts writing page content.
func PageContentBegin(ctx *Context, doc *C.fz_document, page *C.fz_page) (*C.fz_buffer, error) {
	if ctx == nil || doc == nil || page == nil {
		return nil, errors.New("nil arguments")
	}
	var buf *C.fz_buffer
	ctx.WithLock(func() {
		buf = C.gomupdf_pdf_page_write_begin(ctx.ctx, doc, page)
	})
	if buf == nil {
		return nil, fmt.Errorf("failed to begin page content: %s", GetLastError())
	}
	return buf, nil
}

// PageContentEnd finishes writing page content.
func PageContentEnd(ctx *Context, doc *C.fz_document, page *C.fz_page, buf *C.fz_buffer) error {
	if ctx == nil || doc == nil || page == nil || buf == nil {
		return errors.New("nil arguments")
	}
	var rc C.int
	ctx.WithLock(func() {
		rc = C.gomupdf_pdf_page_write_end(ctx.ctx, doc, page, buf)
	})
	if rc != 0 {
		return fmt.Errorf("failed to end page content: %s", GetLastError())
	}
	return nil
}
