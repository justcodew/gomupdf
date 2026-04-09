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

// Pixmap represents a raster image.
type Pixmap struct {
	ctx *Context
	pix *C.fz_pixmap
}

// RenderPage renders a page to a pixmap.
func RenderPage(ctx *Context, page *C.fz_page, a, b, c, d, e, f float64, alpha bool) (*Pixmap, error) {
	var pix *C.fz_pixmap
	var alphaInt C.int
	if alpha {
		alphaInt = 1
	}

	ctx.WithLock(func() {
		pix = C.gomupdf_render_page(ctx.ctx, page, C.float(a), C.float(b), C.float(c), C.float(d), C.float(e), C.float(f), alphaInt)
	})

	if pix == nil {
		return nil, errors.New("failed to render page")
	}

	p := &Pixmap{
		ctx: ctx,
		pix: pix,
	}

	// Note: No SetFinalizer - caller must explicitly call Destroy()
	// This avoids crashes during shutdown when context might be corrupted
	return p, nil
}

// NewPixmap creates a new pixmap with the given dimensions.
func NewPixmap(ctx *Context, width, height int) (*Pixmap, error) {
	var pix *C.fz_pixmap

	ctx.WithLock(func() {
		cs := C.gomupdf_new_colorspace_rgb(ctx.ctx)
		pix = C.gomupdf_new_pixmap(ctx.ctx, cs, C.int(width), C.int(height))
		// Colorspace reference is kept by pixmap, don't free here
	})

	if pix == nil {
		return nil, errors.New("failed to create pixmap")
	}

	p := &Pixmap{
		ctx: ctx,
		pix: pix,
	}

	// Note: No SetFinalizer - caller must explicitly call Destroy()
	return p, nil
}

// Destroy releases the pixmap.
func (p *Pixmap) Destroy() {
	if p.pix != nil && p.ctx != nil {
		p.ctx.WithLock(func() {
			C.gomupdf_drop_pixmap(p.ctx.ctx, p.pix)
		})
		p.pix = nil
	}
}

// Width returns the pixmap width.
func (p *Pixmap) Width() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var w C.int
	p.ctx.WithLock(func() {
		w = C.gomupdf_pixmap_width(p.ctx.ctx, p.pix)
	})
	return int(w)
}

// Height returns the pixmap height.
func (p *Pixmap) Height() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var h C.int
	p.ctx.WithLock(func() {
		h = C.gomupdf_pixmap_height(p.ctx.ctx, p.pix)
	})
	return int(h)
}

// Stride returns the pixmap stride (bytes per row).
func (p *Pixmap) Stride() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var stride C.int
	p.ctx.WithLock(func() {
		stride = C.gomupdf_pixmap_stride(p.ctx.ctx, p.pix)
	})
	return int(stride)
}

// N returns the number of color components.
func (p *Pixmap) N() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var n C.int
	p.ctx.WithLock(func() {
		n = C.gomupdf_pixmap_n(p.ctx.ctx, p.pix)
	})
	return int(n)
}

// Samples returns the raw pixel data.
func (p *Pixmap) Samples() []byte {
	if p.pix == nil || p.ctx == nil {
		return nil
	}

	var ptr *C.uchar
	var len C.int

	p.ctx.WithLock(func() {
		ptr = C.gomupdf_pixmap_samples(p.ctx.ctx, p.pix)
		stride := C.gomupdf_pixmap_stride(p.ctx.ctx, p.pix)
		height := C.gomupdf_pixmap_height(p.ctx.ctx, p.pix)
		len = stride * height
	})

	if ptr == nil || len == 0 {
		return nil
	}

	return C.GoBytes(unsafe.Pointer(ptr), len)
}

// SavePNG saves the pixmap as a PNG file.
func (p *Pixmap) SavePNG(filename string) error {
	if p.pix == nil || p.ctx == nil {
		return errors.New("pixmap is nil")
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	p.ctx.WithLock(func() {
		C.gomupdf_save_pixmap_as_png(p.ctx.ctx, p.pix, cfilename)
	})

	return nil
}

// SaveJPEG saves the pixmap as a JPEG file.
func (p *Pixmap) SaveJPEG(filename string, quality int) error {
	if p.pix == nil || p.ctx == nil {
		return errors.New("pixmap is nil")
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	p.ctx.WithLock(func() {
		C.gomupdf_save_pixmap_as_jpeg(p.ctx.ctx, p.pix, cfilename, C.int(quality))
	})

	return nil
}
