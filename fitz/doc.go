// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（doc.go）封装了文档级操作：打开、创建、保存、元数据、目录、交叉引用表及嵌入文件。
package fitz

import (
	"fmt"
	"io"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Document 表示一个打开的 PDF 或其他格式的文档对象。
// 封装了 MuPDF 的 Context 和 Document 指针，提供文档级操作。
type Document struct {
	ctx *cgo_bindings.Context // MuPDF 上下文，管理内存和异常处理
	doc *cgo_bindings.Document // MuPDF 文档指针
}

// Open 从文件路径打开文档，内部自动创建新的 MuPDF 上下文。
func Open(filename string) (*Document, error) {
	ctx := cgo_bindings.NewContext()
	doc, err := cgo_bindings.Open(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// OpenWithContext 使用指定的 MuPDF 上下文从文件路径打开文档。
func OpenWithContext(ctx *cgo_bindings.Context, filename string) (*Document, error) {
	doc, err := cgo_bindings.Open(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// OpenStream 从 io.Reader 流中读取数据并打开文档。
func OpenStream(r io.Reader, filetype string) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream: %w", err)
	}
	ctx := cgo_bindings.New()
	doc, err := cgo_bindings.OpenStream(ctx, data, filetype)
	if err != nil {
		return nil, fmt.Errorf("failed to open document from stream: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// OpenStreamWithContext 使用指定的 MuPDF 上下文从字节切片打开文档。
func OpenStreamWithContext(ctx *cgo_bindings.Context, data []byte, filetype string) (*Document, error) {
	doc, err := cgo_bindings.OpenStream(ctx, data, filetype)
	if err != nil {
		return nil, fmt.Errorf("failed to open document from stream: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// NewPDF 创建一个新的空白 PDF 文档。
func NewPDF() (*Document, error) {
	ctx := cgo_bindings.New()
	doc, err := cgo_bindings.NewPDF(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF: %w", err)
	}
	return &Document{ctx: ctx, doc: doc}, nil
}

// Close 关闭文档并释放相关资源。
func (d *Document) Close() error {
	if d.doc != nil {
		d.doc.Destroy()
		d.doc = nil
	}
	return nil
}

// PageCount 返回文档的总页数。
func (d *Document) PageCount() int {
	if d.doc == nil {
		return 0
	}
	return d.doc.PageCount()
}

// IsPDF 判断文档是否为 PDF 格式。
func (d *Document) IsPDF() bool {
	if d.doc == nil {
		return false
	}
	return d.doc.IsPDF()
}

// NeedsPassword 判断文档是否需要密码才能访问。
func (d *Document) NeedsPassword() bool {
	if d.doc == nil {
		return false
	}
	return d.doc.NeedsPassword()
}

// Authenticate 使用给定密码尝试认证文档，成功返回 true。
func (d *Document) Authenticate(password string) bool {
	if d.doc == nil {
		return false
	}
	return d.doc.Authenticate(password)
}

// Metadata 返回文档的元数据键值对（如标题、作者、格式等）。
func (d *Document) Metadata() map[string]string {
	if d.doc == nil {
		return nil
	}
	return d.doc.Metadata()
}

// Page 根据页码（从 0 开始）加载并返回对应的页面对象。
func (d *Document) Page(number int) (*Page, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	if number < 0 || number >= d.doc.PageCount() {
		return nil, fmt.Errorf("page number out of range: %d", number)
	}

	page, err := cgo_bindings.LoadPage(d.ctx, d.doc.Doc, number)
	if err != nil {
		return nil, fmt.Errorf("failed to load page: %w", err)
	}

	return &Page{
		ctx:   d.ctx,
		page:  page,
		doc:   d,
		index: number,
	}, nil
}

// SaveOptions 是文档保存选项的类型别名，对应 cgo_bindings.SaveOptions。
type SaveOptions = cgo_bindings.SaveOptions

// Save 将文档保存到指定文件路径。
func (d *Document) Save(filename string, opts *SaveOptions) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.SaveDocument(filename, opts)
}

// SaveToBytes 将文档保存到字节切片并返回。
func (d *Document) SaveToBytes(opts *SaveOptions) ([]byte, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	return d.doc.WriteDocument(opts)
}

// NewPage 在指定位置插入一个新的空白页面。
func (d *Document) NewPage(at int, width, height float64, rotation int) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.InsertPage(at, 0, 0, width, height, rotation)
}

// DeletePage 根据页码删除文档中的指定页面。
func (d *Document) DeletePage(number int) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.DeletePage(number)
}

// SetMetadata 设置文档的指定元数据字段。
func (d *Document) SetMetadata(key, value string) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return d.doc.SetMetadata(key, value)
}

// Permissions 返回文档的权限标志位。
func (d *Document) Permissions() int {
	if d.doc == nil {
		return 0
	}
	return d.doc.Permissions()
}

// OutlineEntry 是目录条目的类型别名，对应 cgo_bindings.OutlineEntry。
type OutlineEntry = cgo_bindings.OutlineEntry

// GetOutline 返回文档的目录（书签）列表。
func (d *Document) GetOutline() ([]OutlineEntry, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	return d.doc.GetOutline()
}

// String 返回文档的字符串描述信息。
func (d *Document) String() string {
	if d.doc == nil {
		return "Document(<closed>)"
	}
	return fmt.Sprintf("Document(%q, %d pages)", d.Metadata()["format"], d.PageCount())
}

// GetDoc 返回底层的 cgo_bindings.Document 指针，用于高级操作。
func (d *Document) GetDoc() *cgo_bindings.Document {
	return d.doc
}

// XRefLength 返回文档交叉引用表的长度。
func (d *Document) XRefLength() int {
	if d.doc == nil {
		return 0
	}
	return cgo_bindings.XRefLength(d.ctx, d.doc.Doc)
}

// XRefGetKey 获取指定交叉引用对象中某个键的值。
func (d *Document) XRefGetKey(xref int, key string) string {
	if d.doc == nil {
		return ""
	}
	return cgo_bindings.XRefGetKey(d.ctx, d.doc.Doc, xref, key)
}

// XRefIsStream 判断指定交叉引用条目是否为流对象。
func (d *Document) XRefIsStream(xref int) bool {
	if d.doc == nil {
		return false
	}
	return cgo_bindings.XRefIsStream(d.ctx, d.doc.Doc, xref)
}

// EmbeddedFileCount 返回文档中嵌入文件的数量。
func (d *Document) EmbeddedFileCount() int {
	if d.doc == nil {
		return 0
	}
	return cgo_bindings.EmbeddedFileCount(d.ctx, d.doc.Doc)
}

// EmbeddedFileName 根据索引返回嵌入文件的名称。
func (d *Document) EmbeddedFileName(idx int) string {
	if d.doc == nil {
		return ""
	}
	return cgo_bindings.EmbeddedFileName(d.ctx, d.doc.Doc, idx)
}

// EmbeddedFileGet 根据索引获取嵌入文件的数据内容。
func (d *Document) EmbeddedFileGet(idx int) ([]byte, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document is closed")
	}
	return cgo_bindings.EmbeddedFileGet(d.ctx, d.doc.Doc, idx)
}

// AddEmbeddedFile 向文档中添加一个嵌入文件。
func (d *Document) AddEmbeddedFile(filename, mimetype string, data []byte) error {
	if d.doc == nil {
		return fmt.Errorf("document is closed")
	}
	return cgo_bindings.AddEmbeddedFile(d.ctx, d.doc.Doc, filename, mimetype, data)
}
