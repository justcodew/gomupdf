package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: modify <input.pdf> <output.pdf>")
		fmt.Println("  Opens input.pdf and saves it as output.pdf")
		fmt.Println("  (This is a placeholder for future modification features)")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Open the document
	doc, err := fitz.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	defer doc.Close()

	fmt.Printf("Opened document: %s\n", doc)
	fmt.Printf("Pages: %d\n", doc.PageCount())
	fmt.Printf("Is PDF: %v\n", doc.IsPDF())

	// For now, just copy the file
	// TODO: implement actual modifications
	if doc.IsPDF() {
		err = doc.Save(outputFile, nil)
		if err != nil {
			log.Fatalf("Failed to save document: %v", err)
		}
		fmt.Printf("Saved document to: %s\n", outputFile)
	} else {
		fmt.Println("Only PDF modification is supported")
	}
}
