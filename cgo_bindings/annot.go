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

// Annot represents a PDF annotation (wraps pdf_annot).
type Annot struct {
	ctx   *Context
	annot *C.pdf_annot
	doc   *C.fz_document
	page  *C.fz_page
}

// FirstAnnot returns the first annotation on a page.
func FirstAnnot(ctx *Context, doc *C.fz_document, page *C.fz_page) *Annot {
	if ctx == nil || doc == nil || page == nil {
		return nil
	}
	var annot *C.pdf_annot
	ctx.WithLock(func() {
		annot = C.gomupdf_pdf_first_annot(ctx.ctx, doc, page)
	})
	if annot == nil {
		return nil
	}
	return &Annot{ctx: ctx, annot: annot, doc: doc, page: page}
}

// Next returns the next annotation.
func (a *Annot) Next() *Annot {
	if a.annot == nil || a.ctx == nil {
		return nil
	}
	var next *C.pdf_annot
	a.ctx.WithLock(func() {
		next = C.gomupdf_pdf_next_annot(a.ctx.ctx, a.annot)
	})
	if next == nil {
		return nil
	}
	return &Annot{ctx: a.ctx, annot: next, doc: a.doc, page: a.page}
}

// Type returns the annotation type.
func (a *Annot) Type() int {
	if a.annot == nil || a.ctx == nil {
		return -1
	}
	var typ C.int
	a.ctx.WithLock(func() {
		typ = C.gomupdf_pdf_annot_type(a.ctx.ctx, a.annot)
	})
	return int(typ)
}

// Rect returns the annotation rectangle.
func (a *Annot) Rect() (x0, y0, x1, y1 float64) {
	if a.annot == nil || a.ctx == nil {
		return 0, 0, 0, 0
	}
	var r C.fz_rect
	a.ctx.WithLock(func() {
		r = C.gomupdf_pdf_annot_rect(a.ctx.ctx, a.annot)
	})
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}

// Contents returns the annotation text content.
func (a *Annot) Contents() string {
	if a.annot == nil || a.ctx == nil {
		return ""
	}
	var s *C.char
	a.ctx.WithLock(func() {
		s = C.gomupdf_pdf_annot_contents(a.ctx.ctx, a.annot)
	})
	return C.GoString(s)
}

// SetContents sets the annotation text content.
func (a *Annot) SetContents(text string) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_contents(a.ctx.ctx, a.annot, ctext)
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot contents: %s", GetLastError())
	}
	return nil
}

// Color returns the annotation color as RGBA.
func (a *Annot) Color() (r, g, b, alpha float64, n int) {
	if a.annot == nil || a.ctx == nil {
		return 0, 0, 0, 1, 0
	}
	var cr, cg, cb, ca C.float
	var cn C.int
	a.ctx.WithLock(func() {
		cn = C.gomupdf_pdf_annot_color(a.ctx.ctx, a.annot, &cr, &cg, &cb, &ca)
	})
	return float64(cr), float64(cg), float64(cb), float64(ca), int(cn)
}

// SetColor sets the annotation color as RGBA.
func (a *Annot) SetColor(r, g, b, alpha float64) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_color(a.ctx.ctx, a.annot, C.float(r), C.float(g), C.float(b), C.float(alpha))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot color: %s", GetLastError())
	}
	return nil
}

// Opacity returns the annotation opacity.
func (a *Annot) Opacity() float64 {
	if a.annot == nil || a.ctx == nil {
		return 1.0
	}
	var op C.float
	a.ctx.WithLock(func() {
		op = C.gomupdf_pdf_annot_opacity(a.ctx.ctx, a.annot)
	})
	return float64(op)
}

// SetOpacity sets the annotation opacity.
func (a *Annot) SetOpacity(opacity float64) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_opacity(a.ctx.ctx, a.annot, C.float(opacity))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot opacity: %s", GetLastError())
	}
	return nil
}

// Flags returns the annotation flags.
func (a *Annot) Flags() int {
	if a.annot == nil || a.ctx == nil {
		return 0
	}
	var flags C.int
	a.ctx.WithLock(func() {
		flags = C.gomupdf_pdf_annot_flags(a.ctx.ctx, a.annot)
	})
	return int(flags)
}

// SetFlags sets the annotation flags.
func (a *Annot) SetFlags(flags int) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_flags(a.ctx.ctx, a.annot, C.int(flags))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot flags: %s", GetLastError())
	}
	return nil
}

// Border returns the annotation border width.
func (a *Annot) Border() float64 {
	if a.annot == nil || a.ctx == nil {
		return 0
	}
	var w C.float
	a.ctx.WithLock(func() {
		w = C.gomupdf_pdf_annot_border(a.ctx.ctx, a.annot)
	})
	return float64(w)
}

// SetBorder sets the annotation border width.
func (a *Annot) SetBorder(width float64) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_border(a.ctx.ctx, a.annot, C.float(width))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot border: %s", GetLastError())
	}
	return nil
}

// Title returns the annotation title.
func (a *Annot) Title() string {
	if a.annot == nil || a.ctx == nil {
		return ""
	}
	var s *C.char
	a.ctx.WithLock(func() {
		s = C.gomupdf_pdf_annot_title(a.ctx.ctx, a.annot)
	})
	return C.GoString(s)
}

// SetTitle sets the annotation title.
func (a *Annot) SetTitle(title string) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_title(a.ctx.ctx, a.annot, ctitle)
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot title: %s", GetLastError())
	}
	return nil
}

// Update saves changes to the annotation.
func (a *Annot) Update() error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_update_annot(a.ctx.ctx, a.doc, a.annot)
	})
	if rc != 0 {
		return fmt.Errorf("failed to update annot: %s", GetLastError())
	}
	return nil
}

// Delete removes the annotation from the page.
func (a *Annot) Delete() error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_delete_annot(a.ctx.ctx, a.doc, a.page, a.annot)
	})
	if rc != 0 {
		return fmt.Errorf("failed to delete annot: %s", GetLastError())
	}
	a.annot = nil
	return nil
}

// SetRect sets the annotation rectangle.
func (a *Annot) SetRect(x0, y0, x1, y1 float64) error {
	if a.annot == nil || a.ctx == nil {
		return errors.New("annot is nil")
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_rect(a.ctx.ctx, a.annot, C.float(x0), C.float(y0), C.float(x1), C.float(y1))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set annot rect: %s", GetLastError())
	}
	return nil
}

// QuadPoints returns the annotation's quad points.
func (a *Annot) QuadPoints() []Rect {
	if a.annot == nil || a.ctx == nil {
		return nil
	}
	var count C.int
	var quads *C.fz_quad
	a.ctx.WithLock(func() {
		quads = C.gomupdf_pdf_annot_quad_points(a.ctx.ctx, a.annot, &count)
	})
	if quads == nil || count <= 0 {
		return nil
	}
	defer C.free(unsafe.Pointer(quads))
	// 转换 fz_quad 数组为 Rect 列表
	// quads 是 C 分配的数组，需要逐个访问
	result := make([]Rect, count)
	for i := 0; i < int(count); i++ {
		q := C.gomupdf_quad_rect(*(*C.fz_quad)(unsafe.Pointer(uintptr(unsafe.Pointer(quads)) + uintptr(i)*unsafe.Sizeof(*quads))))
		result[i] = Rect{X0: float64(q.x0), Y0: float64(q.y0), X1: float64(q.x1), Y1: float64(q.y1)}
	}
	return result
}

// SetQuadPoints sets the annotation's quad points.
func (a *Annot) SetQuadPoints(quads []Rect) error {
	if a.annot == nil || a.ctx == nil || len(quads) == 0 {
		return errors.New("invalid arguments")
	}
	// Convert Rects to fz_quads
	cquads := make([]C.fz_quad, len(quads))
	for i, q := range quads {
		cquads[i] = C.fz_quad{
			ul: C.fz_point{x: C.float(q.X0), y: C.float(q.Y1)},
			ur: C.fz_point{x: C.float(q.X1), y: C.float(q.Y1)},
			ll: C.fz_point{x: C.float(q.X0), y: C.float(q.Y0)},
			lr: C.fz_point{x: C.float(q.X1), y: C.float(q.Y0)},
		}
	}
	var rc C.int
	a.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_annot_quad_points(a.ctx.ctx, a.annot, C.int(len(quads)), &cquads[0])
	})
	if rc != 0 {
		return fmt.Errorf("failed to set quad points: %s", GetLastError())
	}
	return nil
}

// CreateAnnot creates a new annotation on a page.
func CreateAnnot(ctx *Context, doc *C.fz_document, page *C.fz_page, annotType int) (*Annot, error) {
	if ctx == nil || doc == nil || page == nil {
		return nil, errors.New("nil arguments")
	}
	var annot *C.pdf_annot
	ctx.WithLock(func() {
		annot = C.gomupdf_pdf_create_annot(ctx.ctx, doc, page, C.int(annotType))
	})
	if annot == nil {
		return nil, fmt.Errorf("failed to create annotation type %d: %s", annotType, GetLastError())
	}
	return &Annot{ctx: ctx, annot: annot, doc: doc, page: page}, nil
}

// ApplyRedactions applies all redaction annotations on a page.
func ApplyRedactions(ctx *Context, doc *C.fz_document, page *C.fz_page) error {
	if ctx == nil || doc == nil || page == nil {
		return errors.New("nil arguments")
	}
	var rc C.int
	ctx.WithLock(func() {
		rc = C.gomupdf_pdf_apply_redactions(ctx.ctx, doc, page)
	})
	if rc != 0 {
		return fmt.Errorf("failed to apply redactions: %s", GetLastError())
	}
	return nil
}

// PopupRect returns the popup window rectangle.
func (a *Annot) PopupRect() (x0, y0, x1, y1 float64) {
	if a.annot == nil || a.ctx == nil {
		return 0, 0, 0, 0
	}
	var r C.fz_rect
	a.ctx.WithLock(func() {
		r = C.gomupdf_pdf_annot_popup(a.ctx.ctx, a.annot)
	})
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}
