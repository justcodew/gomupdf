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

// Font represents a loaded font.
type Font struct {
	ctx  *Context
	font *C.fz_font
}

// NewFontFromFile loads a font from a file.
func NewFontFromFile(ctx *Context, filename string, index int) (*Font, error) {
	if ctx == nil || filename == "" {
		return nil, errors.New("nil context or empty filename")
	}
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	var font *C.fz_font
	ctx.WithLock(func() {
		font = C.gomupdf_new_font_from_file(ctx.ctx, cfilename, C.int(index))
	})
	if font == nil {
		return nil, fmt.Errorf("failed to load font: %s", GetLastError())
	}
	return &Font{ctx: ctx, font: font}, nil
}

// NewFontFromBuffer loads a font from a byte buffer.
func NewFontFromBuffer(ctx *Context, data []byte, index int) (*Font, error) {
	if ctx == nil || len(data) == 0 {
		return nil, errors.New("nil context or empty data")
	}
	var font *C.fz_font
	ctx.WithLock(func() {
		font = C.gomupdf_new_font_from_buffer(ctx.ctx, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(len(data)), C.int(index))
	})
	if font == nil {
		return nil, fmt.Errorf("failed to load font from buffer: %s", GetLastError())
	}
	return &Font{ctx: ctx, font: font}, nil
}

// Destroy releases the font.
func (f *Font) Destroy() {
	if f.font != nil && f.ctx != nil {
		f.ctx.WithLock(func() {
			C.gomupdf_drop_font(f.ctx.ctx, f.font)
		})
		f.font = nil
	}
}

// Name returns the font name.
func (f *Font) Name() string {
	if f.font == nil || f.ctx == nil {
		return ""
	}
	var name *C.char
	f.ctx.WithLock(func() {
		name = C.gomupdf_font_name(f.ctx.ctx, f.font)
	})
	return C.GoString(name)
}

// Ascender returns the font ascender.
func (f *Font) Ascender() float64 {
	if f.font == nil || f.ctx == nil {
		return 0
	}
	var a C.float
	f.ctx.WithLock(func() {
		a = C.gomupdf_font_ascender(f.ctx.ctx, f.font)
	})
	return float64(a)
}

// Descender returns the font descender.
func (f *Font) Descender() float64 {
	if f.font == nil || f.ctx == nil {
		return 0
	}
	var d C.float
	f.ctx.WithLock(func() {
		d = C.gomupdf_font_descender(f.ctx.ctx, f.font)
	})
	return float64(d)
}

// MeasureText measures the width of text rendered with this font.
func (f *Font) MeasureText(text string, size float64) float64 {
	if f.font == nil || f.ctx == nil || text == "" {
		return 0
	}
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	var w C.float
	f.ctx.WithLock(func() {
		w = C.gomupdf_measure_text(f.ctx.ctx, f.font, ctext, C.float(size))
	})
	return float64(w)
}

// GlyphAdvance returns the advance width of a glyph.
func (f *Font) GlyphAdvance(glyph int, size float64) float64 {
	if f.font == nil || f.ctx == nil {
		return 0
	}
	var w C.float
	f.ctx.WithLock(func() {
		w = C.gomupdf_font_glyph_advance(f.ctx.ctx, f.font, C.int(glyph), C.float(size))
	})
	return float64(w)
}
