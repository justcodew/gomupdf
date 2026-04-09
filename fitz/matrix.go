package fitz

import (
	"fmt"
	"math"
)

// Matrix represents a 2D affine transformation matrix.
// The matrix is stored in the form:
//   [ A B ]
//   [ C D ]
//   [ E F ]
//
// Which represents the transformation:
//   x' = A*x + C*y + E
//   y' = B*x + D*y + F
type Matrix struct {
	A, B float64 // first row
	C, D float64 // second row
	E, F float64 // third row (translation)
}

// String returns a string representation of the Matrix.
func (m Matrix) String() string {
	return fmt.Sprintf("Matrix(%g, %g, %g, %g, %g, %g)", m.A, m.B, m.C, m.D, m.E, m.F)
}

// Identity returns the identity matrix.
func Identity() Matrix {
	return Matrix{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}
}

// NewMatrix creates a matrix with the given values.
func NewMatrix(a, b, c, d, e, f float64) Matrix {
	return Matrix{A: a, B: b, C: c, D: d, E: e, F: f}
}

// NewMatrixZoom creates a scaling matrix.
func NewMatrixZoom(zoomX, zoomY float64) Matrix {
	return Matrix{A: zoomX, B: 0, C: 0, D: zoomY, E: 0, F: 0}
}

// NewMatrixScale creates a scaling matrix (shorthand for uniform zoom).
func NewMatrixScale(s float64) Matrix {
	return NewMatrixZoom(s, s)
}

// NewMatrixRotation creates a rotation matrix (angle in radians).
func NewMatrixRotation(radians float64) Matrix {
	s := math.Sin(radians)
	c := math.Cos(radians)
	return Matrix{A: c, B: s, C: -s, D: c, E: 0, F: 0}
}

// NewMatrixShear creates a shear matrix.
func NewMatrixShear(h, v float64) Matrix {
	return Matrix{A: 1, B: v, C: h, D: 1, E: 0, F: 0}
}

// NewMatrixTranslation creates a translation matrix.
func NewMatrixTranslation(tx, ty float64) Matrix {
	return Matrix{A: 1, B: 0, C: 0, D: 1, E: tx, F: ty}
}

// Invert returns the inverse of the matrix.
func (m Matrix) Invert() (Matrix, error) {
	det := m.A*m.D - m.B*m.C
	if math.Abs(det) < 1e-10 {
		return Matrix{}, fmt.Errorf("matrix is singular")
	}
	invDet := 1.0 / det
	return Matrix{
		A:  m.D * invDet,
		B: -m.B * invDet,
		C: -m.C * invDet,
		D:  m.A * invDet,
		E: (m.C*m.F - m.D*m.E) * invDet,
		F: (m.B*m.E - m.A*m.F) * invDet,
	}, nil
}

// Concat returns the concatenation of two matrices (m * other).
func (m Matrix) Concat(other Matrix) Matrix {
	return Matrix{
		A: m.A*other.A + m.B*other.C,
		B: m.A*other.B + m.B*other.D,
		C: m.C*other.A + m.D*other.C,
		D: m.C*other.B + m.D*other.D,
		E: m.E*other.A + m.F*other.C + other.E,
		F: m.E*other.B + m.F*other.D + other.F,
	}
}

// TransformPoint applies the matrix transformation to a point.
func (m Matrix) TransformPoint(p Point) Point {
	return Point{
		X: m.A*p.X + m.C*p.Y + m.E,
		Y: m.B*p.X + m.D*p.Y + m.F,
	}
}

// TransformRect applies the matrix transformation to a rectangle.
func (m Matrix) TransformRect(r Rect) Rect {
	return r.Transform(m)
}

// Determinant returns the determinant of the matrix.
func (m Matrix) Determinant() float64 {
	return m.A*m.D - m.B*m.C
}

// IsIdentity returns true if the matrix is the identity matrix.
func (m Matrix) IsIdentity() bool {
	const eps = 1e-10
	return math.Abs(m.A-1) < eps && math.Abs(m.B) < eps &&
		math.Abs(m.C) < eps && math.Abs(m.D-1) < eps &&
		math.Abs(m.E) < eps && math.Abs(m.F) < eps
}

// Scale returns the scale factors of the matrix.
func (m Matrix) Scale() (sx, sy float64) {
	sx = math.Sqrt(m.A*m.A + m.B*m.B)
	sy = math.Sqrt(m.C*m.C + m.D*m.D)
	return
}

// Prescaled returns the matrix with scale factors removed.
func (m Matrix) Prescaled() Matrix {
	sx, sy := m.Scale()
	if sx < 1e-10 || sy < 1e-10 {
		return Matrix{}
	}
	return Matrix{
		A: m.A / sx,
		B: m.B / sx,
		C: m.C / sy,
		D: m.D / sy,
		E: m.E,
		F: m.F,
	}
}
