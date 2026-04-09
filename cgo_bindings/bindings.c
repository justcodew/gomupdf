/*
 * gomupdf CGO 绑定层 - MuPDF C API 的 Go CGO 桥接
 *
 * 本文件为 MuPDF 库提供 C 语言包装函数，供 Go 通过 CGO 调用。
 * 核心设计模式：
 *   1. 线程局部错误缓冲区（gomupdf_last_error）：用于在 fz_try/fz_catch 异常
 *      保护下捕获错误信息，因为 MuPDF 使用 setjmp/longjmp 实现异常机制，
 *      无法直接将错误信息返回给 Go 层。
 *   2. fz_try/fz_catch 异常保护：MuPDF 内部使用 longjmp 实现异常跳转，
 *      所有可能抛出异常的调用都必须包裹在 fz_try/fz_catch 中，
 *      否则 longjmp 会导致未定义行为（尤其在 CGO 环境中极为危险）。
 *   3. 所有函数在操作前清除旧错误，操作失败时将错误信息写入线程局部缓冲区。
 */

#include "bindings.h"
#include <string.h>
#include <time.h>

// ============================================================
// 第一阶段：上下文与错误管理 / Phase 1: Context & Error Management
// ============================================================

/*
 * 线程局部错误消息缓冲区（Thread-local error buffer）
 *
 * 设计原因：MuPDF 的 fz_try/fz_catch 基于 setjmp/longjmp 实现，
 * 异常发生时执行 longjmp 跳转，无法通过常规返回值传递错误消息。
 * 使用 __thread 修饰符确保每个线程拥有独立的缓冲区，
 * 避免 Go 并发调用时出现竞态条件。
 */
static __thread char gomupdf_last_error[512] = {0};

/* 获取最近一次操作的错误消息 */
const char *gomupdf_get_last_error(void) {
    return gomupdf_last_error;
}

/* 清除错误缓冲区，在每次新操作开始时调用 */
void gomupdf_clear_error(void) {
    gomupdf_last_error[0] = '\0';
}

/*
 * 自定义警告回调 - 空实现，用于抑制所有 MuPDF 警告输出
 *
 * 设计原因：MuPDF 默认将警告输出到 stderr，在 CGO 环境中
 * 这些警告会混入 Go 程序的标准输出，干扰正常日志。
 * 通过注册空回调函数静默丢弃所有警告。
 */
static void gomupdf_warning_callback(void *user, const char *message) {
    // Suppress all warnings - 静默丢弃所有警告
}

/* 创建新的 MuPDF 上下文，注册文档处理器并设置警告抑制回调 */
fz_context *gomupdf_new_context(void) {
    fz_context *ctx = fz_new_context(NULL, NULL, FZ_STORE_DEFAULT);
    if (ctx) {
        fz_register_document_handlers(ctx);
        fz_set_warning_callback(ctx, gomupdf_warning_callback, NULL);
    }
    return ctx;
}

/* 释放 MuPDF 上下文资源 */
void gomupdf_drop_context(fz_context *ctx) {
    if (ctx) {
        fz_drop_context(ctx);
    }
}

/*
 * ============================================================
 * 文档操作 / Document Operations
 * ============================================================
 */

/* 从文件路径打开文档，使用 fz_try/fz_catch 保护以捕获打开失败等异常 */
fz_document *gomupdf_open_document(fz_context *ctx, const char *filename) {
    fz_document *doc = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        doc = fz_open_document(ctx, filename);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to open document: %s", fz_caught_message(ctx));
        doc = NULL;
    }
    return doc;
}

/* 从内存数据流打开文档，支持通过内存中的 PDF 字节直接加载 */
fz_document *gomupdf_open_document_with_stream(fz_context *ctx, const char *type, unsigned char *data, size_t len) {
    fz_document *doc = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        fz_buffer *buf = fz_new_buffer_from_data(ctx, data, len);
        fz_stream *stm = fz_open_buffer(ctx, buf);
        doc = fz_open_document_with_stream(ctx, type, stm);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to open document from stream: %s", fz_caught_message(ctx));
        doc = NULL;
    }
    return doc;
}

/* 创建空白 PDF 文档，通过 pdf_create_document 生成 */
fz_document *gomupdf_new_pdf_document(fz_context *ctx) {
    fz_document *doc = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        doc = (fz_document *)pdf_create_document(ctx);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to create PDF: %s", fz_caught_message(ctx));
        doc = NULL;
    }
    return doc;
}

/* 关闭并释放文档资源，清理时静默忽略错误避免影响上层逻辑 */
void gomupdf_drop_document(fz_context *ctx, fz_document *doc) {
    if (ctx && doc) {
        fz_try(ctx) {
            fz_drop_document(ctx, doc);
        }
        fz_catch(ctx) {
            // Silently ignore errors during cleanup
        }
    }
}

/* 获取文档总页数，异常时返回 0 */
int gomupdf_page_count(fz_context *ctx, fz_document *doc) {
    int count = 0;
    fz_try(ctx) {
        count = fz_count_pages(ctx, doc);
    }
    fz_catch(ctx) {
        count = 0;
    }
    return count;
}

/* 判断文档是否为 PDF 格式（通过尝试转换为 pdf_document 指针） */
int gomupdf_is_pdf(fz_context *ctx, fz_document *doc) {
    return pdf_document_from_fz_document(ctx, doc) != NULL;
}

/* 查询文档元数据（如标题、作者等），使用静态缓冲区返回结果 */
const char *gomupdf_document_metadata(fz_context *ctx, fz_document *doc, const char *key) {
    if (!ctx || !doc || !key) return "";
    static char buf[256];
    int len = fz_lookup_metadata(ctx, doc, key, buf, sizeof(buf));
    if (len > 0) {
        return buf;
    }
    return "";
}

/* 检查文档是否需要密码才能访问 */
int gomupdf_needs_password(fz_context *ctx, fz_document *doc) {
    return fz_needs_password(ctx, doc);
}

/* 使用密码尝试解锁文档，成功返回 1 */
int gomupdf_authenticate_password(fz_context *ctx, fz_document *doc, const char *password) {
    return fz_authenticate_password(ctx, doc, password);
}

/*
 * ============================================================
 * 页面操作 / Page Operations
 * ============================================================
 */

/* 加载指定编号的页面，失败时返回 NULL 并设置错误信息 */
fz_page *gomupdf_load_page(fz_context *ctx, fz_document *doc, int number) {
    fz_page *page = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        page = fz_load_page(ctx, doc, number);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to load page %d: %s", number, fz_caught_message(ctx));
        page = NULL;
    }
    return page;
}

/* 获取页面边界矩形（MediaBox） */
fz_rect gomupdf_page_rect(fz_context *ctx, fz_page *page) {
    fz_rect rect = {0, 0, 0, 0};
    fz_try(ctx) {
        rect = fz_bound_page(ctx, page);
    }
    fz_catch(ctx) {
        // Return empty rect on error
    }
    return rect;
}

/*
 * 获取页面旋转角度（通用 fz_page 接口，暂返回 0）
 *
 * MuPDF 的 fz_page 抽象接口未直接暴露旋转属性，
 * 需要通过 PDF 层的 /Rotate 字典条目访问。
 * 此函数仅作为占位，Go 层应调用 gomupdf_pdf_page_rotation 获取实际值。
 */
int gomupdf_page_rotation(fz_context *ctx, fz_page *page) {
    (void)ctx;
    (void)page;
    return 0;
}

/* 获取 PDF 页面的实际旋转角度，通过读取页面对象的 /Rotate 字段实现 */
int gomupdf_pdf_page_rotation(fz_context *ctx, fz_document *doc, int page_num) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (pdf) {
        pdf_obj *page_obj = pdf_lookup_page_obj(ctx, pdf, page_num);
        if (page_obj) {
            return pdf_dict_get_int(ctx, page_obj, PDF_NAME(Rotate));
        }
    }
    return 0;
}

/* 释放页面资源 */
void gomupdf_drop_page(fz_context *ctx, fz_page *page) {
    if (ctx && page) {
        fz_drop_page(ctx, page);
    }
}

/*
 * ============================================================
 * 渲染与像素图操作 / Rendering & Pixmap Operations
 * ============================================================
 */

/* 创建新的像素图（pixmap），用于存储渲染结果 */
fz_pixmap *gomupdf_new_pixmap(fz_context *ctx, fz_colorspace *cs, int width, int height) {
    return fz_new_pixmap(ctx, cs, width, height, NULL, 0);
}

/*
 * 渲染页面到像素图
 *
 * 参数 a-f 构成变换矩阵（CTM），控制缩放、旋转、平移等。
 * 渲染流程：计算页面边界 -> 创建像素图 -> 创建绘图设备 -> 运行页面内容。
 * 注意：fz_device_rgb 返回的是借用引用（borrowed reference），不可释放。
 * 异常处理中需要按顺序清理 dev 和 pix，避免资源泄漏。
 */
fz_pixmap *gomupdf_render_page(fz_context *ctx, fz_page *page, float a, float b, float c, float d, float e, float f, int alpha) {
    fz_pixmap *pix = NULL;
    fz_device *dev = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        fz_matrix ctm = fz_make_matrix(a, b, c, d, e, f);
        fz_rect rect = fz_bound_page(ctx, page);
        rect = fz_transform_rect(rect, ctm); /* 用变换矩阵缩放边界矩形 */
        fz_irect irect = fz_round_rect(rect);
        fz_colorspace *cs = fz_device_rgb(ctx);
        pix = fz_new_pixmap(ctx, cs, irect.x1, irect.y1, NULL, alpha ? 1 : 0);
        fz_clear_pixmap(ctx, pix);
        dev = fz_new_draw_device(ctx, fz_identity, pix);
        fz_run_page(ctx, page, dev, ctm, NULL);
        fz_drop_device(ctx, dev);
        dev = NULL;
        // Note: fz_device_rgb returns a borrowed reference, must NOT be dropped
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to render page: %s", fz_caught_message(ctx));
        if (dev) fz_drop_device(ctx, dev);
        if (pix) fz_drop_pixmap(ctx, pix);
        pix = NULL;
    }
    return pix;
}

/* 释放像素图资源 */
void gomupdf_drop_pixmap(fz_context *ctx, fz_pixmap *pix) {
    if (ctx && pix) {
        fz_drop_pixmap(ctx, pix);
    }
}

/* 按名称查找设备色彩空间（RGB/Gray/CMYK） */
fz_colorspace *gomupdf_find_device_colorspace(fz_context *ctx, const char *name) {
    if (strcmp(name, "RGB") == 0) return fz_device_rgb(ctx);
    if (strcmp(name, "Gray") == 0) return fz_device_gray(ctx);
    if (strcmp(name, "CMYK") == 0) return fz_device_cmyk(ctx);
    return NULL;
}

fz_colorspace *gomupdf_pixmap_colorspace(fz_context *ctx, fz_pixmap *pix) {
    return fz_pixmap_colorspace(ctx, pix);
}

int gomupdf_pixmap_width(fz_context *ctx, fz_pixmap *pix) {
    return fz_pixmap_width(ctx, pix);
}

int gomupdf_pixmap_height(fz_context *ctx, fz_pixmap *pix) {
    return fz_pixmap_height(ctx, pix);
}

int gomupdf_pixmap_stride(fz_context *ctx, fz_pixmap *pix) {
    return fz_pixmap_stride(ctx, pix);
}

int gomupdf_pixmap_n(fz_context *ctx, fz_pixmap *pix) {
    return fz_pixmap_components(ctx, pix);
}

unsigned char *gomupdf_pixmap_samples(fz_context *ctx, fz_pixmap *pix) {
    return fz_pixmap_samples(ctx, pix);
}

/* 将像素图保存为 PNG 文件 */
void gomupdf_save_pixmap_as_png(fz_context *ctx, fz_pixmap *pix, const char *filename) {
    fz_try(ctx) {
        fz_save_pixmap_as_png(ctx, pix, filename);
    }
    fz_catch(ctx) {
        // Error silently ignored — caller should check if file exists
    }
}

/*
 * 将像素图保存为 JPEG 文件
 *
 * 使用 fz_buffer + fz_output 中间层生成 JPEG 数据，
 * 再通过标准 C 文件 I/O 写入磁盘。
 * fz_always 块确保 output 和 buffer 资源总是被正确释放。
 */
void gomupdf_save_pixmap_as_jpeg(fz_context *ctx, fz_pixmap *pix, const char *filename, int quality) {
    fz_buffer *buf = NULL;
    fz_output *out = NULL;
    fz_try(ctx) {
        buf = fz_new_buffer(ctx, 4096);
        out = fz_new_output_with_buffer(ctx, buf);
        fz_write_pixmap_as_jpeg(ctx, out, pix, quality, 0);
        fz_close_output(ctx, out);
        FILE *f = fopen(filename, "wb");
        if (f) {
            fwrite(buf->data, 1, buf->len, f);
            fclose(f);
        }
    }
    fz_always(ctx) {
        if (out) fz_drop_output(ctx, out);
        if (buf) fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        // Error silently ignored
    }
}

/*
 * ============================================================
 * 结构化文本提取 / Structured Text Extraction
 *
 * 使用 fz_stext_page 提取页面中的文本内容及其位置、字体等信息。
 * 这是文本搜索、内容导出等功能的基础。
 * ============================================================
 */

/* 从页面创建结构化文本页，用于文本分析和搜索 */
fz_stext_page *gomupdf_new_stext_page_from_page(fz_context *ctx, fz_page *page) {
    fz_stext_page *stpage = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        stpage = fz_new_stext_page_from_page_with_cookie(ctx, page, NULL, NULL);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to create stext page: %s", fz_caught_message(ctx));
        stpage = NULL;
    }
    return stpage;
}

/* 释放结构化文本页资源 */
void gomupdf_drop_stext_page(fz_context *ctx, fz_stext_page *page) {
    if (ctx && page) {
        fz_drop_stext_page(ctx, page);
    }
}

/*
 * 从结构化文本页提取纯文本
 *
 * 遍历 block -> line -> char 层级结构，将每个字符转为 UTF-8 编码。
 * 跳过制表符和换行符，保留空格作为词分隔符。
 * 使用 fz_buffer + fz_output 构建结果字符串，避免频繁内存分配。
 * 调用方负责通过 gomupdf_free 释放返回的字符串。
 */
char *gomupdf_stext_page_text(fz_context *ctx, fz_stext_page *page) {
    fz_buffer *buf = fz_new_buffer(ctx, 256);
    fz_output *out = fz_new_output_with_buffer(ctx, buf);

    fz_stext_page_block_iterator iter = fz_stext_page_block_iterator_begin(page);
    while (!fz_stext_page_block_iterator_eod(iter)) {
        fz_stext_block *block = iter.block;
        if (block && block->type == FZ_STEXT_BLOCK_TEXT) {
            fz_stext_line *line = block->u.t.first_line;
            while (line) {
                fz_stext_char *ch = line->first_char;
                while (ch) {
                    if (ch->c != ' ' && ch->c != '\t' && ch->c != '\n') {
                        char utf8[8];
                        int len = fz_runetochar(utf8, ch->c);
                        fz_write_data(ctx, out, utf8, len);
                    } else if (ch->c == ' ') {
                        fz_write_data(ctx, out, " ", 1);
                    }
                    ch = ch->next;
                }
                line = line->next;
            }
        }
        iter = fz_stext_page_block_iterator_next(iter);
    }

    fz_write_byte(ctx, out, 0);
    fz_close_output(ctx, out);

    char *result = fz_strdup(ctx, (const char *)buf->data);
    fz_drop_output(ctx, out);
    fz_drop_buffer(ctx, buf);
    return result;
}

/* 获取文本块（block）数量，通过迭代器遍历计数 */
int gomupdf_stext_page_block_count(fz_context *ctx, fz_stext_page *page) {
    if (ctx && page) {
        int count = 0;
        fz_stext_page_block_iterator iter = fz_stext_page_block_iterator_begin(page);
        while (!fz_stext_page_block_iterator_eod(iter)) {
            count++;
            iter = fz_stext_page_block_iterator_next(iter);
        }
        return count;
    }
    return 0;
}

/* 按索引获取文本块指针 */
fz_stext_block *gomupdf_stext_page_get_block(fz_context *ctx, fz_stext_page *page, int idx) {
    if (ctx && page) {
        int i = 0;
        fz_stext_page_block_iterator iter = fz_stext_page_block_iterator_begin(page);
        while (!fz_stext_page_block_iterator_eod(iter)) {
            if (i == idx) return iter.block;
            i++;
            iter = fz_stext_page_block_iterator_next(iter);
        }
    }
    return NULL;
}

/* 获取文本块中的行（line）数量 */
int gomupdf_stext_block_line_count(fz_context *ctx, fz_stext_block *block) {
    if (ctx && block && block->type == FZ_STEXT_BLOCK_TEXT) {
        int count = 0;
        fz_stext_line *line = block->u.t.first_line;
        while (line) {
            count++;
            line = line->next;
        }
        return count;
    }
    return 0;
}

/* 按索引获取文本行指针 */
fz_stext_line *gomupdf_stext_block_get_line(fz_context *ctx, fz_stext_block *block, int idx) {
    if (ctx && block && block->type == FZ_STEXT_BLOCK_TEXT) {
        int i = 0;
        fz_stext_line *line = block->u.t.first_line;
        while (line) {
            if (i == idx) return line;
            i++;
            line = line->next;
        }
    }
    return NULL;
}

/* 获取文本行中的字符数量 */
int gomupdf_stext_line_char_count(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        int count = 0;
        fz_stext_char *ch = line->first_char;
        while (ch) {
            count++;
            ch = ch->next;
        }
        return count;
    }
    return 0;
}

/*
 * 获取文本行的第一个字符
 *
 * MuPDF 的 fz_stext_line 通过 first_char 链表直接包含字符，
 * 没有"span"概念（span 是 PyMuPDF 的抽象）。
 * Go 层应根据字符的字体/大小/样式自行分组以构建 span。
 */
fz_stext_char *gomupdf_stext_line_first_char(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->first_char;
    }
    return NULL;
}

/* 获取文本块类型（文本块或图像块） */
int gomupdf_stext_block_type(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block) {
        return block->type;
    }
    return -1;
}

/* 获取文本块的边界框 */
fz_rect gomupdf_stext_block_bbox(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block) {
        return block->bbox;
    }
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

/* 获取文本行的边界框 */
fz_rect gomupdf_stext_line_bbox(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->bbox;
    }
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

/* 获取文本行的书写方向向量 */
fz_point gomupdf_stext_line_dir(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->dir;
    }
    fz_point p = {0, 0};
    return p;
}

/* 获取字符的起点坐标 */
fz_point gomupdf_stext_char_origin(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->origin;
    }
    fz_point p = {0, 0};
    return p;
}

/* 获取字符的 Unicode 码点 */
int gomupdf_stext_char_c(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->c;
    }
    return 0;
}

/* 获取字符的字号 */
float gomupdf_stext_char_size(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->size;
    }
    return 0;
}

/* 获取字符的精确边界框（从 quad 四边形转换为矩形） */
fz_rect gomupdf_stext_char_bbox(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        // Use the quad directly for accurate char bbox
        return fz_rect_from_quad(ch->quad);
    }
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

/* 获取链表中的下一个字符 */
fz_stext_char *gomupdf_stext_char_next(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->next;
    }
    return NULL;
}

/* 获取文本块的第一行 */
fz_stext_line *gomupdf_stext_block_first_line(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block && block->type == FZ_STEXT_BLOCK_TEXT) {
        return block->u.t.first_line;
    }
    return NULL;
}

/* 获取链表中的下一行 */
fz_stext_line *gomupdf_stext_line_next(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->next;
    }
    return NULL;
}

/* 获取图像块中的图像对象 */
fz_image *gomupdf_stext_block_get_image(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block && block->type == FZ_STEXT_BLOCK_IMAGE) {
        return block->u.i.image;
    }
    return NULL;
}

/* 获取图像宽度（像素） */
int gomupdf_image_width(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->w;
    return 0;
}

/* 获取图像高度（像素） */
int gomupdf_image_height(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->h;
    return 0;
}

/* 获取图像颜色分量数（如 RGB=3, CMYK=4） */
int gomupdf_image_n(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->n;
    return 0;
}

/* 获取图像每分量位数（bits per component） */
int gomupdf_image_bpc(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->bpc;
    return 0;
}

/* 获取图像色彩空间名称 */
const char *gomupdf_image_colorspace_name(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img && img->colorspace) {
        return fz_colorspace_name(ctx, img->colorspace);
    }
    return "None";
}

/*
 * ============================================================
 * 色彩空间操作 / Colorspace Operations
 * ============================================================
 */

/* 获取设备 RGB 色彩空间（借用引用，无需释放） */
fz_colorspace *gomupdf_new_colorspace_rgb(fz_context *ctx) {
    return fz_device_rgb(ctx);
}

fz_colorspace *gomupdf_new_colorspace_gray(fz_context *ctx) {
    return fz_device_gray(ctx);
}

fz_colorspace *gomupdf_new_colorspace_cmyk(fz_context *ctx) {
    return fz_device_cmyk(ctx);
}

/* 释放色彩空间引用 */
void gomupdf_drop_colorspace(fz_context *ctx, fz_colorspace *cs) {
    if (ctx && cs) {
        fz_drop_colorspace(ctx, cs);
    }
}

/*
 * ============================================================
 * 矩阵与几何运算 / Matrix & Geometry Operations
 *
 * MuPDF 使用 2D 仿射变换矩阵 [a b c d e f] 表示缩放、旋转、平移等变换。
 * 这些函数提供了矩阵构造、组合、求逆及坐标变换等基础操作。
 * ============================================================
 */

/* 返回单位矩阵（无变换） */
fz_matrix gomupdf_identity_matrix(void) {
    return fz_identity;
}

/* 从 6 个分量构造变换矩阵 */
fz_matrix gomupdf_make_matrix(float a, float b, float c, float d, float e, float f) {
    return fz_make_matrix(a, b, c, d, e, f);
}

/* 构造缩放矩阵 */
fz_matrix gomupdf_scale_matrix(float sx, float sy) {
    return fz_scale(sx, sy);
}

/* 构造旋转矩阵（角度制） */
fz_matrix gomupdf_rotate_matrix(float degrees) {
    return fz_rotate(degrees);
}

/* 构造平移矩阵 */
fz_matrix gomupdf_translate_matrix(float tx, float ty) {
    return fz_translate(tx, ty);
}

/* 矩阵乘法（组合两个变换） */
fz_matrix gomupdf_concat_matrix(fz_matrix left, fz_matrix right) {
    return fz_concat(left, right);
}

/* 矩阵求逆 */
fz_matrix gomupdf_invert_matrix(fz_context *ctx, fz_matrix matrix) {
    (void)ctx;
    return fz_invert_matrix(matrix);
}

/* 使用矩阵变换点坐标 */
fz_point gomupdf_transform_point(fz_point p, fz_matrix m) {
    return fz_transform_point(p, m);
}

/* 使用矩阵变换矩形 */
fz_rect gomupdf_transform_rect(fz_rect r, fz_matrix m) {
    return fz_transform_rect(r, m);
}

/* 使用矩阵变换四边形 */
fz_quad gomupdf_transform_quad(fz_quad q, fz_matrix m) {
    return fz_transform_quad(q, m);
}

fz_rect gomupdf_make_rect(float x0, float y0, float x1, float y1) {
    return fz_make_rect(x0, y0, x1, y1);
}

fz_irect gomupdf_make_irect(int x0, int y0, int x1, int y1) {
    return fz_make_irect(x0, y0, x1, y1);
}

int gomupdf_rect_is_empty(fz_rect r) {
    return fz_is_empty_rect(r);
}

int gomupdf_rect_is_infinite(fz_rect r) {
    return fz_is_infinite_rect(r);
}

fz_point gomupdf_make_point(float x, float y) {
    return fz_make_point(x, y);
}

fz_quad gomupdf_make_quad(fz_point ul, fz_point ur, fz_point ll, fz_point lr) {
    return fz_make_quad(ul.x, ul.y, ur.x, ur.y, ll.x, ll.y, lr.x, lr.y);
}

fz_rect gomupdf_quad_rect(fz_quad q) {
    return fz_rect_from_quad(q);
}

/* 返回 MuPDF 版本号字符串 */
const char *gomupdf_version(void) {
    return FZ_VERSION;
}

// ============================================================
// PDF 保存操作 / PDF Save Operations
// ============================================================

/*
 * 构建 PDF 写入选项结构体
 *
 * 将分散的整数参数打包为 MuPDF 的 pdf_write_options 结构体，
 * 支持垃圾回收、清理、压缩、增量写入、线性化等选项。
 */
static pdf_write_options make_write_opts(
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata) {
    pdf_write_options opts = {0};
    opts.do_garbage = do_garbage;
    opts.do_clean = do_clean;
    opts.do_compress = do_compress;
    opts.do_compress_images = do_compress_images;
    opts.do_compress_fonts = do_compress_fonts;
    opts.do_decompress = do_decompress;
    opts.do_linear = do_linear;
    opts.do_ascii = do_ascii;
    opts.do_incremental = do_incremental;
    opts.do_pretty = do_pretty;
    opts.do_sanitize = do_sanitize;
    opts.do_appearance = do_appearance;
    opts.do_preserve_metadata = do_preserve_metadata;
    return opts;
}

/* 将 PDF 文档保存到文件 */
int gomupdf_pdf_save_document(fz_context *ctx, fz_document *doc, const char *filename,
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata) {
    if (!ctx || !doc || !filename) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;

    pdf_write_options opts = make_write_opts(
        do_garbage, do_clean, do_compress, do_compress_images, do_compress_fonts,
        do_decompress, do_linear, do_ascii, do_incremental, do_pretty,
        do_sanitize, do_appearance, do_preserve_metadata);

    gomupdf_clear_error();
    fz_try(ctx) {
        pdf_save_document(ctx, pdf, filename, &opts);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to save document: %s", fz_caught_message(ctx));
        return -1;
    }
    return 0;
}

/*
 * 将 PDF 文档写入内存缓冲区
 *
 * 使用 fz_buffer + fz_output 中间层生成 PDF 字节流，
 * 然后复制到 malloc 分配的内存中返回给 Go 层。
 * 调用方需通过 gomupdf_free 释放返回的数据。
 */
int gomupdf_pdf_write_document(fz_context *ctx, fz_document *doc,
    unsigned char **out_data, size_t *out_len,
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata) {
    if (!ctx || !doc || !out_data || !out_len) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;

    pdf_write_options opts = make_write_opts(
        do_garbage, do_clean, do_compress, do_compress_images, do_compress_fonts,
        do_decompress, do_linear, do_ascii, do_incremental, do_pretty,
        do_sanitize, do_appearance, do_preserve_metadata);

    fz_buffer *buf = NULL;
    fz_output *out = NULL;
    *out_data = NULL;
    *out_len = 0;

    gomupdf_clear_error();
    fz_try(ctx) {
        buf = fz_new_buffer(ctx, 4096);
        out = fz_new_output_with_buffer(ctx, buf);
        pdf_write_document(ctx, pdf, out, &opts);
        fz_close_output(ctx, out);
        // Copy buffer data for caller
        *out_len = buf->len;
        *out_data = (unsigned char *)malloc(buf->len);
        if (*out_data) {
            memcpy(*out_data, buf->data, buf->len);
        }
    }
    fz_always(ctx) {
        if (out) fz_drop_output(ctx, out);
        if (buf) fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to write document: %s", fz_caught_message(ctx));
        if (*out_data) { free(*out_data); *out_data = NULL; }
        *out_len = 0;
        return -1;
    }
    return 0;
}

/* 释放由 malloc 分配的内存，供 Go 层调用以释放 C 分配的缓冲区 */
void gomupdf_free(void *ptr) {
    if (ptr) free(ptr);
}

// ============================================================
// PDF 页面管理 / PDF Page Management
// ============================================================

/*
 * 在指定位置插入新页面
 *
 * 使用 pdf_page_write 创建页面写入设备，然后通过 pdf_add_page 构建页面对象，
 * 最后 pdf_insert_page 将其插入到文档的页面树中。
 * pdf_insert_page 会接管 page obj 的所有权。
 */
int gomupdf_pdf_insert_page(fz_context *ctx, fz_document *doc, int at, float x0, float y0, float x1, float y1, int rotation) {
    if (!ctx || !doc) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;

    gomupdf_clear_error();
    fz_try(ctx) {
        fz_rect mediabox = fz_make_rect(x0, y0, x1, y1);
        pdf_obj *resources = NULL;
        fz_buffer *contents = NULL;
        fz_device *dev = pdf_page_write(ctx, pdf, mediabox, &resources, &contents);
        fz_drop_device(ctx, dev);
        pdf_obj *page = pdf_add_page(ctx, pdf, mediabox, rotation, resources, contents);
        pdf_insert_page(ctx, pdf, at, page);
        // pdf_insert_page takes ownership of page obj
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to insert page: %s", fz_caught_message(ctx));
        return -1;
    }
    return 0;
}

/* 删除指定编号的页面 */
int gomupdf_pdf_delete_page(fz_context *ctx, fz_document *doc, int number) {
    if (!ctx || !doc) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;

    gomupdf_clear_error();
    fz_try(ctx) {
        pdf_delete_page(ctx, pdf, number);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to delete page: %s", fz_caught_message(ctx));
        return -1;
    }
    return 0;
}

/* 删除指定范围的页面（包含 start，不包含 end） */
int gomupdf_pdf_delete_page_range(fz_context *ctx, fz_document *doc, int start, int end) {
    if (!ctx || !doc) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;

    gomupdf_clear_error();
    fz_try(ctx) {
        pdf_delete_page_range(ctx, pdf, start, end);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to delete page range: %s", fz_caught_message(ctx));
        return -1;
    }
    return 0;
}

// ============================================================
// PDF 元数据设置 / PDF Set Metadata
// ============================================================

/* 设置文档元数据字段（如标题、作者、主题等） */
int gomupdf_pdf_set_metadata(fz_context *ctx, fz_document *doc, const char *key, const char *value) {
    if (!ctx || !doc || !key || !value) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;

    gomupdf_clear_error();
    fz_try(ctx) {
        fz_set_metadata(ctx, doc, key, value);
    }
    fz_catch(ctx) {
        snprintf(gomupdf_last_error, sizeof(gomupdf_last_error),
                 "failed to set metadata: %s", fz_caught_message(ctx));
        return -1;
    }
    return 0;
}

// ============================================================
// PDF 大纲/目录 / PDF Outline / TOC
//
// 将 MuPDF 的树形大纲结构扁平化为线性数组，便于 Go 层按索引访问。
// 使用线程局部存储缓存展开后的条目，避免频繁内存分配。
// ============================================================

/* 大纲条目结构体，包含标题、页码、层级和 URI 信息 */
typedef struct {
    char *title;
    int page;
    int level;
    char *uri;
    int is_open;
} gomupdf_outline_entry;

/* 线程局部大纲缓存，避免多线程竞争 */
static __thread gomupdf_outline_entry *g_outline_entries = NULL;
static __thread int g_outline_count = 0;

/*
 * 递归遍历大纲树，将节点收集到线性数组中
 *
 * 对于每个大纲节点：复制标题和 URI，通过 fz_resolve_link 解析页码，
 * 然后递归处理子节点（level+1）和兄弟节点。
 */
static void collect_outline(fz_context *ctx, fz_document *doc, fz_outline *outline, int level) {
    while (outline) {
        // Reallocate
        gomupdf_outline_entry *new_entries = (gomupdf_outline_entry *)realloc(
            g_outline_entries, (g_outline_count + 1) * sizeof(gomupdf_outline_entry));
        if (!new_entries) return;
        g_outline_entries = new_entries;

        gomupdf_outline_entry *entry = &g_outline_entries[g_outline_count];
        entry->title = outline->title ? fz_strdup(ctx, outline->title) : NULL;
        entry->uri = outline->uri ? fz_strdup(ctx, outline->uri) : NULL;
        entry->level = level;
        entry->is_open = outline->is_open;

        // Resolve page number from fz_location
        entry->page = -1;
        if (outline->uri) {
            fz_location loc = fz_resolve_link(ctx, doc, outline->uri, NULL, NULL);
            entry->page = fz_page_number_from_location(ctx, doc, loc);
        } else {
            // Try the page field directly (fz_location)
            entry->page = fz_page_number_from_location(ctx, doc, outline->page);
        }

        g_outline_count++;

        // Recurse into children
        if (outline->down) {
            collect_outline(ctx, doc, outline->down, level + 1);
        }

        outline = outline->next;
    }
}

/* 释放大纲缓存的内存，包括所有标题和 URI 字符串 */
static void free_outline_entries(fz_context *ctx) {
    if (g_outline_entries) {
        for (int i = 0; i < g_outline_count; i++) {
            if (g_outline_entries[i].title) fz_free(ctx, g_outline_entries[i].title);
            if (g_outline_entries[i].uri) fz_free(ctx, g_outline_entries[i].uri);
        }
        free(g_outline_entries);
        g_outline_entries = NULL;
    }
    g_outline_count = 0;
}

/*
 * 加载大纲并返回条目总数
 *
 * 首先释放旧的缓存数据，然后加载大纲树并递归展开为线性数组。
 * 使用 fz_always 确保 fz_drop_outline 总是被调用以释放大纲资源。
 */
int gomupdf_pdf_outline_count(fz_context *ctx, fz_document *doc) {
    if (!ctx || !doc) return 0;

    free_outline_entries(ctx);
    fz_outline *outline = NULL;

    fz_try(ctx) {
        outline = fz_load_outline(ctx, doc);
        if (outline) {
            collect_outline(ctx, doc, outline, 0);
        }
    }
    fz_always(ctx) {
        if (outline) fz_drop_outline(ctx, outline);
    }
    fz_catch(ctx) {
        // Return whatever we collected so far
    }
    return g_outline_count;
}

/* 按索引获取大纲条目的详细信息（标题、页码、层级、URI、展开状态） */
int gomupdf_pdf_outline_get(fz_context *ctx, fz_document *doc, int idx,
    const char **title, int *page, int *level, const char **uri, int *is_open) {
    if (!g_outline_entries || idx < 0 || idx >= g_outline_count) return -1;
    gomupdf_outline_entry *entry = &g_outline_entries[idx];
    *title = entry->title ? entry->title : "";
    *page = entry->page;
    *level = entry->level;
    *uri = entry->uri ? entry->uri : "";
    *is_open = entry->is_open;
    return 0;
}

// ============================================================
// PDF 链接操作 / PDF Links
// ============================================================

/* 加载页面中的所有链接 */
fz_link *gomupdf_page_load_links(fz_context *ctx, fz_page *page) {
    fz_link *links = NULL;
    fz_try(ctx) {
        links = fz_load_links(ctx, page);
    }
    fz_catch(ctx) {
        links = NULL;
    }
    return links;
}

/* 释放链接资源 */
void gomupdf_drop_link(fz_context *ctx, fz_link *link) {
    if (ctx && link) {
        fz_drop_link(ctx, link);
    }
}

/* 获取链表中的下一个链接 */
fz_link *gomupdf_link_next(fz_link *link) {
    if (link) return link->next;
    return NULL;
}

/* 获取链接的点击区域矩形 */
fz_rect gomupdf_link_rect(fz_link *link) {
    if (link) return link->rect;
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

/* 获取链接的 URI 字符串 */
const char *gomupdf_link_uri(fz_link *link) {
    if (link) return link->uri;
    return NULL;
}

/* 通过 URI 解析链接指向的页码 */
int gomupdf_link_page(fz_context *ctx, fz_document *doc, fz_link *link) {
    if (!ctx || !doc || !link || !link->uri) return -1;
    int pagenum = -1;
    fz_try(ctx) {
        fz_location loc = fz_resolve_link(ctx, doc, link->uri, NULL, NULL);
        pagenum = fz_page_number_from_location(ctx, doc, loc);
    }
    fz_catch(ctx) {
        pagenum = -1;
    }
    return pagenum;
}

// ============================================================
// PDF 文本搜索 / PDF Text Search
// ============================================================

/*
 * 在结构化文本页中搜索关键词
 *
 * 使用 fz_search_stext_page 进行文本搜索，返回匹配位置的四边形数组。
 * max_hits 限制最大匹配数量，防止缓冲区溢出。
 */
int gomupdf_search_text(fz_context *ctx, fz_stext_page *page, const char *needle, int max_hits, fz_quad *hits) {
    if (!ctx || !page || !needle || !hits || max_hits <= 0) return 0;
    int count = 0;
    fz_try(ctx) {
        count = fz_search_stext_page(ctx, page, needle, NULL, hits, max_hits);
    }
    fz_catch(ctx) {
        count = 0;
    }
    return count;
}

// ============================================================
// PDF 权限查询 / PDF Permissions
// ============================================================

/* 获取文档的权限标志位（打印、复制、修改等） */
int gomupdf_pdf_permissions(fz_context *ctx, fz_document *doc) {
    if (!ctx || !doc) return 0;
    int perm = 0;
    fz_try(ctx) {
        perm = pdf_document_permissions(ctx, pdf_document_from_fz_document(ctx, doc));
    }
    fz_catch(ctx) {
        perm = 0;
    }
    return perm;
}

// ============================================================
// 第二阶段：文本输出格式 / Phase 2: Text Output Formats
//
// 将结构化文本页导出为 HTML、XML、XHTML、JSON 等格式，
// 使用 fz_buffer + fz_output 中间层生成内容，再复制到 malloc 缓冲区返回。
// 调用方需通过 gomupdf_free 释放返回的字符串。
// ============================================================

/* 导出结构化文本为 HTML 格式 */
char *gomupdf_stext_page_to_html(fz_context *ctx, fz_stext_page *page) {
    if (!ctx || !page) return NULL;
    fz_buffer *buf = fz_new_buffer(ctx, 4096);
    fz_output *out = fz_new_output_with_buffer(ctx, buf);
    char *result = NULL;
    fz_try(ctx) {
        fz_print_stext_page_as_html(ctx, out, page, 0);
        fz_close_output(ctx, out);
        result = (char *)malloc(buf->len + 1);
        if (result) {
            memcpy(result, buf->data, buf->len);
            result[buf->len] = 0;
        }
    }
    fz_always(ctx) {
        fz_drop_output(ctx, out);
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        if (result) { free(result); result = NULL; }
    }
    return result;
}

/* 导出结构化文本为 XML 格式 */
char *gomupdf_stext_page_to_xml(fz_context *ctx, fz_stext_page *page) {
    if (!ctx || !page) return NULL;
    fz_buffer *buf = fz_new_buffer(ctx, 4096);
    fz_output *out = fz_new_output_with_buffer(ctx, buf);
    char *result = NULL;
    fz_try(ctx) {
        fz_print_stext_page_as_xml(ctx, out, page, 0);
        fz_close_output(ctx, out);
        result = (char *)malloc(buf->len + 1);
        if (result) {
            memcpy(result, buf->data, buf->len);
            result[buf->len] = 0;
        }
    }
    fz_always(ctx) {
        fz_drop_output(ctx, out);
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        if (result) { free(result); result = NULL; }
    }
    return result;
}

/* 导出结构化文本为 XHTML 格式 */
char *gomupdf_stext_page_to_xhtml(fz_context *ctx, fz_stext_page *page) {
    if (!ctx || !page) return NULL;
    fz_buffer *buf = fz_new_buffer(ctx, 4096);
    fz_output *out = fz_new_output_with_buffer(ctx, buf);
    char *result = NULL;
    fz_try(ctx) {
        fz_print_stext_page_as_xhtml(ctx, out, page, 0);
        fz_close_output(ctx, out);
        result = (char *)malloc(buf->len + 1);
        if (result) {
            memcpy(result, buf->data, buf->len);
            result[buf->len] = 0;
        }
    }
    fz_always(ctx) {
        fz_drop_output(ctx, out);
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        if (result) { free(result); result = NULL; }
    }
    return result;
}

/* 导出结构化文本为 JSON 格式，参数 1.0 控制坐标精度缩放 */
char *gomupdf_stext_page_to_json(fz_context *ctx, fz_stext_page *page) {
    if (!ctx || !page) return NULL;
    fz_buffer *buf = fz_new_buffer(ctx, 4096);
    fz_output *out = fz_new_output_with_buffer(ctx, buf);
    char *result = NULL;
    fz_try(ctx) {
        fz_print_stext_page_as_json(ctx, out, page, 1.0f);
        fz_close_output(ctx, out);
        result = (char *)malloc(buf->len + 1);
        if (result) {
            memcpy(result, buf->data, buf->len);
            result[buf->len] = 0;
        }
    }
    fz_always(ctx) {
        fz_drop_output(ctx, out);
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        if (result) { free(result); result = NULL; }
    }
    return result;
}

/* 获取字符所属的字体名称 */
const char *gomupdf_stext_char_font(fz_context *ctx, fz_stext_char *ch) {
    if (!ctx || !ch) return "";
    if (ch->font) {
        return fz_font_name(ctx, ch->font);
    }
    return "";
}

/* 获取字符的标志位（如粗体、斜体、上标、下标等） */
int gomupdf_stext_char_flags(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) return ch->flags;
    return 0;
}

// ============================================================
// 第三阶段：图像处理 / Phase 3: Image Processing
// ============================================================

/* 从图像对象获取像素图（解码图像数据） */
fz_pixmap *gomupdf_image_get_pixmap(fz_context *ctx, fz_image *img) {
    if (!ctx || !img) return NULL;
    fz_pixmap *pix = NULL;
    fz_try(ctx) {
        pix = fz_get_pixmap_from_image(ctx, img, NULL, NULL, NULL, NULL);
    }
    fz_catch(ctx) {
        pix = NULL;
    }
    return pix;
}

/* 将像素图编码为 PNG 格式的字节数组 */
int gomupdf_pixmap_to_png_bytes(fz_context *ctx, fz_pixmap *pix, unsigned char **out_data, size_t *out_len) {
    if (!ctx || !pix || !out_data || !out_len) return -1;
    fz_buffer *buf = NULL;
    fz_output *out = NULL;
    *out_data = NULL;
    *out_len = 0;
    fz_try(ctx) {
        buf = fz_new_buffer(ctx, 4096);
        out = fz_new_output_with_buffer(ctx, buf);
        fz_write_pixmap_as_png(ctx, out, pix);
        fz_close_output(ctx, out);
        *out_len = buf->len;
        *out_data = (unsigned char *)malloc(buf->len);
        if (*out_data) memcpy(*out_data, buf->data, buf->len);
    }
    fz_always(ctx) {
        if (out) fz_drop_output(ctx, out);
        if (buf) fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        if (*out_data) { free(*out_data); *out_data = NULL; }
        *out_len = 0;
        return -1;
    }
    return 0;
}

/* 将像素图编码为 JPEG 格式的字节数组，可指定压缩质量 (1-100) */
int gomupdf_pixmap_to_jpeg_bytes(fz_context *ctx, fz_pixmap *pix, int quality, unsigned char **out_data, size_t *out_len) {
    if (!ctx || !pix || !out_data || !out_len) return -1;
    fz_buffer *buf = NULL;
    fz_output *out = NULL;
    *out_data = NULL;
    *out_len = 0;
    fz_try(ctx) {
        buf = fz_new_buffer(ctx, 4096);
        out = fz_new_output_with_buffer(ctx, buf);
        fz_write_pixmap_as_jpeg(ctx, out, pix, quality, 0);
        fz_close_output(ctx, out);
        *out_len = buf->len;
        *out_data = (unsigned char *)malloc(buf->len);
        if (*out_data) memcpy(*out_data, buf->data, buf->len);
    }
    fz_always(ctx) {
        if (out) fz_drop_output(ctx, out);
        if (buf) fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) {
        if (*out_data) { free(*out_data); *out_data = NULL; }
        *out_len = 0;
        return -1;
    }
    return 0;
}

/* 获取像素图中指定位置的像素值（各通道字节拼接为整数） */
int gomupdf_pixmap_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y) {
    if (!ctx || !pix) return 0;
    int n = fz_pixmap_components(ctx, pix);
    unsigned char *samp = fz_pixmap_samples(ctx, pix) + (size_t)y * fz_pixmap_stride(ctx, pix) + (size_t)x * n;
    unsigned int val = 0;
    for (int i = 0; i < n; i++) val = (val << 8) | samp[i];
    return (int)val;
}

/* 设置像素图中指定位置的像素值 */
void gomupdf_pixmap_set_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y, unsigned int val) {
    if (!ctx || !pix) return;
    int n = fz_pixmap_components(ctx, pix);
    unsigned char *samp = fz_pixmap_samples(ctx, pix) + (size_t)y * fz_pixmap_stride(ctx, pix) + (size_t)x * n;
    for (int i = n - 1; i >= 0; i--) {
        samp[i] = val & 0xFF;
        val >>= 8;
    }
}

/* 使用指定值填充整个像素图（0xFF=白色，0x00=黑色） */
void gomupdf_pixmap_clear_with(fz_context *ctx, fz_pixmap *pix, int value) {
    if (!ctx || !pix) return;
    fz_clear_pixmap_with_value(ctx, pix, value);
}

/* 反转像素图颜色（底片效果） */
void gomupdf_pixmap_invert(fz_context *ctx, fz_pixmap *pix) {
    if (!ctx || !pix) return;
    fz_invert_pixmap(ctx, pix);
}

/* 对像素图应用伽马校正 */
void gomupdf_pixmap_gamma(fz_context *ctx, fz_pixmap *pix, float gamma) {
    if (!ctx || !pix) return;
    fz_gamma_pixmap(ctx, pix, gamma);
}

/* 对像素图应用着色效果（指定黑白两色） */
void gomupdf_pixmap_tint(fz_context *ctx, fz_pixmap *pix, int black, int white) {
    if (!ctx || !pix) return;
    fz_tint_pixmap(ctx, pix, black, white);
}

// ============================================================
// 第四阶段：注释系统 / Phase 4: Annotation System
//
// PDF 注释（Annotation）是附加在页面上的标记，如高亮、备注、链接等。
// 通过 pdf_first_annot/pdf_next_annot 遍历页面的注释链表。
// ============================================================

/* 获取页面的第一个注释对象 */
pdf_annot *gomupdf_pdf_first_annot(fz_context *ctx, fz_document *doc, fz_page *page) {
    if (!ctx || !doc || !page) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return NULL;
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pp) return NULL;
    return pdf_first_annot(ctx, pp);
}

/* 获取链表中的下一个注释 */
pdf_annot *gomupdf_pdf_next_annot(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return NULL;
    return pdf_next_annot(ctx, annot);
}

/* 获取注释类型枚举值（如高亮、文本、链接等） */
int gomupdf_pdf_annot_type(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return -1;
    return (int)pdf_annot_type(ctx, annot);
}

/* 获取注释的边界矩形 */
fz_rect gomupdf_pdf_annot_rect(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) { fz_rect r = {0,0,0,0}; return r; }
    fz_rect r = {0,0,0,0};
    fz_try(ctx) { r = pdf_annot_rect(ctx, annot); }
    fz_catch(ctx) {}
    return r;
}

/* 获取注释的内容文本，使用线程局部静态缓冲区存储结果 */
const char *gomupdf_pdf_annot_contents(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return "";
    static __thread char buf[1024];
    fz_try(ctx) {
        const char *c = pdf_annot_contents(ctx, annot);
        if (c) { snprintf(buf, sizeof(buf), "%s", c); }
        else { buf[0] = 0; }
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

/* 设置注释的内容文本 */
int gomupdf_pdf_set_annot_contents(fz_context *ctx, pdf_annot *annot, const char *text) {
    if (!ctx || !annot || !text) return -1;
    fz_try(ctx) { pdf_set_annot_contents(ctx, annot, text); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取注释的颜色分量（RGBA），返回分量数 n */
int gomupdf_pdf_annot_color(fz_context *ctx, pdf_annot *annot, float *r, float *g, float *b, float *a) {
    if (!ctx || !annot) return -1;
    int n = 0;
    float color[4] = {0,0,0,1};
    fz_try(ctx) {
        pdf_annot_color(ctx, annot, &n, color);
    }
    fz_catch(ctx) {}
    if (r) *r = color[0];
    if (g) *g = n > 1 ? color[1] : 0;
    if (b) *b = n > 2 ? color[2] : 0;
    if (a) *a = n > 3 ? color[3] : 1.0f;
    return n;
}

/* 设置注释颜色（RGBA 四分量） */
int gomupdf_pdf_set_annot_color(fz_context *ctx, pdf_annot *annot, float r, float g, float b, float a) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) {
        float color[4] = {r, g, b, a};
        pdf_set_annot_color(ctx, annot, 4, color);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取注释的不透明度（0.0-1.0） */
float gomupdf_pdf_annot_opacity(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return 1.0f;
    float opacity = 1.0f;
    fz_try(ctx) { opacity = pdf_annot_opacity(ctx, annot); }
    fz_catch(ctx) {}
    return opacity;
}

/* 设置注释的不透明度 */
int gomupdf_pdf_set_annot_opacity(fz_context *ctx, pdf_annot *annot, float opacity) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) { pdf_set_annot_opacity(ctx, annot, opacity); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取注释的标志位（如只读、隐藏、可打印等） */
int gomupdf_pdf_annot_flags(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return 0;
    int flags = 0;
    fz_try(ctx) { flags = pdf_annot_flags(ctx, annot); }
    fz_catch(ctx) {}
    return flags;
}

/* 设置注释的标志位 */
int gomupdf_pdf_set_annot_flags(fz_context *ctx, pdf_annot *annot, int flags) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) { pdf_set_annot_flags(ctx, annot, flags); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取注释的边框宽度 */
float gomupdf_pdf_annot_border(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return 0;
    float w = 0;
    fz_try(ctx) { w = pdf_annot_border(ctx, annot); }
    fz_catch(ctx) {}
    return w;
}

/* 设置注释的边框宽度 */
int gomupdf_pdf_set_annot_border(fz_context *ctx, pdf_annot *annot, float width) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) { pdf_set_annot_border(ctx, annot, width); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 更新注释对象（将内存中的修改同步到 PDF 对象） */
int gomupdf_pdf_update_annot(fz_context *ctx, fz_document *doc, pdf_annot *annot) {
    if (!ctx || !annot) return -1;
    (void)doc;
    fz_try(ctx) { pdf_update_annot(ctx, annot); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 从页面中删除指定注释 */
int gomupdf_pdf_delete_annot(fz_context *ctx, fz_document *doc, fz_page *page, pdf_annot *annot) {
    if (!ctx || !doc || !page || !annot) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return -1;
    fz_try(ctx) { pdf_delete_annot(ctx, pp, annot); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 在页面上创建指定类型的新注释 */
pdf_annot *gomupdf_pdf_create_annot(fz_context *ctx, fz_document *doc, fz_page *page, int annot_type) {
    if (!ctx || !doc || !page) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return NULL;
    pdf_annot *annot = NULL;
    fz_try(ctx) {
        annot = pdf_create_annot(ctx, pp, (enum pdf_annot_type)annot_type);
    }
    fz_catch(ctx) { annot = NULL; }
    return annot;
}

/* 获取注释的四边形点数组（用于高亮、删除线等），调用方需 free 返回值 */
fz_quad *gomupdf_pdf_annot_quad_points(fz_context *ctx, pdf_annot *annot, int *count) {
    if (!ctx || !annot || !count) return NULL;
    fz_quad *quads = NULL;
    fz_try(ctx) {
        int n = pdf_annot_quad_point_count(ctx, annot);
        *count = n;
        if (n > 0) {
            quads = (fz_quad *)malloc(n * sizeof(fz_quad));
            if (quads) {
                for (int i = 0; i < n; i++) {
                    quads[i] = pdf_annot_quad_point(ctx, annot, i);
                }
            }
        }
    }
    fz_catch(ctx) {
        if (quads) { free(quads); quads = NULL; }
        *count = 0;
    }
    return quads;
}

/* 设置注释的四边形点数组 */
int gomupdf_pdf_set_annot_quad_points(fz_context *ctx, pdf_annot *annot, int count, fz_quad *quads) {
    if (!ctx || !annot || !quads || count <= 0) return -1;
    fz_try(ctx) { pdf_set_annot_quad_points(ctx, annot, count, quads); }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 设置注释的边界矩形 */
int gomupdf_pdf_set_annot_rect(fz_context *ctx, pdf_annot *annot, float x0, float y0, float x1, float y1) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) {
        fz_rect r = fz_make_rect(x0, y0, x1, y1);
        pdf_set_annot_rect(ctx, annot, r);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 设置注释的弹出窗口矩形 */
int gomupdf_pdf_set_annot_popup(fz_context *ctx, pdf_annot *annot, float x0, float y0, float x1, float y1) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) {
        fz_rect r = fz_make_rect(x0, y0, x1, y1);
        pdf_set_annot_popup(ctx, annot, r);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取注释的弹出窗口矩形 */
fz_rect gomupdf_pdf_annot_popup(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) { fz_rect r = {0,0,0,0}; return r; }
    fz_rect r = {0,0,0,0};
    fz_try(ctx) { r = pdf_annot_popup(ctx, annot); }
    fz_catch(ctx) {}
    return r;
}

/* 应用页面上的涂黑（Redaction）注释，永久删除被标记的内容 */
int gomupdf_pdf_apply_redactions(fz_context *ctx, fz_document *doc, fz_page *page) {
    if (!ctx || !doc || !page) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        pdf_redact_page(ctx, pdf, (pdf_page *)pdf_page_from_fz_page(ctx, page), NULL);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/*
 * 获取注释标题
 *
 * 设计说明：MuPDF 没有提供 pdf_annot_title 函数。
 * 这里使用 pdf_annot_field_label 作为替代获取字段标签，
 * 因为注释的标题通常存储在 PDF 对象的 /T 字段中，
 * 而 pdf_annot_field_label 内部正是读取该字段。
 */
const char *gomupdf_pdf_annot_title(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return "";
    static __thread char buf[512];
    fz_try(ctx) {
        /* MuPDF 使用 pdf_annot_field_label 获取字段标签，
           注释标题需要从 PDF 对象的 /T 字段读取 */
        const char *t = pdf_annot_field_label(ctx, annot);
        if (t) snprintf(buf, sizeof(buf), "%s", t);
        else buf[0] = 0;
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

/* 设置注释标题，通过直接操作底层 PDF 对象的 /T 字段实现 */
int gomupdf_pdf_set_annot_title(fz_context *ctx, pdf_annot *annot, const char *title) {
    if (!ctx || !annot || !title) return -1;
    fz_try(ctx) {
        /* 通过底层 PDF 对象设置 /T 字段 */
        pdf_obj *obj = pdf_annot_obj(ctx, annot);
        if (obj) pdf_dict_put_text_string(ctx, obj, PDF_NAME(T), title);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

// ============================================================
// 第五阶段：链接创建与删除 / Phase 5: Link Operations
// ============================================================

/*
 * 在页面上创建新的超链接
 *
 * 使用 pdf_create_link 创建指向 URI 的链接。
 * 注意：page_num 参数当前未使用，链接目标由 uri 字符串决定。
 */
int gomupdf_pdf_create_link(fz_context *ctx, fz_document *doc, fz_page *page,
    float x0, float y0, float x1, float y1, const char *uri, int page_num) {
    if (!ctx || !doc || !page) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return -1;
    fz_try(ctx) {
        fz_rect link_rect = fz_make_rect(x0, y0, x1, y1);
        /* pdf_create_link 接受 URI 字符串参数 */
        fz_link *link = pdf_create_link(ctx, pp, link_rect, uri);
        if (link) fz_drop_link(ctx, link);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 从页面中删除指定链接 */
int gomupdf_pdf_delete_link(fz_context *ctx, fz_document *doc, fz_page *page, fz_link *link) {
    if (!ctx || !doc || !page || !link) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return -1;
    fz_try(ctx) { pdf_delete_link(ctx, pp, link); }
    fz_catch(ctx) { return -1; }
    return 0;
}

// ============================================================
// 第六阶段：图形绘制（内容流操作）/ Phase 6: Shape Drawing
//
// 提供页面内容流的写入接口，Go 层可向 buffer 中写入 PDF 操作符
// 来绘制线条、矩形、文本等图形元素。
// ============================================================

/*
 * 开始页面内容流写入，返回用于写入 PDF 操作符的缓冲区
 *
 * 使用 pdf_page_write 创建写入设备和缓冲区，Go 层向缓冲区
 * 写入 PDF 内容流操作符（如 "re" 画矩形、"BT...ET" 写文本等），
 * 最后通过 gomupdf_pdf_page_write_end 提交修改。
 */
fz_buffer *gomupdf_pdf_page_write_begin(fz_context *ctx, fz_document *doc, fz_page *page) {
    if (!ctx || !doc || !page) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return NULL;
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pp) return NULL;
    fz_buffer *buf = NULL;
    fz_try(ctx) {
        buf = fz_new_buffer(ctx, 256);
        pdf_obj *res = pdf_page_resources(ctx, pp);
        fz_device *dev = pdf_page_write(ctx, pdf, fz_bound_page(ctx, page), &res, &buf);
        fz_drop_device(ctx, dev);
    }
    fz_catch(ctx) { if (buf) { fz_drop_buffer(ctx, buf); buf = NULL; } }
    return buf;
}

/*
 * 结束页面内容流写入，将新内容应用到页面对象
 *
 * 使用 pdf_update_stream 将缓冲区中的内容流更新到页面的 PDF 对象中。
 */
int gomupdf_pdf_page_write_end(fz_context *ctx, fz_document *doc, fz_page *page,
    fz_buffer *contents) {
    if (!ctx || !doc || !page || !contents) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pp) return -1;
    fz_try(ctx) {
        /* 将新内容流写入页面对象 */
        pdf_update_stream(ctx, pdf, pp->obj, contents, 0);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

// ============================================================
// 第七阶段：字体操作 / Phase 7: Font Operations
// ============================================================

/* 从字体文件加载字体（支持 TTF、OTF、CFF 等格式），index 用于 TTC 集合中的字体索引 */
fz_font *gomupdf_new_font_from_file(fz_context *ctx, const char *filename, int index) {
    if (!ctx || !filename) return NULL;
    fz_font *font = NULL;
    fz_try(ctx) {
        font = fz_new_font_from_file(ctx, NULL, filename, index, 0);
    }
    fz_catch(ctx) { font = NULL; }
    return font;
}

/* 从内存缓冲区加载字体 */
fz_font *gomupdf_new_font_from_buffer(fz_context *ctx, const char *data, size_t len, int index) {
    if (!ctx || !data || len == 0) return NULL;
    fz_font *font = NULL;
    fz_try(ctx) {
        fz_buffer *buf = fz_new_buffer_from_data(ctx, (unsigned char *)data, len);
        font = fz_new_font_from_buffer(ctx, NULL, buf, index, 0);
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) { font = NULL; }
    return font;
}

/* 释放字体资源 */
void gomupdf_drop_font(fz_context *ctx, fz_font *font) {
    if (ctx && font) fz_drop_font(ctx, font);
}

/* 获取字体名称 */
const char *gomupdf_font_name(fz_context *ctx, fz_font *font) {
    if (!ctx || !font) return "";
    return fz_font_name(ctx, font);
}

/* 获取字体的上升线（ascender）高度 */
float gomupdf_font_ascender(fz_context *ctx, fz_font *font) {
    if (!ctx || !font) return 0;
    return fz_font_ascender(ctx, font);
}

/* 获取字体的下降线（descender）高度 */
float gomupdf_font_descender(fz_context *ctx, fz_font *font) {
    if (!ctx || !font) return 0;
    return fz_font_descender(ctx, font);
}

/*
 * 测量文本在指定字体和字号下的渲染宽度
 *
 * 设计说明：MuPDF v1.27 没有提供 fz_measure_text 便捷函数，
 * 因此这里采用逐字符测量（char-by-char measurement）的方式：
 * 对每个 Unicode 字符查找对应的字形（glyph），获取其水平步进值，
 * 累加后乘以字号得到文本总宽度。这种方式虽然不如整体测量高效，
 * 但结果精确且兼容所有字体类型。
 */
float gomupdf_measure_text(fz_context *ctx, fz_font *font, const char *text, float size) {
    if (!ctx || !font || !text) return 0;
    float width = 0;
    fz_try(ctx) {
        /* 逐字符测量文本宽度 */
        const unsigned char *s = (const unsigned char *)text;
        int chr;
        while (*s) {
            s += fz_chartorune(&chr, (const char *)s);
            int glyph = fz_encode_character(ctx, font, chr);
            width += fz_advance_glyph(ctx, font, glyph, 0) * size;
        }
    }
    fz_catch(ctx) { width = 0; }
    return width;
}

/* 获取单个字形的前进宽度（advance width）乘以字号 */
float gomupdf_font_glyph_advance(fz_context *ctx, fz_font *font, int glyph, float size) {
    if (!ctx || !font) return 0;
    return fz_advance_glyph(ctx, font, glyph, 0) * size;
}

// ============================================================
// 第八阶段：表单控件/Widget 系统 / Phase 8: Widget/Form System
//
// PDF 表单控件（Widget）在 MuPDF 中作为特殊的注释对象（pdf_annot）处理。
// 通过 pdf_first_widget/pdf_next_widget 遍历页面的表单控件链表。
//
// 重要设计决策：MuPDF 不存在 pdf_widget_field_name 函数（尽管名称暗示应该有），
// 因此我们使用 pdf_annot_field_label 作为替代来获取字段名称。
// 两者在内部实现上访问的是同一个 PDF /T 字段。
// ============================================================

/* 获取页面的第一个表单控件 */
pdf_annot *gomupdf_pdf_first_widget(fz_context *ctx, fz_document *doc, fz_page *page) {
    if (!ctx || !doc || !page) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return NULL;
    return pdf_first_widget(ctx, pp);
}

/* 获取链表中的下一个表单控件 */
pdf_annot *gomupdf_pdf_next_widget(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return NULL;
    return pdf_next_widget(ctx, widget);
}

/* 获取表单控件的类型（文本框、复选框、单选按钮、列表框等） */
int gomupdf_pdf_widget_type(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return -1;
    return (int)pdf_widget_type(ctx, widget);
}

/*
 * 获取表单控件的字段名称
 *
 * 使用 pdf_annot_field_label 替代不存在的 pdf_widget_field_name 函数。
 * 两者底层都是读取 PDF 对象的 /T 字段，功能完全等价。
 */
const char *gomupdf_pdf_widget_field_name(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return "";
    static __thread char buf[256];
    fz_try(ctx) {
        /* 使用 pdf_annot_field_label 获取字段名 */
        const char *n = pdf_annot_field_label(ctx, widget);
        if (n) snprintf(buf, sizeof(buf), "%s", n);
        else buf[0] = 0;
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

/* 获取表单控件的字段值（如文本框内容、复选框状态等） */
const char *gomupdf_pdf_widget_field_value(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return "";
    static __thread char buf[4096];
    fz_try(ctx) {
        /* 使用 pdf_annot_field_value 直接返回字段值的字符串 */
        const char *val = pdf_annot_field_value(ctx, widget);
        if (val) snprintf(buf, sizeof(buf), "%s", val);
        else buf[0] = 0;
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

/* 设置文本类型表单控件的字段值 */
int gomupdf_pdf_widget_set_field_value(fz_context *ctx, fz_document *doc, pdf_annot *widget, const char *value) {
    if (!ctx || !doc || !widget || !value) return -1;
    fz_try(ctx) {
        pdf_set_text_field_value(ctx, widget, value);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取表单控件的字段标志位（如只读、必填、多行等） */
int gomupdf_pdf_widget_field_flags(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return 0;
    int flags = 0;
    fz_try(ctx) { flags = pdf_annot_field_flags(ctx, widget); }
    fz_catch(ctx) {}
    return flags;
}

/*
 * 设置表单控件的字段标志位
 *
 * 设计说明：MuPDF 没有提供直接的 set_field_flags 函数，
 * 因此需要通过底层 PDF 对象操作：读取或创建 /Ff 字段，
 * 然后设置其整数值。
 */
int gomupdf_pdf_widget_set_field_flags(fz_context *ctx, pdf_annot *widget, int flags) {
    if (!ctx || !widget) return -1;
    /* MuPDF 没有直接的 set_field_flags，需要通过底层对象操作 */
    fz_try(ctx) {
        pdf_obj *obj = pdf_annot_obj(ctx, widget);
        pdf_obj *ff = pdf_dict_get(ctx, obj, PDF_NAME(Ff));
        if (!ff) {
            ff = pdf_new_int(ctx, flags);
            pdf_dict_put(ctx, obj, PDF_NAME(Ff), ff);
            pdf_drop_obj(ctx, ff);
        } else {
            pdf_set_int(ctx, ff, flags);
        }
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/*
 * 检查复选框/单选按钮控件是否被选中
 *
 * 通过检查 PDF 对象的 /AS（Appearance State）字段判断：
 * 如果 /AS 存在且不等于 /Off，则认为控件处于选中状态。
 */
int gomupdf_pdf_widget_is_checked(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return 0;
    int checked = 0;
    fz_try(ctx) {
        /* 通过 AS 字段检查是否选中 */
        pdf_obj *obj = pdf_annot_obj(ctx, widget);
        pdf_obj *as = pdf_dict_get(ctx, obj, PDF_NAME(AS));
        checked = (as != NULL && !pdf_name_eq(ctx, as, PDF_NAME(Off)));
    }
    fz_catch(ctx) {}
    return checked;
}

/* 切换复选框/单选按钮的选中状态 */
int gomupdf_pdf_widget_toggle(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return -1;
    fz_try(ctx) { pdf_toggle_widget(ctx, widget); }
    fz_catch(ctx) { return -1; }
    return 0;
}

// ============================================================
// 第九阶段：高级功能 / Phase 9: Advanced Features
//
// 包含显示列表（Display List）、页面框（Page Box）、XRef 操作、
// 嵌入文件（Embedded Files）等高级 PDF 操作功能。
// ============================================================

/*
 * --- 显示列表（Display List）---
 *
 * 显示列表是页面内容的录制回放机制：先将页面内容"录制"到列表中，
 * 然后可以多次"回放"到不同的渲染目标（像素图、文本页等），
 * 避免重复解析页面内容流，提高多次渲染的效率。
 */

/* 创建新的显示列表，需指定页面边界 */
fz_display_list *gomupdf_new_display_list(fz_context *ctx, fz_rect bounds) {
    if (!ctx) return NULL;
    fz_display_list *list = NULL;
    fz_try(ctx) { list = fz_new_display_list(ctx, bounds); }
    fz_catch(ctx) { list = NULL; }
    return list;
}

/* 释放显示列表资源 */
void gomupdf_drop_display_list(fz_context *ctx, fz_display_list *list) {
    if (ctx && list) fz_drop_display_list(ctx, list);
}

/*
 * 将页面内容录制到显示列表中
 *
 * 使用 fz_list_device 作为录制设备，运行页面内容流。
 * 嵌套的 fz_try/fz_catch 确保设备在任何情况下都被正确释放。
 */
fz_display_list *gomupdf_run_page_to_list(fz_context *ctx, fz_page *page, fz_display_list *list, fz_matrix ctm) {
    if (!ctx || !page || !list) return NULL;
    fz_try(ctx) {
        fz_device *dev = fz_new_list_device(ctx, list);
        fz_try(ctx) { fz_run_page(ctx, page, dev, ctm, NULL); }
        fz_always(ctx) { fz_drop_device(ctx, dev); }
        fz_catch(ctx) { fz_rethrow(ctx); }
    }
    fz_catch(ctx) { return NULL; }
    return list;
}

/*
 * 将显示列表渲染为像素图
 *
 * 先计算列表边界经变换后的整数矩形，创建匹配尺寸的像素图，
 * 然后使用绘图设备回放显示列表内容。支持透明通道（alpha）。
 */
fz_pixmap *gomupdf_display_list_get_pixmap(fz_context *ctx, fz_display_list *list, fz_matrix ctm, int alpha) {
    if (!ctx || !list) return NULL;
    fz_pixmap *pix = NULL;
    fz_try(ctx) {
        fz_rect bounds = fz_bound_display_list(ctx, list);
        fz_irect ibounds = fz_round_rect(fz_transform_rect(bounds, ctm));
        fz_colorspace *cs = fz_device_rgb(ctx);
        pix = fz_new_pixmap(ctx, cs, ibounds.x1 - ibounds.x0, ibounds.y1 - ibounds.y0, NULL, alpha);
        fz_clear_pixmap(ctx, pix);
        fz_device *dev = fz_new_draw_device(ctx, fz_identity, pix);
        fz_try(ctx) { fz_run_display_list(ctx, list, dev, ctm, fz_infinite_rect, NULL); }
        fz_always(ctx) { fz_drop_device(ctx, dev); }
        fz_catch(ctx) { fz_rethrow(ctx); }
    }
    fz_catch(ctx) { if (pix) { fz_drop_pixmap(ctx, pix); pix = NULL; } }
    return pix;
}

/*
 * --- 页面框（Page Box）操作 ---
 *
 * PDF 页面有多种框定义：MediaBox（媒体框）、CropBox（裁切框）等。
 * 这些函数通过直接读取/写入页面对象的字典字段来操作。
 */

/* 获取页面的裁切框（CropBox） */
fz_rect gomupdf_pdf_page_cropbox(fz_context *ctx, fz_document *doc, int page_num) {
    if (!ctx || !doc) { fz_rect r = {0,0,0,0}; return r; }
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) { fz_rect r = {0,0,0,0}; return r; }
    fz_rect r = {0,0,0,0};
    fz_try(ctx) {
        pdf_obj *page_obj = pdf_lookup_page_obj(ctx, pdf, page_num);
        if (page_obj) r = pdf_dict_get_rect(ctx, page_obj, PDF_NAME(CropBox));
    }
    fz_catch(ctx) {}
    return r;
}

/* 设置页面的裁切框（CropBox） */
int gomupdf_pdf_set_page_cropbox(fz_context *ctx, fz_document *doc, int page_num, float x0, float y0, float x1, float y1) {
    if (!ctx || !doc) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        pdf_obj *page_obj = pdf_lookup_page_obj(ctx, pdf, page_num);
        if (page_obj) {
            fz_rect r = fz_make_rect(x0, y0, x1, y1);
            pdf_dict_put_rect(ctx, page_obj, PDF_NAME(CropBox), r);
        }
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 获取页面的媒体框（MediaBox） */
fz_rect gomupdf_pdf_page_mediabox(fz_context *ctx, fz_document *doc, int page_num) {
    if (!ctx || !doc) { fz_rect r = {0,0,0,0}; return r; }
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) { fz_rect r = {0,0,0,0}; return r; }
    fz_rect r = {0,0,0,0};
    fz_try(ctx) {
        pdf_obj *page_obj = pdf_lookup_page_obj(ctx, pdf, page_num);
        if (page_obj) r = pdf_dict_get_rect(ctx, page_obj, PDF_NAME(MediaBox));
    }
    fz_catch(ctx) {}
    return r;
}

/* 设置页面的媒体框（MediaBox） */
int gomupdf_pdf_set_page_mediabox(fz_context *ctx, fz_document *doc, int page_num, float x0, float y0, float x1, float y1) {
    if (!ctx || !doc) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        pdf_obj *page_obj = pdf_lookup_page_obj(ctx, pdf, page_num);
        if (page_obj) {
            fz_rect r = fz_make_rect(x0, y0, x1, y1);
            pdf_dict_put_rect(ctx, page_obj, PDF_NAME(MediaBox), r);
        }
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/* 设置页面旋转角度（0/90/180/270） */
int gomupdf_pdf_set_page_rotation(fz_context *ctx, fz_document *doc, int page_num, int rotation) {
    if (!ctx || !doc) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        pdf_obj *page_obj = pdf_lookup_page_obj(ctx, pdf, page_num);
        if (page_obj) pdf_dict_put_int(ctx, page_obj, PDF_NAME(Rotate), rotation);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

/*
 * --- XRef（交叉引用表）操作 ---
 *
 * XRef 表是 PDF 内部对象索引的核心数据结构。
 * 这些函数提供对 XRef 条目的低级访问能力。
 */

/* 获取文档 XRef 表的长度（最大对象编号 + 1） */
int gomupdf_pdf_xref_length(fz_context *ctx, fz_document *doc) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return 0;
    return pdf_xref_len(ctx, pdf);
}

/*
 * 获取指定 XRef 对象的字典键值
 *
 * 加载 XRef 对象并查找指定 key 的值，尝试将其转为名称或文本字符串返回。
 * 如果对象不是字典或 key 不存在，返回空字符串。
 */
const char *gomupdf_pdf_xref_get_key(fz_context *ctx, fz_document *doc, int xref, const char *key) {
    if (!ctx || !doc || !key) return "";
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return "";
    static __thread char buf[1024];
    fz_try(ctx) {
        pdf_obj *xobj = pdf_load_object(ctx, pdf, xref);
        if (xobj && pdf_is_dict(ctx, xobj)) {
            pdf_obj *val = pdf_dict_gets(ctx, xobj, key);
            if (val) {
                fz_buffer *b = fz_new_buffer(ctx, 1024);
                fz_try(ctx) {
                    /* 简单地将对象转为字符串表示 */
                    const char *s = pdf_to_name(ctx, val);
                    if (s && s[0]) snprintf(buf, sizeof(buf), "/%s", s);
                    else {
                        s = pdf_to_text_string(ctx, val);
                        if (s) snprintf(buf, sizeof(buf), "%s", s);
                        else buf[0] = 0;
                    }
                }
                fz_always(ctx) { fz_drop_buffer(ctx, b); }
                fz_catch(ctx) { buf[0] = 0; }
            } else {
                buf[0] = 0;
            }
            pdf_drop_obj(ctx, xobj);
        } else {
            if (xobj) pdf_drop_obj(ctx, xobj);
            buf[0] = 0;
        }
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

/* 检查指定 XRef 条目是否包含流数据（通过检查 stm_ofs 偏移量） */
int gomupdf_pdf_xref_is_stream(fz_context *ctx, fz_document *doc, int xref) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return 0;
    int result = 0;
    fz_try(ctx) {
        pdf_xref_entry *entry = pdf_get_xref_entry(ctx, pdf, xref);
        result = (entry && entry->stm_ofs >= 0) ? 1 : 0;
    }
    fz_catch(ctx) {}
    return result;
}

/*
 * --- 嵌入文件（Embedded Files）操作 ---
 *
 * 设计说明：MuPDF 没有提供 pdf_count_embedded_files 等高级 API 来直接
 * 枚举嵌入文件。因此这里采用遍历 PDF Names 字典的方式实现：
 *
 *   路径: trailer -> /Names -> /EmbeddedFiles -> /Names -> [name1, dict1, name2, dict2, ...]
 *
 * Names 数组中的元素以"名称-文件规格字典"对的形式交替排列，
 * 因此嵌入文件数量 = 数组长度 / 2。
 * 这种方式虽然需要理解 PDF 内部结构，但无需依赖 MuPDF 未提供的高级 API。
 */

/* 获取嵌入文件的数量（通过遍历 Names 字典计数） */
int gomupdf_pdf_embedded_file_count(fz_context *ctx, fz_document *doc) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return 0;
    int count = 0;
    fz_try(ctx) {
        pdf_obj *names = pdf_dict_get(ctx, pdf_trailer(ctx, pdf), PDF_NAME(Names));
        if (names) {
            pdf_obj *ef = pdf_dict_get(ctx, names, PDF_NAME(EmbeddedFiles));
            if (ef) {
                pdf_obj *narray = pdf_dict_get(ctx, ef, PDF_NAME(Names));
                if (narray && pdf_is_array(ctx, narray)) {
                    /* Names 数组格式: [name1, dict1, name2, dict2, ...] */
                    count = pdf_array_len(ctx, narray) / 2;
                }
            }
        }
    }
    fz_catch(ctx) {}
    return count;
}

/* 获取指定索引的嵌入文件名称（从 Names 数组的偶数索引位置读取） */
const char *gomupdf_pdf_embedded_file_name(fz_context *ctx, fz_document *doc, int idx) {
    if (!ctx || !doc) return "";
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return "";
    static __thread char buf[512];
    fz_try(ctx) {
        pdf_obj *names = pdf_dict_get(ctx, pdf_trailer(ctx, pdf), PDF_NAME(Names));
        if (names) {
            pdf_obj *ef = pdf_dict_get(ctx, names, PDF_NAME(EmbeddedFiles));
            if (ef) {
                pdf_obj *narray = pdf_dict_get(ctx, ef, PDF_NAME(Names));
                if (narray && pdf_is_array(ctx, narray)) {
                    pdf_obj *name = pdf_array_get(ctx, narray, idx * 2);
                    if (name) {
                        const char *s = pdf_to_text_string(ctx, name);
                        if (s) snprintf(buf, sizeof(buf), "%s", s);
                        else buf[0] = 0;
                    } else buf[0] = 0;
                } else buf[0] = 0;
            } else buf[0] = 0;
        } else buf[0] = 0;
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

/*
 * 获取指定索引的嵌入文件内容
 *
 * 从 Names 数组的奇数索引位置取出文件规格字典（filespec），
 * 然后使用 pdf_load_embedded_file_contents 解码文件内容。
 * 返回的内存由 malloc 分配，调用方需通过 gomupdf_free 释放。
 */
unsigned char *gomupdf_pdf_embedded_file_get(fz_context *ctx, fz_document *doc, int idx, size_t *out_len) {
    if (!ctx || !doc || !out_len) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return NULL;
    unsigned char *result = NULL;
    *out_len = 0;
    fz_try(ctx) {
        pdf_obj *names = pdf_dict_get(ctx, pdf_trailer(ctx, pdf), PDF_NAME(Names));
        if (names) {
            pdf_obj *ef = pdf_dict_get(ctx, names, PDF_NAME(EmbeddedFiles));
            if (ef) {
                pdf_obj *narray = pdf_dict_get(ctx, ef, PDF_NAME(Names));
                if (narray && pdf_is_array(ctx, narray)) {
                    pdf_obj *filespec = pdf_array_get(ctx, narray, idx * 2 + 1);
                    if (filespec) {
                        fz_buffer *buf = pdf_load_embedded_file_contents(ctx, filespec);
                        if (buf) {
                            *out_len = buf->len;
                            result = (unsigned char *)malloc(buf->len);
                            if (result) memcpy(result, buf->data, buf->len);
                            fz_drop_buffer(ctx, buf);
                        }
                    }
                }
            }
        }
    }
    fz_catch(ctx) { if (result) { free(result); result = NULL; } *out_len = 0; }
    return result;
}

/*
 * 添加嵌入文件到 PDF 文档
 *
 * 使用 pdf_add_embedded_file 将文件数据附加到文档的 Names 字典中。
 * mimetype 默认为 "application/octet-stream"，记录当前时间戳。
 */
int gomupdf_pdf_add_embedded_file(fz_context *ctx, fz_document *doc,
    const char *filename, const char *mimetype, const unsigned char *data, size_t len) {
    if (!ctx || !doc || !filename || !data) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        fz_buffer *buf = fz_new_buffer_from_data(ctx, (unsigned char *)data, len);
        pdf_add_embedded_file(ctx, pdf, filename, mimetype ? mimetype : "application/octet-stream",
            buf, 0, time(NULL), 0);
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}