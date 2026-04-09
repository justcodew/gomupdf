package fitz

import (
	"fmt"
	"math"
)

// Rect represents a rectangle using float64 coordinates.
// The rectangle is defined by four coordinates: X0 (left), Y0 (top), X1 (right), Y1 (bottom).
type Rect struct {
	X0 float64
	Y0 float64
	X1 float64
	Y1 float64
}

// NewRect creates a new Rect with the given coordinates.
func NewRect(x0, y0, x1, y1 float64) *Rect {
	return &Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// String returns a string representation of the Rect.
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%g, %g, %g, %g)", r.X0, r.Y0, r.X1, r.Y1)
}

// IsEmpty returns true if the rectangle has zero width and height.
func (r Rect) IsEmpty() bool {
	return r.X0 == r.X1 && r.Y0 == r.Y1
}

// IsInfinite returns true if the rectangle represents infinity.
func (r Rect) IsInfinite() bool {
	return math.IsInf(r.X0, -1) && math.IsInf(r.X1, 1) &&
		math.IsInf(r.Y0, -1) && math.IsInf(r.Y1, 1)
}

// ContainsPoint returns true if the point (x, y) is inside the rectangle.
func (r Rect) ContainsPoint(x, y float64) bool {
	return x >= r.X0 && x <= r.X1 && y >= r.Y0 && y <= r.Y1
}

// ContainsRect returns true if the other rectangle is completely inside this rectangle.
func (r Rect) ContainsRect(other Rect) bool {
	return r.X0 <= other.X0 && r.X1 >= other.X1 && r.Y0 <= other.Y0 && r.Y1 >= other.Y1
}

// Intersects returns true if the two rectangles overlap.
func (r Rect) Intersects(other Rect) bool {
	return r.X0 < other.X1 && r.X1 > other.X0 && r.Y0 < other.Y1 && r.Y1 > other.Y0
}

// Intersection returns the intersection of two rectangles.
func (r Rect) Intersection(other Rect) Rect {
	if !r.Intersects(other) {
		return Rect{}
	}
	return Rect{
		X0: math.Max(r.X0, other.X0),
		Y0: math.Max(r.Y0, other.Y0),
		X1: math.Min(r.X1, other.X1),
		Y1: math.Min(r.Y1, other.Y1),
	}
}

// Inclusion returns the smallest rectangle that contains both rectangles.
func (r Rect) Inclusion(other Rect) Rect {
	return Rect{
		X0: math.Min(r.X0, other.X0),
		Y0: math.Min(r.Y0, other.Y0),
		X1: math.Max(r.X1, other.X1),
		Y1: math.Max(r.Y1, other.Y1),
	}
}

// Transform applies a matrix transformation to the rectangle.
func (r Rect) Transform(m Matrix) Rect {
	p1 := Point{X: r.X0, Y: r.Y0}.Transform(m)
	p2 := Point{X: r.X1, Y: r.Y0}.Transform(m)
	p3 := Point{X: r.X0, Y: r.Y1}.Transform(m)
	p4 := Point{X: r.X1, Y: r.Y1}.Transform(m)

	return Rect{
		X0: math.Min(math.Min(p1.X, p2.X), math.Min(p3.X, p4.X)),
		Y0: math.Min(math.Min(p1.Y, p2.Y), math.Min(p3.Y, p4.Y)),
		X1: math.Max(math.Max(p1.X, p2.X), math.Max(p3.X, p4.X)),
		Y1: math.Max(math.Max(p1.Y, p2.Y), math.Max(p3.Y, p4.Y)),
	}
}

// Width returns the width of the rectangle.
func (r Rect) Width() float64 {
	return r.X1 - r.X0
}

// Height returns the height of the rectangle.
func (r Rect) Height() float64 {
	return r.Y1 - r.Y0
}

// Size returns the width and height of the rectangle.
func (r Rect) Size() (width, height float64) {
	return r.Width(), r.Height()
}

// IRect returns the integer rectangle (rounded coordinates).
func (r Rect) IRect() IRect {
	return IRect{
		X0: int(math.Floor(r.X0)),
		Y0: int(math.Floor(r.Y0)),
		X1: int(math.Ceil(r.X1)),
		Y1: int(math.Ceil(r.Y1)),
	}
}
