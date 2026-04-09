// gomupdf 性能基准测试
//
// 测试项目（与 PyMuPDF 对齐）：
// 1.  文档打开与关闭
// 2.  元数据读取
// 3.  单页加载
// 4.  页面渲染 1x
// 5.  页面渲染 2x
// 6.  全页面渲染 1x
// 7.  文本提取
// 8.  文本搜索
// 9.  大纲读取
// 10. PDF 保存到内存
// 11. 页面插入 10 页
// 12. 页面删除 10 页
// 13. Pixmap→PNG
// 14. Pixmap→JPEG
// 15. 链接读取
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/go-pymupdf/gomupdf/fitz"
)

// BenchResult 单项测试结果
type BenchResult struct {
	MeanMs float64 `json:"mean_ms"`
	MinMs  float64 `json:"min_ms"`
	MaxMs  float64 `json:"max_ms"`
	StdMs  float64 `json:"std_ms,omitempty"`
}

func stats(durations []time.Duration) BenchResult {
	if len(durations) == 0 {
		return BenchResult{}
	}
	var sum float64
	vals := make([]float64, len(durations))
	for i, d := range durations {
		ms := float64(d.Nanoseconds()) / 1e6
		vals[i] = ms
		sum += ms
	}
	sort.Float64s(vals)
	mean := sum / float64(len(vals))

	var variance float64
	for _, v := range vals {
		diff := v - mean
		variance += diff * diff
	}
	std := 0.0
	if len(vals) > 1 {
		std = math.Sqrt(variance / float64(len(vals)-1))
	}

	return BenchResult{
		MeanMs: math.Round(mean*1000) / 1000,
		MinMs:  math.Round(vals[0]*1000) / 1000,
		MaxMs:  math.Round(vals[len(vals)-1]*1000) / 1000,
		StdMs:  math.Round(std*1000) / 1000,
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bench_gomupdf <input.pdf> [iterations]")
		os.Exit(1)
	}

	pdfPath := os.Args[1]
	iterations := 5
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &iterations)
	}

	fmt.Printf("=== gomupdf Benchmark ===\n")
	fmt.Printf("Input: %s\n", pdfPath)
	fmt.Printf("Iterations: %d\n\n", iterations)

	results := make(map[string]interface{})

	// ---- 1. 文档打开与关闭 ----
	var times []time.Duration
	var pageCount int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		doc, err := fitz.Open(pdfPath)
		if err != nil {
			fmt.Printf("Open failed: %v\n", err)
			os.Exit(1)
		}
		pageCount = doc.PageCount()
		doc.Close()
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
	}
	s := stats(times)
	results["open_close"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"pages": pageCount,
	}
	fmt.Printf("[1] Open/Close: mean=%.1fms (pages=%d)\n", s.MeanMs, pageCount)

	// ---- 2. 元数据读取 ----
	doc, err := fitz.Open(pdfPath)
	if err != nil {
		fmt.Printf("Open failed: %v\n", err)
		os.Exit(1)
	}
	times = times[:0]
	var metaKeys int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		meta := doc.Metadata()
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
		metaKeys = len(meta)
	}
	s = stats(times)
	results["metadata"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"keys": metaKeys,
	}
	fmt.Printf("[2] Metadata: mean=%.3fms\n", s.MeanMs)

	// ---- 3. 单页加载 ----
	times = times[:0]
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		page, err := doc.Page(0)
		if err != nil {
			fmt.Printf("Page load failed: %v\n", err)
			os.Exit(1)
		}
		_ = page.Rect()
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
		page.Close()
	}
	s = stats(times)
	results["load_page"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
	}
	fmt.Printf("[3] Load page 0: mean=%.3fms\n", s.MeanMs)

	// ---- 4. 页面渲染 1x ----
	page, _ := doc.Page(0)
	defer page.Close()
	mat1x := fitz.Identity()

	times = times[:0]
	var pixW, pixH int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		pix, err := page.Pixmap(mat1x, false)
		if err != nil {
			fmt.Printf("Render failed: %v\n", err)
			os.Exit(1)
		}
		pixW = pix.Width()
		pixH = pix.Height()
		pix.Close()
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
	}
	s = stats(times)
	results["render_1x"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"width": pixW, "height": pixH,
	}
	fmt.Printf("[4] Render 1x: mean=%.1fms (%dx%d)\n", s.MeanMs, pixW, pixH)

	// ---- 5. 页面渲染 2x ----
	mat2x := fitz.NewMatrixZoom(2, 2)
	times = times[:0]
	var pixW2, pixH2 int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		pix, err := page.Pixmap(mat2x, false)
		if err != nil {
			fmt.Printf("Render 2x failed: %v\n", err)
			os.Exit(1)
		}
		pixW2 = pix.Width()
		pixH2 = pix.Height()
		pix.Close()
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
	}
	s = stats(times)
	results["render_2x"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"width": pixW2, "height": pixH2,
	}
	fmt.Printf("[5] Render 2x: mean=%.1fms (%dx%d)\n", s.MeanMs, pixW2, pixH2)

	// ---- 6. 全页面渲染 1x ----
	t0 := time.Now()
	for i := 0; i < pageCount; i++ {
		p, err := doc.Page(i)
		if err != nil {
			continue
		}
		pix, err := p.Pixmap(mat1x, false)
		if err == nil {
			pix.Close()
		}
		p.Close()
	}
	t1 := time.Now()
	allMs := float64(t1.Sub(t0).Nanoseconds()) / 1e6
	results["render_all_1x"] = map[string]interface{}{
		"total_ms":     math.Round(allMs*10) / 10,
		"per_page_ms":  math.Round(allMs/float64(pageCount)*10) / 10,
		"pages":        pageCount,
	}
	fmt.Printf("[6] Render all 1x: %.0fms total (%.1fms/page)\n", allMs, allMs/float64(pageCount))

	// ---- 7. 文本提取 ----
	times = times[:0]
	var textLen int
	for i := 0; i < iterations; i++ {
		p, _ := doc.Page(0)
		t0 := time.Now()
		text, err := p.GetText(nil)
		t1 := time.Now()
		if err == nil {
			textLen = len(text)
		}
		times = append(times, t1.Sub(t0))
		p.Close()
	}
	s = stats(times)
	results["text_extract"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"text_length": textLen,
	}
	fmt.Printf("[7] Text extract: mean=%.3fms (len=%d)\n", s.MeanMs, textLen)

	// ---- 8. 文本搜索 ----
	times = times[:0]
	searchHits := 0
	for i := 0; i < iterations; i++ {
		p, _ := doc.Page(0)
		t0 := time.Now()
		rects, err := p.SearchFor("Logics", 10)
		t1 := time.Now()
		if err == nil {
			searchHits = len(rects)
		}
		times = append(times, t1.Sub(t0))
		p.Close()
	}
	s = stats(times)
	results["text_search"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"hits": searchHits,
	}
	fmt.Printf("[8] Text search: mean=%.3fms (hits=%d)\n", s.MeanMs, searchHits)

	// ---- 9. 大纲读取 ----
	times = times[:0]
	outlineCount := 0
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		outline, err := doc.GetOutline()
		t1 := time.Now()
		if err == nil {
			outlineCount = len(outline)
		}
		times = append(times, t1.Sub(t0))
	}
	s = stats(times)
	results["outline"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"count": outlineCount,
	}
	fmt.Printf("[9] Outline: mean=%.3fms (entries=%d)\n", s.MeanMs, outlineCount)

	// ---- 10. PDF 保存到内存 ----
	times = times[:0]
	var saveSize int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		data, err := doc.SaveToBytes(nil)
		t1 := time.Now()
		if err == nil {
			saveSize = len(data)
		}
		times = append(times, t1.Sub(t0))
	}
	s = stats(times)
	results["save_bytes"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"size_mb": math.Round(float64(saveSize)/1024/1024*100) / 100,
	}
	fmt.Printf("[10] Save to bytes: mean=%.1fms (%.1fMB)\n", s.MeanMs, float64(saveSize)/1024/1024)

	// ---- 11. 页面插入 10 页 ----
	times = times[:0]
	for i := 0; i < iterations; i++ {
		doc2, _ := fitz.NewPDF()
		t0 := time.Now()
		for j := 0; j < 10; j++ {
			doc2.NewPage(-1, 612, 792, 0)
		}
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
		doc2.Close()
	}
	s = stats(times)
	results["insert_10pages"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
	}
	fmt.Printf("[11] Insert 10 pages: mean=%.3fms\n", s.MeanMs)

	// ---- 12. 页面删除 10 页 ----
	times = times[:0]
	for i := 0; i < iterations; i++ {
		doc2, _ := fitz.NewPDF()
		for j := 0; j < 10; j++ {
			doc2.NewPage(-1, 612, 792, 0)
		}
		t0 := time.Now()
		for j := 0; j < 10; j++ {
			doc2.DeletePage(0)
		}
		t1 := time.Now()
		times = append(times, t1.Sub(t0))
		doc2.Close()
	}
	s = stats(times)
	results["delete_10pages"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
	}
	fmt.Printf("[12] Delete 10 pages: mean=%.3fms\n", s.MeanMs)

	// ---- 13. Pixmap→PNG ----
	pix, _ := page.Pixmap(mat2x, false)
	defer pix.Close()
	times = times[:0]
	var pngSize int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		data, err := pix.PNG()
		t1 := time.Now()
		if err == nil {
			pngSize = len(data)
		}
		times = append(times, t1.Sub(t0))
	}
	s = stats(times)
	results["pixmap_png"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"size_kb": math.Round(float64(pngSize)/1024*10) / 10,
	}
	fmt.Printf("[13] Pixmap→PNG: mean=%.1fms (%dKB)\n", s.MeanMs, pngSize/1024)

	// ---- 14. Pixmap→JPEG ----
	times = times[:0]
	var jpegSize int
	for i := 0; i < iterations; i++ {
		t0 := time.Now()
		data, err := pix.JPEG(85)
		t1 := time.Now()
		if err == nil {
			jpegSize = len(data)
		}
		times = append(times, t1.Sub(t0))
	}
	s = stats(times)
	results["pixmap_jpeg"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"size_kb": math.Round(float64(jpegSize)/1024*10) / 10,
	}
	fmt.Printf("[14] Pixmap→JPEG: mean=%.1fms (%dKB)\n", s.MeanMs, jpegSize/1024)

	// ---- 15. 链接读取 ----
	times = times[:0]
	linkCount := 0
	for i := 0; i < iterations; i++ {
		p, _ := doc.Page(0)
		t0 := time.Now()
		links, err := p.GetLinks()
		t1 := time.Now()
		if err == nil {
			linkCount = len(links)
		}
		times = append(times, t1.Sub(t0))
		p.Close()
	}
	s = stats(times)
	results["links"] = map[string]interface{}{
		"mean_ms": s.MeanMs, "min_ms": s.MinMs, "max_ms": s.MaxMs, "std_ms": s.StdMs,
		"count": linkCount,
	}
	fmt.Printf("[15] Links: mean=%.3fms (count=%d)\n", s.MeanMs, linkCount)

	doc.Close()

	// 元信息
	fi, _ := os.Stat(pdfPath)
	results["tool"] = "gomupdf"
	results["version"] = "" // MuPDF via CGO
	results["file"] = pdfPath
	results["file_size_mb"] = math.Round(float64(fi.Size())/1024/1024*100) / 100
	results["iterations"] = iterations

	// 保存 JSON
	outFile := "benchmark_gomupdf_result.json"
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(outFile, jsonData, 0644)
	fmt.Printf("\nResults saved to %s\n", outFile)
}
