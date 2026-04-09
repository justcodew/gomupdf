// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（advanced.go）包含 PDF 高级功能的 CGO 封装函数，涵盖：
//   - 显示列表（DisplayList）：缓存页面绘制命令，支持离线渲染
//   - 页面框（PageBox）：读写 CropBox、MediaBox 等页面边界
//   - 页面旋转（SetPageRotation）
//   - XRef 表操作（XRefLength、XRefGetKey、XRefIsStream）
//   - 嵌入文件（EmbeddedFileCount、EmbeddedFileName、EmbeddedFileGet、AddEmbeddedFile）
//   - 链接创建（CreateLink）
//   - 内容流写入（PageContentBegin、PageContentEnd）
//
// CGO 模式说明：
//   - C.fz_display_list 是 MuPDF 的显示列表类型，用于缓存页面绘制命令
//   - C.fz_buffer 是 MuPDF 的缓冲区类型，用于内容流写入
//   - Go []byte 通过 unsafe.Pointer(&data[0]) 获取底层数据指针传给 C 端
//   - C 端分配的输出缓冲区通过 C.free 或 C.GoBytes + defer C.free 管理
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

// DisplayList 表示页面的缓存显示列表，用于离线渲染（无需保留原始页面即可重复渲染）。
type DisplayList struct {
	ctx  *Context               // MuPDF 上下文
	list *C.fz_display_list    // C 端 fz_display_list 指针
}

// NewDisplayList 创建一个新的空白显示列表，x0/y0/x1/y1 为其边界矩形。
// 边界矩形通过 C.fz_rect 结构体传入 C 端。
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

// RunPageToDisplayList 将页面内容渲染到显示列表中（记录绘制命令而非直接光栅化）。
// a,b,c,d,e,f 为仿射变换矩阵的 6 个分量。
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

// GetPixmap 将显示列表渲染为 Pixmap。
// 使用记录的绘制命令和指定的变换矩阵进行光栅化，alpha 控制是否包含透明通道。
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

// Destroy 释放显示列表的 C 端资源。调用后 DisplayList 不可再使用。
func (dl *DisplayList) Destroy() {
	if dl.list != nil && dl.ctx != nil {
		dl.ctx.WithLock(func() {
			C.gomupdf_drop_display_list(dl.ctx.ctx, dl.list)
		})
		dl.list = nil
	}
}

// PageBox 返回指定页面的页面框（如 CropBox、MediaBox）。
// boxType 为框类型名称，目前支持 "CropBox" 和 "MediaBox"。
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

// SetPageBox 设置指定页面的页面框（如 CropBox、MediaBox）。
// 坐标值通过 C.float() 转换为 C 端浮点数。
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

// SetPageRotation 设置页面的旋转角度。
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

// XRefLength 返回文档 XRef（交叉引用）表的条目数量。
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

// XRefGetKey 返回 XRef 对象中指定键的值。
// Go 字符串 key 通过 C.CString 转换，使用后通过 defer C.free 释放。
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

// XRefIsStream 判断指定 XRef 条目是否为流对象。C.int 通过 != 0 转换为 Go bool。
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

// EmbeddedFileCount 返回文档中嵌入文件的数量。
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

// EmbeddedFileName 按索引返回嵌入文件的名称。
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

// EmbeddedFileGet 按索引获取嵌入文件的数据内容。
// C 端通过输出参数（outData 指针 + outLen 长度）返回数据，
// Go 侧通过 C.GoBytes 拷贝为 []byte，然后通过 defer C.free 释放 C 端缓冲区。
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

// AddEmbeddedFile 向文档添加一个嵌入文件。
// filename 为文件名，mimetype 为 MIME 类型（可为空），data 为文件内容。
// Go 字符串通过 C.CString 转换，Go []byte 通过 unsafe.Pointer(&data[0]) 获取底层指针。
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

// CreateLink 在页面上创建一个新的超链接。
// x0/y0/x1/y1 为链接区域矩形，uri 为链接地址，pageNum 为目标页码。
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

// PageContentBegin 开始写入页面内容流，返回 fz_buffer 指针用于后续写入。
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

// PageContentEnd 结束页面内容流的写入，将缓冲区中的内容应用到页面。
// buf 为 PageContentBegin 返回的 fz_buffer 指针。
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
