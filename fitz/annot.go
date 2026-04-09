package fitz

// Annot represents a PDF annotation.
type Annot struct {
	// TODO: implement
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
