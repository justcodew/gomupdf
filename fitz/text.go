// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（text.go）封装了文本提取功能，支持纯文本、HTML、XML、XHTML、JSON 等多种输出格式。
package fitz

import (
	"fmt"
	"unicode/utf8"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// TextBlock 表示一个文本块，包含边界框、文本内容及其下属的行和片段信息。
type TextBlock struct {
	Bbox   Rect       // 文本块的边界矩形
	Text   string     // 文本内容
	Number int        // 块编号
	Type   int        // 块类型（0=文本，1=图像）
	Lines  []TextLine // 文本行列表
}

// TextLine 表示文本块中的一行文字。
type TextLine struct {
	Bbox  Rect           // 行的边界矩形
	WDir  Point          // 书写方向向量
	Frags []TextFragment // 文本片段列表
}

// TextFragment 表示一个文本片段（通常对应单个字符）。
type TextFragment struct {
	Bbox   Rect    // 片段的边界矩形
	Origin Point   // 字符原点（基线位置）
	Font   string  // 字体名称
	Size   float64 // 字号
	Color  Color   // 文本颜色（RGBA）
	Text   string  // 文本内容
}

// TextWord 表示一个单词及其边界框。
type TextWord struct {
	Bbox     Rect    // 单词的边界矩形
	Origin   Point   // 单词原点
	Word     string  // 单词文本
	FontName string  // 字体名称
	FontSize float64 // 字号
	Color    Color   // 文本颜色
	BlockNum int     // 所属块编号
	LineNum  int     // 所属行编号
	WordNum  int     // 单词编号
}

// Color 表示一个 RGBA 颜色值。
type Color struct {
	R, G, B, A float64 // 红、绿、蓝、透明度分量（0.0~1.0）
}

// TextOptions 包含文本提取的选项参数。
type TextOptions struct {
	// Flags 文本提取标志位（参见 MuPDF 的 FzStextOptions）
	Flags int
}

// GetTextBlocks 返回页面的文本块列表，包含每个块的边界框和文本内容。
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

		if blockType == 0 { // 文本块
			// 从文本块中提取行
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

				// 从行中提取字符
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

			// 从片段中构建文本内容
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

// GetText 返回页面的纯文本内容。
func (p *Page) GetText(options *TextOptions) (string, error) {
	if p.page == nil {
		return "", fmt.Errorf("page is nil")
	}

	// 从页面创建结构化文本页
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return "", fmt.Errorf("failed to create text page: %w", err)
	}
	defer stextPage.Destroy()

	return stextPage.Text(), nil
}

// GetTextHTML 返回页面的 HTML 格式文本内容。
func (p *Page) GetTextHTML() (string, error) {
	if p.page == nil {
		return "", fmt.Errorf("page is nil")
	}
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return "", err
	}
	defer stextPage.Destroy()
	return stextPage.HTML(), nil
}

// GetTextXML 返回页面的 XML 格式文本内容。
func (p *Page) GetTextXML() (string, error) {
	if p.page == nil {
		return "", fmt.Errorf("page is nil")
	}
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return "", err
	}
	defer stextPage.Destroy()
	return stextPage.XML(), nil
}

// GetTextXHTML 返回页面的 XHTML 格式文本内容。
func (p *Page) GetTextXHTML() (string, error) {
	if p.page == nil {
		return "", fmt.Errorf("page is nil")
	}
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return "", err
	}
	defer stextPage.Destroy()
	return stextPage.XHTML(), nil
}

// GetTextJSON 返回页面的 JSON 格式文本内容。
func (p *Page) GetTextJSON() (string, error) {
	if p.page == nil {
		return "", fmt.Errorf("page is nil")
	}
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return "", err
	}
	defer stextPage.Destroy()
	return stextPage.JSON(), nil
}

// GetTextWords 返回页面的单词列表及其边界框。
func (p *Page) GetTextWords() ([]TextWord, error) {
	blocks, err := p.GetTextBlocks()
	if err != nil {
		return nil, err
	}

	var words []TextWord
	for bi, block := range blocks {
		if block.Type != 0 { // 跳过图像块
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

// GetTextDict 返回带有坐标的结构化文本提取结果，以字典形式表示。
// 返回的字典中 "blocks" 键包含文本和图像块列表。
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

// Close 释放页面相关资源。
func (p *Page) Close() error {
	// TODO: 待实现
	return nil
}