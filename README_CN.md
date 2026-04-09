# GoMuPDF - Go 语言的 MuPDF 绑定

GoMuPDF 是一个用 Go 语言编写的 MuPDF C 库的绑定，用于 PDF 文档处理。

**状态：早期开发**

这是初始实现，CGO 绑定和核心类型已就绪。

## 功能特性

### 已实现

**核心功能 (fitz 包):**
- 文档打开（文件和流）
- 页面枚举
- 基本页面渲染为 pixmap
- Pixmap 保存（PNG、JPEG）
- 基本几何类型（Rect、Point、Matrix、Quad）
- 文档元数据

**PDF 工具 (pdf 包):**
- 带坐标的文本提取（OCR 风格）
- 图片位置提取
- SVG/绘图位置提取
- 不同页面方向的坐标旋转
- 乱码文本检测
- 字体损坏检测
- 纯白/纯黑图片检测
- PDF 分割和合并
- 图片处理（调整大小、旋转）

### 尚未完全实现
- 完整文本提取（所有样式信息）
- 注解操作
- TOC/大纲操作
- 表单字段（widgets）
- 嵌入式文件
- 搜索功能

## 安装要求

- Go 1.21+
- MuPDF C 库 (v1.27.2)
- GCC 或 Clang 编译器

### macOS (使用 Homebrew)

```bash
brew install go mupdf
```

### Linux (使用包管理器)

```bash
# Ubuntu/Debian
sudo apt-get install libmupdf-dev golang

# Fedora
sudo dnf install mupdf-devel golang
```

## 构建

```bash
cd gomupdf
go build ./...
```

## 使用示例

### 渲染 PDF 页面为图片

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

### 获取 PDF 信息

```go
package main

import (
    "fmt"
    "log"
    "github.com/go-pymupdf/gomupdf/fitz"
)

func main() {
    doc, err := fitz.Open("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer doc.Close()

    fmt.Printf("Pages: %d\n", doc.PageCount())
    fmt.Printf("Is PDF: %v\n", doc.IsPDF())

    metadata := doc.Metadata()
    for k, v := range metadata {
        fmt.Printf("%s: %s\n", k, v)
    }
}
```

## 项目结构

```
gomupdf/
├── cgo_bindings/    # MuPDF 的低级 CGO 绑定
│   ├── bindings.h   # C 头文件
│   ├── bindings.c   # C 实现
│   ├── context.go   # 上下文管理
│   ├── document.go  # 文档操作
│   ├── page.go     # 页面操作
│   ├── pixmap.go   # Pixmap 操作
│   └── text.go     # 文本操作
├── fitz/            # Go 原生 API 封装
│   ├── doc.go       # Document 类型
│   ├── page.go      # Page 类型
│   ├── pixmap.go    # Pixmap 类型
│   ├── rect.go      # Rect 类型
│   ├── point.go     # Point 类型
│   ├── matrix.go    # Matrix 类型
│   ├── quad.go      # Quad 类型
│   └── text.go      # 文本类型
├── pdf/             # 完整 PDF 工具（类似 pdf_utils.py）
│   ├── api.go       # 主 API
│   ├── structs.go   # 数据结构
│   ├── text.go      # 文本处理和乱码检测
│   ├── position.go  # 坐标旋转
│   ├── image.go     # 图片处理
│   └── merge_split.go # PDF 分割/合并
└── examples/        # 示例程序
    ├── render/      # 页面渲染
    ├── extract/    # 文本/图片提取
    ├── pdfinfo/     # PDF 信息
    └── visualize/   # 可视化
```

## 基准测试

运行基准测试：

```bash
# Python 基准测试
python benchmark_python.py document.pdf

# Go 基准测试
cd gomupdf && go run benchmark/main.go document.pdf

# 比较结果
python benchmark_compare.py
```

## 可视化

生成 PDF 解析结果的可视化：

```bash
# Go 可视化（生成 PNG）
./go_visualize -i document.pdf -o visualization -p 0-5

# Python 可视化（生成 PDF）
python visualize_to_pdf.py document.pdf -o visualization.pdf
```

## 与 PyMuPDF 的对比

| 操作 | PyMuPDF | GoMuPDF | 备注 |
|------|---------|---------|------|
| 打开文档 | 较慢 | 更快 | Go 使用 CGO 直接调用 |
| 加载页面 | 更快 | 较慢 | - |
| 页面渲染 | 相当 | 相当 | - |
| 保存文件 | 相当 | 相当 | - |

## 已知问题

1. **多页处理稳定性**: MuPDF 使用 `setjmp`/`longjmp` 进行错误处理，与 Go GC 存在兼容性。解决方案：使用并发限制和线程锁定。

2. **文本提取**: 部分功能尚未在 fitz 包中实现，需要进一步封装 cgo_bindings 中的底层函数。

3. **错误处理**: MuPDF 的错误处理机制与 Go 的 panic/recover 模型不同，需要通过返回值传递错误。

## 许可证

本项目使用与 MuPDF 相同的许可证（AGPL v3 或商业许可证）。
如需商业许可，请联系 Artifex Software。

## 参考链接

- [MuPDF 官网](https://mupdf.com/)
- [PyMuPDF](https://pymupdf.readthedocs.io/)
- [Go 官方文档](https://go.dev/)
