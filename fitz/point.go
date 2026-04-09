package fitz

import (
	"fmt"
	"math"
)

// Point represents a 2D point using float64 coordinates.
type Point struct {
	X float64
	Y float64
}

// NewPoint creates a new Point with the given coordinates.
func NewPoint(x, y float64) *Point {
	return &Point{X: x, Y: y}
}

// String returns a string representation of the Point.
func (p Point) String() string {
	return fmt.Sprintf("Point(%g, %g)", p.X, p.Y)
}

// DistanceTo returns the Euclidean distance to another point.
func (p Point) DistanceTo(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Transform applies a matrix transformation to the point.
func (p Point) Transform(m Matrix) Point {
	return Point{
		X: m.A*p.X + m.C*p.Y + m.E,
		Y: m.B*p.X + m.D*p.Y + m.F,
	}
}

// Add returns the sum of two points.
func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub returns the difference of two points.
func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

// Mul returns the point multiplied by a scalar.
func (p Point) Mul(scalar float64) Point {
	return Point{X: p.X * scalar, Y: p.Y * scalar}
}
