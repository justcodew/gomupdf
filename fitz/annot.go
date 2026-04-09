package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// Annot represents a PDF annotation.
type Annot struct {
	ctx  *cgo_bindings.Context
	annot *cgo_bindings.Annot
}

// AnnotType represents the type of annotation.
type AnnotType int

const (
	AnnotText           AnnotType = 0
	AnnotLink           AnnotType = 1
	AnnotFreeText       AnnotType = 2
	AnnotLine           AnnotType = 3
	AnnotSquare         AnnotType = 4
	AnnotCircle         AnnotType = 5
	AnnotPolygon        AnnotType = 6
	AnnotPolyLine       AnnotType = 7
	AnnotHighlight      AnnotType = 8
	AnnotUnderline      AnnotType = 9
	AnnotSquiggly       AnnotType = 10
	AnnotStrikeOut      AnnotType = 11
	AnnotRedact         AnnotType = 12
	AnnotStamp          AnnotType = 13
	AnnotCaret          AnnotType = 14
	AnnotInk            AnnotType = 15
	AnnotPopup          AnnotType = 16
	AnnotFileAttachment AnnotType = 17
	AnnotSound          AnnotType = 18
	AnnotMovie          AnnotType = 19
	AnnotWidget         AnnotType = 20
	AnnotScreen         AnnotType = 21
	AnnotPrinterMark    AnnotType = 22
	AnnotTrapNet        AnnotType = 23
	AnnotWatermark      AnnotType = 24
	Annot3D             AnnotType = 25
	AnnotProjection     AnnotType = 26
)

// Type returns the annotation type.
func (a *Annot) Type() AnnotType {
	if a.annot == nil {
		return -1
	}
	return AnnotType(a.annot.Type())
}

// Rect returns the annotation rectangle.
func (a *Annot) Rect() Rect {
	if a.annot == nil {
		return Rect{}
	}
	x0, y0, x1, y1 := a.annot.Rect()
	return Rect{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// SetRect sets the annotation rectangle.
func (a *Annot) SetRect(r Rect) error {
	return a.annot.SetRect(r.X0, r.Y0, r.X1, r.Y1)
}

// Contents returns the annotation text content.
func (a *Annot) Contents() string {
	if a.annot == nil {
		return ""
	}
	return a.annot.Contents()
}

// SetContents sets the annotation text content.
func (a *Annot) SetContents(text string) error {
	return a.annot.SetContents(text)
}

// Color returns the annotation color as RGBA.
func (a *Annot) Color() Color {
	if a.annot == nil {
		return Color{}
	}
	r, g, b, alpha, _ := a.annot.Color()
	return Color{R: r, G: g, B: b, A: alpha}
}

// SetColor sets the annotation color.
func (a *Annot) SetColor(c Color) error {
	return a.annot.SetColor(c.R, c.G, c.B, c.A)
}

// Opacity returns the annotation opacity.
func (a *Annot) Opacity() float64 {
	if a.annot == nil {
		return 1.0
	}
	return a.annot.Opacity()
}

// SetOpacity sets the annotation opacity.
func (a *Annot) SetOpacity(opacity float64) error {
	return a.annot.SetOpacity(opacity)
}

// Flags returns the annotation flags.
func (a *Annot) Flags() int {
	if a.annot == nil {
		return 0
	}
	return a.annot.Flags()
}

// SetFlags sets the annotation flags.
func (a *Annot) SetFlags(flags int) error {
	return a.annot.SetFlags(flags)
}

// Border returns the annotation border width.
func (a *Annot) Border() float64 {
	if a.annot == nil {
		return 0
	}
	return a.annot.Border()
}

// SetBorder sets the annotation border width.
func (a *Annot) SetBorder(width float64) error {
	return a.annot.SetBorder(width)
}

// Title returns the annotation title.
func (a *Annot) Title() string {
	if a.annot == nil {
		return ""
	}
	return a.annot.Title()
}

// SetTitle sets the annotation title.
func (a *Annot) SetTitle(title string) error {
	return a.annot.SetTitle(title)
}

// Update saves changes to the annotation.
func (a *Annot) Update() error {
	return a.annot.Update()
}

// Delete removes the annotation from the page.
func (a *Annot) Delete() error {
	return a.annot.Delete()
}

// QuadPoints returns the annotation's quad points.
func (a *Annot) QuadPoints() []Rect {
	if a.annot == nil {
		return nil
	}
	return a.annot.QuadPoints()
}

// SetQuadPoints sets the annotation's quad points.
func (a *Annot) SetQuadPoints(quads []Rect) error {
	cgoQuads := make([]cgo_bindings.Rect, len(quads))
	for i, q := range quads {
		cgoQuads[i] = cgo_bindings.Rect{X0: q.X0, Y0: q.Y0, X1: q.X1, Y1: q.Y1}
	}
	return a.annot.SetQuadPoints(cgoQuads)
}

// String returns a string representation.
func (a *Annot) String() string {
	if a.annot == nil {
		return "Annot(<nil>)"
	}
	return fmt.Sprintf("Annot(type=%d, rect=%v)", a.Type(), a.Rect())
}
