package cgo

import (
	"os"
	"testing"
)

const testPDF = "../arxiv_2603.09677_translated.pdf"

func getTestPDF(t *testing.T) string {
	t.Helper()
	// Try multiple paths
	candidates := []string{
		testPDF,
		"../../arxiv_2603.09677_translated.pdf",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Fatal("test PDF not found")
	return ""
}

func TestContext(t *testing.T) {
	ctx := New()
	if ctx == nil {
		t.Fatal("New() returned nil")
	}
	defer ctx.Destroy()

	v := Version()
	if v == "" {
		t.Fatal("Version() returned empty string")
	}
	t.Logf("MuPDF version: %s", v)
}

func TestOpenDocument(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer doc.Destroy()

	if doc.PageCount() != 43 {
		t.Errorf("expected 43 pages, got %d", doc.PageCount())	}
	if !doc.IsPDF() {
		t.Error("expected IsPDF() == true")
	}
	if doc.NeedsPassword() {
		t.Error("expected NeedsPassword() == false")
	}
}

func TestMetadata(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	meta := doc.Metadata()
	if meta["format"] != "PDF 1.5" {
		t.Errorf("expected format=PDF 1.5, got %q", meta["format"])
	}
	t.Logf("Metadata: %v", meta)
}

func TestPageLoad(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	page, err := LoadPage(ctx, doc.Doc, 0)
	if err != nil {
		t.Fatalf("LoadPage failed: %v", err)
	}
	defer page.Destroy()

	x0, y0, x1, y1 := page.Rect()
	if x0 != 0 || y0 != 0 || x1 != 612 || y1 != 792 {
		t.Errorf("unexpected page rect: (%.0f, %.0f, %.0f, %.0f)", x0, y0, x1, y1)
	}
}

func TestPageRotation(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	rot := PDFPageRotation(ctx, doc.Doc, 0)
	t.Logf("Page 0 rotation: %d", rot)
}

func TestRenderPage(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	page, err := LoadPage(ctx, doc.Doc, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer page.Destroy()

	pix, err := RenderPage(ctx, page.Page, 1, 0, 0, 1, 0, 0, false)
	if err != nil {
		t.Fatalf("RenderPage failed: %v", err)
	}
	defer pix.Destroy()

	if pix.Width() != 612 {
		t.Errorf("expected width 612, got %d", pix.Width())
	}
	if pix.Height() != 792 {
		t.Errorf("expected height 792, got %d", pix.Height())
	}
	if pix.N() < 3 {
		t.Errorf("expected at least 3 components, got %d", pix.N())
	}

	samples := pix.Samples()
	if len(samples) == 0 {
		t.Error("Samples() returned empty")
	}
	t.Logf("Rendered: %dx%d, %d components, %d bytes", pix.Width(), pix.Height(), pix.N(), len(samples))
}

func TestSavePNG(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	page, err := LoadPage(ctx, doc.Doc, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer page.Destroy()

	pix, err := RenderPage(ctx, page.Page, 2, 0, 0, 2, 0, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	defer pix.Destroy()

	tmpfile := "/tmp/gomupdf_test_save.png"
	err = pix.SavePNG(tmpfile)
	if err != nil {
		t.Fatalf("SavePNG failed: %v", err)
	}

	fi, err := os.Stat(tmpfile)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if fi.Size() == 0 {
		t.Error("saved PNG is empty")
	}
	t.Logf("Saved PNG: %d bytes", fi.Size())
}

func TestTextExtraction(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	page, err := LoadPage(ctx, doc.Doc, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer page.Destroy()

	tp, err := NewTextPage(page)
	if err != nil {
		t.Fatalf("NewTextPage failed: %v", err)
	}
	defer tp.Destroy()

	if tp.BlockCount() == 0 {
		t.Error("BlockCount() returned 0")
	}
	t.Logf("Blocks: %d", tp.BlockCount())

	text := tp.Text()
	if len(text) == 0 {
		t.Error("Text() returned empty")
	}
	t.Logf("Text (first 100 chars): %q", truncate(text, 100))
}

func TestPDFSave(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	tmpfile := "/tmp/gomupdf_test_save.pdf"
	err = doc.SaveDocument(tmpfile, &SaveOptions{Garbage: 1, Clean: true})
	if err != nil {
		t.Fatalf("SaveDocument failed: %v", err)
	}

	fi, err := os.Stat(tmpfile)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if fi.Size() == 0 {
		t.Error("saved PDF is empty")
	}
	t.Logf("Saved PDF: %d bytes", fi.Size())
}

func TestPDFWriteToBytes(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	data, err := doc.WriteDocument(nil)
	if err != nil {
		t.Fatalf("WriteDocument failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("WriteDocument returned empty bytes")
	}
	t.Logf("Written to bytes: %d bytes", len(data))
}

func TestPDFInsertPage(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := NewPDF(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	if doc.PageCount() != 0 {
		t.Errorf("expected 0 pages, got %d", doc.PageCount())
	}

	err = doc.InsertPage(-1, 0, 0, 612, 792, 0)
	if err != nil {
		t.Fatalf("InsertPage failed: %v", err)
	}

	if doc.PageCount() != 1 {
		t.Errorf("expected 1 page, got %d", doc.PageCount())
	}
}

func TestPDFDeletePage(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	before := doc.PageCount()
	err = doc.DeletePage(0)
	if err != nil {
		t.Fatalf("DeletePage failed: %v", err)
	}
	after := doc.PageCount()

	if after != before-1 {
		t.Errorf("expected %d pages, got %d", before-1, after)
	}
}

func TestSetMetadata(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	err = doc.SetMetadata("info:Title", "Test Title")
	if err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Save and re-read to verify metadata persistence
	tmpfile := "/tmp/gomupdf_test_metadata.pdf"
	err = doc.SaveDocument(tmpfile, nil)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Re-open and check
	doc2, err := Open(ctx, tmpfile)
	if err != nil {
		t.Fatal(err)
	}
	defer doc2.Destroy()

	meta := doc2.Metadata()
	if meta["title"] != "Test Title" {
		t.Errorf("expected title=Test Title after save/reload, got %q", meta["title"])
	}
}

func TestGetOutline(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	outline, err := doc.GetOutline()
	if err != nil {
		t.Fatalf("GetOutline failed: %v", err)
	}

	if len(outline) == 0 {
		t.Error("GetOutline returned empty")
	}
	t.Logf("Outline entries: %d", len(outline))
	for i, e := range outline {
		if i > 5 {
			break
		}
		t.Logf("  [%d] %q page=%d level=%d", i, e.Title, e.Page, e.Level)
	}
}

func TestSearchText(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	page, err := LoadPage(ctx, doc.Doc, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer page.Destroy()

	tp, err := NewTextPage(page)
	if err != nil {
		t.Fatal(err)
	}
	defer tp.Destroy()

	rects := SearchText(ctx, tp.CStextPage(), "Logics", 10)
	if len(rects) == 0 {
		t.Error("SearchText returned no results")
	}
	t.Logf("Search 'Logics': %d hits", len(rects))
}

func TestPermissions(t *testing.T) {
	ctx := New()
	defer ctx.Destroy()

	doc, err := Open(ctx, getTestPDF(t))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Destroy()

	perm := doc.Permissions()
	t.Logf("Permissions: %d", perm)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
