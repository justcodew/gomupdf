// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（doc_ops.go）包含 PDF 文档级别的操作函数，涵盖：
//   - 文档保存（SaveDocument、WriteDocument）
//   - 页面管理（InsertPage、DeletePage、DeletePageRange）
//   - 元数据读写（SetMetadata、Permissions）
//   - 大纲/书签（OutlineEntry、GetOutline）
//   - 超链接（LinkInfo、LoadLinks）
//   - 文本搜索（SearchText）
//
// CGO 调用模式说明：
//   - Go 字符串通过 C.CString 转换为 *C.char，使用后必须通过 C.free 释放
//   - C 分配的内存通过 C.gomupdf_free / C.free 释放
//   - C.GoBytes 将 C 内存中的字节数组拷贝为 Go []byte
//   - 所有 MuPDF 调用均在 ctx.WithLock 回调中执行，以保证线程安全
//   - 返回值通过 C.int 传递错误码，非零表示失败，可通过 GetLastError 获取详情
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

// SaveOptions 包含 PDF 保存时的全部选项。
// 对应 MuPDF 的 pdf_write_options 结构体，控制压缩、清理、增量写入等行为。
type SaveOptions struct {
	Garbage          int  // 垃圾回收级别：0=关闭, 1=GC, 2=重编号, 3=去重
	Clean            bool // 是否清理内容流
	Compress         int  // 压缩方式：0=关闭, 1=zlib, 2=brotli
	CompressImages   bool // 是否压缩图像流
	CompressFonts    bool // 是否压缩字体流
	Decompress       bool // 是否解压所有流
	Linear           bool // 是否线性化（适用于 Web 展示）
	ASCII            bool // 是否使用 ASCII hex 编码
	Incremental      bool // 是否增量写入（仅写入变更的对象）
	Pretty           bool // 是否美化打印字典
	Sanitize         bool // 是否清理内容流
	Appearance       bool // 是否重新生成外观流
	PreserveMetadata bool // 是否保持元数据不变
}

// SaveDocument 将 PDF 文档保存到指定路径的文件。
// opts 为保存选项，传入 nil 则使用默认选项。
// 内部通过 C.CString 将 Go 字符串转为 C 字符串，调用完毕后通过 defer C.free 释放。
func (d *Document) SaveDocument(filename string, opts *SaveOptions) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var cOpts SaveOptions
	if opts != nil {
		cOpts = *opts
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_save_document(
			d.Ctx.ctx, d.Doc, cfilename,
			C.int(cOpts.Garbage), boolToInt(cOpts.Clean), C.int(cOpts.Compress),
			boolToInt(cOpts.CompressImages), boolToInt(cOpts.CompressFonts),
			boolToInt(cOpts.Decompress), boolToInt(cOpts.Linear), boolToInt(cOpts.ASCII),
			boolToInt(cOpts.Incremental), boolToInt(cOpts.Pretty),
			boolToInt(cOpts.Sanitize), boolToInt(cOpts.Appearance), boolToInt(cOpts.PreserveMetadata),
		)
	})

	if rc != 0 {
		return fmt.Errorf("failed to save document: %s", GetLastError())
	}
	return nil
}

// WriteDocument 将 PDF 文档序列化为字节切片（[]byte）。
// 内部调用 C 端函数分配输出缓冲区，通过 C.GoBytes 将 C 内存拷贝为 Go 字节切片，
// 最后通过 defer C.gomupdf_free 释放 C 端分配的缓冲区。
func (d *Document) WriteDocument(opts *SaveOptions) ([]byte, error) {
	if d.Doc == nil || d.Ctx == nil {
		return nil, errors.New("document is nil")
	}

	var cOpts SaveOptions
	if opts != nil {
		cOpts = *opts
	}

	var outData *C.uchar
	var outLen C.size_t

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_write_document(
			d.Ctx.ctx, d.Doc,
			&outData, &outLen,
			C.int(cOpts.Garbage), boolToInt(cOpts.Clean), C.int(cOpts.Compress),
			boolToInt(cOpts.CompressImages), boolToInt(cOpts.CompressFonts),
			boolToInt(cOpts.Decompress), boolToInt(cOpts.Linear), boolToInt(cOpts.ASCII),
			boolToInt(cOpts.Incremental), boolToInt(cOpts.Pretty),
			boolToInt(cOpts.Sanitize), boolToInt(cOpts.Appearance), boolToInt(cOpts.PreserveMetadata),
		)
	})

	if rc != 0 || outData == nil {
		return nil, fmt.Errorf("failed to write document: %s", GetLastError())
	}
	defer C.gomupdf_free(unsafe.Pointer(outData))

	return C.GoBytes(unsafe.Pointer(outData), C.int(outLen)), nil
}

// InsertPage 在指定位置创建并插入一个新的空白页面。
// at 为插入位置（页码索引），x0/y0/x1/y1 为页面边界框坐标，rotation 为旋转角度。
func (d *Document) InsertPage(at int, x0, y0, x1, y1 float64, rotation int) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_insert_page(
			d.Ctx.ctx, d.Doc, C.int(at),
			C.float(x0), C.float(y0), C.float(x1), C.float(y1), C.int(rotation),
		)
	})

	if rc != 0 {
		return fmt.Errorf("failed to insert page: %s", GetLastError())
	}
	return nil
}

// DeletePage 按页码删除指定页面。
func (d *Document) DeletePage(number int) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_delete_page(d.Ctx.ctx, d.Doc, C.int(number))
	})

	if rc != 0 {
		return fmt.Errorf("failed to delete page: %s", GetLastError())
	}
	return nil
}

// DeletePageRange 删除从 start 到 end（不含 end）的连续页面范围。
func (d *Document) DeletePageRange(start, end int) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_delete_page_range(d.Ctx.ctx, d.Doc, C.int(start), C.int(end))
	})

	if rc != 0 {
		return fmt.Errorf("failed to delete page range: %s", GetLastError())
	}
	return nil
}

// SetMetadata 设置文档的元数据键值对（如 title、author、subject 等）。
// Go 字符串通过 C.CString 转换为 C 字符串，使用后通过 defer C.free(unsafe.Pointer(...)) 释放。
func (d *Document) SetMetadata(key, value string) error {
	if d.Doc == nil || d.Ctx == nil {
		return errors.New("document is nil")
	}

	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	var rc C.int
	d.Ctx.WithLock(func() {
		rc = C.gomupdf_pdf_set_metadata(d.Ctx.ctx, d.Doc, ckey, cvalue)
	})

	if rc != 0 {
		return fmt.Errorf("failed to set metadata: %s", GetLastError())
	}
	return nil
}

// Permissions 返回文档的权限标志位（打印、复制、修改等权限的组合）。
func (d *Document) Permissions() int {
	if d.Doc == nil || d.Ctx == nil {
		return 0
	}
	var perm C.int
	d.Ctx.WithLock(func() {
		perm = C.gomupdf_pdf_permissions(d.Ctx.ctx, d.Doc)
	})
	return int(perm)
}

// OutlineEntry 表示文档书签/大纲中的一个条目。
type OutlineEntry struct {
	Title  string // 条目标题
	Page   int    // 目标页码（-1 表示外部链接）
	Level  int    // 层级深度（0 为顶层）
	URI    string // 关联的 URI（可为空）
	IsOpen bool   // 在大纲树中是否展开
}

// GetOutline 加载并返回文档的书签/目录大纲。
// 先通过 C 端获取条目数量，再逐条读取 title、page、level、uri、isOpen 等字段。
// C.GoString 将 C 字符串转换为 Go 字符串（自动拷贝内存）。
func (d *Document) GetOutline() ([]OutlineEntry, error) {
	if d.Doc == nil || d.Ctx == nil {
		return nil, errors.New("document is nil")
	}

	var count C.int
	d.Ctx.WithLock(func() {
		count = C.gomupdf_pdf_outline_count(d.Ctx.ctx, d.Doc)
	})

	if count <= 0 {
		return nil, nil
	}

	entries := make([]OutlineEntry, count)
	for i := 0; i < int(count); i++ {
		var title, uri *C.char
		var page, level, isOpen C.int

		d.Ctx.WithLock(func() {
			C.gomupdf_pdf_outline_get(d.Ctx.ctx, d.Doc, C.int(i),
				&title, &page, &level, &uri, &isOpen)
		})

		entries[i] = OutlineEntry{
			Title:  C.GoString(title),
			Page:   int(page),
			Level:  int(level),
			URI:    C.GoString(uri),
			IsOpen: isOpen != 0,
		}
	}

	return entries, nil
}

// LinkInfo 表示页面上的一个超链接信息。
type LinkInfo struct {
	Rect    Rect      // 链接的矩形区域（PDF 坐标系）
	URI     string    // 链接 URI
	Page    int      // 解析后的目标页码（-1 表示外部链接）
	Next    *LinkInfo // 指向下一个链接的指针（链表结构）
}

// Rect 表示 PDF 坐标系中的一个矩形，由左上角 (X0, Y0) 和右下角 (X1, Y1) 定义。
type Rect struct {
	X0, Y0, X1, Y1 float64
}

// LoadLinks 加载页面上的所有超链接。
// 返回的 C.fz_link 链表通过 C.gomupdf_link_next 遍历，遍历完毕后通过 C.gomupdf_drop_link 释放。
// C.fz_rect 中的字段直接通过 float64() 类型转换映射为 Go 的 float64。
func LoadLinks(ctx *Context, page *Page) ([]LinkInfo, error) {
	if ctx == nil || page == nil {
		return nil, errors.New("nil context or page")
	}

	var clink *C.fz_link
	ctx.WithLock(func() {
		clink = C.gomupdf_page_load_links(ctx.ctx, page.Page)
	})
	if clink == nil {
		return nil, nil
	}
	defer C.gomupdf_drop_link(ctx.ctx, clink)

	var links []LinkInfo
	for l := clink; l != nil; l = C.gomupdf_link_next(l) {
		r := C.gomupdf_link_rect(l)
		uri := C.gomupdf_link_uri(l)

		link := LinkInfo{
			Rect: Rect{X0: float64(r.x0), Y0: float64(r.y0), X1: float64(r.x1), Y1: float64(r.y1)},
			URI:  C.GoString(uri),
			Page: -1,
		}
		links = append(links, link)
	}

	return links, nil
}

// SearchText 在结构化文本页面中搜索指定文本，返回所有匹配区域的矩形列表。
// needle 为搜索关键词，maxHits 为最大命中数。
// 搜索结果为 fz_quad 四边形，通过 C.gomupdf_quad_rect 转换为包围矩形 fz_rect。
// Go 侧预分配 []C.fz_quad 切片，并通过 &hits[0] 将其首地址传给 C 端填充结果。
func SearchText(ctx *Context, stextPage *C.fz_stext_page, needle string, maxHits int) []Rect {
	if ctx == nil || stextPage == nil || needle == "" || maxHits <= 0 {
		return nil
	}

	cneedle := C.CString(needle)
	defer C.free(unsafe.Pointer(cneedle))

	hits := make([]C.fz_quad, maxHits)

	var count C.int
	ctx.WithLock(func() {
		count = C.gomupdf_search_text(ctx.ctx, stextPage, cneedle, C.int(maxHits), &hits[0])
	})

	if count <= 0 {
		return nil
	}

	rects := make([]Rect, count)
	for i := 0; i < int(count); i++ {
		r := C.gomupdf_quad_rect(hits[i])
		rects[i] = Rect{X0: float64(r.x0), Y0: float64(r.y0), X1: float64(r.x1), Y1: float64(r.y1)}
	}
	return rects
}

// boolToInt 将 Go 的 bool 值转换为 C.int（true -> 1, false -> 0），
// 用于传递布尔选项给 C 端函数。
func boolToInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
