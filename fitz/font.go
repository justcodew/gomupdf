// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（font.go）封装了字体加载和文本测量功能。
package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Font 表示一个已加载的字体对象，封装了 MuPDF 的 Font 指针。
type Font struct {
	ctx  *cgo_bindings.Context  // MuPDF 上下文
	font *cgo_bindings.Font     // MuPDF 字体指针
}

// LoadFontFromFile 从文件加载字体，index 用于指定字体集合中的字体索引。
func LoadFontFromFile(filename string, index int) (*Font, error) {
	ctx := cgo_bindings.NewContext()
	font, err := cgo_bindings.NewFontFromFile(ctx, filename, index)
	if err != nil {
		ctx.Destroy()
		return nil, fmt.Errorf("failed to load font: %w", err)
	}
	return &Font{ctx: ctx, font: font}, nil
}

// LoadFontFromBuffer 从字节缓冲区加载字体。
func LoadFontFromBuffer(data []byte, index int) (*Font, error) {
	ctx := cgo_bindings.NewContext()
	font, err := cgo_bindings.NewFontFromBuffer(ctx, data, index)
	if err != nil {
		ctx.Destroy()
		return nil, fmt.Errorf("failed to load font: %w", err)
	}
	return &Font{ctx: ctx, font: font}, nil
}

// Close 释放字体及相关资源。
func (f *Font) Close() {
	if f.font != nil {
		f.font.Destroy()
		f.font = nil
	}
	if f.ctx != nil {
		f.ctx.Destroy()
		f.ctx = nil
	}
}

// Name 返回字体名称。
func (f *Font) Name() string {
	if f.font == nil {
		return ""
	}
	return f.font.Name()
}

// Ascender 返回字体的上升量（基线到字体顶部的距离）。
func (f *Font) Ascender() float64 {
	if f.font == nil {
		return 0
	}
	return f.font.Ascender()
}

// Descender 返回字体的下降量（基线到字体底部的距离）。
func (f *Font) Descender() float64 {
	if f.font == nil {
		return 0
	}
	return f.font.Descender()
}

// MeasureText 测量使用该字体渲染指定文本时的宽度。
func (f *Font) MeasureText(text string, size float64) float64 {
	if f.font == nil {
		return 0
	}
	return f.font.MeasureText(text, size)
}

// GlyphAdvance 返回指定字形的步进宽度。
func (f *Font) GlyphAdvance(glyph int, size float64) float64 {
	if f.font == nil {
		return 0
	}
	return f.font.GlyphAdvance(glyph, size)
}

// String 返回字体的字符串描述信息。
func (f *Font) String() string {
	if f.font == nil {
		return "Font(<nil>)"
	}
	return fmt.Sprintf("Font(%q)", f.Name())
}
