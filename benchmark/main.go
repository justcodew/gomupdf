package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: benchmark <input.pdf>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	fmt.Printf("=== GoMuPDF Benchmark ===\n")
	fmt.Printf("File: %s\n\n", inputFile)

	// Test 1: Open document
	fmt.Println("Test 1: Open document")
	start := time.Now()
	doc, err := fitz.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	openTime := time.Since(start)
	fmt.Printf("  Time: %v\n", openTime)
	fmt.Printf("  Pages: %d\n", doc.PageCount())

	// Test 2: Get metadata
	fmt.Println("\nTest 2: Get metadata")
	start = time.Now()
	metadata := doc.Metadata()
	metadataTime := time.Since(start)
	fmt.Printf("  Time: %v\n", metadataTime)
	if len(metadata) > 0 {
		fmt.Printf("  Keys: %d\n", len(metadata))
	}

	// Test 3: Get page count
	fmt.Println("\nTest 3: Get page count")
	start = time.Now()
	pageCount := doc.PageCount()
	pageCountTime := time.Since(start)
	fmt.Printf("  Time: %v\n", pageCountTime)

	// Test 4: Load first page
	fmt.Println("\nTest 4: Load first page")
	start = time.Now()
	page, err := doc.Page(0)
	loadPageTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}
	fmt.Printf("  Time: %v\n", loadPageTime)

	// Test 5: Get page rect
	fmt.Println("\nTest 5: Get page rect")
	start = time.Now()
	rect := page.Rect()
	rectTime := time.Since(start)
	fmt.Printf("  Time: %v\n", rectTime)
	fmt.Printf("  Rect: %v\n", rect)

	// Test 6: Render page to pixmap (1x zoom)
	fmt.Println("\nTest 6: Render page to pixmap (1x zoom)")
	start = time.Now()
	matrix := fitz.Identity()
	pix, err := page.Pixmap(matrix, false)
	renderTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to render page: %v", err)
	}
	fmt.Printf("  Time: %v\n", renderTime)
	fmt.Printf("  Pixmap: %v\n", pix)

	// Test 7: Get pixmap size
	fmt.Println("\nTest 7: Get pixmap properties")
	start = time.Now()
	width := pix.Width()
	height := pix.Height()
	n := pix.N()
	propsTime := time.Since(start)
	fmt.Printf("  Time: %v\n", propsTime)
	fmt.Printf("  Size: %dx%d, channels: %d\n", width, height, n)

	// Test 8: Save pixmap
	fmt.Println("\nTest 8: Save pixmap as PNG")
	start = time.Now()
	err = pix.Save("benchmark_output.png")
	saveTime := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to save pixmap: %v", err)
	}
	fmt.Printf("  Time: %v\n", saveTime)

	// Clean up
	pix.Close()
	page.Close()
	doc.Close()

	// Summary
	fmt.Println("\n=== Summary ===")
	totalTime := openTime + metadataTime + pageCountTime + loadPageTime + rectTime + renderTime + propsTime + saveTime
	fmt.Printf("Total time (excl. cleanup): %v\n", totalTime)

	// Output JSON for comparison
	result := map[string]interface{}{
		"tool":           "gomupdf",
		"file":           inputFile,
		"pages":          pageCount,
		"open_time_ms":   openTime.Milliseconds(),
		"metadata_time_ms": metadataTime.Milliseconds(),
		"page_count_time_ms": pageCountTime.Milliseconds(),
		"load_page_time_ms": loadPageTime.Milliseconds(),
		"rect_time_ms":   rectTime.Milliseconds(),
		"render_time_ms": renderTime.Milliseconds(),
		"props_time_ms":  propsTime.Milliseconds(),
		"save_time_ms":   saveTime.Milliseconds(),
		"total_time_ms":  totalTime.Milliseconds(),
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("\n=== JSON Output ===")
	fmt.Println(string(jsonData))

	// Save JSON to file for comparison
	jsonFile, err := os.Create("benchmark_go_result.json")
	if err != nil {
		log.Printf("Failed to create JSON file: %v", err)
	} else {
		jsonFile.Write(jsonData)
		jsonFile.Close()
		fmt.Println("\nResult saved to: benchmark_go_result.json")
	}
}
