package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Font represents a loaded font.
type Font struct {
	ctx  *cgo_bindings.Context
	font *cgo_bindings.Font
}

// LoadFontFromFile loads a font from a file.
func LoadFontFromFile(filename string, index int) (*Font, error) {
	ctx := cgo_bindings.NewContext()
	font, err := cgo_bindings.NewFontFromFile(ctx, filename, index)
	if err != nil {
		ctx.Destroy()
		return nil, fmt.Errorf("failed to load font: %w", err)
	}
	return &Font{ctx: ctx, font: font}, nil
}

// LoadFontFromBuffer loads a font from a byte buffer.
func LoadFontFromBuffer(data []byte, index int) (*Font, error) {
	ctx := cgo_bindings.NewContext()
	font, err := cgo_bindings.NewFontFromBuffer(ctx, data, index)
	if err != nil {
		ctx.Destroy()
		return nil, fmt.Errorf("failed to load font: %w", err)
	}
	return &Font{ctx: ctx, font: font}, nil
}

// Close releases the font resources.
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

// Name returns the font name.
func (f *Font) Name() string {
	if f.font == nil {
		return ""
	}
	return f.font.Name()
}

// Ascender returns the font ascender.
func (f *Font) Ascender() float64 {
	if f.font == nil {
		return 0
	}
	return f.font.Ascender()
}

// Descender returns the font descender.
func (f *Font) Descender() float64 {
	if f.font == nil {
		return 0
	}
	return f.font.Descender()
}

// MeasureText measures the width of text rendered with this font.
func (f *Font) MeasureText(text string, size float64) float64 {
	if f.font == nil {
		return 0
	}
	return f.font.MeasureText(text, size)
}

// GlyphAdvance returns the advance width of a glyph.
func (f *Font) GlyphAdvance(glyph int, size float64) float64 {
	if f.font == nil {
		return 0
	}
	return f.font.GlyphAdvance(glyph, size)
}

// String returns a string representation.
func (f *Font) String() string {
	if f.font == nil {
		return "Font(<nil>)"
	}
	return fmt.Sprintf("Font(%q)", f.Name())
}
