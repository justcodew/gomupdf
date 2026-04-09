// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（page.go）封装了页面级操作：渲染、注释、链接、搜索、页面框及表单控件。
package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Page 表示文档中的一个页面，提供渲染、注释、链接等操作。
type Page struct {
	ctx   *cgo_bindings.Context  // MuPDF 上下文
	page  *cgo_bindings.Page     // MuPDF 页面指针
	doc   *Document              // 所属文档
	index int                    // 页码（从 0 开始）
}

// Rect 返回页面的边界矩形。
func (p *Page) Rect() Rect {
	if p.page == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := p.page.Rect()
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// Rotation 返回页面的旋转角度（度数）。
func (p *Page) Rotation() int {
	if p.page == nil || p.doc == nil {
		return 0
	}
	if p.doc.doc != nil {
		rot := cgo_bindings.PDFPageRotation(p.ctx, p.doc.doc.Doc, p.index)
		if rot != 0 {
			return rot
		}
	}
	return p.page.Rotation()
}

// SetRotation 设置页面的旋转角度。
func (p *Page) SetRotation(rotation int) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.SetPageRotation(p.ctx, p.doc.doc.Doc, p.index, rotation)
}

// Number 返回页面的页码（从 0 开始）。
func (p *Page) Number() int {
	return p.index
}

// Pixmap 根据给定的变换矩阵渲染页面，返回像素图对象。alpha 控制是否包含透明通道。
func (p *Page) Pixmap(matrix Matrix, alpha bool) (*Pixmap, error) {
	if p.page == nil || p.ctx == nil {
		return nil, fmt.Errorf("page is nil")
	}

	pix, err := cgo_bindings.RenderPage(p.ctx, p.page.Page,
		matrix.A, matrix.B, matrix.C, matrix.D, matrix.E, matrix.F, alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to render page: %w", err)
	}

	return &Pixmap{
		ctx:  p.ctx,
		pixmap: pix,
	}, nil
}

// Annots 返回页面上所有注释的列表。
func (p *Page) Annots() ([]*Annot, error) {
	if p.page == nil || p.doc == nil {
		return nil, fmt.Errorf("page is nil")
	}

	var annots []*Annot
	a := cgo_bindings.FirstAnnot(p.ctx, p.doc.doc.Doc, p.page.Page)
	for a != nil {
		annots = append(annots, &Annot{ctx: p.ctx, annot: a})
		a = a.Next()
	}
	return annots, nil
}

// AddAnnot 在页面上创建并添加指定类型的新注释。
func (p *Page) AddAnnot(annotType int) (*Annot, error) {
	if p.page == nil || p.doc == nil {
		return nil, fmt.Errorf("page is nil")
	}
	annot, err := cgo_bindings.CreateAnnot(p.ctx, p.doc.doc.Doc, p.page.Page, annotType)
	if err != nil {
		return nil, err
	}
	return &Annot{ctx: p.ctx, annot: annot}, nil
}

// AddHighlightAnnot 添加高亮注释，quads 指定高亮区域。
func (p *Page) AddHighlightAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotHighlight))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		if err := annot.annot.SetQuadPoints(cgoQuads); err != nil {
			annot.Delete()
			return nil, err
		}
	}
	annot.Update()
	return annot, nil
}

// AddStrikeoutAnnot 添加删除线注释。
func (p *Page) AddStrikeoutAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotStrikeOut))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		annot.annot.SetQuadPoints(cgoQuads)
	}
	annot.Update()
	return annot, nil
}

// AddUnderlineAnnot 添加下划线注释。
func (p *Page) AddUnderlineAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotUnderline))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		annot.annot.SetQuadPoints(cgoQuads)
	}
	annot.Update()
	return annot, nil
}

// AddSquigglyAnnot 添加波浪线注释。
func (p *Page) AddSquigglyAnnot(quads []Rect) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotSquiggly))
	if err != nil {
		return nil, err
	}
	if len(quads) > 0 {
		cgoQuads := make([]cgo_bindings.Rect, len(quads))
		for i, q := range quads {
			cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
		}
		annot.annot.SetQuadPoints(cgoQuads)
	}
	annot.Update()
	return annot, nil
}

// AddTextAnnot 添加文本（便签）注释。
func (p *Page) AddTextAnnot(point Point, text string) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotText))
	if err != nil {
		return nil, err
	}
	annot.annot.SetContents(text)
	annot.Update()
	return annot, nil
}

// AddFreeTextAnnot 添加自由文本注释。
func (p *Page) AddFreeTextAnnot(rect Rect, text string) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotFreeText))
	if err != nil {
		return nil, err
	}
	annot.annot.SetRect(rect.X0, rect.Y0, rect.X1, rect.Y1)
	annot.annot.SetContents(text)
	annot.Update()
	return annot, nil
}

// AddRedactAnnot 添加涂黑（修订）注释。
func (p *Page) AddRedactAnnot(rect Rect, text string) (*Annot, error) {
	annot, err := p.AddAnnot(int(AnnotRedact))
	if err != nil {
		return nil, err
	}
	annot.annot.SetRect(rect.X0, rect.Y0, rect.X1, rect.Y1)
	if text != "" {
		annot.annot.SetContents(text)
	}
	annot.Update()
	return annot, nil
}

// ApplyRedactions 应用页面上所有涂黑注释，永久移除被涂黑的内容。
func (p *Page) ApplyRedactions() error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.ApplyRedactions(p.ctx, p.doc.doc.Doc, p.page.Page)
}

// LinkInfo 是页面超链接信息的类型别名，对应 cgo_bindings.LinkInfo。
type LinkInfo = cgo_bindings.LinkInfo

// GetLinks 加载页面上的所有超链接。
func (p *Page) GetLinks() ([]LinkInfo, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	return cgo_bindings.LoadLinks(p.ctx, p.page)
}

// AddLink 在页面的指定区域添加一个指向 URI 的链接。
func (p *Page) AddLink(rect Rect, uri string) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.CreateLink(p.ctx, p.doc.doc.Doc, p.page.Page,
		rect.X0, rect.Y0, rect.X1, rect.Y1, uri, -1)
}

// SearchFor 在页面中搜索指定文本，返回匹配区域的边界矩形列表。
func (p *Page) SearchFor(text string, maxHits int) ([]Rect, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	if maxHits <= 0 {
		maxHits = 100
	}

	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, err
	}
	defer stextPage.Destroy()

	cgoRects := cgo_bindings.SearchText(p.ctx, stextPage.CStextPage(), text, maxHits)

	rects := make([]Rect, len(cgoRects))
	for i, r := range cgoRects {
		rects[i] = Rect{X0: r.X0, Y0: r.Y0, X1: r.X1, Y1: r.Y1}
	}
	return rects, nil
}

// CropBox 返回页面的裁剪框。
func (p *Page) CropBox() Rect {
	if p.page == nil || p.doc == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := cgo_bindings.PageBox(p.ctx, p.doc.doc.Doc, p.index, "CropBox")
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// SetCropBox 设置页面的裁剪框。
func (p *Page) SetCropBox(r Rect) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.SetPageBox(p.ctx, p.doc.doc.Doc, p.index, "CropBox", r.X0, r.Y0, r.X1, r.Y1)
}

// MediaBox 返回页面的媒体框。
func (p *Page) MediaBox() Rect {
	if p.page == nil || p.doc == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := cgo_bindings.PageBox(p.ctx, p.doc.doc.Doc, p.index, "MediaBox")
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// SetMediaBox 设置页面的媒体框。
func (p *Page) SetMediaBox(r Rect) error {
	if p.page == nil || p.doc == nil {
		return fmt.Errorf("page is nil")
	}
	return cgo_bindings.SetPageBox(p.ctx, p.doc.doc.Doc, p.index, "MediaBox", r.X0, r.Y0, r.X1, r.Y1)
}

// Widgets 返回页面上所有表单控件的列表。
func (p *Page) Widgets() ([]*Widget, error) {
	if p.page == nil || p.doc == nil {
		return nil, fmt.Errorf("page is nil")
	}
	var widgets []*Widget
	w := cgo_bindings.FirstWidget(p.ctx, p.doc.doc.Doc, p.page.Page)
	for w != nil {
		widgets = append(widgets, &Widget{ctx: p.ctx, widget: w})
		w = w.Next()
	}
	return widgets, nil
}

// String 返回页面的字符串描述信息。
func (p *Page) String() string {
	if p.page == nil {
		return "Page(<closed>)"
	}
	return fmt.Sprintf("Page(%v)", p.Rect())
}
