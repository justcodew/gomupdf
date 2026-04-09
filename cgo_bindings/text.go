// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（text.go）包含文本提取与结构化文本页面操作函数，涵盖：
//   - 结构化文本页面的创建与销毁（NewTextPage、Destroy）
//   - 文本块的访问（BlockCount、GetBlock、BlockType、BlockBbox）
//   - 文本行的遍历（LineCount、GetLine、FirstLine、NextLine、LineBbox、LineDir）
//   - 字符级别的操作（CharCount、FirstChar、NextChar、CharOrigin、CharC、CharSize、CharBbox）
//   - 多格式文本输出（纯文本 Text、HTML、XML、XHTML、JSON）
//   - 字符字体与标志（CharFont、CharFlags）
//
// CGO 模式说明：
//   - C.fz_stext_page / C.fz_stext_block / C.fz_stext_line / C.fz_stext_char
//     分别对应 MuPDF 的结构化文本页面、块、行、字符类型
//   - 所有 MuPDF 调用均在 ctx.WithLock 回调中执行，保证线程安全
//   - C 端返回的字符串指针通过 C.GoString 转换为 Go string（自动内存拷贝）
//   - C 端分配的字符串通过 defer C.free(unsafe.Pointer(...)) 释放
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

// TextBlockType 常量定义文本块的类型。
const (
	TextBlockTypeText   = 0 // 文本块
	TextBlockTypeImage  = 1 // 图像块
)

// TextPage 表示一个结构化文本页面，包含从 PDF 页面提取的文本结构（块 -> 行 -> 字符）。
// 持有 owningPage 引用以防止底层 fz_page 被过早回收。
type TextPage struct {
	ctx      *Context            // MuPDF 上下文
	page     *C.fz_stext_page   // C 端结构化文本页面指针
	owningPage *Page             // 持有的 Page 引用，保持底层 fz_page 存活
}

// NewTextPage 从 Page 创建结构化文本页面，用于后续的文本提取和分析。
// 调用者在使用完毕后必须显式调用 Destroy() 释放资源（不使用 SetFinalizer 以避免关闭时崩溃）。
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
		owningPage: page,  // 持有 Page 引用以防止底层 fz_page 被过早回收
	}

	// 注意：不使用 SetFinalizer，调用者必须显式调用 Destroy()
	return tp, nil
}

// Destroy 释放结构化文本页面的 C 端资源。调用后 TextPage 不可再使用。
func (tp *TextPage) Destroy() {
	if tp.page != nil && tp.ctx != nil {
		tp.ctx.WithLock(func() {
			C.gomupdf_drop_stext_page(tp.ctx.ctx, tp.page)
		})
		tp.page = nil
	}
}

// CStextPage 返回底层的 C.fz_stext_page 指针，供其他 CGO 函数直接使用。
func (tp *TextPage) CStextPage() *C.fz_stext_page {
	return tp.page
}

// BlockCount 返回文本页面中的块数量。
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

// Text 返回文本页面的纯文本内容。
// C 端返回的字符串通过 C.GoString 拷贝，并通过 defer C.free 释放 C 端内存。
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

// BlockType 返回文本块的类型（0=文本, 1=图像）。
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

// BlockBbox 返回文本块的包围矩形（x0, y0, x1, y1）。
// C.fz_rect 的字段通过 float64() 类型转换映射为 Go 的 float64。
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

// GetBlock 按索引返回文本页面中的文本块指针。
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

// LineBbox 返回文本行的包围矩形。
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

// LineDir 返回文本行的书写方向向量。
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

// CharCount 返回文本行中的字符数量。
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

// CharOrigin 返回字符的原点坐标（基线位置）。
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

// CharC 返回字符的 Unicode 码点值。
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

// CharSize 返回字符的字号大小。
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

// CharBbox 返回字符的近似包围矩形。
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

// LineCount 返回文本块中的行数。
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

// GetLine 按索引返回文本块中的行指针。
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

// FirstLine 返回文本块中的第一行（链表头）。
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

// NextLine 返回链表中的下一行。
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

// FirstChar 返回文本行中的第一个字符。直接访问 C 结构体字段 line.first_char。
func (tp *TextPage) FirstChar(line *C.fz_stext_line) *C.fz_stext_char {
	if line == nil {
		return nil
	}
	return line.first_char
}

// NextChar 返回链表中的下一个字符。
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

// HTML 将文本页面内容导出为 HTML 格式字符串。
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

// XML 将文本页面内容导出为 XML 格式字符串。
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

// XHTML 将文本页面内容导出为 XHTML 格式字符串。
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

// JSON 将文本页面内容导出为 JSON 格式字符串。
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

// CharFont 返回字符所使用的字体名称。
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

// CharFlags 返回字符的标志位（用于判断上标、下标、合成等属性）。
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
