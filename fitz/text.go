package fitz

import (
	"fmt"
	"unicode/utf8"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// TextBlock represents a block of text with its bounding box.
type TextBlock struct {
	Bbox   Rect   // Bounding box of the text
	Text   string // The text content
	Number int    // Block number
	Type   int    // Block type (0=text, 1=image)
	Lines  []TextLine
}

// TextLine represents a line of text within a block.
type TextLine struct {
	Bbox  Rect
	WDir  Point // Writing direction
	Frags []TextFragment
}

// TextFragment represents a text fragment (typically a word or character).
type TextFragment struct {
	Bbox   Rect
	Origin Point
	Font   string
	Size   float64
	Color  Color // Text color as RGBA
	Text   string
}

// TextWord represents a word with its bounding box.
type TextWord struct {
	Bbox     Rect
	Origin   Point
	Word     string
	FontName string
	FontSize float64
	Color    Color
	BlockNum int
	LineNum  int
	WordNum  int
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A float64
}

// TextOptions contains options for text extraction.
type TextOptions struct {
	// Flags for text extraction (see MuPDF FzStextOptions)
	Flags int
}

// GetTextBlocks returns text blocks with their bounding boxes.
func (p *Page) GetTextBlocks() ([]TextBlock, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	// Create structured text page from the page
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, fmt.Errorf("failed to create text page: %w", err)
	}
	defer stextPage.Destroy()

	blockCount := stextPage.BlockCount()
	blocks := make([]TextBlock, 0, blockCount)

	for i := 0; i < blockCount; i++ {
		block := stextPage.GetBlock(i)
		if block == nil {
			continue
		}

		blockType := stextPage.BlockType(block)

		bx0, by0, bx1, by1 := stextPage.BlockBbox(block)

		textBlock := TextBlock{
			Bbox:   Rect{X0: bx0, Y0: by0, X1: bx1, Y1: by1},
			Number: i,
			Type:   blockType,
		}

		if blockType == 0 { // Text block
			// Extract lines from text block
			lineCount := stextPage.LineCount(block)
			textBlock.Lines = make([]TextLine, 0, lineCount)

			for j := 0; j < lineCount; j++ {
				line := stextPage.GetLine(block, j)
				if line == nil {
					continue
				}

				lx0, ly0, lx1, ly1 := stextPage.LineBbox(line)
				dx, dy := stextPage.LineDir(line)

				textLine := TextLine{
					Bbox: Rect{X0: lx0, Y0: ly0, X1: lx1, Y1: ly1},
					WDir: Point{X: dx, Y: dy},
				}

				// Extract characters from line
				var chars []TextFragment
				ch := stextPage.FirstChar(line)
				for ch != nil {
					c := stextPage.CharC(ch)
					if c > 0 {
						cx0, cy0, cx1, cy1 := stextPage.CharBbox(ch)
						ox, oy := stextPage.CharOrigin(ch)
						size := stextPage.CharSize(ch)

						// Convert Unicode code point to UTF-8 string
					 buf := make([]byte, utf8.UTFMax)
					 n := utf8.EncodeRune(buf, rune(c))
					 text := string(buf[:n])

					 frag := TextFragment{
						 Bbox:   Rect{X0: cx0, Y0: cy0, X1: cx1, Y1: cy1},
						 Origin: Point{X: ox, Y: oy},
						 Size:   size,
						 Text:   text,
					 }
					 chars = append(chars, frag)
					}
				 ch = stextPage.NextChar(ch)
				}

				textLine.Frags = chars
				textBlock.Lines = append(textBlock.Lines, textLine)
			}

			// Build text content from fragments
			var textContent string
			for _, line := range textBlock.Lines {
				for _, frag := range line.Frags {
					textContent += frag.Text
				}
			}
			textBlock.Text = textContent
		}

		blocks = append(blocks, textBlock)
	}

	return blocks, nil
}

// GetText returns the text content of the page.
func (p *Page) GetText(options *TextOptions) (string, error) {
	if p.page == nil {
		return "", fmt.Errorf("page is nil")
	}

	// Create structured text page from the page
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return "", fmt.Errorf("failed to create text page: %w", err)
	}
	defer stextPage.Destroy()

	return stextPage.Text(), nil
}

// GetTextWords returns words with their bounding boxes.
func (p *Page) GetTextWords() ([]TextWord, error) {
	blocks, err := p.GetTextBlocks()
	if err != nil {
		return nil, err
	}

	var words []TextWord
	for bi, block := range blocks {
		if block.Type != 0 { // Skip image blocks
			continue
		}
		for li, line := range block.Lines {
			for _, frag := range line.Frags {
				word := TextWord{
					Bbox:     frag.Bbox,
					Origin:   frag.Origin,
					Word:     frag.Text,
					FontSize: frag.Size,
					Color:    frag.Color,
					BlockNum: bi,
					LineNum:  li,
				}
				words = append(words, word)
			}
		}
	}

	return words, nil
}

// GetTextDict returns structured text extraction with coordinates.
// Returns a dictionary with "blocks" containing text and image blocks.
func (p *Page) GetTextDict() (map[string]interface{}, error) {
	blocks, err := p.GetTextBlocks()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"blocks": blocks,
	}
	return result, nil
}

// Close releases the page resources.
func (p *Page) Close() error {
	// TODO: implement
	return nil
}