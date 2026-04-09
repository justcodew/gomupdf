// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（image.go）封装了从 PDF 页面中提取图像信息及像素数据的功能。
package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// ImageInfo 表示 PDF 页面中图像的元数据信息。
type ImageInfo struct {
	Number     int    // 图像编号
	Bbox       Rect   // 图像在页面上的边界矩形
	Matrix     Matrix // 图像变换矩阵
	Width      int    // 图像宽度（像素）
	Height     int    // 图像高度（像素）
	Colorspace string // 颜色空间名称
	N          int    // 颜色分量数
	Xres       int    // 水平分辨率（DPI）
	Yres       int    // 垂直分辨率（DPI）
	BPC        int    // 每个颜色分量的位数
	Size       int    // 图像数据大小（字节）
	HasMask    bool   // 是否包含遮罩
	Digest     []byte // 图像数据摘要
}

// Image 表示从 PDF 中提取的图像，包含像素数据。
type Image struct {
	Width      int    // 图像宽度（像素）
	Height     int    // 图像高度（像素）
	N          int    // 颜色分量数
	BPC        int    // 每个颜色分量的位数
	Colorspace string // 颜色空间名称
	Samples    []byte // 原始像素数据
}

// GetImages 返回页面上所有图像的信息列表。
func (p *Page) GetImages() ([]ImageInfo, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	// 使用结构化文本页面查找图像块
	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, err
	}
	defer stextPage.Destroy()

	blockCount := stextPage.BlockCount()
	var images []ImageInfo

	for i := 0; i < blockCount; i++ {
		block := stextPage.GetBlock(i)
		if block == nil {
			continue
		}
		if stextPage.BlockType(block) != 1 { // 跳过非图像块
			continue
		}

		bx0, by0, bx1, by1 := stextPage.BlockBbox(block)
		images = append(images, ImageInfo{
			Number: i,
			Bbox:   Rect{X0: bx0, Y0: by0, X1: bx1, Y1: by1},
		})
	}

	return images, nil
}

// ExtractImage 从指定文本块中提取图像的像素数据。
func (p *Page) ExtractImage(blockIdx int) (*Image, error) {
	if p.page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	stextPage, err := cgo_bindings.NewTextPage(p.page)
	if err != nil {
		return nil, err
	}
	defer stextPage.Destroy()

	block := stextPage.GetBlock(blockIdx)
	if block == nil {
		return nil, fmt.Errorf("block %d not found", blockIdx)
	}
	if stextPage.BlockType(block) != 1 {
		return nil, fmt.Errorf("block %d is not an image block", blockIdx)
	}

	// 从块中获取图像并渲染为像素图
	img := cgo_bindings.BlockGetImage(p.ctx, block)
	if img == nil {
		return nil, fmt.Errorf("failed to get image from block")
	}

	w := cgo_bindings.ImageWidth(p.ctx, img)
	h := cgo_bindings.ImageHeight(p.ctx, img)
	n := cgo_bindings.ImageN(p.ctx, img)
	bpc := cgo_bindings.ImageBPC(p.ctx, img)
	csName := cgo_bindings.ImageColorspaceName(p.ctx, img)

	pix, err := cgo_bindings.NewPixmapFromImage(p.ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to extract image pixels: %w", err)
	}
	defer pix.Destroy()

	return &Image{
		Width:      w,
		Height:     h,
		N:          n,
		BPC:        bpc,
		Colorspace: csName,
		Samples:    pix.Samples(),
	}, nil
}

// GetImageXObjects 返回页面中 XObject 图像的列表（尚未实现）。
func (p *Page) GetImageXObjects() ([]ImageInfo, error) {
	// TODO: 实现 XObject 枚举
	return nil, fmt.Errorf("not yet implemented")
}
