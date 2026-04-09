// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（geometry.go）包含几何相关的操作函数，涵盖：
//   - 矩阵运算（单位矩阵、缩放、旋转、平移、拼接、求逆）
//   - 点运算（创建点、矩阵变换点）
//   - 矩形运算（创建矩形、判空、判无限、矩阵变换矩形）
//   - 四边形运算（创建四边形、包围矩形、矩阵变换四边形）
//   - 颜色空间查找
//   - 图像信息查询（宽高、分量数、位深度、颜色空间名称）
//
// CGO 模式说明：
//   - C.fz_matrix 是 6 元素的仿射矩阵 {a, b, c, d, e, f}
//   - C.fz_point 是二维点 {x, y}
//   - C.fz_rect 是矩形 {x0, y0, x1, y1}
//   - C.fz_quad 是四边形 {ul, ur, ll, lr}，每个角为 fz_point
//   - Go float64 通过 C.float() 转换传入 C 端，C 端返回值通过 float64() 转回 Go
//   - 所有函数返回多个 float64 值（元组风格），避免定义大量小型 Go 结构体
package cgo

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/mupdf/lib -lmupdf -lmupdfcpp
#cgo CFLAGS: -I/opt/homebrew/opt/mupdf/include

#include "bindings.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// ==================== 矩阵运算 ====================

// IdentityMatrix 返回单位变换矩阵 (1, 0, 0, 1, 0, 0)，即不进行任何变换。
func IdentityMatrix() (a, b, c, d, e, f float64) {
	m := C.gomupdf_identity_matrix()
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// MakeMatrix 从 6 个分量 (a, b, c, d, e, f) 构造仿射变换矩阵。
func MakeMatrix(a, b, c, d, e, f float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_make_matrix(C.float(a), C.float(b), C.float(c), C.float(d), C.float(e), C.float(f))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// ScaleMatrix 创建缩放矩阵，sx 和 sy 分别为水平和垂直方向的缩放因子。
func ScaleMatrix(sx, sy float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_scale_matrix(C.float(sx), C.float(sy))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// RotateMatrix 创建旋转矩阵，degrees 为顺时针旋转角度。
func RotateMatrix(degrees float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_rotate_matrix(C.float(degrees))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// TranslateMatrix 创建平移矩阵，tx 和 ty 分别为水平和垂直方向的平移量。
func TranslateMatrix(tx, ty float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_translate_matrix(C.float(tx), C.float(ty))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// ConcatMatrix 将两个仿射矩阵拼接（相乘），返回组合后的矩阵。
// 左矩阵和右矩阵均以 6 个 float64 分量传入，通过 C.fz_matrix 构造体传递给 C 端。
func ConcatMatrix(la, lb, lc, ld, le, lf float64, ra, rb, rc, rd, re, rf float64) (float64, float64, float64, float64, float64, float64) {
	left := C.fz_matrix{a: C.float(la), b: C.float(lb), c: C.float(lc), d: C.float(ld), e: C.float(le), f: C.float(lf)}
	right := C.fz_matrix{a: C.float(ra), b: C.float(rb), c: C.float(rc), d: C.float(rd), e: C.float(re), f: C.float(rf)}
	m := C.gomupdf_concat_matrix(left, right)
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// InvertMatrix 对给定矩阵求逆，返回逆矩阵。需要 Context 因为内部可能使用 MuPDF 的异常处理。
func (ctx *Context) InvertMatrix(a, b, c, d, e, f float64) (float64, float64, float64, float64, float64, float64) {
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	inv := C.gomupdf_invert_matrix(ctx.ctx, m)
	return float64(inv.a), float64(inv.b), float64(inv.c), float64(inv.d), float64(inv.e), float64(inv.f)
}

// ==================== 点运算 ====================

// MakePoint 创建一个二维点 (x, y)。
func MakePoint(x, y float64) (float64, float64) {
	p := C.gomupdf_make_point(C.float(x), C.float(y))
	return float64(p.x), float64(p.y)
}

// TransformPoint 使用仿射矩阵变换一个点，返回变换后的坐标。
func TransformPoint(px, py float64, a, b, c, d, e, f float64) (float64, float64) {
	p := C.fz_point{x: C.float(px), y: C.float(py)}
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	r := C.gomupdf_transform_point(p, m)
	return float64(r.x), float64(r.y)
}

// ==================== 矩形运算 ====================

// MakeRect 创建一个浮点矩形 (x0, y0, x1, y1)。
func MakeRect(x0, y0, x1, y1 float64) (float64, float64, float64, float64) {
	r := C.gomupdf_make_rect(C.float(x0), C.float(y0), C.float(x1), C.float(y1))
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}

// MakeIRect 创建一个整数矩形 (x0, y0, x1, y1)。
func MakeIRect(x0, y0, x1, y1 int) (int, int, int, int) {
	r := C.gomupdf_make_irect(C.int(x0), C.int(y0), C.int(x1), C.int(y1))
	return int(r.x0), int(r.y0), int(r.x1), int(r.y1)
}

// RectIsEmpty 判断矩形是否为空（面积为零或无效）。
func RectIsEmpty(x0, y0, x1, y1 float64) bool {
	r := C.fz_rect{x0: C.float(x0), y0: C.float(y0), x1: C.float(x1), y1: C.float(y1)}
	return C.gomupdf_rect_is_empty(r) != 0
}

// RectIsInfinite 判断矩形是否为无限大。
func RectIsInfinite(x0, y0, x1, y1 float64) bool {
	r := C.fz_rect{x0: C.float(x0), y0: C.float(y0), x1: C.float(x1), y1: C.float(y1)}
	return C.gomupdf_rect_is_infinite(r) != 0
}

// TransformRect 使用仿射矩阵变换矩形，返回变换后的包围矩形。
func TransformRect(rx0, ry0, rx1, ry1 float64, a, b, c, d, e, f float64) (float64, float64, float64, float64) {
	r := C.fz_rect{x0: C.float(rx0), y0: C.float(ry0), x1: C.float(rx1), y1: C.float(ry1)}
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	res := C.gomupdf_transform_rect(r, m)
	return float64(res.x0), float64(res.y0), float64(res.x1), float64(res.y1)
}

// ==================== 四边形运算 ====================

// MakeQuad 由四个顶点创建一个四边形。
// ul=左上, ur=右上, ll=左下, lr=右下。
// 返回 8 个 float64 值，依次为四角坐标。
func MakeQuad(ulx, uly, urx, ury, llx, lly, lrx, lry float64) (float64, float64, float64, float64, float64, float64, float64, float64) {
	ul := C.fz_point{x: C.float(ulx), y: C.float(uly)}
	ur := C.fz_point{x: C.float(urx), y: C.float(ury)}
	ll := C.fz_point{x: C.float(llx), y: C.float(lly)}
	lr := C.fz_point{x: C.float(lrx), y: C.float(lry)}
	q := C.gomupdf_make_quad(ul, ur, ll, lr)
	return float64(q.ul.x), float64(q.ul.y), float64(q.ur.x), float64(q.ur.y),
		float64(q.ll.x), float64(q.ll.y), float64(q.lr.x), float64(q.lr.y)
}

// QuadRect 返回四边形的最小包围矩形。
func QuadRect(ulx, uly, urx, ury, llx, lly, lrx, lry float64) (float64, float64, float64, float64) {
	ul := C.fz_point{x: C.float(ulx), y: C.float(uly)}
	ur := C.fz_point{x: C.float(urx), y: C.float(ury)}
	ll := C.fz_point{x: C.float(llx), y: C.float(lly)}
	lr := C.fz_point{x: C.float(lrx), y: C.float(lry)}
	q := C.gomupdf_make_quad(ul, ur, ll, lr)
	r := C.gomupdf_quad_rect(q)
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}

// TransformQuad 使用仿射矩阵变换四边形，返回变换后的四角坐标。
func TransformQuad(ulx, uly, urx, ury, llx, lly, lrx, lry float64, a, b, c, d, e, f float64) (float64, float64, float64, float64, float64, float64, float64, float64) {
	ul := C.fz_point{x: C.float(ulx), y: C.float(uly)}
	ur := C.fz_point{x: C.float(urx), y: C.float(ury)}
	ll := C.fz_point{x: C.float(llx), y: C.float(lly)}
	lr := C.fz_point{x: C.float(lrx), y: C.float(lry)}
	q := C.gomupdf_make_quad(ul, ur, ll, lr)
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	res := C.gomupdf_transform_quad(q, m)
	return float64(res.ul.x), float64(res.ul.y), float64(res.ur.x), float64(res.ur.y),
		float64(res.ll.x), float64(res.ll.y), float64(res.lr.x), float64(res.lr.y)
}

// ==================== 颜色空间操作 ====================

// FindDeviceColorspace 按名称查找设备颜色空间，支持 "RGB"、"Gray"、"CMYK"。
// 返回 *C.fz_colorspace 指针，可直接传给其他需要颜色空间的 CGO 函数。
// Go 字符串通过 C.CString 转换，通过 defer C.free 释放。
func FindDeviceColorspace(ctx *Context, name string) *C.fz_colorspace {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return C.gomupdf_find_device_colorspace(ctx.ctx, cname)
}

// PixmapColorspace 返回 Pixmap 的颜色空间指针。
func (ctx *Context) PixmapColorspace(pix *C.fz_pixmap) *C.fz_colorspace {
	return C.gomupdf_pixmap_colorspace(ctx.ctx, pix)
}

// ==================== 图像信息操作 ====================

// ImageWidth 返回 fz_image 的宽度（像素）。
func ImageWidth(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_width(ctx.ctx, img))
}

// ImageHeight 返回 fz_image 的高度（像素）。
func ImageHeight(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_height(ctx.ctx, img))
}

// ImageN 返回 fz_image 的颜色分量数（如 RGB 为 3，RGBA 为 4）。
func ImageN(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_n(ctx.ctx, img))
}

// ImageBPC 返回 fz_image 每个分量的位深度（bits per component）。
func ImageBPC(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_bpc(ctx.ctx, img))
}

// ImageColorspaceName 返回 fz_image 颜色空间的名称字符串。
// 通过 C.GoString 将 C 端返回的 *C.char 转换为 Go string。
func ImageColorspaceName(ctx *Context, img *C.fz_image) string {
	return C.GoString(C.gomupdf_image_colorspace_name(ctx.ctx, img))
}

// BlockGetImage 从图像类型的 stext_block 中提取 fz_image 指针。
// 仅当 BlockType() == TextBlockTypeImage 时有效。
func BlockGetImage(ctx *Context, block *C.fz_stext_block) *C.fz_image {
	return C.gomupdf_stext_block_get_image(ctx.ctx, block)
}
