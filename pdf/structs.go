package pdf

import (
	"image"
	"regexp"
	"sync"
)

// Common Chinese and English font names for validation
var commonFontNames = map[string]bool{
	"宋体": true, "黑体": true, "楷体": true, "仿宋": true,
	"微软雅黑": true, "方正": true, "华文": true, "思源": true,
	"苹方": true, "阿里巴巴": true, "站酷": true, "汉仪": true,
	"文泉驿": true, "源样": true,
	"Arial": true, "Times New Roman": true, "Calibri": true,
	"Helvetica": true, "Verdana": true, "Tahoma": true,
	"Georgia": true, "Courier": true, "Impact": true,
	"Comic Sans": true, "Consolas": true,
}

// Chinese character regex
var chineseRegex = regexp.MustCompile(`[\x{4E00}-\x{9FFF}]`)

// Garbled text pattern - 5+ consecutive non-ASCII characters (excluding common Chinese)
var garbledPattern = regexp.MustCompile(`[^\x00-\x7F\x{4E00}-\x{9FFF}]{5,}`)

// Font name validation regex
var fontNameRegex = regexp.MustCompile(`[\x{4E00}-\x{9FFF}A-Za-z0-9 _\-]+`)

// Font keyword list
var fontKeywords = []string{
	"体", "黑", "宋", "楷", "仿", "圆", "雅", "简", "繁", "明",
	"bold", "italic", "light",
}

// PixmapLock protects concurrent pixmap operations
var PixmapLock sync.Mutex

// ContentType represents the type of content
type ContentType int

const (
	ContentTypeText ContentType = iota
	ContentTypeImage
	ContentTypeSVG
)

// Span represents a text span with styling information
type Span struct {
	Type     ContentType // Content type
	Poly     []float64   // Polygon coordinates (8 points)
	Score    float64     // Confidence score
	Text     string      // Text content
	Flags    int         // Style flags
	Size     float64     // Font size
	Font     string       // Font name
	Color    int         // Text color
	Ascender float64     // Font ascender
	Descender float64    // Font descender
	Origin   Point       // Text origin
	Dir      Point       // Writing direction
	Garbled  bool        // Whether text is garbled
}

// ImagePosition represents the position of an image on a page
type ImagePosition struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// PageResult represents parsed results for a single page
type PageResult struct {
	Img         image.Image       `json:"-"` // PIL Image, skip in JSON
	PosImgs     []ImagePosition   `json:"pos_imgs"`
	PosSVGs     []ImagePosition   `json:"pos_svgs"`
	PosFigures  []ImagePosition   `json:"pos_figures"`
	OCRInfo     []Span            `json:"ocr_info"`
	OCRCharsInfo []CharInfo       `json:"ocr_chars_info"`
}

// CharInfo represents character-level information
type CharInfo struct {
	Bbox   []float64 `json:"bbox"`   // [x0, y0, x1, y1]
	Origin Point     `json:"origin"`
	Point  Point     `json:"point"`
	Size   float64   `json:"size"`
	Font   string    `json:"font"`
	Color  int       `json:"color"`
	Flags  int       `json:"flags"`
	Text   string    `json:"text"`
}

// PDFInfo represents comprehensive PDF information
type PDFInfo struct {
	IsResolvable    bool           `json:"is_resolvable"`    // Can be opened
	IsScanned       bool           `json:"is_scanned"`       // Is scanned PDF
	IsNeedsPassword bool           `json:"is_needs_password"` // Requires password
	IsEncrypted     bool           `json:"is_encrypted"`     // Is encrypted
	TotalPage       int            `json:"total_page"`       // Total page count
	ImgsPerPage     [][]ImagePosition `json:"imgs_per_page"`   // Image positions per page
	SVGsPerPage     [][]ImagePosition `json:"svgs_per_page"`   // SVG positions per page
	OCRPerPage      [][]Span       `json:"ocr_per_page"`     // OCR info per page
	OCRCharsPerPage [][]CharInfo   `json:"ocr_chars_per_page"` // OCR chars per page
	Pngs            [][]byte       `json:"-"`                // PNG bytes, skip in JSON
}
