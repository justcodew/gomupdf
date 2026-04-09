# gomupdf 开发工作日志

> 开始日期：2026-04-09

---

## 阶段 0：基础设施修复

### 0.1 已知 Bug 修复（已完成）

**bindings.c 修复：**

| Bug | 修复方式 | 状态 |
|-----|---------|------|
| Metadata 永远返回空字符串 | 改用 `fz_lookup_metadata(ctx, doc, key, buf, size)` 正确读取元数据 | 已完成 |
| `fz_drop_colorspace` 释放设备颜色空间 | 移除对 `fz_device_rgb` 返回值的 `fz_drop_colorspace` 调用（借用引用，不可释放） | 已完成 |
| JPEG 保存 `fz_output` 泄漏 | 添加 `fz_close_output` + `fz_drop_output`，使用 `fz_always` 确保清理 | 已完成 |
| 文本提取 `fz_output` 泄漏 | 添加 `fz_close_output` + `fz_drop_output` | 已完成 |
| `char_count` 使用错误的指针运算 | 将 `(char*)ch + sizeof(void*)` 改为 `ch->next` | 已完成 |
| `stext_line_get_span` 返回 NULL | 替换为 `stext_line_first_char`，因为 MuPDF 的 C 层没有 span 概念 | 已完成 |
| `image_colorspace_name` 总返回 DeviceRGB | 改用 `fz_colorspace_name()` 返回真实颜色空间名 | 已完成 |
| `page_rotation` 硬编码返回 0 | 新增 `gomupdf_pdf_page_rotation` 通过 PDF 字典 `/Rotate` 读取真实旋转值 | 已完成 |

**Go 层修复：**

| Bug | 修复方式 | 状态 |
|-----|---------|------|
| `OpenPDFStream` 不初始化 fitz 字段 | 新增 `fitz.OpenStreamWithContext()`，同时初始化 cgo 和 fitz 文档 | 已完成 |
| `GetPixmap` alpha 处理错误 | 使用 `pixmap.N()` 获取实际组件数，按 n 步进而非硬编码 3 | 已完成 |
| 90 度旋转坐标缺少 pageHeight | 添加 `pageHeight - y - height` 正确计算 90 度旋转 | 已完成 |
| `fitz/Page` 不存储页码 | 新增 `index` 字段，`Number()` 返回真实页码 | 已完成 |
| `fitz/Page.Rotation()` 只查 fz_page | 改为优先调用 `PDFPageRotation` 读取 PDF /Rotate 字典项 | 已完成 |

### 0.2 MuPDF 异常保护（已完成）

在 `bindings.c` 中为所有关键 MuPDF 调用增加 `fz_try`/`fz_catch` 保护：

- `gomupdf_open_document`
- `gomupdf_open_document_with_stream`
- `gomupdf_new_pdf_document`
- `gomupdf_drop_document`
- `gomupdf_page_count`
- `gomupdf_load_page`
- `gomupdf_page_rect`
- `gomupdf_render_page`（含设备清理）
- `gomupdf_new_stext_page_from_page`
- `gomupdf_save_pixmap_as_png`
- `gomupdf_save_pixmap_as_jpeg`（含 `fz_always` 确保资源释放）

新增错误传递机制：
- `gomupdf_get_last_error()` — 获取最近一次 MuPDF 错误信息
- `gomupdf_clear_error()` — 清除错误
- Go 层 `cgo.GetLastError()` / `cgo.ClearError()`

### 0.3 激活已有 C 绑定（已完成）

新建 `cgo_bindings/geometry.go`，暴露 29 个已有 C wrapper 的 Go 入口：

**矩阵操作（7 个）：**
- `IdentityMatrix`, `MakeMatrix`, `ScaleMatrix`, `RotateMatrix`
- `TranslateMatrix`, `ConcatMatrix`, `InvertMatrix`

**点操作（2 个）：**
- `MakePoint`, `TransformPoint`

**矩形操作（5 个）：**
- `MakeRect`, `MakeIRect`, `RectIsEmpty`, `RectIsInfinite`, `TransformRect`

**四边形操作（3 个）：**
- `MakeQuad`, `QuadRect`, `TransformQuad`

**颜色空间操作（2 个）：**
- `FindDeviceColorspace`, `PixmapColorspace`

**图像信息操作（6 个）：**
- `ImageWidth`, `ImageHeight`, `ImageN`, `ImageBPC`
- `ImageColorspaceName`, `BlockGetImage`

### 验证

所有修复通过集成测试：
```
$ go run examples/render/main.go test.pdf output.png
# 正常输出：metadata 显示 "PDF 1.5"，渲染正常
$ go run examples/extract/main.go test.pdf
# 文本提取正常
```

### 0.4 测试框架（已完成）

新建 `cgo_bindings/cgo_test.go`，包含 16 个测试用例：

| 测试 | 验证内容 |
|------|---------|
| TestContext | Context 创建/销毁、版本号 |
| TestOpenDocument | 文档打开/关闭/页面计数/IsPDF/NeedsPassword |
| TestMetadata | 元数据读取（format/creator/producer 等） |
| TestPageLoad | 页面加载、矩形尺寸 |
| TestPageRotation | 页面旋转查询 |
| TestRenderPage | 页面渲染为 Pixmap |
| TestSavePNG | Pixmap 保存为 PNG |
| TestTextExtraction | 文本提取（块计数、文本内容） |
| TestPDFSave | PDF 保存到文件 |
| TestPDFWriteToBytes | PDF 保存到内存字节 |
| TestPDFInsertPage | 新建空白页 |
| TestPDFDeletePage | 删除页面 |
| TestSetMetadata | 设置元数据（保存后验证） |
| TestGetOutline | 目录/大纲读取（31 条） |
| TestSearchText | 文本搜索（6 个命中） |
| TestPermissions | 权限标志 |

运行结果：**16/16 PASS，耗时 0.48s**

---

## 阶段 1：核心文档操作（已完成）

### 新增 C 函数（bindings.c/bindings.h）

| 函数 | 说明 |
|------|------|
| `gomupdf_pdf_save_document` | PDF 保存到文件，支持全部 13 项保存选项 |
| `gomupdf_pdf_write_document` | PDF 保存到内存 buffer |
| `gomupdf_free` | 释放 C 分配的内存 |
| `gomupdf_pdf_insert_page` | 创建并插入空白页（指定 mediabox 和旋转） |
| `gomupdf_pdf_delete_page` | 删除单页 |
| `gomupdf_pdf_delete_page_range` | 删除页面范围 |
| `gomupdf_pdf_set_metadata` | 设置元数据（通过 `fz_set_metadata`） |
| `gomupdf_pdf_permissions` | 读取 PDF 权限标志 |
| `gomupdf_pdf_outline_count` | 加载并展平大纲树 |
| `gomupdf_pdf_outline_get` | 获取第 N 个大纲条目 |
| `gomupdf_page_load_links` | 加载页面链接 |
| `gomupdf_drop_link` | 释放链接 |
| `gomupdf_link_next/rect/uri` | 链接属性访问 |
| `gomupdf_link_page` | 解析链接目标页码 |
| `gomupdf_search_text` | 在 TextPage 中搜索文本 |

### 新增 Go 文件

| 文件 | 说明 |
|------|------|
| `cgo_bindings/doc_ops.go` | PDF 操作 Go 绑定：保存、页面管理、元数据、大纲、链接、搜索 |
| `cgo_bindings/geometry.go` | 几何类型操作：矩阵、点、矩形、四边形、颜色空间、图像信息 |

### fitz 层更新

- `fitz/doc.go` — 新增 `Save`, `SaveToBytes`, `NewPage`, `DeletePage`, `SetMetadata`, `Permissions`, `GetOutline`
- `fitz/page.go` — 新增 `GetLinks`, `SearchFor`

### pdf 层修复

- `pdf/api.go` — 修复 `OpenPDFStream`（同时初始化 fitz）、`GetPixmap`（正确处理 alpha/N 组件）、`IsEncrypted`
- `pdf/position.go` — 修复 90 度旋转坐标计算
- `pdf/merge_split.go` — 暂未修复（需要 PDF 页面复制能力，将在后续阶段实现）

### Metadata 改进

- `Metadata()` 现在同时尝试短键名和 `info:` 前缀长键名
- 正确返回 `creationDate`, `modDate`, `creator`, `producer` 等字段

---

## 阶段 2：文本处理完整化（进行中）
