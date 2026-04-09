// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（pixmap.go）包含 Pixmap（光栅图像）相关的操作函数，涵盖：
//   - Pixmap 创建（RenderPage、NewPixmap、NewPixmapFromImage）
//   - Pixmap 销毁（Destroy）
//   - Pixmap 属性查询（Width、Height、Stride、N）
//   - 像素数据访问（Samples、Pixel、SetPixel）
//   - 文件保存（SavePNG、SaveJPEG）
//   - 内存编码输出（PNGBytes、JPEGBytes）
//   - 像素操作（ClearWith、Invert、Gamma、Tint）
//
// CGO 模式说明：
//   - C.fz_pixmap 是 MuPDF 的光栅图像类型
//   - 像素数据通过 C.GoBytes(unsafe.Pointer(ptr), len) 从 C 内存拷贝到 Go []byte
//   - C 端分配的输出缓冲区通过 C.free(unsafe.Pointer(...)) 释放
//   - 不使用 SetFinalizer，调用者必须显式调用 Destroy() 释放资源
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

// Pixmap 表示一个光栅图像（像素图），封装 MuPDF 的 fz_pixmap 指针。
type Pixmap struct {
	ctx *Context        // MuPDF 上下文
	pix *C.fz_pixmap   // C 端 fz_pixmap 指针
}

// RenderPage 将 PDF 页面渲染为 Pixmap。
// a,b,c,d,e,f 为仿射变换矩阵的 6 个分量（控制缩放、旋转、平移等）。
// alpha 控制是否生成带透明通道的图像。
func RenderPage(ctx *Context, page *C.fz_page, a, b, c, d, e, f float64, alpha bool) (*Pixmap, error) {
	var pix *C.fz_pixmap
	var alphaInt C.int
	if alpha {
		alphaInt = 1
	}

	ctx.WithLock(func() {
		pix = C.gomupdf_render_page(ctx.ctx, page, C.float(a), C.float(b), C.float(c), C.float(d), C.float(e), C.float(f), alphaInt)
	})

	if pix == nil {
		return nil, errors.New("failed to render page")
	}

	p := &Pixmap{
		ctx: ctx,
		pix: pix,
	}

	// 注意：不使用 SetFinalizer，调用者必须显式调用 Destroy()
	// 这可以避免关闭时上下文可能已损坏导致的崩溃
	return p, nil
}

// NewPixmap 创建指定宽高的空白 RGB Pixmap。
// 内部先通过 C 端获取 RGB 颜色空间，再创建 Pixmap（颜色空间引用由 Pixmap 持有，无需单独释放）。
func NewPixmap(ctx *Context, width, height int) (*Pixmap, error) {
	var pix *C.fz_pixmap

	ctx.WithLock(func() {
		cs := C.gomupdf_new_colorspace_rgb(ctx.ctx)
		pix = C.gomupdf_new_pixmap(ctx.ctx, cs, C.int(width), C.int(height))
		// Colorspace reference is kept by pixmap, don't free here
	})

	if pix == nil {
		return nil, errors.New("failed to create pixmap")
	}

	p := &Pixmap{
		ctx: ctx,
		pix: pix,
	}

	// 注意：不使用 SetFinalizer，调用者必须显式调用 Destroy()
	return p, nil
}

// NewPixmapFromImage 从 fz_image 指针创建 Pixmap（将图像解码为像素数据）。
func NewPixmapFromImage(ctx *Context, img *C.fz_image) (*Pixmap, error) {
	if ctx == nil || img == nil {
		return nil, errors.New("nil context or image")
	}
	var pix *C.fz_pixmap
	ctx.WithLock(func() {
		pix = C.gomupdf_image_get_pixmap(ctx.ctx, img)
	})
	if pix == nil {
		return nil, errors.New("failed to get pixmap from image")
	}
	return &Pixmap{ctx: ctx, pix: pix}, nil
}

// Destroy 释放 Pixmap 的 C 端资源。调用后 Pixmap 不可再使用。
func (p *Pixmap) Destroy() {
	if p.pix != nil && p.ctx != nil {
		p.ctx.WithLock(func() {
			C.gomupdf_drop_pixmap(p.ctx.ctx, p.pix)
		})
		p.pix = nil
	}
}

// Width 返回 Pixmap 的宽度（像素）。
func (p *Pixmap) Width() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var w C.int
	p.ctx.WithLock(func() {
		w = C.gomupdf_pixmap_width(p.ctx.ctx, p.pix)
	})
	return int(w)
}

// Height 返回 Pixmap 的高度（像素）。
func (p *Pixmap) Height() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var h C.int
	p.ctx.WithLock(func() {
		h = C.gomupdf_pixmap_height(p.ctx.ctx, p.pix)
	})
	return int(h)
}

// Stride 返回 Pixmap 的行跨度（每行字节数，可能大于 width * n）。
func (p *Pixmap) Stride() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var stride C.int
	p.ctx.WithLock(func() {
		stride = C.gomupdf_pixmap_stride(p.ctx.ctx, p.pix)
	})
	return int(stride)
}

// N 返回 Pixmap 的颜色分量数（如 RGB=3, RGBA=4, Gray=1）。
func (p *Pixmap) N() int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var n C.int
	p.ctx.WithLock(func() {
		n = C.gomupdf_pixmap_n(p.ctx.ctx, p.pix)
	})
	return int(n)
}

// Samples 返回原始像素数据（[]byte）。
// 通过 C.GoBytes 将 C 内存中的像素缓冲区拷贝为 Go 字节切片。
// 数据总长度 = stride * height。
func (p *Pixmap) Samples() []byte {
	if p.pix == nil || p.ctx == nil {
		return nil
	}

	var ptr *C.uchar
	var len C.int

	p.ctx.WithLock(func() {
		ptr = C.gomupdf_pixmap_samples(p.ctx.ctx, p.pix)
		stride := C.gomupdf_pixmap_stride(p.ctx.ctx, p.pix)
		height := C.gomupdf_pixmap_height(p.ctx.ctx, p.pix)
		len = stride * height
	})

	if ptr == nil || len == 0 {
		return nil
	}

	return C.GoBytes(unsafe.Pointer(ptr), len)
}

// SavePNG 将 Pixmap 保存为 PNG 文件。
// Go 字符串文件名通过 C.CString 转换，使用后通过 defer C.free 释放。
func (p *Pixmap) SavePNG(filename string) error {
	if p.pix == nil || p.ctx == nil {
		return errors.New("pixmap is nil")
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	p.ctx.WithLock(func() {
		C.gomupdf_save_pixmap_as_png(p.ctx.ctx, p.pix, cfilename)
	})

	return nil
}

// SaveJPEG 将 Pixmap 保存为 JPEG 文件，quality 为 JPEG 质量（1-100）。
func (p *Pixmap) SaveJPEG(filename string, quality int) error {
	if p.pix == nil || p.ctx == nil {
		return errors.New("pixmap is nil")
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	p.ctx.WithLock(func() {
		C.gomupdf_save_pixmap_as_jpeg(p.ctx.ctx, p.pix, cfilename, C.int(quality))
	})

	return nil
}

// PNGBytes 将 Pixmap 编码为 PNG 格式的字节切片。
// C 端通过输出参数（outData 指针 + outLen 长度）返回编码结果，
// Go 侧通过 C.GoBytes 拷贝为 []byte，然后通过 defer C.free 释放 C 端缓冲区。
func (p *Pixmap) PNGBytes() ([]byte, error) {
	if p.pix == nil || p.ctx == nil {
		return nil, errors.New("pixmap is nil")
	}
	var outData *C.uchar
	var outLen C.size_t
	var rc C.int
	p.ctx.WithLock(func() {
		rc = C.gomupdf_pixmap_to_png_bytes(p.ctx.ctx, p.pix, &outData, &outLen)
	})
	if rc != 0 || outData == nil {
		return nil, fmt.Errorf("failed to encode PNG: %s", GetLastError())
	}
	defer C.free(unsafe.Pointer(outData))
	return C.GoBytes(unsafe.Pointer(outData), C.int(outLen)), nil
}

// JPEGBytes 将 Pixmap 编码为 JPEG 格式的字节切片，quality 为 JPEG 质量（1-100）。
// 内存管理模式与 PNGBytes 相同。
func (p *Pixmap) JPEGBytes(quality int) ([]byte, error) {
	if p.pix == nil || p.ctx == nil {
		return nil, errors.New("pixmap is nil")
	}
	var outData *C.uchar
	var outLen C.size_t
	var rc C.int
	p.ctx.WithLock(func() {
		rc = C.gomupdf_pixmap_to_jpeg_bytes(p.ctx.ctx, p.pix, C.int(quality), &outData, &outLen)
	})
	if rc != 0 || outData == nil {
		return nil, fmt.Errorf("failed to encode JPEG: %s", GetLastError())
	}
	defer C.free(unsafe.Pointer(outData))
	return C.GoBytes(unsafe.Pointer(outData), C.int(outLen)), nil
}

// Pixel 返回指定坐标 (x, y) 处的像素值。
func (p *Pixmap) Pixel(x, y int) int {
	if p.pix == nil || p.ctx == nil {
		return 0
	}
	var val C.int
	p.ctx.WithLock(func() {
		val = C.gomupdf_pixmap_pixel(p.ctx.ctx, p.pix, C.int(x), C.int(y))
	})
	return int(val)
}

// SetPixel 设置指定坐标 (x, y) 处的像素值。
// Go int 通过 C.uint() 转换为 C.uint（无符号整型）传入 C 端。
func (p *Pixmap) SetPixel(x, y, val int) {
	if p.pix == nil || p.ctx == nil {
		return
	}
	p.ctx.WithLock(func() {
		C.gomupdf_pixmap_set_pixel(p.ctx.ctx, p.pix, C.int(x), C.int(y), C.uint(val))
	})
}

// ClearWith 使用指定值填充整个 Pixmap 的所有像素。
func (p *Pixmap) ClearWith(value int) {
	if p.pix == nil || p.ctx == nil {
		return
	}
	p.ctx.WithLock(func() {
		C.gomupdf_pixmap_clear_with(p.ctx.ctx, p.pix, C.int(value))
	})
}

// Invert 反转 Pixmap 的所有颜色（生成底片效果）。
func (p *Pixmap) Invert() {
	if p.pix == nil || p.ctx == nil {
		return
	}
	p.ctx.WithLock(func() {
		C.gomupdf_pixmap_invert(p.ctx.ctx, p.pix)
	})
}

// Gamma 对 Pixmap 应用伽马校正。gamma 参数通过 C.float() 转换传入 C 端。
func (p *Pixmap) Gamma(gamma float64) {
	if p.pix == nil || p.ctx == nil {
		return
	}
	p.ctx.WithLock(func() {
		C.gomupdf_pixmap_gamma(p.ctx.ctx, p.pix, C.float(gamma))
	})
}

// Tint 使用指定的黑/白值对 Pixmap 进行着色处理。
func (p *Pixmap) Tint(black, white int) {
	if p.pix == nil || p.ctx == nil {
		return
	}
	p.ctx.WithLock(func() {
		C.gomupdf_pixmap_tint(p.ctx.ctx, p.pix, C.int(black), C.int(white))
	})
}
