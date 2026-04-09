package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-pymupdf/gomupdf/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pdfinfo <input.pdf> [dpi]")
		fmt.Println("  Extracts comprehensive PDF information")
		fmt.Println("  dpi: optional DPI for rendering (default: 72)")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	dpi := 72.0
	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%f", &dpi)
	}

	fmt.Printf("Opening: %s\n", inputFile)
	fmt.Printf("DPI: %.0f\n", dpi)

	// Get PDF info
	info, pngs, err := pdf.GetPDFInfo(inputFile, dpi, false)
	if err != nil {
		log.Fatalf("Failed to get PDF info: %v", err)
	}

	// Print summary
	fmt.Printf("\n=== PDF Information ===\n")
	fmt.Printf("Total Pages: %d\n", info.TotalPage)
	fmt.Printf("Is Encrypted: %v\n", info.IsEncrypted)
	fmt.Printf("Needs Password: %v\n", info.IsNeedsPassword)
	fmt.Printf("Is Resolvable: %v\n", info.IsResolvable)
	fmt.Printf("Is Scanned: %v\n", info.IsScanned)

	// Print page details
	for pageIdx := 0; pageIdx < info.TotalPage; pageIdx++ {
		fmt.Printf("\n=== Page %d ===\n", pageIdx+1)

		// Image positions
		if pageIdx < len(info.ImgsPerPage) && len(info.ImgsPerPage[pageIdx]) > 0 {
			fmt.Printf("Images: %d\n", len(info.ImgsPerPage[pageIdx]))
			for i, pos := range info.ImgsPerPage[pageIdx] {
				if i < 5 {
					fmt.Printf("  Image %d: x=%d y=%d w=%d h=%d\n",
						i, pos.X, pos.Y, pos.Width, pos.Height)
				}
			}
			if len(info.ImgsPerPage[pageIdx]) > 5 {
				fmt.Printf("  ... and %d more\n", len(info.ImgsPerPage[pageIdx])-5)
			}
		}

		// SVG positions
		if pageIdx < len(info.SVGsPerPage) && len(info.SVGsPerPage[pageIdx]) > 0 {
			fmt.Printf("SVGs: %d\n", len(info.SVGsPerPage[pageIdx]))
		}

		// OCR info
		if pageIdx < len(info.OCRPerPage) && len(info.OCRPerPage[pageIdx]) > 0 {
			fmt.Printf("OCR spans: %d\n", len(info.OCRPerPage[pageIdx]))
			for i, span := range info.OCRPerPage[pageIdx] {
				if i < 5 {
					text := span.Text
					if len(text) > 30 {
						text = text[:30] + "..."
					}
					fmt.Printf("  Span %d: %q bbox=%v font=%s size=%.1f\n",
						i, text, span.Poly, span.Font, span.Size)
				}
			}
			if len(info.OCRPerPage[pageIdx]) > 5 {
				fmt.Printf("  ... and %d more\n", len(info.OCRPerPage[pageIdx])-5)
			}
		}

		// OCR chars
		if pageIdx < len(info.OCRCharsPerPage) && len(info.OCRCharsPerPage[pageIdx]) > 0 {
			fmt.Printf("OCR chars: %d\n", len(info.OCRCharsPerPage[pageIdx]))
		}

		// PNG
		if pageIdx < len(pngs) && len(pngs[pageIdx]) > 0 {
			fmt.Printf("PNG size: %d bytes\n", len(pngs[pageIdx]))
		}
	}

	// Output JSON
	fmt.Printf("\n=== JSON Output ===\n")
	jsonData, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(jsonData))
}
