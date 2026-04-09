// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（font.go）包含字体相关的 CGO 封装函数，涵盖：
//   - 从文件加载字体（NewFontFromFile）
//   - 从内存缓冲区加载字体（NewFontFromBuffer）
//   - 字体销毁（Destroy）
//   - 字体属性查询（Name、Ascender、Descender）
//   - 文本测量（MeasureText、GlyphAdvance）
//
// CGO 模式说明：
//   - C.fz_font 是 MuPDF 的字体对象指针
//   - Go 字符串通过 C.CString 转换为 *C.char，使用后通过 C.free(unsafe.Pointer(...)) 释放
//   - Go []byte 通过 (*C.char)(unsafe.Pointer(&data[0])) 获取首元素地址传给 C 端
//   - C.float() 用于将 Go float64 转为 C 端的 float 类型
//   - 所有 MuPDF 调用均在 ctx.WithLock 回调中执行，保证线程安全
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

// Font 表示一个已加载的字体对象，封装 MuPDF 的 fz_font 指针。
type Font struct {
	ctx  *Context      // MuPDF 上下文
	font *C.fz_font   // C 端 fz_font 指针
}

// NewFontFromFile 从文件路径加载字体。index 为字体集合中的子字体索引（通常为 0）。
// Go 文件名通过 C.CString 转换，使用后通过 defer C.free 释放。
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

// NewFontFromBuffer 从内存缓冲区加载字体。
// data 为字体文件的原始字节内容，通过 unsafe.Pointer(&data[0]) 获取底层内存地址传给 C 端。
// 注意：调用期间 data 不能被 GC 回收，C 端会在内部拷贝数据。
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

// Destroy 释放字体的 C 端资源。调用后 Font 不可再使用。
func (f *Font) Destroy() {
	if f.font != nil && f.ctx != nil {
		f.ctx.WithLock(func() {
			C.gomupdf_drop_font(f.ctx.ctx, f.font)
		})
		f.font = nil
	}
}

// Name 返回字体名称。通过 C.GoString 将 C 端返回的字符串指针转换为 Go string。
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

// Ascender 返回字体的上升线（ascender）高度。C.float 通过 float64() 转为 Go 类型。
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

// Descender 返回字体的下降线（descender）高度。
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

// MeasureText 测量指定文本在给定字号下的渲染宽度。
// text 通过 C.CString 转换为 C 字符串，size 通过 C.float() 转换为 C 浮点数。
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

// GlyphAdvance 返回指定字形（glyph）的前进宽度。
// glyph 为字形索引，size 为字号大小。
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
