package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-pymupdf/gomupdf/pdf"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	chiFont font.Face
)

func initFonts(ttfPath string) error {
	// Load TTF font if available
	if _, err := os.Stat(ttfPath); err == nil {
		data, err := os.ReadFile(ttfPath)
		if err == nil {
			f, err := parseTTF(data)
			if err == nil {
				chiFont = newFont(f, 18)
				return nil
			}
		}
	}

	// Fallback to basic font
	chiFont = basicfont.Face7x13
	return nil
}

// Simple TTF parser and font renderer
type ttfFont struct {
	data []byte
}

func parseTTF(data []byte) (*ttfFont, error) {
	return &ttfFont{data: data}, nil
}

func newFont(f *ttfFont, size float64) font.Face {
	// Use basic font as fallback since proper TTF rendering is complex
	return basicfont.Face7x13
}

func drawString(img *image.RGBA, x, y int, text string, col color.RGBA) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{col},
		Face: chiFont,
		Dot:  point,
	}
	d.DrawString(text)
}

func main() {
	inputFile := flag.String("i", "", "Input PDF file")
	outputDir := flag.String("o", "", "Output directory (default: <input>_vis)")
	fontPath := flag.String("f", "/Users/xiongzhaolong/Downloads/code/code_2026/STKAITI.TTF", "Chinese TTF font path")
	dpi := flag.Float64("d", 144.0, "DPI for rendering")
	pagesFlag := flag.String("p", "", "Pages to process (e.g., 0,1,2 or 0-5)")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: visualize -i <input.pdf> [-o output_dir] [-f font.ttf] [-d dpi] [-p pages]")
		flag.PrintDefaults()
		return
	}

	// Initialize fonts
	initFonts(*fontPath)

	// Default output dir
	if *outputDir == "" {
		*outputDir = strings.TrimSuffix(*inputFile, ".pdf") + "_vis"
	}

	// Parse pages
	var pages []int
	if *pagesFlag != "" {
		parts := strings.Split(*pagesFlag, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if strings.Contains(p, "-") {
				rangeParts := strings.Split(p, "-")
				if len(rangeParts) == 2 {
					start, _ := strconv.Atoi(rangeParts[0])
					end, _ := strconv.Atoi(rangeParts[1])
					for i := start; i <= end; i++ {
						pages = append(pages, i)
					}
				}
			} else {
				num, _ := strconv.Atoi(p)
				pages = append(pages, num)
			}
		}
	}

	fmt.Printf("Opening: %s\n", *inputFile)

	// Open doc to get page count
	doc, err := pdf.OpenPDF(*inputFile)
	if err != nil {
		fmt.Printf("Failed to open document: %v\n", err)
		os.Exit(1)
	}
	totalPages := doc.PageCount()
	doc.Close()

	fmt.Printf("Pages: %d\n", totalPages)
	fmt.Printf("DPI: %.0f\n", *dpi)
	fmt.Printf("Output directory: %s\n\n", *outputDir)

	// Create output directory
	os.MkdirAll(*outputDir, 0755)

	// If no pages specified, process first 3 pages
	if pages == nil {
		for i := 0; i < totalPages && i < 3; i++ {
			pages = append(pages, i)
		}
	}

	for _, pageNum := range pages {
		if pageNum >= totalPages {
			fmt.Printf("Warning: Page %d out of range, skipping\n", pageNum)
			continue
		}

		fmt.Printf("Processing page %d...\n", pageNum+1)

		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  PANIC recovered: %v\n", r)
			}
		}()

		outputPath := filepath.Join(*outputDir, fmt.Sprintf("page_%03d_vis.png", pageNum+1))
		err := processPage(*inputFile, pageNum, *dpi, outputPath)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		fmt.Printf("  Saved: %s\n", outputPath)
	}

	fmt.Printf("\nDone! Visualizations saved to: %s/\n", *outputDir)
}

func processPage(pdfPath string, pageNum int, dpi float64, outputPath string) error {
	// Re-open doc for each page to avoid state corruption
	doc, err := pdf.OpenPDF(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open: %v", err)
	}
	defer doc.Close()

	// Get page info (includes OCR, images, SVGs)
	pageResult, err := pdf.GetPageInfo(doc, pageNum, dpi, false, "")
	if err != nil {
		return fmt.Errorf("failed to get page info: %v", err)
	}

	// Get page for rendering
	page, err := doc.GetPage(pageNum)
	if err != nil {
		return fmt.Errorf("failed to load page: %v", err)
	}
	defer page.Close()

	// Render page to image
	img, width, height, err := page.GetPixmap(dpi)
	if err != nil || img == nil {
		return fmt.Errorf("failed to render: %v", err)
	}

	// Convert to image.RGBA
	original := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			original.Set(x, y, img.At(x, y))
		}
	}

	// Create annotated copy
	annotated := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Copy(annotated, image.Point{}, original, original.Bounds(), draw.Src, nil)

	// Draw image positions (red rectangles)
	for _, pos := range pageResult.PosImgs {
		drawRectOutline(annotated, pos.X, pos.Y, pos.Width, pos.Height, color.RGBA{255, 0, 0, 200}, 3)
	}

	// Draw SVG positions (blue rectangles)
	for _, pos := range pageResult.PosSVGs {
		drawRectOutline(annotated, pos.X, pos.Y, pos.Width, pos.Height, color.RGBA{0, 0, 255, 200}, 3)
	}

	// Draw OCR text positions (green rectangles)
	for _, span := range pageResult.OCRInfo {
		if len(span.Poly) >= 8 {
			x := int(span.Poly[0])
			y := int(span.Poly[1])
			x2 := int(span.Poly[2])
			y2 := int(span.Poly[7])
			w := x2 - x
			h := y2 - y
			if w > 0 && h > 0 {
				drawRectOutline(annotated, x, y, w, h, color.RGBA{0, 200, 0, 150}, 1)
			}
		}
	}

	// Draw merged figures (yellow rectangles)
	if len(pageResult.PosFigures) > 0 && len(pageResult.PosFigures) < 20 {
		for _, pos := range pageResult.PosFigures {
			drawRectOutline(annotated, pos.X, pos.Y, pos.Width, pos.Height, color.RGBA{255, 255, 0, 180}, 2)
		}
	}

	// Create side-by-side: original | annotated
	gap := 40
	combinedWidth := width*2 + gap
	combined := image.NewRGBA(image.Rect(0, 0, combinedWidth, height))

	// Fill background gray
	drawRectFill(combined, 0, 0, combinedWidth, height, color.RGBA{220, 220, 220, 255})

	// Copy original to left
	draw.Copy(combined, image.Point{}, original, original.Bounds(), draw.Src, nil)

	// Copy annotated to right
	draw.Copy(combined, image.Point{X: width + gap}, annotated, annotated.Bounds(), draw.Src, nil)

	// Draw labels on left (original)
	drawRectFill(combined, 10, 10, 60, 28, color.RGBA{0, 0, 0, 200})
	drawString(combined, 18, 28, "Original", color.RGBA{255, 255, 255, 255})

	// Draw labels on right (annotated)
	drawRectFill(combined, width+gap+10, 10, 120, 28, color.RGBA{0, 0, 0, 200})
	drawString(combined, width+gap+18, 28, "Annotated", color.RGBA{255, 255, 255, 255})

	// Draw legend at bottom
	legendY := height - 80
	legendHeight := 70

	// Legend background
	drawRectFill(combined, 0, legendY, combinedWidth, legendHeight, color.RGBA{245, 245, 245, 255})

	// Separator line
	for x := 0; x < combinedWidth; x++ {
		if legendY > 0 && legendY < combined.Bounds().Dy() {
			combined.Set(x, legendY, color.RGBA{180, 180, 180, 255})
		}
	}

	// Legend items
	legendStartX := 30
	legendYPos := legendY + 25

	// Red - Images
	drawRectFill(combined, legendStartX, legendYPos-10, 20, 16, color.RGBA{255, 0, 0, 200})
	drawString(combined, legendStartX+28, legendYPos, fmt.Sprintf("Images: %d", len(pageResult.PosImgs)), color.RGBA{0, 0, 0, 255})

	// Blue - SVGs
	legendStartX += 150
	drawRectFill(combined, legendStartX, legendYPos-10, 20, 16, color.RGBA{0, 0, 255, 200})
	drawString(combined, legendStartX+28, legendYPos, fmt.Sprintf("SVGs: %d", len(pageResult.PosSVGs)), color.RGBA{0, 0, 0, 255})

	// Green - OCR
	legendStartX += 130
	drawRectFill(combined, legendStartX, legendYPos-10, 20, 16, color.RGBA{0, 200, 0, 150})
	drawString(combined, legendStartX+28, legendYPos, fmt.Sprintf("Text: %d", len(pageResult.OCRInfo)), color.RGBA{0, 0, 0, 255})

	// Yellow - Figures
	legendStartX += 130
	drawRectFill(combined, legendStartX, legendYPos-10, 20, 16, color.RGBA{255, 255, 0, 180})
	drawString(combined, legendStartX+28, legendYPos, fmt.Sprintf("Figures: %d", len(pageResult.PosFigures)), color.RGBA{0, 0, 0, 255})

	// Page number - right side
	pageStr := fmt.Sprintf("Page %d", pageNum+1)
	pageStrX := combinedWidth - 120
	drawString(combined, pageStrX, legendYPos, pageStr, color.RGBA{100, 100, 100, 255})

	// Save
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output: %v", err)
	}
	defer f.Close()

	return png.Encode(f, combined)
}

func drawRectOutline(img *image.RGBA, x, y, w, h int, col color.RGBA, thickness int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > img.Bounds().Dx() {
		w = img.Bounds().Dx() - x
	}
	if y+h > img.Bounds().Dy() {
		h = img.Bounds().Dy() - y
	}
	if w <= 0 || h <= 0 {
		return
	}

	for t := 0; t < thickness; t++ {
		// Top
		for px := x; px < x+w && px < img.Bounds().Dx(); px++ {
			if y+t < img.Bounds().Dy() {
				img.Set(px, y+t, col)
			}
		}
		// Bottom
		for px := x; px < x+w && px < img.Bounds().Dx(); px++ {
			if y+h-1-t < img.Bounds().Dy() && y+h-1-t >= 0 {
				img.Set(px, y+h-1-t, col)
			}
		}
		// Left
		for py := y; py < y+h && py < img.Bounds().Dy(); py++ {
			if x+t < img.Bounds().Dx() {
				img.Set(x+t, py, col)
			}
		}
		// Right
		for py := y; py < y+h && py < img.Bounds().Dy(); py++ {
			if x+w-1-t < img.Bounds().Dx() && x+w-1-t >= 0 {
				img.Set(x+w-1-t, py, col)
			}
		}
	}
}

func drawRectFill(img *image.RGBA, x, y, w, h int, col color.RGBA) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > img.Bounds().Dx() {
		w = img.Bounds().Dx() - x
	}
	if y+h > img.Bounds().Dy() {
		h = img.Bounds().Dy() - y
	}
	if w <= 0 || h <= 0 {
		return
	}

	for px := x; px < x+w; px++ {
		for py := y; py < y+h; py++ {
			img.Set(px, py, col)
		}
	}
}