# gomupdf 完全复现 PyMuPDF 实施计划

> 创建日期：2026-04-09
> 目标：将 gomupdf 从当前 ~15-20% 功能覆盖率提升至与 PyMuPDF 对等的完整实现

---

## 现状概要

| 指标 | 当前值 |
|------|--------|
| Go 源码行数 | ~3,941 行 |
| 已绑定 C 函数 | 58 个（其中 29 个无 Go 入口） |
| 已实现核心类 | Document, Page, Pixmap, TextPage（基础） |
| 已实现辅助类 | Rect, IRect, Point, Matrix, Quad |
| 已知 Bug 数 | 3+（颜色空间释放、指针运算、fz_output 泄漏） |
| 测试文件 | 0 |
| 目标参照 | PyMuPDF ~25 个类、500+ 方法 |

---

## 阶段 0：基础设施修复（前置条件）

> 不新增功能，但为后续所有工作奠定可靠基础。

### 0.1 修复已知 Bug

| 任务 | 文件 | 说明 |
|------|------|------|
| 修复 colorspace 释放 | `cgo_bindings/bindings.c:107` | `fz_drop_colorspace` 不应释放 `fz_device_rgb` 设备颜色空间，会导致 use-after-free |
| 修复 char_count 指针运算 | `cgo_bindings/bindings.c:260-272` | 用 `ch->next` 替代手动指针偏移 `(char*)ch + sizeof(void*)` |
| 修复 fz_output 泄漏 | `cgo_bindings/bindings.c:153-204` | JPEG 保存和文本提取路径缺少 `fz_drop_output` |
| 修复 Metadata 返回空 | `cgo_bindings/bindings.c:53-63` | `pdf_metadata` 结果被丢弃，永远返回空字符串 |
| 修复 page_rotation 硬编码 | `cgo_bindings/bindings.c:84` | 应调用 `fz_page_rotation` 而非返回 0 |
| 修复 OpenPDFStream 不初始化 fitz 字段 | `pdf/api.go` | 流式打开后 GetPage/IsScannedMode 会静默失败 |
| 修复 GetPixmap alpha 处理 | `pdf/api.go:173-211` | 硬编码 alpha=255，4 通道时按 3 通道步进导致图像损坏 |
| 修复 90 度旋转坐标计算 | `pdf/position.go:61` | 未考虑 pageWidth/pageHeight |
| 修复 SaveTo 读源文件而非内存 | `pdf/merge_split.go:143-154` | 应保存当前文档而非重读磁盘文件 |

### 0.2 增加 MuPDF 异常保护

| 任务 | 说明 |
|------|------|
| 在所有 C wrapper 中增加 `fz_try`/`fz_catch` | 当前任何 MuPDF 异常都会导致进程崩溃而非返回 Go error |
| 统一错误返回机制 | 定义错误码或使用 Go error 包裹 MuPDF 异常信息 |

### 0.3 激活已有但未接入的 C 绑定

已有 29 个 C wrapper 无 Go 入口，需逐一在 `cgo_bindings/*.go` 中暴露：

| C 函数组 | Go 文件 | 数量 |
|---------|---------|------|
| 颜色空间操作（gray/cmyk/find/drop） | `context.go` 或新文件 | 5 |
| 矩阵操作（identity/make/scale/rotate/translate/concat/invert） | 新文件 `matrix.go` | 7 |
| 几何变换（transform_point/rect/quad） | 新文件 `geometry.go` | 3 |
| 几何构造（make_rect/irect/point/quad, rect_is_empty/infinite） | 新文件 `geometry.go` | 7 |
| 图像信息（width/height/n/bpc/colorspace_name） | `pixmap.go` 或新文件 | 5 |
| 文本块图像提取（stext_block_get_image） | `text.go` | 1 |
| 设备颜色空间查找（find_device_colorspace） | `pixmap.go` | 1 |

### 0.4 建立测试框架

| 任务 | 说明 |
|------|------|
| 创建 `testdata/` 目录 | 放入测试用 PDF（加密、多页、含注释、含表单等） |
| 编写 `cgo_bindings/context_test.go` | 测试 Context 创建/销毁、版本号 |
| 编写 `cgo_bindings/document_test.go` | 测试打开/关闭/页面计数/元数据 |
| 编写 `cgo_bindings/page_test.go` | 测试加载/矩形/旋转 |
| 编写 `cgo_bindings/pixmap_test.go` | 测试渲染/保存 PNG/JPEG |
| 编写 `cgo_bindings/text_test.go` | 测试文本提取/块迭代 |
| 编写 `fitz/` 包测试 | 端到端测试每个 fitz 类型 |
| 编写 `pdf/` 包测试 | 测试高级 API |
| 配置 CI（GitHub Actions） | go test + go vet + golangci-lint |

---

## 阶段 1：核心文档操作（Document 完整化）

> 目标：Document 类达到 PyMuPDF 基本对等。

### 1.1 PDF 保存能力

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| 新增 C wrapper `gomupdf_save_document` | `pdf_save_document` | 保存 PDF 到文件 |
| 新增 C wrapper `gomupdf_save_document_to_buffer` | `pdf_save_document` + buffer | 保存 PDF 到内存 |
| 在 `cgo_bindings/document.go` 暴露 Save/SaveToBuffer | | CGO 入口 |
| 在 `fitz/doc.go` 实现 `Document.Save(opts)` | | Go 高级 API |
| 支持 SaveOptions（deflate, clean, garbage, encryption） | `fz_write_options` | 保存选项 |
| 在 `pdf/api.go` 修复 `PDFDocument.Save` | | 接入 fitz 层 |

### 1.2 页面管理

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| `Document.NewPage(pno, width, height)` | 无直接 API，需操作 PDF 对象 | 新建空白页 |
| `Document.DeletePage(pno)` | `pdf_delete_page` | 删除页面 |
| `Document.CopyPage(pno, to)` | `pdf_copy_page` 或 grmap 方式 | 复制页面 |
| `Document.MovePage(pno, to)` | `pdf_move_page` | 移动页面 |
| `Document.Select(pages)` | `pdf_select_pages` 或重排页面树 | 选择性保留页面 |
| `Document.InsertPDF(src, from, to, start_at)` | `pdf_insert_page` + graftmap | 插入其他 PDF 的页面 |

### 1.3 元数据与属性

| 任务 | 说明 |
|------|------|
| 修复 Metadata 提取 | 使 `pdf_metadata` 正确返回 format/title/author 等字段 |
| `Document.SetMetadata(m)` | 设置文档元数据 |
| `Document.PageCount` 属性化 | 改为缓存值 + 属性访问 |
| `Document.IsPDF/IsEncrypted/NeedsPassword` 完善 | 独立实现而非互相委托 |
| `Document.Permissions` | 读取 PDF 权限标志 |
| `Document.Version` | PDF 版本号 |
| `Document.IsDirty` | 文档是否被修改 |

### 1.4 加密与权限

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Document.Encrypt(owner_pw, user_pw, permissions)` | `pdf_encrypt_document` |
| 支持加密级别（RC4-40/128, AES-128/256） | `PDF_ENCRYPT_*` 常量 |
| `Document.Permissions` 返回详细权限 | `pdf_document_permissions` |

---

## 阶段 2：文本处理完整化

> 目标：文本提取支持所有 PyMuPDF 输出格式。

### 2.1 TextPage 多格式输出

| 任务 | 说明 | 优先级 |
|------|------|--------|
| `TextPage.ExtractText()` | 纯文本（已实现） | 已完成 |
| `TextPage.ExtractHTML()` | HTML 格式 | 高 |
| `TextPage.ExtractXML()` | XML 格式 | 高 |
| `TextPage.ExtractXHTML()` | XHTML 格式 | 中 |
| `TextPage.ExtractDICT()` | 结构化字典 | 高 |
| `TextPage.ExtractRAWDICT()` | 原始字典 | 中 |
| `TextPage.ExtractJSON()` | JSON 格式 | 中 |
| `TextPage.ExtractBLOCKS()` | 块格式（已实现） | 已完成 |
| `TextPage.ExtractWORDS()` | 词列表 | 高 |

需要新增 C wrapper：
- `gomupdf_stext_page_to_text`（已有）
- `gomupdf_stext_page_to_html` → `fz_print_stext_page_as_html`
- `gomupdf_stext_page_to_xml` → `fz_print_stext_page_as_xml`
- `gomupdf_stext_page_to_xhtml` → `fz_print_stext_page_as_xhtml`
- `gomupdf_stext_page_to_json` → `fz_print_stext_page_as_json`

### 2.2 文本搜索

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_search_text` | `fz_search_stext_page` | 在 TextPage 中搜索文本 |
| `Page.SearchFor(text)` → `[]Quad` | | 返回所有匹配位置的 Quad |
| `Page.SearchPageFor(pno, text)` | | Document 级别便捷方法 |

### 2.3 文本选择

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| `TextPage.ExtractSelection(p1, p2)` | `fz_copy_selection` / `fz_highlight_selection` | 两点间文本选择 |
| `TextPage.ExtractTextbox(rect)` | `fz_copy_selection` + clip | 矩形区域文本 |

### 2.4 修复 span 级迭代

| 任务 | 说明 |
|------|------|
| 修复 `gomupdf_stext_line_get_span` | 当前返回 NULL，需遍历 line->first_span |
| 新增 `TextPage.SpanBbox(span)` | span 边界框 |
| 新增 `TextPage.SpanText(span)` | span 文本内容 |
| 新增 `TextPage.SpanFont(span)` | span 字体名称 |
| 新增 `TextPage.SpanSize(span)` | span 字号 |
| 新增 `TextPage.SpanColor(span)` | span 颜色 |
| 在 fitz 层暴露完整的 TextFragment 信息 | 填充 Color/Flags 字段 |

---

## 阶段 3：图像处理完整化

> 目标：图像提取、插入、替换全部可用。

### 3.1 图像提取

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_stext_block_get_image` | 已有 wrapper，需 Go 入口 | 从文本块提取图像 |
| C wrapper `gomupdf_page_get_images` | 遍历页面资源字典 | 获取页面所有图像引用 |
| `Document.ExtractImage(xref)` | `fz_compressed_image_data` | 按 xref 提取原始图像 |
| `Document.ExtractFont(xref)` | 字体提取 | 提取嵌入字体 |
| `Page.GetImages(full)` | | 页面级图像列表 |

### 3.2 图像插入

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_page_insert_image` | `fz_new_image_from_buffer` + 内容流操作 | 插入图像到指定矩形 |
| `Page.InsertImage(rect, filename/stream/pixmap, ...)` | | Go API |
| `Page.ReplaceImage(xref, pixmap, ...)` | 替换已有图像 | |
| `Page.DeleteImage(xref)` | 删除图像 | |

### 3.3 Pixmap 增强

| 任务 | 说明 |
|------|------|
| `Pixmap.Pixel(x, y)` | 获取单像素值 |
| `Pixmap.SetPixel(x, y, color)` | 设置单像素 |
| `Pixmap.SetRect(rect, color)` | 矩形填色 |
| `Pixmap.ClearWith(value)` | 整体清空 |
| `Pixmap.CopyFrom(src, rect)` | 区域复制 |
| `Pixmap.GammaWith(gamma)` | 伽马校正 |
| `Pixmap.InvertIRect(rect)` | 反色 |
| `Pixmap.TintWith(black, white)` | 着色 |
| `Pixmap.Shrink(factor)` | 缩小 |
| `Pixmap.ColorCount()` | 颜色计数 |
| `Pixmap.Tobytes()` | 内存中编码为 PNG/JPEG 字节（不经过临时文件） |
| 支持多种颜色空间（GRAY, CMYK） | 当前仅 RGB |
| 支持 alpha 通道 | 当前强制忽略 |

### 3.4 SVG 输出

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| C wrapper `gomupdf_page_to_svg` | `fz_new_svg_writer` / `fz_print_stext_page_as_svg` |
| `Page.GetSVGImage()` → string | 页面转 SVG |

---

## 阶段 4：注释系统（Annot）

> 目标：支持所有主要注释类型。这是 PyMuPDF 最庞大的子系统之一（~1,400 行）。

### 4.1 注释基础设施

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_page_first_annot` | `pdf_first_annot` | 获取首个注释 |
| C wrapper `gomupdf_annot_next` | `pdf_next_annot` | 遍历注释 |
| C wrapper `gomupdf_annot_type` | `pdf_annot_type` | 注释类型 |
| C wrapper `gomupdf_annot_rect` | `pdf_annot_rect` | 注释矩形 |
| C wrapper `gomupdf_update_annot` | `pdf_update_annot` | 更新注释 |
| C wrapper `gomupdf_delete_annot` | `pdf_delete_annot` | 删除注释 |
| 在 `cgo_bindings/` 新建 `annot.go` | CGO 封装 | |
| 在 `fitz/annot.go` 实现完整 `Annot` 类型 | Go 高级 API | |

### 4.2 文本标记注释

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Page.AddHighlightAnnot(quads)` | `PDF_ANNOT_HIGHLIGHT` |
| `Page.AddStrikeoutAnnot(quads)` | `PDF_ANNOT_STRIKEOUT` |
| `Page.AddUnderlineAnnot(quads)` | `PDF_ANNOT_UNDERLINE` |
| `Page.AddSquigglyAnnot(quads)` | `PDF_ANNOT_SQUIGGLY` |

### 4.3 文本/便签注释

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Page.AddTextAnnot(point, text, icon)` | `PDF_ANNOT_TEXT` |
| `Page.AddFreeTextAnnot(rect, text, ...)` | `PDF_ANNOT_FREE_TEXT` |

### 4.4 形状注释

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Page.AddLineAnnot(p1, p2)` | `PDF_ANNOT_LINE` |
| `Page.AddRectAnnot(rect)` | `PDF_ANNOT_SQUARE` / `PDF_ANNOT_SQUARE` |
| `Page.AddCircleAnnot(rect)` | `PDF_ANNOT_CIRCLE` |
| `Page.AddPolygonAnnot(points)` | `PDF_ANNOT_POLYGON` |
| `Page.AddPolylineAnnot(points)` | `PDF_ANNOT_POLYLINE` |

### 4.5 其他注释

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Page.AddInkAnnot(handwriting)` | `PDF_ANNOT_INK` |
| `Page.AddStampAnnot(rect, stamp)` | `PDF_ANNOT_STAMP`（14 种标准图章） |
| `Page.AddFileAnnot(point, data, filename)` | `PDF_ANNOT_FILE_ATTACHMENT` |
| `Page.AddCaretAnnot(point)` | `PDF_ANNOT_CARET` |
| `Page.AddRedactAnnot(quad, text, ...)` | `PDF_ANNOT_REDACT` |
| `Page.ApplyRedactions()` | `pdf_redact_page` |

### 4.6 注释属性操作

| 任务 | 说明 |
|------|------|
| `Annot.Rect()` / `SetRect()` | 注释矩形 |
| `Annot.Colors()` / `SetColors()` | 边框/填充颜色 |
| `Annot.Border()` / `SetBorder()` | 边框样式 |
| `Annot.Opacity()` / `SetOpacity()` | 透明度 |
| `Annot.Info()` / `SetInfo()` | 标题/内容/日期 |
| `Annot.Flags()` / `SetFlags()` | 注释标志 |
| `Annot.BlendMode()` / `SetBlendMode()` | 混合模式 |
| `Annot.Update()` | 更新注释外观 |
| `Annot.GetPixmap()` | 渲染注释为图像 |
| `Annot.GetText()` | 提取注释中的文本 |

---

## 阶段 5：链接与导航

### 5.1 链接操作

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_page_first_link` | `fz_load_links` | 获取首个链接 |
| C wrapper `gomupdf_link_next` | 遍历链接链表 | |
| C wrapper `gomupdf_link_rect/uri/dest` | 链接属性 | |
| `Page.GetLinks()` → `[]Link` | | 获取所有链接 |
| `Page.InsertLink(linkdict)` | `pdf_insert_link` | 插入链接 |
| `Page.DeleteLink(linkdict)` | `pdf_delete_link` | 删除链接 |
| `Page.UpdateLink(linkdict)` | `pdf_update_link` | 更新链接 |

### 5.2 目录/大纲

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_load_outline` | `fz_load_outline` | 加载大纲 |
| C wrapper 迭代 outline 节点 | `fz_outline_title/page/uri/next/down` | |
| `Document.GetTOC()` → `[]Outline` | | 获取目录树 |
| `Document.SetTOC(toc)` | `pdf_set_outline` | 设置目录 |
| `Document.DelTOCItem(idx)` | | 删除目录项 |

### 5.3 书签

| 任务 | 说明 |
|------|------|
| `Document.MakeBookmark(loc)` | 创建书签 |
| `Document.FindBookmark(bm)` | 查找书签 |

---

## 阶段 6：Shape 绘图系统

> 目标：完整的矢量绘图能力。

### 6.1 Shape 类

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| 新建 `fitz/shape.go` | | Shape 类型定义 |
| `Shape.DrawLine(p1, p2)` | 内容流写入 | |
| `Shape.DrawPolyline(points)` | | |
| `Shape.DrawBezier(p1, p2, p3)` | | |
| `Shape.DrawOval(quad)` | | |
| `Shape.DrawCircle(center, radius)` | | |
| `Shape.DrawCurve(p1, p2, p3)` | | |
| `Shape.DrawSector(center, point, beta)` | | |
| `Shape.DrawRect(rect)` | | |
| `Shape.DrawQuad(quad)` | | |
| `Shape.DrawZigzag(p1, p2, breadth)` | | |
| `Shape.DrawSquiggle(p1, p2, breadth)` | | |
| `Shape.Finish(color, fill, ...)` | 路径属性设置 | |
| `Shape.Commit(overlay, ...)` | 提交到页面内容流 | |

### 6.2 Page 绘图便捷方法

| 任务 | 说明 |
|------|------|
| `Page.DrawLine(p1, p2, ...)` | 页面级绘图快捷方法 |
| `Page.DrawRect(rect, ...)` | |
| `Page.DrawCircle(center, radius, ...)` | |
| `Page.DrawBezier(p1, p2, p3, ...)` | |
| ... 其他绘图方法 | 对应 PyMuPDF Page 的 draw_* 系列 |

### 6.3 矢量绘图提取

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Page.GetDrawings()` | 遍历内容流中的路径操作 |
| `Page.GetCDrawings()` | 扩展版绘图提取 |
| 修复 `extractSVGPositions` | 依赖此功能 |

---

## 阶段 7：文本写入与字体

### 7.1 Font 类

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_new_font` | `fz_new_font_from_buffer` | 从文件/缓冲区创建字体 |
| `fitz/font.go` 新文件 | Font 类型 | |
| `Font.TextLength(text, fontsize)` | `fz_measure_text` | 测量文本宽度 |
| `Font.CharLengths(text, fontsize)` | 逐字符宽度 | |
| `Font.GlyphAdvance(chr)` | `fz_advance_glyph` | 字符步进宽度 |
| `Font.GlyphBbox(chr)` | `fz_bound_glyph` | 字符边界框 |
| `Font.HasGlyph(chr)` | `fz_encode_character` | 字符是否存在 |
| `Font.Name/Ascender/Descender/Bbox` | 属性访问 | |
| `Document.GetPageFonts(pno)` | 列出页面使用的字体 | |
| `Page.GetFonts()` | 页面级字体列表 | |
| `Page.InsertFont(name, file)` | 插入字体到页面 | |

### 7.2 TextWriter 类

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_new_text_writer` | `fz_new_text_writer` | 创建文本写入器 |
| `fitz/textwriter.go` 新文件 | TextWriter 类型 | |
| `TextWriter.Append(pos, text, font, fontsize)` | 追加文本 | |
| `TextWriter.FillTextbox(rect, text, font, ...)` | 文本填框 | |
| `TextWriter.WriteText(page, ...)` | 写入页面 | |

### 7.3 Page 文本插入

| 任务 | 说明 |
|------|------|
| `Page.InsertText(point, text, font, fontsize, ...)` | 插入文本 |
| `Page.InsertTextbox(rect, text, font, fontsize, ...)` | 文本填框 |
| `Page.InsertHTMLBox(rect, html, css, ...)` | HTML 文本框（依赖 Story） |

---

## 阶段 8：Widget 表单系统

### 8.1 Widget 基础设施

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_page_first_widget` | `pdf_first_widget` | 获取首个控件 |
| C wrapper `gomupdf_widget_next` | 遍历控件 | |
| C wrapper `gomupdf_widget_type` | 控件类型 | |
| C wrapper `gomupdf_update_widget` | `pdf_update_annot` | 更新控件 |
| 新建 `cgo_bindings/widget.go` | CGO 封装 | |
| 新建 `fitz/widget.go` | Widget 类型 | |

### 8.2 Widget 属性

| 任务 | 说明 |
|------|------|
| `Widget.FieldType()` | 字段类型（text/checkbox/radio/list/choice） |
| `Widget.FieldName()` | 字段名称 |
| `Widget.FieldValue()` / `SetFieldValue()` | 字段值 |
| `Widget.FieldFlags()` | 字段标志 |
| `Widget.Rect()` / `SetRect()` | 控件矩形 |
| `Widget.TextFont()` / `TextFontSize()` | 文本字体和大小 |
| `Widget.TextColor()` / `SetTextColor()` | 文本颜色 |
| `Widget.FillColor()` / `SetFillColor()` | 填充颜色 |
| `Widget.BorderColor()` / `SetBorderColor()` | 边框颜色 |
| `Widget.ChoiceValues()` | 选择类字段值 |
| `Widget.CheckboxChecked()` | 复选框状态 |
| `Page.AddWidget(widget)` | 添加控件 |
| `Page.DeleteWidget(widget)` | 删除控件 |
| `Page.LoadWidget(xref)` | 按 xref 加载控件 |
| `Document.IsFormPDF()` | 判断是否为表单 PDF |

---

## 阶段 9：高级功能

### 9.1 Story（HTML 转 PDF）

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_new_story` | `fz_new_story` | 创建 Story |
| C wrapper `gomupdf_story_place` | `fz_place_story` | 放置内容 |
| C wrapper `gomupdf_story_draw` | `fz_draw_story` | 绘制到设备 |
| 新建 `fitz/story.go` | Story 类型 | |
| `Story.Place(where)` | | |
| `Story.Draw(device, matrix)` | | |
| `Story.Write(writer, rectfn, ...)` | | |
| `Story.Body()` → Xml 节点 | DOM 访问 | |

### 9.2 Archive 与 Xml

| 任务 | 说明 |
|------|------|
| `Archive` 类型 | 资源归档（字体、图片等） |
| `Xml` 类型 | XML/HTML DOM 操作 |
| `DocumentWriter` 类型 | 文档写入器 |

### 9.3 DisplayList

| 任务 | 需要的 MuPDF API | 说明 |
|------|-----------------|------|
| C wrapper `gomupdf_new_display_list` | `fz_new_display_list` | 创建显示列表 |
| C wrapper `gomupdf_run_page_to_list` | `fz_run_page` → list | 页面渲染到列表 |
| `fitz/displaylist.go` | DisplayList 类型 | |
| `DisplayList.GetPixmap(matrix)` | 从缓存列表渲染 | |
| `DisplayList.GetTextPage()` | 从列表提取文本 | |

### 9.4 嵌入文件

| 任务 | 需要的 MuPDF API |
|------|-----------------|
| `Document.EmbfileAdd(name, data, filename, desc)` | `pdf_add_embedded_file` |
| `Document.EmbfileCount()` | |
| `Document.EmbfileGet(item)` | `pdf_load_embedded_file_contents` |
| `Document.EmbfileDel(item)` | |
| `Document.EmbfileInfo(item)` | |
| `Document.EmbfileNames()` | |
| `Document.EmbfileUpd(item, data)` | |

### 9.5 可选内容（图层）

| 任务 | 说明 |
|------|------|
| `Document.AddOCG(name, config, on, ...)` | 添加可选内容组 |
| `Document.AddLayer(name, creator, on)` | 添加图层 |
| `Document.GetOCGs()` | 获取图层信息 |
| `Document.GetLayers()` / `SetLayer()` | 图层操作 |
| `Document.SwitchLayer(config)` | 切换图层可见性 |
| `Page.GetOCItems()` | 页面级可选内容项 |

### 9.6 页面属性操作

| 任务 | 说明 |
|------|------|
| `Page.SetRotation(angle)` | 设置页面旋转 |
| `Page.SetCropBox(rect)` | 设置裁剪框 |
| `Page.SetMediaBox(rect)` | 设置媒体框 |
| `Page.CropBox/MediaBox/BleedBox/TrimBox/ArtBox` | 各种页面框属性 |
| `Page.CleanContents()` | 清理内容流 |
| `Page.GetContents()` / `SetContents()` | 内容流操作 |

### 9.7 XRef 低级操作

| 任务 | 说明 |
|------|------|
| `Document.XRefGetKey(xref, key)` | 获取对象字典键值 |
| `Document.XRefSetKey(xref, key, value)` | 设置对象字典键值 |
| `Document.XRefObject(xref)` | 获取对象源码 |
| `Document.XRefStream(xref)` | 获取流数据 |
| `Document.XRefLength()` | XRef 表长度 |
| `Document.XRefIsStream(xref)` | 判断是否为流对象 |
| `Document.GetNewXRef()` | 创建新 XRef |
| `Document.UpdateObject(xref, text)` | 更新对象定义 |
| `Document.UpdateStream(xref, data)` | 更新流 |

### 9.8 Journalling（日志/撤销）

| 任务 | 说明 |
|------|------|
| `Document.JournalEnable()` | 启用操作日志 |
| `Document.JournalStartOp(name)` | 开始操作 |
| `Document.JournalStopOp()` | 结束操作 |
| `Document.JournalUndo()` | 撤销 |
| `Document.JournalRedo()` | 重做 |
| `Document.JournalSave(filename)` | 保存日志 |

---

## 阶段 10：OCR 与表格

### 10.1 OCR 集成

| 任务 | 说明 |
|------|------|
| Tesseract CGO 绑定 | 调用 Tesseract C API |
| `Page.GetTextPageOCR(...)` | OCR 文本提取 |
| `Pixmap.PDFOCRSave(filename, language)` | 保存为可搜索 PDF |
| `Pixmap.PDFOCRTobytes(language)` | OCR PDF 字节 |

### 10.2 表格检测

| 任务 | 说明 |
|------|------|
| 纯算法实现，无 MuPDF 依赖 | 参照 PyMuPDF 的 `table.py`（2,730 行） |
| `Page.FindTables(**kwargs)` | 检测页面中的表格 |
| `Table.Extract()` | 提取表格数据 |
| `Table.ToMarkdown()` | 转 Markdown |
| `TableHeader` 类型 | 表头识别 |

---

## 阶段 11：跨平台与发布

### 11.1 跨平台支持

| 任务 | 说明 |
|------|------|
| Linux 支持 | CGO 路径改为动态检测（pkg-config） |
| Windows 支持 | MSVC/MinGW 编译配置 |
| 去除 Homebrew 硬编码路径 | 使用 `pkg-config --cflags --libs mupdf` |
| 条件编译 | `#cgo` 指令区分平台 |

### 11.2 CLI 工具

| 任务 | 对应 PyMuPDF 子命令 |
|------|---------------------|
| `gomupdf clean input.pdf output.pdf` | `pymupdf clean` |
| `gomupdf join a.pdf b.pdf -o merged.pdf` | `pymupdf join` |
| `gomupdf extract input.pdf` | `pymupdf extract` |
| `gomupdf gettext input.pdf` | `pymupdf gettext` |
| `gomupdf show input.pdf` | `pymupdf show` |
| 嵌入文件操作子命令 | `embed-add/del/upd/extract/copy/info` |

### 11.3 文档与示例

| 任务 | 说明 |
|------|------|
| GoDoc 文档注释 | 所有导出类型和方法 |
| 完善每个阶段的示例程序 | 覆盖所有主要功能 |
| 中文文档 | README_CN.md 同步更新 |
| 迁移指南 | PyMuPDF → gomupdf 映射表 |

---

## 工作量估算与优先级

| 阶段 | 预估 Go 代码行数 | 预估 C 代码行数 | 优先级 | 依赖 |
|------|-----------------|----------------|--------|------|
| **0. 基础修复** | ~500 | ~200 | P0 | 无 |
| **1. Document 完整化** | ~1,500 | ~300 | P0 | 阶段 0 |
| **2. 文本完整化** | ~1,200 | ~200 | P0 | 阶段 0 |
| **3. 图像完整化** | ~1,000 | ~200 | P1 | 阶段 1 |
| **4. 注释系统** | ~2,000 | ~400 | P1 | 阶段 1 |
| **5. 链接导航** | ~800 | ~200 | P1 | 阶段 1 |
| **6. Shape 绘图** | ~1,200 | ~300 | P2 | 阶段 1 |
| **7. 字体与文本写入** | ~1,500 | ~300 | P2 | 阶段 2, 6 |
| **8. Widget 表单** | ~1,200 | ~300 | P2 | 阶段 4 |
| **9. 高级功能** | ~3,000 | ~600 | P3 | 阶段 1-6 |
| **10. OCR 与表格** | ~2,000 | ~200 | P3 | 阶段 2 |
| **11. 跨平台发布** | ~500 | ~100 | P3 | 阶段 0-10 |
| **合计** | **~16,400** | **~3,300** | | |

### 总量预估

- **当前**：~4,000 行 Go + ~600 行 C
- **目标**：~20,000 行 Go + ~4,000 行 C
- **测试代码**：~5,000 行 Go
- **约 5 倍代码量增长**

### 建议开发顺序

```
阶段 0（基础修复）
  ↓
阶段 1（Document）+ 阶段 2（文本）  ← 可并行
  ↓
阶段 3（图像）+ 阶段 4（注释）+ 阶段 5（链接）  ← 可并行
  ↓
阶段 6（Shape）+ 阶段 7（字体/TextWriter）
  ↓
阶段 8（Widget）
  ↓
阶段 9（高级功能）  ← 多个子系统可并行
  ↓
阶段 10（OCR/表格）+ 阶段 11（跨平台）
```

---

## 关键风险与挑战

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| MuPDF API 文档不全 | 开发效率低 | 参考 PyMuPDF SWIG 绑定逆向工程 |
| CGO 调用开销 | 页面加载慢 60x（2ms vs 0.03ms） | 批量操作减少跨边界调用 |
| 内存管理 | 手动 Destroy 易泄漏 | 添加 defer 模式 + finalizer 兜底 |
| 线程安全 | Mutex 锁粒度粗 | 考虑 context 池化 |
| Story/HTML-PDF 复杂度 | 可能需要完整的 XML/DOM 支持 | 分阶段实现，先支持基础 HTML |
| Tesseract CGO 绑定 | 额外 C 依赖 | 作为可选模块，编译标签控制 |
| MuPDF 版本更新 | API 可能变化 | 锁定 v1.27.2，逐步升级 |

---

## 里程碑定义

| 里程碑 | 覆盖率 | 可用场景 |
|--------|--------|---------|
| **M1 — 基础可用** | ~30% | 打开/保存/渲染/文本提取/基础元数据 |
| **M2 — 文档处理** | ~50% | PDF 合并/拆分/页面管理/加密/链接 |
| **M3 — 注释标注** | ~65% | 添加/修改/删除各类注释和绘图 |
| **M4 — 表单编辑** | ~75% | 表单填写、Widget 操作、文本写入 |
| **M5 — 高级功能** | ~90% | Story/OCR/表格/图层/嵌入文件 |
| **M6 — 完全对等** | ~95%+ | 跨平台/CLI/全部 API 对齐 |
