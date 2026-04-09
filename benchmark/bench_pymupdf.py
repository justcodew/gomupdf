"""
PyMuPDF 性能基准测试脚本

测试项目：
1. 文档打开与关闭
2. 元数据读取
3. 页面加载
4. 页面渲染（1x / 2x）
5. 全页面渲染
6. 文本提取
7. 文本搜索
8. 大纲读取
9. 页面保存
10. 页面插入/删除
"""

import sys
import os
import json
import time
import statistics

def run_benchmark(pdf_path, iterations=5):
    import pymupdf

    print(f"PyMuPDF version: {pymupdf.__version__}")
    print(f"Input: {pdf_path}")
    print(f"Iterations: {iterations}\n")

    results = {}

    # ---- 1. 文档打开与关闭 ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        doc = pymupdf.open(pdf_path)
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
        page_count = len(doc)
        doc.close()
    results["open_close"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "pages": page_count,
    }
    print(f"[1] Open/Close: mean={results['open_close']['mean_ms']:.1f}ms "
          f"(pages={page_count})")

    # ---- 2. 元数据读取 ----
    doc = pymupdf.open(pdf_path)
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        meta = doc.metadata
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
    results["metadata"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "keys": len(meta) if meta else 0,
    }
    print(f"[2] Metadata: mean={results['metadata']['mean_ms']:.3f}ms")

    # ---- 3. 单页加载 ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        page = doc[0]
        _ = page.rect
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
    results["load_page"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
    }
    print(f"[3] Load page 0: mean={results['load_page']['mean_ms']:.3f}ms")

    # ---- 4. 页面渲染 1x ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        pix = page.get_pixmap(matrix=pymupdf.Matrix(1, 1))
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
        w, h = pix.width, pix.height
    results["render_1x"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "width": w, "height": h,
    }
    print(f"[4] Render 1x: mean={results['render_1x']['mean_ms']:.1f}ms "
          f"({w}x{h})")

    # ---- 5. 页面渲染 2x ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        pix = page.get_pixmap(matrix=pymupdf.Matrix(2, 2))
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
        w2, h2 = pix.width, pix.height
    results["render_2x"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "width": w2, "height": h2,
    }
    print(f"[5] Render 2x: mean={results['render_2x']['mean_ms']:.1f}ms "
          f"({w2}x{h2})")

    # ---- 6. 全页面渲染 1x ----
    t0 = time.perf_counter()
    for i in range(page_count):
        p = doc[i]
        pix = p.get_pixmap(matrix=pymupdf.Matrix(1, 1))
    t1 = time.perf_counter()
    all_ms = (t1 - t0) * 1000
    results["render_all_1x"] = {
        "total_ms": round(all_ms, 1),
        "per_page_ms": round(all_ms / page_count, 1),
        "pages": page_count,
    }
    print(f"[6] Render all 1x: {all_ms:.0f}ms total "
          f"({all_ms/page_count:.1f}ms/page)")

    # ---- 7. 文本提取 ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        text = doc[0].get_text()
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
    results["text_extract"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "text_length": len(text),
    }
    print(f"[7] Text extract: mean={results['text_extract']['mean_ms']:.3f}ms "
          f"(len={len(text)})")

    # ---- 8. 文本搜索 ----
    times = []
    search_hits = 0
    for _ in range(iterations):
        t0 = time.perf_counter()
        hits = doc[0].search_for("Logics")
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
        search_hits = len(hits)
    results["text_search"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "hits": search_hits,
    }
    print(f"[8] Text search: mean={results['text_search']['mean_ms']:.3f}ms "
          f"(hits={search_hits})")

    # ---- 9. 大纲读取 ----
    times = []
    outline_count = 0
    for _ in range(iterations):
        t0 = time.perf_counter()
        toc = doc.get_toc()
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
        outline_count = len(toc)
    results["outline"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "count": outline_count,
    }
    print(f"[9] Outline: mean={results['outline']['mean_ms']:.3f}ms "
          f"(entries={outline_count})")

    # ---- 10. PDF 保存到内存 ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        data = doc.tobytes()
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
    results["save_bytes"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "std_ms": round(statistics.stdev(times), 3) if len(times) > 1 else 0,
        "size_mb": round(len(data) / 1024 / 1024, 2),
    }
    print(f"[10] Save to bytes: mean={results['save_bytes']['mean_ms']:.1f}ms "
          f"({len(data)/1024/1024:.1f}MB)")
    doc.close()

    # ---- 11. 页面插入/删除 ----
    times_ins = []
    times_del = []
    for _ in range(iterations):
        doc2 = pymupdf.open()
        # 插入 10 页
        t0 = time.perf_counter()
        for i in range(10):
            doc2.new_page(width=612, height=792)
        t1 = time.perf_counter()
        times_ins.append((t1 - t0) * 1000)

        # 删除 10 页
        t0 = time.perf_counter()
        for i in range(10):
            doc2.delete_page(0)
        t1 = time.perf_counter()
        times_del.append((t1 - t0) * 1000)
        doc2.close()

    results["insert_10pages"] = {
        "mean_ms": round(statistics.mean(times_ins), 3),
        "min_ms": round(min(times_ins), 3),
        "max_ms": round(max(times_ins), 3),
    }
    results["delete_10pages"] = {
        "mean_ms": round(statistics.mean(times_del), 3),
        "min_ms": round(min(times_del), 3),
        "max_ms": round(max(times_del), 3),
    }
    print(f"[11] Insert 10 pages: mean={results['insert_10pages']['mean_ms']:.3f}ms")
    print(f"[12] Delete 10 pages: mean={results['delete_10pages']['mean_ms']:.3f}ms")

    # ---- 13. Pixmap 保存 PNG ----
    doc = pymupdf.open(pdf_path)
    page = doc[0]
    pix = page.get_pixmap(matrix=pymupdf.Matrix(2, 2))
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        data = pix.tobytes("png")
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
    results["pixmap_png"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "size_kb": round(len(data) / 1024, 1),
    }
    print(f"[13] Pixmap→PNG: mean={results['pixmap_png']['mean_ms']:.1f}ms "
          f"({len(data)/1024:.0f}KB)")

    # ---- 14. Pixmap 保存 JPEG ----
    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        data = pix.tobytes("jpeg", jpg_quality=85)
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
    results["pixmap_jpeg"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "size_kb": round(len(data) / 1024, 1),
    }
    print(f"[14] Pixmap→JPEG: mean={results['pixmap_jpeg']['mean_ms']:.1f}ms "
          f"({len(data)/1024:.0f}KB)")

    # ---- 15. 链接读取 ----
    times = []
    link_count = 0
    for _ in range(iterations):
        t0 = time.perf_counter()
        links = page.get_links()
        t1 = time.perf_counter()
        times.append((t1 - t0) * 1000)
        link_count = len(links)
    results["links"] = {
        "mean_ms": round(statistics.mean(times), 3),
        "min_ms": round(min(times), 3),
        "max_ms": round(max(times), 3),
        "count": link_count,
    }
    print(f"[15] Links: mean={results['links']['mean_ms']:.3f}ms "
          f"(count={link_count})")

    doc.close()

    results["tool"] = "PyMuPDF"
    results["version"] = pymupdf.__version__
    results["file"] = os.path.basename(pdf_path)
    results["file_size_mb"] = round(os.path.getsize(pdf_path) / 1024 / 1024, 2)
    results["iterations"] = iterations

    return results


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python bench_pymupdf.py <input.pdf> [iterations]")
        sys.exit(1)

    pdf_path = sys.argv[1]
    iterations = int(sys.argv[2]) if len(sys.argv) > 2 else 5

    results = run_benchmark(pdf_path, iterations)

    out_file = "benchmark_pymupdf_result.json"
    with open(out_file, "w") as f:
        json.dump(results, f, indent=2, ensure_ascii=False)
    print(f"\nResults saved to {out_file}")
