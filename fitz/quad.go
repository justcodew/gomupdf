package fitz

import (
	"fmt"
)

// Quad represents a quadrilateral (four-sided polygon) defined by four points.
type Quad struct {
	UL Point // Upper-left corner
	UR Point // Upper-right corner
	LL Point // Lower-left corner
	LR Point // Lower-right corner
}

// NewQuad creates a new Quad with the given four corners.
func NewQuad(ul, ur, ll, lr Point) *Quad {
	return &Quad{UL: ul, UR: ur, LL: ll, LR: lr}
}

// String returns a string representation of the Quad.
func (q Quad) String() string {
	return fmt.Sprintf("Quad(%v, %v, %v, %v)", q.UL, q.UR, q.LL, q.LR)
}

// Rect returns the bounding rectangle of the quad.
func (q Quad) Rect() Rect {
	minX := min(q.UL.X, q.UR.X, q.LL.X, q.LR.X)
	maxX := max(q.UL.X, q.UR.X, q.LL.X, q.LR.X)
	minY := min(q.UL.Y, q.UR.Y, q.LL.Y, q.LR.Y)
	maxY := max(q.UL.Y, q.UR.Y, q.LL.Y, q.LR.Y)
	return Rect{X0: minX, Y0: minY, X1: maxX, Y1: maxY}
}

// Transform applies a matrix transformation to the quad.
func (q Quad) Transform(m Matrix) Quad {
	return Quad{
		UL: q.UL.Transform(m),
		UR: q.UR.Transform(m),
		LL: q.LL.Transform(m),
		LR: q.LR.Transform(m),
	}
}

// QuadFromRect creates a quad from a rectangle.
func QuadFromRect(r Rect) Quad {
	return Quad{
		UL: Point{X: r.X0, Y: r.Y0},
		UR: Point{X: r.X1, Y: r.Y0},
		LL: Point{X: r.X0, Y: r.Y1},
		LR: Point{X: r.X1, Y: r.Y1},
	}
}

func min(values ...float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func max(values ...float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
