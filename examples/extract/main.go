package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: extract <input.pdf> [page_num]")
		fmt.Println("  Extracts text and image coordinates from input.pdf")
		fmt.Println("  page_num is optional (default: 0)")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	pageNum := 0
	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%d", &pageNum)
	}

	// Open the document
	doc, err := fitz.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open document: %v", err)
	}
	defer doc.Close()

	fmt.Printf("Document: %s\n", doc)
	fmt.Printf("Pages: %d\n", doc.PageCount())

	if pageNum < 0 || pageNum >= doc.PageCount() {
		log.Fatalf("Page number out of range: %d", pageNum)
	}

	// Load the specified page
	page, err := doc.Page(pageNum)
	if err != nil {
		log.Fatalf("Failed to load page %d: %v", pageNum, err)
	}
	defer page.Close()

	fmt.Printf("\n=== Page %d ===\n", pageNum)
	fmt.Printf("Rect: %v\n", page.Rect())

	// Extract text blocks with coordinates
	fmt.Println("\n--- Text Blocks ---")
	blocks, err := page.GetTextBlocks()
	if err != nil {
		log.Printf("Failed to get text blocks: %v", err)
	} else if len(blocks) == 0 {
		fmt.Println("(no text blocks found)")
	} else {
		for i, block := range blocks {
			fmt.Printf("Block %d: bbox=%v text=%q\n", i, block.Bbox, truncate(block.Text, 50))
		}
	}

	// Extract words with coordinates
	fmt.Println("\n--- Text Words ---")
	words, err := page.GetTextWords()
	if err != nil {
		log.Printf("Failed to get text words: %v", err)
	} else if len(words) == 0 {
		fmt.Println("(no words found)")
	} else {
		for i, word := range words {
			if i < 20 { // Print first 20 words
				fmt.Printf("Word: %q bbox=%v font=%s size=%.1f\n",
					word.Word, word.Bbox, word.FontName, word.FontSize)
			}
		}
		if len(words) > 20 {
			fmt.Printf("... and %d more words\n", len(words)-20)
		}
	}

	// Extract image information with coordinates
	fmt.Println("\n--- Images ---")
	images, err := page.GetImages()
	if err != nil {
		log.Printf("Failed to get images: %v", err)
	} else if len(images) == 0 {
		fmt.Println("(no images found)")
	} else {
		for i, img := range images {
			fmt.Printf("Image %d: bbox=%v size=%dx%d colorspace=%s\n",
				i, img.Bbox, img.Width, img.Height, img.Colorspace)
		}
	}

	// Output structured JSON
	fmt.Println("\n--- Structured JSON Output ---")
	data := map[string]interface{}{
		"page": pageNum,
		"rect": page.Rect(),
	}

	if blocks != nil {
		blockList := make([]map[string]interface{}, len(blocks))
		for i, b := range blocks {
			blockList[i] = map[string]interface{}{
				"bbox": b.Bbox,
				"text": b.Text,
				"type": b.Type,
			}
		}
		data["blocks"] = blockList
	}

	if words != nil {
		wordList := make([]map[string]interface{}, len(words))
		for i, w := range words {
			wordList[i] = map[string]interface{}{
				"word":     w.Word,
				"bbox":     w.Bbox,
				"font":     w.FontName,
				"size":     w.FontSize,
				"origin":   w.Origin,
			}
		}
		data["words"] = wordList
	}

	if images != nil {
		imgList := make([]map[string]interface{}, len(images))
		for i, img := range images {
			imgList[i] = map[string]interface{}{
				"bbox":       img.Bbox,
				"width":      img.Width,
				"height":     img.Height,
				"colorspace": img.Colorspace,
			}
		}
		data["images"] = imgList
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
