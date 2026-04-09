package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
	inputFile := flag.String("i", "", "Input PDF file")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: extract_benchmark -i <input.pdf>")
		flag.PrintDefaults()
		return
	}

	fmt.Printf("=== GoMuPDF Extraction Benchmark ===\n")
	fmt.Printf("File: %s\n\n", *inputFile)

	// Open document
	fmt.Println("Test 1: Open document")
	start := time.Now()
	doc, err := fitz.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	openTime := time.Since(start)
	fmt.Printf("  Time: %v\n", openTime)
	fmt.Printf("  Pages: %d\n", doc.PageCount())

	// Load first page
	fmt.Println("\nTest 2: Load first page")
	start = time.Now()
	page, err := doc.Page(0)
	loadTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}
	fmt.Printf("  Time: %v\n", loadTime)

	// Get page info
	fmt.Println("\nTest 3: Get page rect")
	start = time.Now()
	rect := page.Rect()
	rectTime := time.Since(start)
	fmt.Printf("  Time: %v\n", rectTime)
	fmt.Printf("  Rect: %v\n", rect)

	// Render page (for comparison with text extraction)
	fmt.Println("\nTest 4: Render page to pixmap")
	start = time.Now()
	matrix := fitz.Matrix{A: 1, D: 1}
	pix, err := page.Pixmap(matrix, false)
	renderTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to render page: %v", err)
	}
	fmt.Printf("  Time: %v\n", renderTime)
	fmt.Printf("  Pixmap: %dx%d\n", pix.Width(), pix.Height())
	pix.Close()

	// Get images
	fmt.Println("\nTest 5: Get images")
	start = time.Now()
	// Note: GetImages not yet implemented in fitz package
	imgTime := time.Since(start)
	fmt.Printf("  Time: %v\n", imgTime)

	// Get text (placeholder - not yet fully implemented)
	fmt.Println("\nTest 6: Get text")
	start = time.Now()
	text, err := page.GetText(nil)
	textTime := time.Since(start)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Time: %v\n", textTime)
		fmt.Printf("  Text length: %d\n", len(text))
	}

	// Summary
	totalTime := openTime + loadTime + rectTime + renderTime + imgTime + textTime
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Open: %v\n", openTime)
	fmt.Printf("Load page: %v\n", loadTime)
	fmt.Printf("Get rect: %v\n", rectTime)
	fmt.Printf("Render: %v\n", renderTime)
	fmt.Printf("Images: %v\n", imgTime)
	fmt.Printf("Get text: %v\n", textTime)
	fmt.Printf("Total: %v\n", totalTime)

	doc.Close()
}