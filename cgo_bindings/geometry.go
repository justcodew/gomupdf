package cgo

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/mupdf/lib -lmupdf -lmupdfcpp
#cgo CFLAGS: -I/opt/homebrew/opt/mupdf/include

#include "bindings.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// Matrix operations

// IdentityMatrix returns the identity transformation matrix.
func IdentityMatrix() (a, b, c, d, e, f float64) {
	m := C.gomupdf_identity_matrix()
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// MakeMatrix creates a transformation matrix from individual components.
func MakeMatrix(a, b, c, d, e, f float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_make_matrix(C.float(a), C.float(b), C.float(c), C.float(d), C.float(e), C.float(f))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// ScaleMatrix creates a scaling matrix.
func ScaleMatrix(sx, sy float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_scale_matrix(C.float(sx), C.float(sy))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// RotateMatrix creates a rotation matrix.
func RotateMatrix(degrees float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_rotate_matrix(C.float(degrees))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// TranslateMatrix creates a translation matrix.
func TranslateMatrix(tx, ty float64) (float64, float64, float64, float64, float64, float64) {
	m := C.gomupdf_translate_matrix(C.float(tx), C.float(ty))
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// ConcatMatrix concatenates two matrices.
func ConcatMatrix(la, lb, lc, ld, le, lf float64, ra, rb, rc, rd, re, rf float64) (float64, float64, float64, float64, float64, float64) {
	left := C.fz_matrix{a: C.float(la), b: C.float(lb), c: C.float(lc), d: C.float(ld), e: C.float(le), f: C.float(lf)}
	right := C.fz_matrix{a: C.float(ra), b: C.float(rb), c: C.float(rc), d: C.float(rd), e: C.float(re), f: C.float(rf)}
	m := C.gomupdf_concat_matrix(left, right)
	return float64(m.a), float64(m.b), float64(m.c), float64(m.d), float64(m.e), float64(m.f)
}

// InvertMatrix inverts a matrix.
func (ctx *Context) InvertMatrix(a, b, c, d, e, f float64) (float64, float64, float64, float64, float64, float64) {
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	inv := C.gomupdf_invert_matrix(ctx.ctx, m)
	return float64(inv.a), float64(inv.b), float64(inv.c), float64(inv.d), float64(inv.e), float64(inv.f)
}

// Point operations

// MakePoint creates a point.
func MakePoint(x, y float64) (float64, float64) {
	p := C.gomupdf_make_point(C.float(x), C.float(y))
	return float64(p.x), float64(p.y)
}

// TransformPoint transforms a point by a matrix.
func TransformPoint(px, py float64, a, b, c, d, e, f float64) (float64, float64) {
	p := C.fz_point{x: C.float(px), y: C.float(py)}
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	r := C.gomupdf_transform_point(p, m)
	return float64(r.x), float64(r.y)
}

// Rect operations

// MakeRect creates a rectangle.
func MakeRect(x0, y0, x1, y1 float64) (float64, float64, float64, float64) {
	r := C.gomupdf_make_rect(C.float(x0), C.float(y0), C.float(x1), C.float(y1))
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}

// MakeIRect creates an integer rectangle.
func MakeIRect(x0, y0, x1, y1 int) (int, int, int, int) {
	r := C.gomupdf_make_irect(C.int(x0), C.int(y0), C.int(x1), C.int(y1))
	return int(r.x0), int(r.y0), int(r.x1), int(r.y1)
}

// RectIsEmpty returns true if the rectangle is empty.
func RectIsEmpty(x0, y0, x1, y1 float64) bool {
	r := C.fz_rect{x0: C.float(x0), y0: C.float(y0), x1: C.float(x1), y1: C.float(y1)}
	return C.gomupdf_rect_is_empty(r) != 0
}

// RectIsInfinite returns true if the rectangle is infinite.
func RectIsInfinite(x0, y0, x1, y1 float64) bool {
	r := C.fz_rect{x0: C.float(x0), y0: C.float(y0), x1: C.float(x1), y1: C.float(y1)}
	return C.gomupdf_rect_is_infinite(r) != 0
}

// TransformRect transforms a rectangle by a matrix.
func TransformRect(rx0, ry0, rx1, ry1 float64, a, b, c, d, e, f float64) (float64, float64, float64, float64) {
	r := C.fz_rect{x0: C.float(rx0), y0: C.float(ry0), x1: C.float(rx1), y1: C.float(ry1)}
	m := C.fz_matrix{a: C.float(a), b: C.float(b), c: C.float(c), d: C.float(d), e: C.float(e), f: C.float(f)}
	res := C.gomupdf_transform_rect(r, m)
	return float64(res.x0), float64(res.y0), float64(res.x1), float64(res.y1)
}

// Quad operations

// MakeQuad creates a quadrilateral from four points.
func MakeQuad(ulx, uly, urx, ury, llx, lly, lrx, lry float64) (float64, float64, float64, float64, float64, float64, float64, float64) {
	ul := C.fz_point{x: C.float(ulx), y: C.float(uly)}
	ur := C.fz_point{x: C.float(urx), y: C.float(ury)}
	ll := C.fz_point{x: C.float(llx), y: C.float(lly)}
	lr := C.fz_point{x: C.float(lrx), y: C.float(lry)}
	q := C.gomupdf_make_quad(ul, ur, ll, lr)
	return float64(q.ul.x), float64(q.ul.y), float64(q.ur.x), float64(q.ur.y),
		float64(q.ll.x), float64(q.ll.y), float64(q.lr.x), float64(q.lr.y)
}

// QuadRect returns the bounding rectangle of a quad.
func QuadRect(ulx, uly, urx, ury, llx, lly, lrx, lry float64) (float64, float64, float64, float64) {
	ul := C.fz_point{x: C.float(ulx), y: C.float(uly)}
	ur := C.fz_point{x: C.float(urx), y: C.float(ury)}
	ll := C.fz_point{x: C.float(llx), y: C.float(lly)}
	lr := C.fz_point{x: C.float(lrx), y: C.float(lry)}
	q := C.gomupdf_make_quad(ul, ur, ll, lr)
	r := C.gomupdf_quad_rect(q)
	return float64(r.x0), float64(r.y0), float64(r.x1), float64(r.y1)
}

// TransformQuad transforms a quad by a matrix.
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

// Colorspace operations

// FindDeviceColorspace returns a device colorspace by name ("RGB", "Gray", "CMYK").
func FindDeviceColorspace(ctx *Context, name string) *C.fz_colorspace {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return C.gomupdf_find_device_colorspace(ctx.ctx, cname)
}

// PixmapColorspace returns the colorspace of a pixmap.
func (ctx *Context) PixmapColorspace(pix *C.fz_pixmap) *C.fz_colorspace {
	return C.gomupdf_pixmap_colorspace(ctx.ctx, pix)
}

// Image operations

// ImageWidth returns the width of an image.
func ImageWidth(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_width(ctx.ctx, img))
}

// ImageHeight returns the height of an image.
func ImageHeight(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_height(ctx.ctx, img))
}

// ImageN returns the number of components in an image.
func ImageN(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_n(ctx.ctx, img))
}

// ImageBPC returns the bits per component of an image.
func ImageBPC(ctx *Context, img *C.fz_image) int {
	return int(C.gomupdf_image_bpc(ctx.ctx, img))
}

// ImageColorspaceName returns the colorspace name of an image.
func ImageColorspaceName(ctx *Context, img *C.fz_image) string {
	return C.GoString(C.gomupdf_image_colorspace_name(ctx.ctx, img))
}

// BlockGetImage extracts the image from an image-type stext block.
func BlockGetImage(ctx *Context, block *C.fz_stext_block) *C.fz_image {
	return C.gomupdf_stext_block_get_image(ctx.ctx, block)
}
