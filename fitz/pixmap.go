// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（pixmap.go）封装了像素图操作：导出 PNG/JPEG、像素读写、颜色变换等。
package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Pixmap 表示一个光栅图像，封装了 MuPDF 的 Pixmap 指针。
type Pixmap struct {
	ctx    *cgo_bindings.Context   // MuPDF 上下文
	pixmap *cgo_bindings.Pixmap    // MuPDF 像素图指针
}

// Width 返回像素图的宽度（像素数）。
func (p *Pixmap) Width() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Width()
}

// Height 返回像素图的高度（像素数）。
func (p *Pixmap) Height() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Height()
}

// Stride 返回像素图的跨距（每行字节数）。
func (p *Pixmap) Stride() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Stride()
}

// N 返回颜色分量数（例如 RGBA 为 4）。
func (p *Pixmap) N() int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.N()
}

// Samples 返回原始像素数据字节切片。
func (p *Pixmap) Samples() []byte {
	if p.pixmap == nil {
		return nil
	}
	return p.pixmap.Samples()
}

// Save 将像素图保存为图像文件，根据文件扩展名自动选择格式（.png/.jpg/.jpeg）。
func (p *Pixmap) Save(filename string) error {
	if p.pixmap == nil {
		return fmt.Errorf("pixmap is nil")
	}

	switch ext := getExt(filename); ext {
	case ".png":
		return p.pixmap.SavePNG(filename)
	case ".jpg", ".jpeg":
		return p.pixmap.SaveJPEG(filename, 90)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
}

// Close 释放像素图资源。
func (p *Pixmap) Close() {
	if p.pixmap != nil {
		p.pixmap.Destroy()
		p.pixmap = nil
	}
}

// PNG 将像素图编码为 PNG 格式的字节切片。
func (p *Pixmap) PNG() ([]byte, error) {
	if p.pixmap == nil {
		return nil, fmt.Errorf("pixmap is nil")
	}
	return p.pixmap.PNGBytes()
}

// JPEG 将像素图编码为 JPEG 格式的字节切片，quality 指定压缩质量。
func (p *Pixmap) JPEG(quality int) ([]byte, error) {
	if p.pixmap == nil {
		return nil, fmt.Errorf("pixmap is nil")
	}
	return p.pixmap.JPEGBytes(quality)
}

// Pixel 返回指定坐标 (x, y) 处的像素值。
func (p *Pixmap) Pixel(x, y int) int {
	if p.pixmap == nil {
		return 0
	}
	return p.pixmap.Pixel(x, y)
}

// SetPixel 设置指定坐标 (x, y) 处的像素值。
func (p *Pixmap) SetPixel(x, y, val int) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.SetPixel(x, y, val)
}

// ClearWith 使用给定值清空整个像素图。
func (p *Pixmap) ClearWith(value int) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.ClearWith(value)
}

// Invert 反转像素图的所有颜色。
func (p *Pixmap) Invert() {
	if p.pixmap == nil {
		return
	}
	p.pixmap.Invert()
}

// Gamma 对像素图应用伽马校正。
func (p *Pixmap) Gamma(gamma float64) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.Gamma(gamma)
}

// Tint 对像素图应用着色处理，black 和 white 分别指定暗色和亮色值。
func (p *Pixmap) Tint(black, white int) {
	if p.pixmap == nil {
		return
	}
	p.pixmap.Tint(black, white)
}

// String 返回像素图的字符串描述信息。
func (p *Pixmap) String() string {
	if p.pixmap == nil {
		return "Pixmap(<nil>)"
	}
	return fmt.Sprintf("Pixmap(%dx%d, %d channels)", p.Width(), p.Height(), p.N())
}

// getExt 从文件名中提取扩展名（含点号），用于判断图像保存格式。
func getExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
		if filename[i] == '/' || filename[i] == '\\' {
			break
		}
	}
	return ""
}
