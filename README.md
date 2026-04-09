# GoMuPDF - Go binding for MuPDF

A Go library for PDF manipulation, built on top of the [MuPDF](https://mupdf.com/) C library.

**Status: Early Development**

This is an initial implementation. The CGO bindings and core types are in place.

## Requirements

- Go 1.21+
- MuPDF C library (v1.27.2)
- GCC or Clang compiler

## Installing MuPDF

### macOS (using Homebrew)
```bash
brew install go
brew install mupdf
```

### Linux (using package manager)
```bash
# Ubuntu/Debian
sudo apt-get install libmupdf-dev golang

# Fedora
sudo dnf install mupdf-devel golang
```

### From Source
```bash
git clone https://github.com/ArtifexSoftware/mupdf.git
cd mupdf
make
sudo make install
```

## Building

```bash
cd gomupdf
go build ./...
```

## Example Usage

### Get comprehensive PDF info (text and image coordinates)

```go
package main

import (
    "fmt"
    "log"
    "github.com/go-pymupdf/gomupdf/pdf"
)

func main() {
    info, pngs, err := pdf.GetPDFInfo("document.pdf", 72.0, false)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Pages: %d\n", info.TotalPage)
    fmt.Printf("Is Scanned: %v\n", info.IsScanned)

    for pageIdx := range info.OCRPerPage {
        fmt.Printf("Page %d: %d OCR spans, %d images\n",
            pageIdx, len(info.OCRPerPage[pageIdx]), len(info.ImgsPerPage[pageIdx]))
    }
}
```

### Render a PDF page to image

```go
package main

import (
    "log"
    "github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
    doc, err := fitz.Open("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer doc.Close()

    page, err := doc.Page(0)
    if err != nil {
        log.Fatal(err)
    }
    defer page.Close()

    pix, err := page.Pixmap(fitz.Identity(), false)
    if err != nil {
        log.Fatal(err)
    }
    defer pix.Close()

    if err := pix.Save("output.png"); err != nil {
        log.Fatal(err)
    }
}
```

## Project Structure

```
gomupdf/
├── cgo_bindings/    # Low-level CGO bindings to MuPDF
│   ├── bindings.h   # C header
│   ├── bindings.c   # C implementation
│   ├── context.go   # Context management
│   ├── document.go  # Document operations
│   ├── page.go      # Page operations
│   └── pixmap.go    # Pixmap operations
├── fitz/            # Go-native API layer (basic)
│   ├── doc.go       # Document type
│   ├── page.go     # Page type
│   ├── pixmap.go   # Pixmap type
│   ├── rect.go     # Rect type
│   ├── point.go    # Point type
│   ├── matrix.go   # Matrix type
│   ├── quad.go     # Quad type
│   ├── text.go     # Text types
│   └── image.go    # Image types
├── pdf/             # Complete PDF utilities (mirrors pdf_utils.py)
│   ├── api.go      # Main API
│   ├── structs.go  # Data structures
│   ├── text.go     # Text processing & garbled detection
│   ├── position.go # Coordinate rotation
│   ├── image.go    # Image processing
│   └── merge_split.go # PDF split/merge
└── examples/        # Example programs
    ├── render/     # Page rendering
    ├── extract/    # Text/image extraction
    ├── pdfinfo/    # Comprehensive PDF info
    └── modify/      # PDF modification
```

## Features

### Implemented

**Core (fitz package):**
- Document opening (file and stream)
- Page enumeration
- Basic page rendering to pixmap
- Pixmap saving (PNG, JPEG)
- Basic geometric types (Rect, Point, Matrix, Quad)
- Document metadata

**Complete PDF Utilities (pdf package):**
- Text extraction with coordinates (OCR-style)
- Image position extraction
- SVG/drawing position extraction
- Coordinate rotation for different page orientations
- Garbled text detection
- Font corruption detection
- Pure white/black image detection
- PDF split and merge
- Image processing (resize, rotate)

### Not Yet Fully Implemented
- Full text extraction with all style information
- Annotation manipulation
- TOC/Outline operations
- Form fields (widgets)
- Embedded files
- Search functionality

## License

This project is licensed under the same terms as MuPDF (AGPL v3 or commercial license).
For commercial licensing, contact Artifex Software.
