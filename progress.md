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

## 阶段 2：文本处理完整化（已完成）

### 2.1 多格式文本输出

| 格式 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 纯文本 | `gomupdf_stext_page_text`（已有） | `TextPage.Text()` | 已完成 |
| HTML | `gomupdf_stext_page_to_html` | `TextPage.HTML()` | 已完成 |
| XML | `gomupdf_stext_page_to_xml` | `TextPage.XML()` | 已完成 |
| XHTML | `gomupdf_stext_page_to_xhtml` | `TextPage.XHTML()` | 已完成 |
| JSON | `gomupdf_stext_page_to_json` | `TextPage.JSON()` | 已完成 |

### 2.2 字符级属性

| 属性 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 字体名 | `gomupdf_stext_char_font` | `TextPage.CharFont()` | 已完成 |
| 字符标志 | `gomupdf_stext_char_flags` | `TextPage.CharFlags()` | 已完成 |

### fitz 层更新

- `fitz/text.go` — 新增 `GetTextHTML()`, `GetTextXML()`, `GetTextXHTML()`, `GetTextJSON()`

### 测试

新增 3 个测试：`TestTextHTML`, `TestTextXML`, `TestTextJSON` — 全部通过

---

## 阶段 3：图像处理完整化（已完成）

### 3.1 图像像素提取

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 图像转 Pixmap | `gomupdf_image_get_pixmap` | `NewPixmapFromImage()` | 已完成 |
| Pixmap → PNG 字节 | `gomupdf_pixmap_to_png_bytes` | `Pixmap.PNGBytes()` | 已完成 |
| Pixmap → JPEG 字节 | `gomupdf_pixmap_to_jpeg_bytes` | `Pixmap.JPEGBytes()` | 已完成 |

### 3.2 Pixmap 操作

| 操作 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 读取像素 | `gomupdf_pixmap_pixel` | `Pixmap.Pixel()` | 已完成 |
| 设置像素 | `gomupdf_pixmap_set_pixel` | `Pixmap.SetPixel()` | 已完成 |
| 清空 | `gomupdf_pixmap_clear_with` | `Pixmap.ClearWith()` | 已完成 |
| 反色 | `gomupdf_pixmap_invert` | `Pixmap.Invert()` | 已完成 |
| Gamma 校正 | `gomupdf_pixmap_gamma` | `Pixmap.Gamma()` | 已完成 |
| 着色 | `gomupdf_pixmap_tint` | `Pixmap.Tint()` | 已完成 |

### fitz 层更新

- `fitz/pixmap.go` — 实现 `PNG()`, `JPEG()`, `Pixel()`, `SetPixel()`, `ClearWith()`, `Invert()`, `Gamma()`, `Tint()`
- `fitz/image.go` — 实现 `GetImages()`, `ExtractImage()` 从图像块提取像素数据

### 测试

新增 3 个测试：`TestPixmapPNGBytes`, `TestPixmapJPEGBytes`, `TestPixmapPixelOps` — 全部通过

---

## 阶段 4：注释系统（已完成）

### 4.1 注释基础设施

| 功能 | C 函数 | 状态 |
|------|--------|------|
| 获取首个注释 | `gomupdf_pdf_first_annot` | 已完成 |
| 遍历注释 | `gomupdf_pdf_next_annot` | 已完成 |
| 注释类型 | `gomupdf_pdf_annot_type` | 已完成 |
| 注释矩形 | `gomupdf_pdf_annot_rect` / `gomupdf_pdf_set_annot_rect` | 已完成 |
| 注释内容 | `gomupdf_pdf_annot_contents` / `gomupdf_pdf_set_annot_contents` | 已完成 |
| 注释颜色 | `gomupdf_pdf_annot_color` / `gomupdf_pdf_set_annot_color` | 已完成 |
| 注释透明度 | `gomupdf_pdf_annot_opacity` / `gomupdf_pdf_set_annot_opacity` | 已完成 |
| 注释标志 | `gomupdf_pdf_annot_flags` / `gomupdf_pdf_set_annot_flags` | 已完成 |
| 注释边框 | `gomupdf_pdf_annot_border` / `gomupdf_pdf_set_annot_border` | 已完成 |
| 注释标题 | `gomupdf_pdf_annot_title` / `gomupdf_pdf_set_annot_title` | 已完成 |
| 弹出窗口 | `gomupdf_pdf_annot_popup` / `gomupdf_pdf_set_annot_popup` | 已完成 |
| Quad Points | `gomupdf_pdf_annot_quad_points` / `gomupdf_pdf_set_annot_quad_points` | 已完成 |
| 更新注释 | `gomupdf_pdf_update_annot` | 已完成 |
| 删除注释 | `gomupdf_pdf_delete_annot` | 已完成 |
| 创建注释 | `gomupdf_pdf_create_annot` | 已完成 |
| 应用涂黑 | `gomupdf_pdf_apply_redactions` | 已完成 |

### 新增 Go 文件

- `cgo_bindings/annot.go` — 注释 CGO 封装
- `fitz/annot.go` — 完整的 Annot Go 类型（含所有属性方法）

### fitz 层 Page 方法

- `Page.Annots()` — 获取注释列表
- `Page.AddAnnot()` — 创建注释
- `Page.AddHighlightAnnot()` — 高亮注释
- `Page.AddStrikeoutAnnot()` — 删除线注释
- `Page.AddUnderlineAnnot()` — 下划线注释
- `Page.AddSquigglyAnnot()` — 波浪线注释
- `Page.AddTextAnnot()` — 文本（便签）注释
- `Page.AddFreeTextAnnot()` — 自由文本注释
- `Page.AddRedactAnnot()` — 涂黑注释
- `Page.ApplyRedactions()` — 应用涂黑

### 测试

新增 `TestAnnotationCRUD` — 通过

---

## 阶段 5：链接与导航（已完成）

### 新增功能

| 功能 | C 函数 | 状态 |
|------|--------|------|
| 创建链接 | `gomupdf_pdf_create_link` | 已完成 |
| 删除链接 | `gomupdf_pdf_delete_link` | 已完成 |

### fitz 层

- `Page.AddLink()` — 添加链接

---

## 阶段 6：Shape 绘图系统（已完成）

### 内容流操作

| 功能 | C 函数 | 状态 |
|------|--------|------|
| 开始写入内容流 | `gomupdf_pdf_page_write_begin` | 已完成 |
| 结束写入内容流 | `gomupdf_pdf_page_write_end` | 已完成 |

---

## 阶段 7：文本写入与字体（已完成）

### 字体操作

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 从文件加载字体 | `gomupdf_new_font_from_file` | `NewFontFromFile()` | 已完成 |
| 从缓冲区加载字体 | `gomupdf_new_font_from_buffer` | `NewFontFromBuffer()` | 已完成 |
| 字体名称 | `gomupdf_font_name` | `Font.Name()` | 已完成 |
| 升部线 | `gomupdf_font_ascender` | `Font.Ascender()` | 已完成 |
| 降部线 | `gomupdf_font_descender` | `Font.Descender()` | 已完成 |
| 文本测量 | `gomupdf_measure_text` | `Font.MeasureText()` | 已完成 |
| 字形步进 | `gomupdf_font_glyph_advance` | `Font.GlyphAdvance()` | 已完成 |

### 新增文件

- `cgo_bindings/font.go` — 字体 CGO 封装
- `fitz/font.go` — Font Go 类型

### 测试

新增 `TestFontMeasure` — 通过（系统 Helvetica 字体，测量 "Hello, World!" = 68.68pt @12pt）

---

## 阶段 8：Widget 表单系统（已完成）

### 表单操作

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 获取首个控件 | `gomupdf_pdf_first_widget` | `FirstWidget()` | 已完成 |
| 遍历控件 | `gomupdf_pdf_next_widget` | `Widget.Next()` | 已完成 |
| 控件类型 | `gomupdf_pdf_widget_type` | `Widget.Type()` | 已完成 |
| 字段名 | `pdf_annot_field_label` | `Widget.FieldName()` | 已完成 |
| 字段值 | `pdf_annot_field_value` | `Widget.FieldValue()` | 已完成 |
| 设置字段值 | `pdf_set_text_field_value` | `Widget.SetFieldValue()` | 已完成 |
| 字段标志 | `pdf_annot_field_flags` | `Widget.FieldFlags()` | 已完成 |
| 复选框选中状态 | 通过 /AS 字段 | `Widget.IsChecked()` | 已完成 |
| 切换复选框 | `pdf_toggle_widget` | `Widget.Toggle()` | 已完成 |

### 新增文件

- `cgo_bindings/widget.go` — Widget CGO 封装
- `fitz/widget.go` — Widget Go 类型

### fitz 层

- `Page.Widgets()` — 获取页面所有表单控件

### 测试

新增 `TestWidgets` — 通过

---

## 阶段 9：高级功能（已完成）

### 9.1 DisplayList

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 创建显示列表 | `gomupdf_new_display_list` | `NewDisplayList()` | 已完成 |
| 页面渲染到列表 | `gomupdf_run_page_to_list` | `RunPageToDisplayList()` | 已完成 |
| 列表渲染为 Pixmap | `gomupdf_display_list_get_pixmap` | `DisplayList.GetPixmap()` | 已完成 |

### 9.2 页面属性

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| MediaBox | `gomupdf_pdf_page_mediabox` | `Page.MediaBox()` | 已完成 |
| CropBox | `gomupdf_pdf_page_cropbox` | `Page.CropBox()` | 已完成 |
| 设置 MediaBox | `gomupdf_pdf_set_page_mediabox` | `Page.SetMediaBox()` | 已完成 |
| 设置 CropBox | `gomupdf_pdf_set_page_cropbox` | `Page.SetCropBox()` | 已完成 |
| 设置旋转 | `gomupdf_pdf_set_page_rotation` | `Page.SetRotation()` | 已完成 |

### 9.3 XRef 操作

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| XRef 长度 | `gomupdf_pdf_xref_length` | `Document.XRefLength()` | 已完成 |
| 获取对象键 | `gomupdf_pdf_xref_get_key` | `Document.XRefGetKey()` | 已完成 |
| 判断流对象 | `gomupdf_pdf_xref_is_stream` | `Document.XRefIsStream()` | 已完成 |

### 9.4 嵌入文件

| 功能 | C 函数 | Go 方法 | 状态 |
|------|--------|---------|------|
| 嵌入文件数量 | 通过 Names 字典 | `Document.EmbeddedFileCount()` | 已完成 |
| 文件名 | 通过 Names 字典 | `Document.EmbeddedFileName()` | 已完成 |
| 获取文件数据 | `pdf_load_embedded_file_contents` | `Document.EmbeddedFileGet()` | 已完成 |
| 添加嵌入文件 | `pdf_add_embedded_file` | `Document.AddEmbeddedFile()` | 已完成 |

### 新增文件

- `cgo_bindings/advanced.go` — DisplayList、页面属性、XRef、嵌入文件 CGO 封装

### 测试

新增 4 个测试：`TestDisplayList`, `TestPageBox`, `TestXRef`, `TestEmbeddedFiles` — 全部通过

---

## 测试总结

| 阶段 | 新增测试 | 状态 |
|------|---------|------|
| 阶段 0-1（已有） | 16 个 | 全部通过 |
| 阶段 2 文本格式 | 3 个 | 全部通过 |
| 阶段 3 图像处理 | 3 个 | 全部通过 |
| 阶段 4 注释系统 | 1 个 | 全部通过 |
| 阶段 7 字体操作 | 1 个 | 全部通过 |
| 阶段 8 Widget | 1 个 | 全部通过 |
| 阶段 9 高级功能 | 4 个 | 全部通过 |
| **合计** | **29 个** | **全部通过（1.2s）** |

---

## 代码统计

| 指标 | 阶段 0-1 | 阶段 2-9 | 合计 |
|------|---------|---------|------|
| C 代码 (bindings.c) | ~995 行 | ~900 行 | ~1895 行 |
| C 头文件 (bindings.h) | ~166 行 | ~170 行 | ~336 行 |
| cgo_bindings/*.go | ~1300 行 | ~1300 行 | ~2600 行 |
| fitz/*.go | ~650 行 | ~900 行 | ~1550 行 |
| 测试代码 | ~426 行 | ~400 行 | ~826 行 |
| **总计** | **~3537 行** | **~3670 行** | **~7207 行** |
