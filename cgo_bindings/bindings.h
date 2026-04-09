/*
 * gomupdf C 语言绑定头文件
 *
 * 本文件定义了 MuPDF C API 的 Go CGO 封装函数。
 * 所有函数以 gomupdf_ 为前缀，通过 CGO 暴露给 Go 层使用。
 *
 * 架构层次：MuPDF C 库 - bindings.c (本层) - cgo_bindings/go - fitz/go - pdf/go
 *
 * 关键设计：
 *   - 使用 fz_try/fz_catch 保护所有 MuPDF 调用，防止异常穿透 Go 栈
 *   - 通过线程局部变量 gomupdf_last_error 传递错误信息
 *   - 返回指针的函数在失败时返回 NULL，调用方需检查
 *   - 返回 int 的函数在失败时返回 -1，成功返回 0
 *
 * 线程安全：所有函数通过 Go 层的 Context.WithLock() 保证互斥访问
 */
#ifndef GOMUPDF_BINDINGS_H
#define GOMUPDF_BINDINGS_H

#include <mupdf/fitz.h>
#include <mupdf/pdf.h>

/* ============================================================
 * 上下文管理
 * MuPDF 的 fz_context 是所有操作的基础，包含内存分配器和异常栈
 * ============================================================ */
fz_context *gomupdf_new_context(void);
void gomupdf_drop_context(fz_context *ctx);

/* ============================================================
 * 文档操作
 * 支持打开文件、内存流、新建空白 PDF、元数据读取、密码认证
 * ============================================================ */
fz_document *gomupdf_open_document(fz_context *ctx, const char *filename);
fz_document *gomupdf_open_document_with_stream(fz_context *ctx, const char *type, unsigned char *data, size_t len);
fz_document *gomupdf_new_pdf_document(fz_context *ctx);
void gomupdf_drop_document(fz_context *ctx, fz_document *doc);
int gomupdf_page_count(fz_context *ctx, fz_document *doc);
int gomupdf_is_pdf(fz_context *ctx, fz_document *doc);
const char *gomupdf_document_metadata(fz_context *ctx, fz_document *doc, const char *key);
int gomupdf_needs_password(fz_context *ctx, fz_document *doc);
int gomupdf_authenticate_password(fz_context *ctx, fz_document *doc, const char *password);

/* ============================================================
 * 页面操作
 * 页面加载、边界查询、旋转角度（PDF 专用）、页面释放
 * ============================================================ */
fz_page *gomupdf_load_page(fz_context *ctx, fz_document *doc, int number);
fz_rect gomupdf_page_rect(fz_context *ctx, fz_page *page);
int gomupdf_page_rotation(fz_context *ctx, fz_page *page);
int gomupdf_pdf_page_rotation(fz_context *ctx, fz_document *doc, int page_num);
void gomupdf_drop_page(fz_context *ctx, fz_page *page);

/* ============================================================
 * Pixmap 基础操作
 * Pixmap 是 MuPDF 的像素缓冲区，用于渲染和图像处理
 * 包括：创建、渲染、尺寸查询、像素数据访问、文件保存
 * ============================================================ */
fz_pixmap *gomupdf_new_pixmap(fz_context *ctx, fz_colorspace *cs, int width, int height);
fz_pixmap *gomupdf_render_page(fz_context *ctx, fz_page *page, float a, float b, float c, float d, float e, float f, int alpha);
void gomupdf_drop_pixmap(fz_context *ctx, fz_pixmap *pix);
fz_colorspace *gomupdf_find_device_colorspace(fz_context *ctx, const char *name);
fz_colorspace *gomupdf_pixmap_colorspace(fz_context *ctx, fz_pixmap *pix);
int gomupdf_pixmap_width(fz_context *ctx, fz_pixmap *pix);
int gomupdf_pixmap_height(fz_context *ctx, fz_pixmap *pix);
int gomupdf_pixmap_stride(fz_context *ctx, fz_pixmap *pix);
int gomupdf_pixmap_n(fz_context *ctx, fz_pixmap *pix);
unsigned char *gomupdf_pixmap_samples(fz_context *ctx, fz_pixmap *pix);
void gomupdf_save_pixmap_as_png(fz_context *ctx, fz_pixmap *pix, const char *filename);
void gomupdf_save_pixmap_as_jpeg(fz_context *ctx, fz_pixmap *pix, const char *filename, int quality);

/* ============================================================
 * 文本提取
 * fz_stext_page 是 MuPDF 的结构化文本页面，包含块/行/字符层级
 * 支持纯文本提取和块级迭代
 * ============================================================ */
fz_stext_page *gomupdf_new_stext_page_from_page(fz_context *ctx, fz_page *page);
void gomupdf_drop_stext_page(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_text(fz_context *ctx, fz_stext_page *page);

/* 文本块迭代 — 文本块是页面的顶级结构单元 */
int gomupdf_stext_page_block_count(fz_context *ctx, fz_stext_page *page);
fz_stext_block *gomupdf_stext_page_get_block(fz_context *ctx, fz_stext_page *page, int idx);

/* 块类型：FZ_STEXT_BLOCK_TEXT 或 FZ_STEXT_BLOCK_IMAGE */
int gomupdf_stext_block_type(fz_context *ctx, fz_stext_block *block);

/* 块的边界矩形 */
fz_rect gomupdf_stext_block_bbox(fz_context *ctx, fz_stext_block *block);

/* 文本行迭代 — 每个文本块包含多行文本 */
int gomupdf_stext_block_line_count(fz_context *ctx, fz_stext_block *block);
fz_stext_line *gomupdf_stext_block_get_line(fz_context *ctx, fz_stext_block *block, int idx);

/* 行的边界矩形和书写方向 */
fz_rect gomupdf_stext_line_bbox(fz_context *ctx, fz_stext_line *line);
fz_point gomupdf_stext_line_dir(fz_context *ctx, fz_stext_line *line);

/* 字符迭代 — MuPDF 的 stext_char 通过链表连接（ch->next） */
int gomupdf_stext_line_char_count(fz_context *ctx, fz_stext_line *line);
fz_stext_char *gomupdf_stext_line_first_char(fz_context *ctx, fz_stext_line *line);

/* 字符属性：原点坐标、Unicode 码点、字号、边界框 */
fz_point gomupdf_stext_char_origin(fz_context *ctx, fz_stext_char *ch);
int gomupdf_stext_char_c(fz_context *ctx, fz_stext_char *ch);
float gomupdf_stext_char_size(fz_context *ctx, fz_stext_char *ch);
fz_rect gomupdf_stext_char_bbox(fz_context *ctx, fz_stext_char *ch);
fz_stext_char *gomupdf_stext_char_next(fz_context *ctx, fz_stext_char *ch);
fz_stext_line *gomupdf_stext_block_first_line(fz_context *ctx, fz_stext_block *block);
fz_stext_line *gomupdf_stext_line_next(fz_context *ctx, fz_stext_line *line);

/* ============================================================
 * 图像提取
 * 从文本块中提取嵌入图像，查询图像属性（宽高、分量数、位深、颜色空间）
 * ============================================================ */
fz_image *gomupdf_stext_block_get_image(fz_context *ctx, fz_stext_block *block);
int gomupdf_image_width(fz_context *ctx, fz_image *img);
int gomupdf_image_height(fz_context *ctx, fz_image *img);
int gomupdf_image_n(fz_context *ctx, fz_image *img);
int gomupdf_image_bpc(fz_context *ctx, fz_image *img);
const char *gomupdf_image_colorspace_name(fz_context *ctx, fz_image *img);

/* ============================================================
 * 颜色空间
 * 支持设备颜色空间：RGB、Gray、CMYK
 * 注意：fz_device_rgb 等返回借用引用，不可调用 fz_drop_colorspace
 * ============================================================ */
fz_colorspace *gomupdf_new_colorspace_rgb(fz_context *ctx);
fz_colorspace *gomupdf_new_colorspace_gray(fz_context *ctx);
fz_colorspace *gomupdf_new_colorspace_cmyk(fz_context *ctx);
void gomupdf_drop_colorspace(fz_context *ctx, fz_colorspace *cs);

/* ============================================================
 * 矩阵变换
 * 仿射矩阵 [a b c d e f] 用于缩放、旋转、平移、拼接、求逆
 * 配合点、矩形、四边形的变换使用
 * ============================================================ */
fz_matrix gomupdf_identity_matrix(void);
fz_matrix gomupdf_make_matrix(float a, float b, float c, float d, float e, float f);
fz_matrix gomupdf_scale_matrix(float sx, float sy);
fz_matrix gomupdf_rotate_matrix(float degrees);
fz_matrix gomupdf_translate_matrix(float tx, float ty);
fz_matrix gomupdf_concat_matrix(fz_matrix left, fz_matrix right);
fz_matrix gomupdf_invert_matrix(fz_context *ctx, fz_matrix matrix);
fz_point gomupdf_transform_point(fz_point p, fz_matrix m);
fz_rect gomupdf_transform_rect(fz_rect r, fz_matrix m);
fz_quad gomupdf_transform_quad(fz_quad q, fz_matrix m);

/* 矩形和整数矩形操作 */
fz_rect gomupdf_make_rect(float x0, float y0, float x1, float y1);
fz_irect gomupdf_make_irect(int x0, int y0, int x1, int y1);
int gomupdf_rect_is_empty(fz_rect r);
int gomupdf_rect_is_infinite(fz_rect r);

/* 点操作 */
fz_point gomupdf_make_point(float x, float y);

/* 四边形操作 — 用于文本搜索命中区域和注释 QuadPoints */
fz_quad gomupdf_make_quad(fz_point ul, fz_point ur, fz_point ll, fz_point lr);
fz_rect gomupdf_quad_rect(fz_quad q);

/* ============================================================
 * 工具函数
 * MuPDF 版本号、错误信息获取与清除
 * ============================================================ */
const char *gomupdf_version(void);
const char *gomupdf_get_last_error(void);
void gomupdf_clear_error(void);

/* ============================================================
 * PDF 保存操作
 * 支持全部 13 项 pdf_write_options 选项
 * 包括：垃圾回收、清理、压缩/解压、线性化、ASCII 编码、增量保存等
 * ============================================================ */
int gomupdf_pdf_save_document(fz_context *ctx, fz_document *doc, const char *filename,
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata);

/* PDF 写入内存缓冲区，调用方需用 gomupdf_free() 释放返回的数据 */
int gomupdf_pdf_write_document(fz_context *ctx, fz_document *doc,
    unsigned char **out_data, size_t *out_len,
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata);
void gomupdf_free(void *ptr);

/* ============================================================
 * PDF 页面管理
 * 插入空白页（指定 MediaBox 和旋转）、删除单页、删除页面范围
 * ============================================================ */
int gomupdf_pdf_insert_page(fz_context *ctx, fz_document *doc, int at, float x0, float y0, float x1, float y1, int rotation);
int gomupdf_pdf_delete_page(fz_context *ctx, fz_document *doc, int number);
int gomupdf_pdf_delete_page_range(fz_context *ctx, fz_document *doc, int start, int end);

/* 设置文档元数据（如 info:Title, info:Author 等） */
int gomupdf_pdf_set_metadata(fz_context *ctx, fz_document *doc, const char *key, const char *value);

/* ============================================================
 * PDF 大纲/目录
 * 大纲树被展平为线性数组，每项包含标题、页码、层级、URI
 * 先调用 outline_count 获取总数，再通过 outline_get 按索引读取
 * ============================================================ */
int gomupdf_pdf_outline_count(fz_context *ctx, fz_document *doc);
int gomupdf_pdf_outline_get(fz_context *ctx, fz_document *doc, int idx,
    const char **title, int *page, int *level, const char **uri, int *is_open);

/* ============================================================
 * PDF 链接
 * 页面内超链接的加载、遍历、属性查询（矩形、URI、目标页码）
 * ============================================================ */
fz_link *gomupdf_page_load_links(fz_context *ctx, fz_page *page);
void gomupdf_drop_link(fz_context *ctx, fz_link *link);
fz_link *gomupdf_link_next(fz_link *link);
fz_rect gomupdf_link_rect(fz_link *link);
const char *gomupdf_link_uri(fz_link *link);
int gomupdf_link_page(fz_context *ctx, fz_document *doc, fz_link *link);

/* 在结构化文本页面中搜索关键词，返回命中区域的四边形列表 */
int gomupdf_search_text(fz_context *ctx, fz_stext_page *page, const char *needle,
    int max_hits, fz_quad *hits);

/* 读取 PDF 文档权限标志（打印、复制、修改等） */
int gomupdf_pdf_permissions(fz_context *ctx, fz_document *doc);

/* ============================================================
 * 文本多格式输出
 * 将结构化文本页面导出为 HTML、XML、XHTML、JSON 格式
 * 字符级属性：字体名称和标志位
 * ============================================================ */
char *gomupdf_stext_page_to_html(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_to_xml(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_to_xhtml(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_to_json(fz_context *ctx, fz_stext_page *page);
const char *gomupdf_stext_char_font(fz_context *ctx, fz_stext_char *ch);
int gomupdf_stext_char_flags(fz_context *ctx, fz_stext_char *ch);

/* ============================================================
 * 图像处理
 * 图像转 Pixmap、Pixmap 编码为 PNG/JPEG 字节流
 * 像素级操作：读写像素、清空、反色、Gamma 校正、着色
 * ============================================================ */
fz_pixmap *gomupdf_image_get_pixmap(fz_context *ctx, fz_image *img);
int gomupdf_pixmap_to_png_bytes(fz_context *ctx, fz_pixmap *pix, unsigned char **out_data, size_t *out_len);
int gomupdf_pixmap_to_jpeg_bytes(fz_context *ctx, fz_pixmap *pix, int quality, unsigned char **out_data, size_t *out_len);
int gomupdf_pixmap_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y);
void gomupdf_pixmap_set_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y, unsigned int val);
void gomupdf_pixmap_clear_with(fz_context *ctx, fz_pixmap *pix, int value);
void gomupdf_pixmap_invert(fz_context *ctx, fz_pixmap *pix);
void gomupdf_pixmap_gamma(fz_context *ctx, fz_pixmap *pix, float gamma);
void gomupdf_pixmap_tint(fz_context *ctx, fz_pixmap *pix, int black, int white);

/* ============================================================
 * 注释系统
 * PDF 注释的完整 CRUD：创建、读取属性、修改、删除
 * 支持类型：文本、链接、自由文本、高亮、删除线、下划线、涂黑等
 * 属性包括：矩形、内容文本、颜色/透明度、标志、边框、标题、弹出窗口、QuadPoints
 * ============================================================ */
pdf_annot *gomupdf_pdf_first_annot(fz_context *ctx, fz_document *doc, fz_page *page);
pdf_annot *gomupdf_pdf_next_annot(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_annot_type(fz_context *ctx, pdf_annot *annot);
fz_rect gomupdf_pdf_annot_rect(fz_context *ctx, pdf_annot *annot);
const char *gomupdf_pdf_annot_contents(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_set_annot_contents(fz_context *ctx, pdf_annot *annot, const char *text);
int gomupdf_pdf_annot_color(fz_context *ctx, pdf_annot *annot, float *r, float *g, float *b, float *a);
int gomupdf_pdf_set_annot_color(fz_context *ctx, pdf_annot *annot, float r, float g, float b, float a);
float gomupdf_pdf_annot_opacity(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_set_annot_opacity(fz_context *ctx, pdf_annot *annot, float opacity);
int gomupdf_pdf_annot_flags(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_set_annot_flags(fz_context *ctx, pdf_annot *annot, int flags);
float gomupdf_pdf_annot_border(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_set_annot_border(fz_context *ctx, pdf_annot *annot, float width);
int gomupdf_pdf_update_annot(fz_context *ctx, fz_document *doc, pdf_annot *annot);
int gomupdf_pdf_delete_annot(fz_context *ctx, fz_document *doc, fz_page *page, pdf_annot *annot);
pdf_annot *gomupdf_pdf_create_annot(fz_context *ctx, fz_document *doc, fz_page *page, int annot_type);
fz_quad *gomupdf_pdf_annot_quad_points(fz_context *ctx, pdf_annot *annot, int *count);
int gomupdf_pdf_set_annot_quad_points(fz_context *ctx, pdf_annot *annot, int count, fz_quad *quads);
int gomupdf_pdf_set_annot_rect(fz_context *ctx, pdf_annot *annot, float x0, float y0, float x1, float y1);
int gomupdf_pdf_set_annot_popup(fz_context *ctx, pdf_annot *annot, float x0, float y0, float x1, float y1);
fz_rect gomupdf_pdf_annot_popup(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_apply_redactions(fz_context *ctx, fz_document *doc, fz_page *page);
const char *gomupdf_pdf_annot_title(fz_context *ctx, pdf_annot *annot);
int gomupdf_pdf_set_annot_title(fz_context *ctx, pdf_annot *annot, const char *title);

/* ============================================================
 * 链接操作
 * 创建和删除页面内的超链接
 * ============================================================ */
int gomupdf_pdf_create_link(fz_context *ctx, fz_document *doc, fz_page *page,
    float x0, float y0, float x1, float y1, const char *uri, int page_num);
int gomupdf_pdf_delete_link(fz_context *ctx, fz_document *doc, fz_page *page, fz_link *link);

/* ============================================================
 * 内容流操作（Shape 绘图基础）
 * begin/end 模式：获取内容流缓冲区，写入绘图指令，然后更新到页面
 * ============================================================ */
fz_buffer *gomupdf_pdf_page_write_begin(fz_context *ctx, fz_document *doc, fz_page *page);
int gomupdf_pdf_page_write_end(fz_context *ctx, fz_document *doc, fz_page *page, fz_buffer *contents);

/* ============================================================
 * 字体操作
 * 从文件或内存缓冲区加载字体，查询字体属性，测量文本宽度
 * 文本测量通过逐字符累加字形步进宽度实现（MuPDF v1.27 无 fz_measure_text）
 * ============================================================ */
fz_font *gomupdf_new_font_from_file(fz_context *ctx, const char *filename, int index);
fz_font *gomupdf_new_font_from_buffer(fz_context *ctx, const char *data, size_t len, int index);
void gomupdf_drop_font(fz_context *ctx, fz_font *font);
const char *gomupdf_font_name(fz_context *ctx, fz_font *font);
float gomupdf_font_ascender(fz_context *ctx, fz_font *font);
float gomupdf_font_descender(fz_context *ctx, fz_font *font);
float gomupdf_measure_text(fz_context *ctx, fz_font *font, const char *text, float size);
float gomupdf_font_glyph_advance(fz_context *ctx, fz_font *font, int glyph, float size);

/* ============================================================
 * Widget 表单系统
 * PDF 表单控件（文本框、复选框、单选按钮等）
 * 控件在 MuPDF 中复用 pdf_annot 类型，但使用专用的字段访问 API
 * ============================================================ */
pdf_annot *gomupdf_pdf_first_widget(fz_context *ctx, fz_document *doc, fz_page *page);
pdf_annot *gomupdf_pdf_next_widget(fz_context *ctx, pdf_annot *widget);
int gomupdf_pdf_widget_type(fz_context *ctx, pdf_annot *widget);
const char *gomupdf_pdf_widget_field_name(fz_context *ctx, pdf_annot *widget);
const char *gomupdf_pdf_widget_field_value(fz_context *ctx, pdf_annot *widget);
int gomupdf_pdf_widget_set_field_value(fz_context *ctx, fz_document *doc, pdf_annot *widget, const char *value);
int gomupdf_pdf_widget_field_flags(fz_context *ctx, pdf_annot *widget);
int gomupdf_pdf_widget_set_field_flags(fz_context *ctx, pdf_annot *widget, int flags);
int gomupdf_pdf_widget_is_checked(fz_context *ctx, pdf_annot *widget);
int gomupdf_pdf_widget_toggle(fz_context *ctx, pdf_annot *widget);

/* ============================================================
 * 高级功能
 *
 * DisplayList — 缓存的页面显示列表，支持多次高效渲染
 * 页面属性 — MediaBox、CropBox、旋转角度的读取和设置
 * XRef — 交叉引用表操作，用于底层 PDF 对象访问
 * 嵌入文件 — PDF 附件的读取和添加（通过 Names 字典迭代）
 * ============================================================ */

/* DisplayList：创建、页面渲染到列表、从列表渲染为 Pixmap */
fz_display_list *gomupdf_new_display_list(fz_context *ctx, fz_rect bounds);
void gomupdf_drop_display_list(fz_context *ctx, fz_display_list *list);
fz_display_list *gomupdf_run_page_to_list(fz_context *ctx, fz_page *page, fz_display_list *list, fz_matrix ctm);
fz_pixmap *gomupdf_display_list_get_pixmap(fz_context *ctx, fz_display_list *list, fz_matrix ctm, int alpha);

/* 页面属性：CropBox、MediaBox、旋转 */
fz_rect gomupdf_pdf_page_cropbox(fz_context *ctx, fz_document *doc, int page_num);
int gomupdf_pdf_set_page_cropbox(fz_context *ctx, fz_document *doc, int page_num, float x0, float y0, float x1, float y1);
fz_rect gomupdf_pdf_page_mediabox(fz_context *ctx, fz_document *doc, int page_num);
int gomupdf_pdf_set_page_mediabox(fz_context *ctx, fz_document *doc, int page_num, float x0, float y0, float x1, float y1);
int gomupdf_pdf_set_page_rotation(fz_context *ctx, fz_document *doc, int page_num, int rotation);

/* XRef 操作 */
int gomupdf_pdf_xref_length(fz_context *ctx, fz_document *doc);
const char *gomupdf_pdf_xref_get_key(fz_context *ctx, fz_document *doc, int xref, const char *key);
int gomupdf_pdf_xref_is_stream(fz_context *ctx, fz_document *doc, int xref);

/* 嵌入文件操作 */
int gomupdf_pdf_embedded_file_count(fz_context *ctx, fz_document *doc);
const char *gomupdf_pdf_embedded_file_name(fz_context *ctx, fz_document *doc, int idx);
unsigned char *gomupdf_pdf_embedded_file_get(fz_context *ctx, fz_document *doc, int idx, size_t *out_len);
int gomupdf_pdf_add_embedded_file(fz_context *ctx, fz_document *doc,
    const char *filename, const char *mimetype, const unsigned char *data, size_t len);

#endif /* GOMUPDF_BINDINGS_H */
