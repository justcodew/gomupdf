package fitz

import (
	"fmt"
)

// IRect represents a rectangle using integer coordinates.
type IRect struct {
	X0 int
	Y0 int
	X1 int
	Y1 int
}

// String returns a string representation of the IRect.
func (r IRect) String() string {
	return fmt.Sprintf("IRect(%d, %d, %d, %d)", r.X0, r.Y0, r.X1, r.Y1)
}

// IsEmpty returns true if the rectangle has zero width and height.
func (r IRect) IsEmpty() bool {
	return r.X0 == r.X1 && r.Y0 == r.Y1
}

// Width returns the width of the rectangle.
func (r IRect) Width() int {
	return r.X1 - r.X0
}

// Height returns the height of the rectangle.
func (r IRect) Height() int {
	return r.Y1 - r.Y0
}

// Rect returns the floating-point Rect representation.
func (r IRect) Rect() Rect {
	return Rect{
		X0: float64(r.X0),
		Y0: float64(r.Y0),
		X1: float64(r.X1),
		Y1: float64(r.Y1),
	}
}
