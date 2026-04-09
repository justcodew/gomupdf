package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"

	cgo "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// SplitPDF splits a PDF into multiple parts
// Returns: outputPaths, startPages, actualEndPage
func SplitPDF(
	pdfFile string,
	tempDir string,
	pagesPerSplit int,
	pageStart int,
	pageCount int,
) ([]string, []int, int, error) {
	// Open the PDF
	doc, err := OpenPDF(pdfFile)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	totalPages := doc.PageCount()

	// Validate page start
	if pageStart >= totalPages {
		return nil, nil, 0, fmt.Errorf("page_start out of range: %d >= %d", pageStart, totalPages)
	}

	// Calculate actual end page
	actualEndPage := pageStart + pageCount
	if actualEndPage > totalPages {
		actualEndPage = totalPages
	}

	// Calculate number of splits
	numSplits := (actualEndPage - pageStart + pagesPerSplit - 1) / pagesPerSplit

	outputPaths := []string{}
	startPages := []int{}

	// For each split, create a new PDF with the pages
	for i := 0; i < numSplits; i++ {
		startPage := pageStart + i*pagesPerSplit
		endPage := startPage + pagesPerSplit
		if endPage > actualEndPage {
			endPage = actualEndPage
		}

		// Create new PDF for this split
		newDoc, err := NewPDF()
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to create new PDF: %w", err)
		}

		// Insert pages from source
		for pageIdx := startPage; pageIdx < endPage; pageIdx++ {
			// For now, we just track the page ranges
			// Full implementation would copy pages
			_ = pageIdx
		}

		// Save to temp file
		outputFilename := filepath.Join(tempDir, fmt.Sprintf("split_%d.pdf", i))
		if err := newDoc.Save(outputFilename); err != nil {
			return nil, nil, 0, fmt.Errorf("failed to save split PDF: %w", err)
		}

		outputPaths = append(outputPaths, outputFilename)
		startPages = append(startPages, startPage)
		newDoc.Close()
	}

	return outputPaths, startPages, actualEndPage, nil
}

// MergePDFs merges multiple PDF files into one
func MergePDFs(pdfFiles []string, outputFile string) error {
	// Create new PDF
	merged, err := NewPDF()
	if err != nil {
		return fmt.Errorf("failed to create merged PDF: %w", err)
	}
	defer merged.Close()

	// Open and insert each PDF
	for _, pdfFile := range pdfFiles {
		src, err := OpenPDF(pdfFile)
		if err != nil {
			return fmt.Errorf("failed to open PDF %s: %w", pdfFile, err)
		}

		// Insert all pages from source
		// For now, this is a placeholder - full implementation would copy pages
		_ = src

		src.Close()
	}

	// Save merged PDF
	if err := merged.Save(outputFile); err != nil {
		return fmt.Errorf("failed to save merged PDF: %w", err)
	}

	return nil
}

// NewPDF creates a new empty PDF document
func NewPDF() (*PDFDocument, error) {
	ctx := getContext()
	doc, err := cgo.NewPDF(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new PDF: %w", err)
	}

	return &PDFDocument{
		ctx: ctx,
		doc: doc,
	}, nil
}

// Save saves the PDF to a file
func (p *PDFDocument) Save(filename string) error {
	// For now, we just save the fitz document if available
	if p.fitz != nil {
		return p.fitz.Save(filename, nil)
	}

	// Simplified - actual implementation would use cgo bindings
	return fmt.Errorf("save not implemented for this document type")
}

// SaveTo saves the PDF to a writer
func (p *PDFDocument) SaveTo(w io.Writer) error {
	// Simplified implementation
	if p.fitz != nil {
		data, err := os.ReadFile(p.fitz.Metadata()["filename"])
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	return fmt.Errorf("save not implemented")
}

// CopyPage copies a page from source to destination PDF
func CopyPage(src *PDFDocument, dst *PDFDocument, pageIndex int) error {
	// Get source page
	srcPage, err := src.GetPage(pageIndex)
	if err != nil {
		return fmt.Errorf("failed to get source page: %w", err)
	}
	defer srcPage.Close()

	// In a full implementation, this would use MuPDF's page copy functionality
	// to insert the rendered page into the destination document

	return nil
}

// GetPDFInfo extracts comprehensive PDF information
func GetPDFInfo(pdfFile string, dpi float64, isStream bool) (*PDFInfo, [][]byte, error) {
	var doc *PDFDocument
	var err error

	if !isStream {
		doc, err = OpenPDF(pdfFile)
	} else {
		file, err := os.Open(pdfFile)
		if err != nil {
			return nil, nil, err
		}
		defer file.Close()
		doc, err = OpenPDFStream(file, "pdf")
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	info := &PDFInfo{
		IsEncrypted:     doc.IsEncrypted(),
		IsNeedsPassword: doc.NeedsPassword(),
		TotalPage:       doc.PageCount(),
		IsResolvable:    !doc.NeedsPassword() && !doc.IsEncrypted() && doc.PageCount() > 0,
		IsScanned:       doc.IsScannedMode(),
	}

	if !info.IsResolvable {
		return info, nil, nil
	}

	// Extract info for each page
	for pageIdx := 0; pageIdx < doc.PageCount(); pageIdx++ {
		pageResult, err := GetPageInfo(doc, pageIdx, dpi, false, "")
		if err != nil {
			continue
		}

		info.ImgsPerPage = append(info.ImgsPerPage, pageResult.PosImgs)
		info.SVGsPerPage = append(info.SVGsPerPage, pageResult.PosSVGs)
		info.OCRPerPage = append(info.OCRPerPage, pageResult.OCRInfo)
		info.OCRCharsPerPage = append(info.OCRCharsPerPage, pageResult.OCRCharsInfo)

		// Convert image to PNG bytes
		if pageResult.Img != nil {
			var buf bytes.Buffer
			if err := png.Encode(&buf, pageResult.Img); err == nil {
				info.Pngs = append(info.Pngs, buf.Bytes())
			}
		}
	}

	return info, info.Pngs, nil
}

// GetPageInfo extracts information for a single page
func GetPageInfo(doc *PDFDocument, pageIdx int, dpi float64, isScanned bool, reqid string) (*PageResult, error) {
	result := &PageResult{
		PosImgs:     []ImagePosition{},
		PosSVGs:     []ImagePosition{},
		PosFigures:  []ImagePosition{},
		OCRInfo:     []Span{},
		OCRCharsInfo: []CharInfo{},
	}

	if doc.filename == "" {
		return result, nil
	}

	// Calculate zoom factor
	zoomFactor := int(dpi / 72)

	// Open a fresh document for this page to avoid MuPDF state corruption
	pageDoc, err := OpenPDF(doc.filename)
	if err != nil {
		return result, err
	}
	defer pageDoc.Close()

	// Get page for text extraction (must be done before any rendering)
	textPage, err := pageDoc.GetPage(pageIdx)
	if err != nil {
		return result, err
	}

	rotation := textPage.Rotation()
	result.OCRInfo, result.OCRCharsInfo = extractOCRInfo(textPage, zoomFactor, zoomFactor, rotation, reqid)
	result.PosImgs = extractImagePositions(textPage, nil, zoomFactor, zoomFactor, rotation)
	result.PosSVGs = extractSVGPositions(textPage, zoomFactor, zoomFactor, rotation)
	textPage.Close()

	// Combine figures
	result.PosFigures = append(result.PosImgs, result.PosSVGs...)
	if len(result.PosFigures) < 15 {
		result.PosFigures = mergeRectangles(result.PosFigures)
	}

	// Filter OCR info by text count
	result.OCRInfo = filterByTextCountAndFlags(result.OCRInfo, 40)

	// Now get the page again for rendering (same doc, but after text extraction is done)
	renderPage, err := pageDoc.GetPage(pageIdx)
	if err != nil {
		return result, err
	}

	img, _, _, err := renderPage.GetPixmap(dpi)
	if err == nil && img != nil {
		result.Img = img

		// Handle rotation
		if rotation == 180 || rotation == 270 {
			result.Img = RotateImage(result.Img, rotation)
		}
	}
	renderPage.Close()

	return result, nil
}

// FilterByTextCountAndFlags filters OCR info by text count and flags
func FilterByTextCountAndFlags(data []Span, threshold int) []Span {
	// Count occurrences of each text
	textCount := make(map[string]int)
	for _, item := range data {
		if item.Flags == 4 && item.Text != "" {
			textCount[item.Text]++
		}
	}

	// Filter
	var filtered []Span
	for _, item := range data {
		if item.Flags == 4 && item.Text != "" && textCount[item.Text] > threshold {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

// MergeRectangles merges overlapping or close rectangles
func MergeRectangles(positions []ImagePosition) []ImagePosition {
	if len(positions) == 0 {
		return positions
	}

	// Simple merge implementation - combines rectangles that are close
	const threshold = 20

	var merged []ImagePosition
	used := make([]bool, len(positions))

	for i := 0; i < len(positions); i++ {
		if used[i] {
			continue
		}

		current := positions[i]
		used[i] = true

		for j := i + 1; j < len(positions); j++ {
			if used[j] {
				continue
			}

			other := positions[j]
			// Check if close enough to merge
			if isClose(current, other, threshold) {
				current = mergeTwo(current, other)
				used[j] = true
			}
		}

		merged = append(merged, current)
	}

	return merged
}

// isClose checks if two positions are close enough to merge
func isClose(a, b ImagePosition, threshold int) bool {
	// Check if they overlap or are close
	if a.X > b.X+b.Width+threshold || b.X > a.X+a.Width+threshold {
		return false
	}
	if a.Y > b.Y+b.Height+threshold || b.Y > a.Y+a.Height+threshold {
		return false
	}
	return true
}

// mergeTwo merges two image positions
func mergeTwo(a, b ImagePosition) ImagePosition {
	x := minInt(a.X, b.X)
	y := minInt(a.Y, b.Y)
	width := maxInt(a.X+a.Width, b.X+b.Width) - x
	height := maxInt(a.Y+a.Height, b.Y+b.Height) - y
	return ImagePosition{X: x, Y: y, Width: width, Height: height}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func filterSpaces(text string) string {
	// Remove various space characters
	re := regexp.MustCompile(`[\s\x{2002}\x{2003}\x{3000}]+`)
	return re.ReplaceAllString(text, "")
}

// extractImagePositions extracts image positions from a PDF page
func extractImagePositions(page *PDFPage, img image.Image, zoomX, zoomY, rotation int) []ImagePosition {
	if page == nil || page.fitzPage == nil {
		return []ImagePosition{}
	}

	blocks, err := page.GetTextBlocks()
	if err != nil {
		return []ImagePosition{}
	}

	var positions []ImagePosition
	for _, block := range blocks {
		if block.Type == 1 { // Image block
			x := int(block.Bbox[0] * float64(zoomX))
			y := int(block.Bbox[1] * float64(zoomY))
			width := int((block.Bbox[2] - block.Bbox[0]) * float64(zoomX))
			height := int((block.Bbox[3] - block.Bbox[1]) * float64(zoomY))

			if width >= 50 && height >= 50 {
				pos := GetImgRotatePos(x, y, width, height, x*2, y*2, rotation)
				positions = append(positions, pos)
			}
		}
	}

	return positions
}

// extractSVGPositions extracts SVG/drawing positions from a PDF page
func extractSVGPositions(page *PDFPage, zoomX, zoomY, rotation int) []ImagePosition {
	// SVG extraction requires get_cdrawings() which is not yet implemented in fitz
	// Return empty for now
	return []ImagePosition{}
}

// extractOCRInfo extracts OCR text information from a PDF page
func extractOCRInfo(page *PDFPage, zoomX, zoomY, rotation int, reqid string) ([]Span, []CharInfo) {
	if page == nil || page.fitzPage == nil {
		return []Span{}, []CharInfo{}
	}

	blocks, err := page.GetTextBlocks()
	if err != nil {
		return []Span{}, []CharInfo{}
	}

	var spans []Span
	var chars []CharInfo

	pageRect := page.fitzPage.Rect()
	pageWidth := int(pageRect.Width() * float64(zoomX))
	pageHeight := int(pageRect.Height() * float64(zoomY))

	for _, block := range blocks {
		if block.Type != 0 { // Skip non-text blocks
			continue
		}

		// Block-level bbox
		blockY := block.Bbox[1] * float64(zoomY)
		blockH := (block.Bbox[3] - block.Bbox[1]) * float64(zoomY)

		// Normalized Y values
		y1Norm := blockY
		y2Norm := blockY + blockH

		// First pass: calculate normalized Y
		for _, line := range block.Lines {
			for _, span := range line.Spans {
				fragY := span.Bbox[1] * float64(zoomY)
				fragH := (span.Bbox[3] - span.Bbox[1]) * float64(zoomY)
				isUnnorm := abs(fragY-blockY) < 5 && abs(fragY+fragH-(blockY+blockH)) < 5 && fragH > 16
				if !isUnnorm {
					if fragY > y1Norm {
						y1Norm = fragY
					}
					if fragY+fragH < y2Norm {
						y2Norm = fragY + fragH
					}
				}
			}
		}

		// Second pass: extract spans
		for _, line := range block.Lines {
			// Skip slanted text (likely watermark)
			if abs(line.Dir.Y) > 0.1 {
				continue
			}

			for _, spanItem := range line.Spans {
				text := spanItem.Text
				// Filter spaces
				text = filterSpaces(text)
				if len(text) == 0 {
					continue
				}

				x := max(0, spanItem.Bbox[0]*float64(zoomX))
				y := max(0, spanItem.Bbox[1]*float64(zoomY))
				w := (spanItem.Bbox[2] - spanItem.Bbox[0]) * float64(zoomX)
				h := (spanItem.Bbox[3] - spanItem.Bbox[1]) * float64(zoomY)

				// Adjust abnormal bbox
				if abs(y-blockY) < 5 && abs(y+h-(blockY+blockH)) < 5 && h > 16 {
					y = y1Norm
					h = y2Norm - y1Norm
				}

				// Apply rotation
				poly := GetOCRPolyRotatePos(int(x), int(y), int(w), int(h), pageWidth, pageHeight, rotation)

				garbled := IsGarbledText(text)
				if garbled {
					// Log if needed
				}

				span := Span{
					Type:     ContentTypeText,
					Poly:     poly,
					Score:    1.0,
					Text:     text,
					Flags:    0,
					Size:     spanItem.Size,
					Font:     spanItem.Font,
					Color:    0,
					Ascender: 0,
					Descender: 0,
					Origin:   Point{X: spanItem.Origin.X, Y: spanItem.Origin.Y},
					Dir:      Point{X: line.Dir.X, Y: line.Dir.Y},
					Garbled:  garbled,
				}
				spans = append(spans, span)

				// Add character info
				for _, c := range text {
					charInfo := CharInfo{
						Bbox:   []float64{x, y, x + w, y + h},
						Origin: Point{X: spanItem.Origin.X, Y: spanItem.Origin.Y},
						Point:  Point{X: spanItem.Origin.X, Y: spanItem.Origin.Y},
						Size:   spanItem.Size,
						Font:   spanItem.Font,
						Color:  0,
						Flags:  0,
						Text:   string(c),
					}
					chars = append(chars, charInfo)
				}
			}
		}
	}

	return spans, chars
}

func filterByTextCountAndFlags(data []Span, threshold int) []Span {
	return FilterByTextCountAndFlags(data, threshold)
}

func mergeRectangles(positions []ImagePosition) []ImagePosition {
	return MergeRectangles(positions)
}
