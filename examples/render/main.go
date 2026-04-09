package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: render <input.pdf> [output.png]")
		fmt.Println("  Renders the first page of input.pdf to output.png (or page0.png by default)")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "page0.png"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	// Open the document
	doc, err := fitz.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	defer doc.Close()

	fmt.Printf("Opened document: %s\n", doc)
	fmt.Printf("Pages: %d\n", doc.PageCount())
	fmt.Printf("Is PDF: %v\n", doc.IsPDF())

	// Print metadata
	metadata := doc.Metadata()
	fmt.Println("\nMetadata:")
	for k, v := range metadata {
		if v != "" {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	// Check if document needs password
	if doc.NeedsPassword() {
		fmt.Println("\nDocument requires password")
	}

	// Load the first page
	if doc.PageCount() == 0 {
		log.Fatal("Document has no pages")
	}

	page, err := doc.Page(0)
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}
	defer page.Close()

	fmt.Printf("\nPage rect: %v\n", page.Rect())
	fmt.Printf("Page rotation: %d\n", page.Rotation())

	// Render the page to a pixmap
	matrix := fitz.Identity()
	matrix = fitz.NewMatrixScale(2.0) // 2x zoom

	pix, err := page.Pixmap(matrix, false)
	if err != nil {
		log.Fatalf("Failed to render page: %v", err)
	}
	defer pix.Close()

	fmt.Printf("Rendered pixmap: %v\n", pix)

	// Save the pixmap
	if err := pix.Save(outputFile); err != nil {
		log.Fatalf("Failed to save pixmap: %v", err)
	}

	fmt.Printf("Saved page to: %s\n", outputFile)
}
