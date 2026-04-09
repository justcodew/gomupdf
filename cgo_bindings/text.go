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

// TextBlockType represents the type of text block.
const (
	TextBlockTypeText   = 0
	TextBlockTypeImage  = 1
)

// TextPage represents a structured text page.
type TextPage struct {
	ctx      *Context
	page     *C.fz_stext_page
	owningPage *Page  // Reference to keep the underlying fz_page alive
}

// NewTextPage creates a structured text page from a page.
func NewTextPage(page *Page) (*TextPage, error) {
	if page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	ctx := page.Ctx
	fzPage := page.Page

	var stextPage *C.fz_stext_page

	ctx.WithLock(func() {
		stextPage = C.gomupdf_new_stext_page_from_page(ctx.ctx, fzPage)
	})

	if stextPage == nil {
		return nil, errors.New("failed to create structured text page")
	}

	tp := &TextPage{
		ctx:        ctx,
		page:       stextPage,
		owningPage: page,  // Keep the Page alive
	}

	// Note: No SetFinalizer - caller must explicitly call Destroy()
	return tp, nil
}

// Destroy releases the text page.
func (tp *TextPage) Destroy() {
	if tp.page != nil && tp.ctx != nil {
		tp.ctx.WithLock(func() {
			C.gomupdf_drop_stext_page(tp.ctx.ctx, tp.page)
		})
		tp.page = nil
	}
}

// CStextPage returns the underlying C fz_stext_page pointer (for use by other CGO functions).
func (tp *TextPage) CStextPage() *C.fz_stext_page {
	return tp.page
}

// BlockCount returns the number of blocks in the text page.
func (tp *TextPage) BlockCount() int {
	if tp.page == nil || tp.ctx == nil {
		return 0
	}
	var count C.int
	tp.ctx.WithLock(func() {
		count = C.gomupdf_stext_page_block_count(tp.ctx.ctx, tp.page)
	})
	return int(count)
}

// Text returns the plain text content of the page.
func (tp *TextPage) Text() string {
	if tp.page == nil || tp.ctx == nil {
		return ""
	}
	var text *C.char
	tp.ctx.WithLock(func() {
		text = C.gomupdf_stext_page_text(tp.ctx.ctx, tp.page)
	})
	if text == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(text))
	return C.GoString(text)
}

// BlockType returns the type of a text block (0=text, 1=image).
func (tp *TextPage) BlockType(block *C.fz_stext_block) int {
	if block == nil || tp.ctx == nil {
		return -1
	}
	var typ C.int
	tp.ctx.WithLock(func() {
		typ = C.gomupdf_stext_block_type(tp.ctx.ctx, block)
	})
	return int(typ)
}

// BlockBbox returns the bounding box of a text block.
func (tp *TextPage) BlockBbox(block *C.fz_stext_block) (x0, y0, x1, y1 float64) {
	if block == nil || tp.ctx == nil {
		return 0, 0, 0, 0
	}
	var bbox C.fz_rect
	tp.ctx.WithLock(func() {
		bbox = C.gomupdf_stext_block_bbox(tp.ctx.ctx, block)
	})
	return float64(bbox.x0), float64(bbox.y0), float64(bbox.x1), float64(bbox.y1)
}

// GetBlock returns the block at the given index.
func (tp *TextPage) GetBlock(idx int) *C.fz_stext_block {
	if tp.page == nil || tp.ctx == nil {
		return nil
	}
	var block *C.fz_stext_block
	tp.ctx.WithLock(func() {
		block = C.gomupdf_stext_page_get_block(tp.ctx.ctx, tp.page, C.int(idx))
	})
	return block
}

// LineBbox returns the bounding box of a text line.
func (tp *TextPage) LineBbox(line *C.fz_stext_line) (x0, y0, x1, y1 float64) {
	if line == nil || tp.ctx == nil {
		return 0, 0, 0, 0
	}
	var bbox C.fz_rect
	tp.ctx.WithLock(func() {
		bbox = C.gomupdf_stext_line_bbox(tp.ctx.ctx, line)
	})
	return float64(bbox.x0), float64(bbox.y0), float64(bbox.x1), float64(bbox.y1)
}

// LineDir returns the direction vector of a text line.
func (tp *TextPage) LineDir(line *C.fz_stext_line) (x, y float64) {
	if line == nil || tp.ctx == nil {
		return 0, 0
	}
	var dir C.fz_point
	tp.ctx.WithLock(func() {
		dir = C.gomupdf_stext_line_dir(tp.ctx.ctx, line)
	})
	return float64(dir.x), float64(dir.y)
}

// CharCount returns the number of characters in a line.
func (tp *TextPage) CharCount(line *C.fz_stext_line) int {
	if line == nil || tp.ctx == nil {
		return 0
	}
	var count C.int
	tp.ctx.WithLock(func() {
		count = C.gomupdf_stext_line_char_count(tp.ctx.ctx, line)
	})
	return int(count)
}

// CharOrigin returns the origin point of a character.
func (tp *TextPage) CharOrigin(ch *C.fz_stext_char) (x, y float64) {
	if ch == nil || tp.ctx == nil {
		return 0, 0
	}
	var origin C.fz_point
	tp.ctx.WithLock(func() {
		origin = C.gomupdf_stext_char_origin(tp.ctx.ctx, ch)
	})
	return float64(origin.x), float64(origin.y)
}

// CharC returns the Unicode code point of a character.
func (tp *TextPage) CharC(ch *C.fz_stext_char) int {
	if ch == nil || tp.ctx == nil {
		return 0
	}
	var c C.int
	tp.ctx.WithLock(func() {
		c = C.gomupdf_stext_char_c(tp.ctx.ctx, ch)
	})
	return int(c)
}

// CharSize returns the font size of a character.
func (tp *TextPage) CharSize(ch *C.fz_stext_char) float64 {
	if ch == nil || tp.ctx == nil {
		return 0
	}
	var size C.float
	tp.ctx.WithLock(func() {
		size = C.gomupdf_stext_char_size(tp.ctx.ctx, ch)
	})
	return float64(size)
}

// CharBbox returns the approximate bounding box of a character.
func (tp *TextPage) CharBbox(ch *C.fz_stext_char) (x0, y0, x1, y1 float64) {
	if ch == nil || tp.ctx == nil {
		return 0, 0, 0, 0
	}
	var bbox C.fz_rect
	tp.ctx.WithLock(func() {
		bbox = C.gomupdf_stext_char_bbox(tp.ctx.ctx, ch)
	})
	return float64(bbox.x0), float64(bbox.y0), float64(bbox.x1), float64(bbox.y1)
}

// LineCount returns the number of lines in a text block.
func (tp *TextPage) LineCount(block *C.fz_stext_block) int {
	if block == nil || tp.ctx == nil {
		return 0
	}
	var count C.int
	tp.ctx.WithLock(func() {
		count = C.gomupdf_stext_block_line_count(tp.ctx.ctx, block)
	})
	return int(count)
}

// GetLine returns the line at the given index in a text block.
func (tp *TextPage) GetLine(block *C.fz_stext_block, idx int) *C.fz_stext_line {
	if block == nil || tp.ctx == nil {
		return nil
	}
	var line *C.fz_stext_line
	tp.ctx.WithLock(func() {
		line = C.gomupdf_stext_block_get_line(tp.ctx.ctx, block, C.int(idx))
	})
	return line
}

// FirstLine returns the first line in a text block.
func (tp *TextPage) FirstLine(block *C.fz_stext_block) *C.fz_stext_line {
	if block == nil || tp.ctx == nil {
		return nil
	}
	var line *C.fz_stext_line
	tp.ctx.WithLock(func() {
		line = C.gomupdf_stext_block_first_line(tp.ctx.ctx, block)
	})
	return line
}

// NextLine returns the next line in the linked list.
func (tp *TextPage) NextLine(line *C.fz_stext_line) *C.fz_stext_line {
	if line == nil || tp.ctx == nil {
		return nil
	}
	var next *C.fz_stext_line
	tp.ctx.WithLock(func() {
		next = C.gomupdf_stext_line_next(tp.ctx.ctx, line)
	})
	return next
}

// FirstChar returns the first character in a line.
func (tp *TextPage) FirstChar(line *C.fz_stext_line) *C.fz_stext_char {
	if line == nil {
		return nil
	}
	return line.first_char
}

// NextChar returns the next character in the linked list.
func (tp *TextPage) NextChar(ch *C.fz_stext_char) *C.fz_stext_char {
	if ch == nil || tp.ctx == nil {
		return nil
	}
	var next *C.fz_stext_char
	tp.ctx.WithLock(func() {
		next = C.gomupdf_stext_char_next(tp.ctx.ctx, ch)
	})
	return next
}

// HTML returns the text page content as HTML.
func (tp *TextPage) HTML() string {
	if tp.page == nil || tp.ctx == nil {
		return ""
	}
	var html *C.char
	tp.ctx.WithLock(func() {
		html = C.gomupdf_stext_page_to_html(tp.ctx.ctx, tp.page)
	})
	if html == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(html))
	return C.GoString(html)
}

// XML returns the text page content as XML.
func (tp *TextPage) XML() string {
	if tp.page == nil || tp.ctx == nil {
		return ""
	}
	var xml *C.char
	tp.ctx.WithLock(func() {
		xml = C.gomupdf_stext_page_to_xml(tp.ctx.ctx, tp.page)
	})
	if xml == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(xml))
	return C.GoString(xml)
}

// XHTML returns the text page content as XHTML.
func (tp *TextPage) XHTML() string {
	if tp.page == nil || tp.ctx == nil {
		return ""
	}
	var xhtml *C.char
	tp.ctx.WithLock(func() {
		xhtml = C.gomupdf_stext_page_to_xhtml(tp.ctx.ctx, tp.page)
	})
	if xhtml == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(xhtml))
	return C.GoString(xhtml)
}

// JSON returns the text page content as JSON.
func (tp *TextPage) JSON() string {
	if tp.page == nil || tp.ctx == nil {
		return ""
	}
	var json *C.char
	tp.ctx.WithLock(func() {
		json = C.gomupdf_stext_page_to_json(tp.ctx.ctx, tp.page)
	})
	if json == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(json))
	return C.GoString(json)
}

// CharFont returns the font name of a character.
func (tp *TextPage) CharFont(ch *C.fz_stext_char) string {
	if ch == nil || tp.ctx == nil {
		return ""
	}
	var name *C.char
	tp.ctx.WithLock(func() {
		name = C.gomupdf_stext_char_font(tp.ctx.ctx, ch)
	})
	if name == nil {
		return ""
	}
	return C.GoString(name)
}

// CharFlags returns the flags of a character.
func (tp *TextPage) CharFlags(ch *C.fz_stext_char) int {
	if ch == nil || tp.ctx == nil {
		return 0
	}
	var flags C.int
	tp.ctx.WithLock(func() {
		flags = C.gomupdf_stext_char_flags(tp.ctx.ctx, ch)
	})
	return int(flags)
}
