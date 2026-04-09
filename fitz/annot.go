// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（annot.go）封装了 PDF 注释的类型常量及其增删改查操作。
package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Annot 表示一个 PDF 注释对象，封装了 MuPDF 的 Annot 指针。
type Annot struct {
	ctx   *cgo_bindings.Context  // MuPDF 上下文
	annot *cgo_bindings.Annot    // MuPDF 注释指针
}

// AnnotType 表示 PDF 注释的类型。
type AnnotType int

const (
	AnnotText           AnnotType = 0  // 文本（便签）注释
	AnnotLink           AnnotType = 1  // 链接注释
	AnnotFreeText       AnnotType = 2  // 自由文本注释
	AnnotLine           AnnotType = 3  // 直线注释
	AnnotSquare         AnnotType = 4  // 矩形注释
	AnnotCircle         AnnotType = 5  // 圆形注释
	AnnotPolygon        AnnotType = 6  // 多边形注释
	AnnotPolyLine       AnnotType = 7  // 折线注释
	AnnotHighlight      AnnotType = 8  // 高亮注释
	AnnotUnderline      AnnotType = 9  // 下划线注释
	AnnotSquiggly       AnnotType = 10 // 波浪线注释
	AnnotStrikeOut      AnnotType = 11 // 删除线注释
	AnnotRedact         AnnotType = 12 // 涂黑（修订）注释
	AnnotStamp          AnnotType = 13 // 图章注释
	AnnotCaret          AnnotType = 14 // 插入符注释
	AnnotInk            AnnotType = 15 // 墨迹注释
	AnnotPopup          AnnotType = 16 // 弹出窗口注释
	AnnotFileAttachment AnnotType = 17 // 文件附件注释
	AnnotSound          AnnotType = 18 // 声音注释
	AnnotMovie          AnnotType = 19 // 影片注释
	AnnotWidget         AnnotType = 20 // 表单控件注释
	AnnotScreen         AnnotType = 21 // 屏幕注释
	AnnotPrinterMark    AnnotType = 22 // 打印标记注释
	AnnotTrapNet        AnnotType = 23 // 陷印网络注释
	AnnotWatermark      AnnotType = 24 // 水印注释
	Annot3D             AnnotType = 25 // 3D 注释
	AnnotProjection     AnnotType = 26 // 投影注释
)

// Type 返回注释的类型。
func (a *Annot) Type() AnnotType {
	if a.annot == nil {
		return -1
	}
	return AnnotType(a.annot.Type())
}

// Rect 返回注释的边界矩形。
func (a *Annot) Rect() Rect {
	if a.annot == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := a.annot.Rect()
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// SetRect 设置注释的边界矩形。
func (a *Annot) SetRect(r Rect) error {
	return a.annot.SetRect(r.X0, r.Y0, r.X1, r.Y1)
}

// Contents 返回注释的文本内容。
func (a *Annot) Contents() string {
	if a.annot == nil {
		return ""
	}
	return a.annot.Contents()
}

// SetContents 设置注释的文本内容。
func (a *Annot) SetContents(text string) error {
	return a.annot.SetContents(text)
}

// Color 返回注释的颜色（RGBA 格式）。
func (a *Annot) Color() Color {
	if a.annot == nil {
		return Color{}
	}
	r, g, b, alpha, _ := a.annot.Color()
	return Color{R: r, G: g, B: b, A: alpha}
}

// SetColor 设置注释的颜色。
func (a *Annot) SetColor(c Color) error {
	return a.annot.SetColor(c.R, c.G, c.B, c.A)
}

// Opacity 返回注释的不透明度（0.0~1.0）。
func (a *Annot) Opacity() float64 {
	if a.annot == nil {
		return 1.0
	}
	return a.annot.Opacity()
}

// SetOpacity 设置注释的不透明度。
func (a *Annot) SetOpacity(opacity float64) error {
	return a.annot.SetOpacity(opacity)
}

// Flags 返回注释的标志位。
func (a *Annot) Flags() int {
	if a.annot == nil {
		return 0
	}
	return a.annot.Flags()
}

// SetFlags 设置注释的标志位。
func (a *Annot) SetFlags(flags int) error {
	return a.annot.SetFlags(flags)
}

// Border 返回注释的边框宽度。
func (a *Annot) Border() float64 {
	if a.annot == nil {
		return 0
	}
	return a.annot.Border()
}

// SetBorder 设置注释的边框宽度。
func (a *Annot) SetBorder(width float64) error {
	return a.annot.SetBorder(width)
}

// Title 返回注释的标题。
func (a *Annot) Title() string {
	if a.annot == nil {
		return ""
	}
	return a.annot.Title()
}

// SetTitle 设置注释的标题。
func (a *Annot) SetTitle(title string) error {
	return a.annot.SetTitle(title)
}

// Update 将注释的修改保存到文档中。
func (a *Annot) Update() error {
	return a.annot.Update()
}

// Delete 从页面上删除该注释。
func (a *Annot) Delete() error {
	return a.annot.Delete()
}

// QuadPoints 返回注释的四边形点列表，用于标记注释覆盖的文本区域。
func (a *Annot) QuadPoints() []Rect {
	if a.annot == nil {
		return nil
	}
	cgoRects := a.annot.QuadPoints()
	rects := make([]Rect, len(cgoRects))
	for i, r := range cgoRects {
		rects[i] = Rect{X0: r.X0, Y0: r.Y0, X1: r.X1, Y1: r.Y1}
	}
	return rects
}

// SetQuadPoints 设置注释的四边形点列表。
func (a *Annot) SetQuadPoints(quads []Rect) error {
	cgoQuads := make([]cgo_bindings.Rect, len(quads))
	for i, q := range quads {
		cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
	}
	return a.annot.SetQuadPoints(cgoQuads)
}

// String 返回注释的字符串描述信息。
func (a *Annot) String() string {
	if a.annot == nil {
		return "Annot(<nil>)"
	}
	return fmt.Sprintf("Annot(type=%d, rect=%v)", a.Type(), a.Rect())
}
