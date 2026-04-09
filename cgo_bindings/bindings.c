#include "bindings.h"
#include <string.h>

// Thread-local error message buffer for fz_try/fz_catch error propagation
static __thread char gomupdf_last_error[512] = {0};

const char *gomupdf_get_last_error(void) {
    return gomupdf_last_error;
}

void gomupdf_clear_error(void) {
    gomupdf_last_error[0] = '\0';
}

// Custom warning handler - does nothing to suppress warnings
static void gomupdf_warning_callback(void *user, const char *message) {
    // Suppress all warnings
}

fz_context *gomupdf_new_context(void) {
    fz_context *ctx = fz_new_context(NULL, NULL, FZ_STORE_DEFAULT);
    if (ctx) {
        fz_register_document_handlers(ctx);
        fz_set_warning_callback(ctx, gomupdf_warning_callback, NULL);
    }
    return ctx;
}

void gomupdf_drop_context(fz_context *ctx) {
    if (ctx) {
        fz_drop_context(ctx);
    }
}

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

int gomupdf_is_pdf(fz_context *ctx, fz_document *doc) {
    return pdf_document_from_fz_document(ctx, doc) != NULL;
}

const char *gomupdf_document_metadata(fz_context *ctx, fz_document *doc, const char *key) {
    if (!ctx || !doc || !key) return "";
    static char buf[256];
    int len = fz_lookup_metadata(ctx, doc, key, buf, sizeof(buf));
    if (len > 0) {
        return buf;
    }
    return "";
}

int gomupdf_needs_password(fz_context *ctx, fz_document *doc) {
    return fz_needs_password(ctx, doc);
}

int gomupdf_authenticate_password(fz_context *ctx, fz_document *doc, const char *password) {
    return fz_authenticate_password(ctx, doc, password);
}

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

int gomupdf_page_rotation(fz_context *ctx, fz_page *page) {
    // Rotation is not directly exposed in MuPDF's fz_page API.
    // For PDF documents, we would need to access the /Rotate entry in the page dict.
    // This requires the pdf_document pointer which is not available from fz_page alone.
    // For now, return 0; the Go layer should query this separately.
    (void)ctx;
    (void)page;
    return 0;
}

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

void gomupdf_drop_page(fz_context *ctx, fz_page *page) {
    if (ctx && page) {
        fz_drop_page(ctx, page);
    }
}

fz_pixmap *gomupdf_new_pixmap(fz_context *ctx, fz_colorspace *cs, int width, int height) {
    return fz_new_pixmap(ctx, cs, width, height, NULL, 0);
}

fz_pixmap *gomupdf_render_page(fz_context *ctx, fz_page *page, float a, float b, float c, float d, float e, float f, int alpha) {
    fz_pixmap *pix = NULL;
    fz_device *dev = NULL;
    gomupdf_clear_error();
    fz_try(ctx) {
        fz_matrix ctm = fz_make_matrix(a, b, c, d, e, f);
        fz_rect rect = fz_bound_page(ctx, page);
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

void gomupdf_drop_pixmap(fz_context *ctx, fz_pixmap *pix) {
    if (ctx && pix) {
        fz_drop_pixmap(ctx, pix);
    }
}

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

void gomupdf_save_pixmap_as_png(fz_context *ctx, fz_pixmap *pix, const char *filename) {
    fz_try(ctx) {
        fz_save_pixmap_as_png(ctx, pix, filename);
    }
    fz_catch(ctx) {
        // Error silently ignored — caller should check if file exists
    }
}

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

void gomupdf_drop_stext_page(fz_context *ctx, fz_stext_page *page) {
    if (ctx && page) {
        fz_drop_stext_page(ctx, page);
    }
}

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

// In MuPDF, fz_stext_line contains characters directly via first_char linked list.
// There is no separate "span" concept in the C struct — spans are a PyMuPDF abstraction.
// We return NULL to indicate spans are not directly available at the C level.
// Higher-level Go code should group characters by font/size/style to create spans.
fz_stext_char *gomupdf_stext_line_first_char(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->first_char;
    }
    return NULL;
}

int gomupdf_stext_block_type(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block) {
        return block->type;
    }
    return -1;
}

fz_rect gomupdf_stext_block_bbox(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block) {
        return block->bbox;
    }
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

fz_rect gomupdf_stext_line_bbox(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->bbox;
    }
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

fz_point gomupdf_stext_line_dir(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->dir;
    }
    fz_point p = {0, 0};
    return p;
}

fz_point gomupdf_stext_char_origin(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->origin;
    }
    fz_point p = {0, 0};
    return p;
}

int gomupdf_stext_char_c(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->c;
    }
    return 0;
}

float gomupdf_stext_char_size(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->size;
    }
    return 0;
}

fz_rect gomupdf_stext_char_bbox(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        // Use the quad directly for accurate char bbox
        return fz_rect_from_quad(ch->quad);
    }
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

fz_stext_char *gomupdf_stext_char_next(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) {
        return ch->next;
    }
    return NULL;
}

fz_stext_line *gomupdf_stext_block_first_line(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block && block->type == FZ_STEXT_BLOCK_TEXT) {
        return block->u.t.first_line;
    }
    return NULL;
}

fz_stext_line *gomupdf_stext_line_next(fz_context *ctx, fz_stext_line *line) {
    (void)ctx;
    if (line) {
        return line->next;
    }
    return NULL;
}

fz_image *gomupdf_stext_block_get_image(fz_context *ctx, fz_stext_block *block) {
    (void)ctx;
    if (block && block->type == FZ_STEXT_BLOCK_IMAGE) {
        return block->u.i.image;
    }
    return NULL;
}

int gomupdf_image_width(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->w;
    return 0;
}

int gomupdf_image_height(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->h;
    return 0;
}

int gomupdf_image_n(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->n;
    return 0;
}

int gomupdf_image_bpc(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img) return img->bpc;
    return 0;
}

const char *gomupdf_image_colorspace_name(fz_context *ctx, fz_image *img) {
    (void)ctx;
    if (img && img->colorspace) {
        return fz_colorspace_name(ctx, img->colorspace);
    }
    return "None";
}

fz_colorspace *gomupdf_new_colorspace_rgb(fz_context *ctx) {
    return fz_device_rgb(ctx);
}

fz_colorspace *gomupdf_new_colorspace_gray(fz_context *ctx) {
    return fz_device_gray(ctx);
}

fz_colorspace *gomupdf_new_colorspace_cmyk(fz_context *ctx) {
    return fz_device_cmyk(ctx);
}

void gomupdf_drop_colorspace(fz_context *ctx, fz_colorspace *cs) {
    if (ctx && cs) {
        fz_drop_colorspace(ctx, cs);
    }
}

fz_matrix gomupdf_identity_matrix(void) {
    return fz_identity;
}

fz_matrix gomupdf_make_matrix(float a, float b, float c, float d, float e, float f) {
    return fz_make_matrix(a, b, c, d, e, f);
}

fz_matrix gomupdf_scale_matrix(float sx, float sy) {
    return fz_scale(sx, sy);
}

fz_matrix gomupdf_rotate_matrix(float degrees) {
    return fz_rotate(degrees);
}

fz_matrix gomupdf_translate_matrix(float tx, float ty) {
    return fz_translate(tx, ty);
}

fz_matrix gomupdf_concat_matrix(fz_matrix left, fz_matrix right) {
    return fz_concat(left, right);
}

fz_matrix gomupdf_invert_matrix(fz_context *ctx, fz_matrix matrix) {
    (void)ctx;
    return fz_invert_matrix(matrix);
}

fz_point gomupdf_transform_point(fz_point p, fz_matrix m) {
    return fz_transform_point(p, m);
}

fz_rect gomupdf_transform_rect(fz_rect r, fz_matrix m) {
    return fz_transform_rect(r, m);
}

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

const char *gomupdf_version(void) {
    return FZ_VERSION;
}

// ============================================================
// PDF Save Operations
// ============================================================

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

void gomupdf_free(void *ptr) {
    if (ptr) free(ptr);
}

// ============================================================
// PDF Page Management
// ============================================================

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
// PDF Set Metadata
// ============================================================

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
// PDF Outline / TOC
// ============================================================

// Helper: flatten outline tree into a linear array
typedef struct {
    char *title;
    int page;
    int level;
    char *uri;
    int is_open;
} gomupdf_outline_entry;

// Thread-local outline cache
static __thread gomupdf_outline_entry *g_outline_entries = NULL;
static __thread int g_outline_count = 0;

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
// PDF Links
// ============================================================

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

void gomupdf_drop_link(fz_context *ctx, fz_link *link) {
    if (ctx && link) {
        fz_drop_link(ctx, link);
    }
}

fz_link *gomupdf_link_next(fz_link *link) {
    if (link) return link->next;
    return NULL;
}

fz_rect gomupdf_link_rect(fz_link *link) {
    if (link) return link->rect;
    fz_rect empty = {0, 0, 0, 0};
    return empty;
}

const char *gomupdf_link_uri(fz_link *link) {
    if (link) return link->uri;
    return NULL;
}

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
// PDF Text Search
// ============================================================

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
// PDF Permissions
// ============================================================

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
// Phase 2: Text Output Formats
// ============================================================

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

const char *gomupdf_stext_char_font(fz_context *ctx, fz_stext_char *ch) {
    if (!ctx || !ch) return "";
    if (ch->font) {
        return fz_font_name(ctx, ch->font);
    }
    return "";
}

int gomupdf_stext_char_flags(fz_context *ctx, fz_stext_char *ch) {
    (void)ctx;
    if (ch) return ch->flags;
    return 0;
}

// ============================================================
// Phase 3: Image Processing
// ============================================================

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

int gomupdf_pixmap_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y) {
    if (!ctx || !pix) return 0;
    int n = fz_pixmap_components(ctx, pix);
    unsigned char *samp = fz_pixmap_samples(ctx, pix) + (size_t)y * fz_pixmap_stride(ctx, pix) + (size_t)x * n;
    unsigned int val = 0;
    for (int i = 0; i < n; i++) val = (val << 8) | samp[i];
    return (int)val;
}

void gomupdf_pixmap_set_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y, unsigned int val) {
    if (!ctx || !pix) return;
    int n = fz_pixmap_components(ctx, pix);
    unsigned char *samp = fz_pixmap_samples(ctx, pix) + (size_t)y * fz_pixmap_stride(ctx, pix) + (size_t)x * n;
    for (int i = n - 1; i >= 0; i--) {
        samp[i] = val & 0xFF;
        val >>= 8;
    }
}

void gomupdf_pixmap_clear_with(fz_context *ctx, fz_pixmap *pix, int value) {
    if (!ctx || !pix) return;
    fz_clear_pixmap_with_value(ctx, pix, value);
}

void gomupdf_pixmap_invert(fz_context *ctx, fz_pixmap *pix) {
    if (!ctx || !pix) return;
    fz_invert_pixmap(ctx, pix);
}

void gomupdf_pixmap_gamma(fz_context *ctx, fz_pixmap *pix, float gamma) {
    if (!ctx || !pix) return;
    fz_gamma_pixmap(ctx, pix, gamma);
}

void gomupdf_pixmap_tint(fz_context *ctx, fz_pixmap *pix, int black, int white) {
    if (!ctx || !pix) return;
    fz_tint_pixmap(ctx, pix, black, white);
}

// ============================================================
// Phase 4: Annotation System
// ============================================================

pdf_annot *gomupdf_pdf_first_annot(fz_context *ctx, fz_document *doc, fz_page *page) {
    if (!ctx || !doc || !page) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return NULL;
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pp) return NULL;
    return pdf_first_annot(ctx, pp);
}

pdf_annot *gomupdf_pdf_next_annot(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return NULL;
    return pdf_next_annot(ctx, annot);
}

int gomupdf_pdf_annot_type(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return -1;
    return (int)pdf_annot_type(ctx, annot);
}

fz_rect gomupdf_pdf_annot_rect(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) { fz_rect r = {0,0,0,0}; return r; }
    fz_rect r = {0,0,0,0};
    fz_try(ctx) { r = pdf_annot_rect(ctx, annot); }
    fz_catch(ctx) {}
    return r;
}

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

int gomupdf_pdf_set_annot_contents(fz_context *ctx, pdf_annot *annot, const char *text) {
    if (!ctx || !annot || !text) return -1;
    fz_try(ctx) { pdf_set_annot_contents(ctx, annot, text); }
    fz_catch(ctx) { return -1; }
    return 0;
}

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

int gomupdf_pdf_set_annot_color(fz_context *ctx, pdf_annot *annot, float r, float g, float b, float a) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) {
        float color[4] = {r, g, b, a};
        pdf_set_annot_color(ctx, annot, 4, color);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

float gomupdf_pdf_annot_opacity(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return 1.0f;
    float opacity = 1.0f;
    fz_try(ctx) { opacity = pdf_annot_opacity(ctx, annot); }
    fz_catch(ctx) {}
    return opacity;
}

int gomupdf_pdf_set_annot_opacity(fz_context *ctx, pdf_annot *annot, float opacity) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) { pdf_set_annot_opacity(ctx, annot, opacity); }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_annot_flags(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return 0;
    int flags = 0;
    fz_try(ctx) { flags = pdf_annot_flags(ctx, annot); }
    fz_catch(ctx) {}
    return flags;
}

int gomupdf_pdf_set_annot_flags(fz_context *ctx, pdf_annot *annot, int flags) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) { pdf_set_annot_flags(ctx, annot, flags); }
    fz_catch(ctx) { return -1; }
    return 0;
}

float gomupdf_pdf_annot_border(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) return 0;
    float w = 0;
    fz_try(ctx) { w = pdf_annot_border(ctx, annot); }
    fz_catch(ctx) {}
    return w;
}

int gomupdf_pdf_set_annot_border(fz_context *ctx, pdf_annot *annot, float width) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) { pdf_set_annot_border(ctx, annot, width); }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_update_annot(fz_context *ctx, fz_document *doc, pdf_annot *annot) {
    if (!ctx || !annot) return -1;
    (void)doc;
    fz_try(ctx) { pdf_update_annot(ctx, annot); }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_delete_annot(fz_context *ctx, fz_document *doc, fz_page *page, pdf_annot *annot) {
    if (!ctx || !doc || !page || !annot) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return -1;
    fz_try(ctx) { pdf_delete_annot(ctx, pp, annot); }
    fz_catch(ctx) { return -1; }
    return 0;
}

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

int gomupdf_pdf_set_annot_quad_points(fz_context *ctx, pdf_annot *annot, int count, fz_quad *quads) {
    if (!ctx || !annot || !quads || count <= 0) return -1;
    fz_try(ctx) { pdf_set_annot_quad_points(ctx, annot, count, quads); }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_set_annot_rect(fz_context *ctx, pdf_annot *annot, float x0, float y0, float x1, float y1) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) {
        fz_rect r = fz_make_rect(x0, y0, x1, y1);
        pdf_set_annot_rect(ctx, annot, r);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_set_annot_popup(fz_context *ctx, pdf_annot *annot, float x0, float y0, float x1, float y1) {
    if (!ctx || !annot) return -1;
    fz_try(ctx) {
        fz_rect r = fz_make_rect(x0, y0, x1, y1);
        pdf_set_annot_popup(ctx, annot, r);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

fz_rect gomupdf_pdf_annot_popup(fz_context *ctx, pdf_annot *annot) {
    if (!ctx || !annot) { fz_rect r = {0,0,0,0}; return r; }
    fz_rect r = {0,0,0,0};
    fz_try(ctx) { r = pdf_annot_popup(ctx, annot); }
    fz_catch(ctx) {}
    return r;
}

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
// Phase 5: Link Operations
// ============================================================

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
// Phase 6: Shape Drawing (content stream operations)
// ============================================================

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

int gomupdf_pdf_page_write_end(fz_context *ctx, fz_document *doc, fz_page *page,
    fz_buffer *contents) {
    if (!ctx || !doc || !page || !contents) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pp) return -1;
    fz_try(ctx) {
        /* 将新内容流写入页面对象 */
        pdf_update_stream(ctx, pdf, pp->obj, contents);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

// ============================================================
// Phase 7: Font Operations
// ============================================================

fz_font *gomupdf_new_font_from_file(fz_context *ctx, const char *filename, int index) {
    if (!ctx || !filename) return NULL;
    fz_font *font = NULL;
    fz_try(ctx) {
        font = fz_new_font_from_file(ctx, NULL, filename, index, 0);
    }
    fz_catch(ctx) { font = NULL; }
    return font;
}

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

void gomupdf_drop_font(fz_context *ctx, fz_font *font) {
    if (ctx && font) fz_drop_font(ctx, font);
}

const char *gomupdf_font_name(fz_context *ctx, fz_font *font) {
    if (!ctx || !font) return "";
    return fz_font_name(ctx, font);
}

float gomupdf_font_ascender(fz_context *ctx, fz_font *font) {
    if (!ctx || !font) return 0;
    return fz_font_ascender(ctx, font);
}

float gomupdf_font_descender(fz_context *ctx, fz_font *font) {
    if (!ctx || !font) return 0;
    return fz_font_descender(ctx, font);
}

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

float gomupdf_font_glyph_advance(fz_context *ctx, fz_font *font, int glyph, float size) {
    if (!ctx || !font) return 0;
    return fz_advance_glyph(ctx, font, glyph, 0) * size;
}

// ============================================================
// Phase 8: Widget/Form System
// ============================================================

pdf_annot *gomupdf_pdf_first_widget(fz_context *ctx, fz_document *doc, fz_page *page) {
    if (!ctx || !doc || !page) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    pdf_page *pp = pdf_page_from_fz_page(ctx, page);
    if (!pdf || !pp) return NULL;
    return pdf_first_widget(ctx, pp);
}

pdf_annot *gomupdf_pdf_next_widget(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return NULL;
    return pdf_next_widget(ctx, widget);
}

int gomupdf_pdf_widget_type(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return -1;
    return (int)pdf_widget_type(ctx, widget);
}

const char *gomupdf_pdf_widget_field_name(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return "";
    static __thread char buf[256];
    fz_try(ctx) {
        const char *n = pdf_widget_field_name(ctx, widget);
        if (n) snprintf(buf, sizeof(buf), "%s", n);
        else buf[0] = 0;
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

const char *gomupdf_pdf_widget_field_value(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return "";
    static __thread char buf[4096];
    fz_try(ctx) {
        pdf_obj *val = pdf_widget_field_value(ctx, widget);
        if (val) {
            const char *s = pdf_to_text_string(ctx, val);
            if (s) snprintf(buf, sizeof(buf), "%s", s);
            else buf[0] = 0;
        } else {
            buf[0] = 0;
        }
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

int gomupdf_pdf_widget_set_field_value(fz_context *ctx, fz_document *doc, pdf_annot *widget, const char *value) {
    if (!ctx || !doc || !widget || !value) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        pdf_widget_set_text_field_value(ctx, widget, value);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_widget_field_flags(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return 0;
    int flags = 0;
    fz_try(ctx) { flags = pdf_widget_field_flags(ctx, widget); }
    fz_catch(ctx) {}
    return flags;
}

int gomupdf_pdf_widget_set_field_flags(fz_context *ctx, pdf_annot *widget, int flags) {
    if (!ctx || !widget) return -1;
    fz_try(ctx) { pdf_widget_set_field_flags(ctx, widget, flags); }
    fz_catch(ctx) { return -1; }
    return 0;
}

int gomupdf_pdf_widget_is_checked(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return 0;
    int checked = 0;
    fz_try(ctx) { checked = pdf_widget_is_checked(ctx, widget); }
    fz_catch(ctx) {}
    return checked;
}

int gomupdf_pdf_widget_toggle(fz_context *ctx, pdf_annot *widget) {
    if (!ctx || !widget) return -1;
    fz_try(ctx) { pdf_toggle_widget(ctx, widget); }
    fz_catch(ctx) { return -1; }
    return 0;
}

// ============================================================
// Phase 9: Advanced Features
// ============================================================

// Display List
fz_display_list *gomupdf_new_display_list(fz_context *ctx, fz_rect bounds) {
    if (!ctx) return NULL;
    fz_display_list *list = NULL;
    fz_try(ctx) { list = fz_new_display_list(ctx, bounds); }
    fz_catch(ctx) { list = NULL; }
    return list;
}

void gomupdf_drop_display_list(fz_context *ctx, fz_display_list *list) {
    if (ctx && list) fz_drop_display_list(ctx, list);
}

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

// Page Box Operations
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

// XRef Operations
int gomupdf_pdf_xref_length(fz_context *ctx, fz_document *doc) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return 0;
    return pdf_xref_len(ctx, pdf);
}

const char *gomupdf_pdf_xref_get_key(fz_context *ctx, fz_document *doc, int xref, const char *key) {
    if (!ctx || !doc || !key) return "";
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return "";
    static __thread char buf[1024];
    fz_try(ctx) {
        pdf_obj *obj = pdf_xref_get_key(ctx, pdf, xref, key);
        if (obj) {
            fz_buffer *b = fz_new_buffer(ctx, 1024);
            fz_try(ctx) {
                fz_append_pdf_obj_string(ctx, b, obj);
                size_t len = b->len;
                if (len >= sizeof(buf)) len = sizeof(buf) - 1;
                memcpy(buf, b->data, len);
                buf[len] = 0;
            }
            fz_always(ctx) { fz_drop_buffer(ctx, b); }
            fz_catch(ctx) { buf[0] = 0; }
        } else {
            buf[0] = 0;
        }
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

int gomupdf_pdf_xref_is_stream(fz_context *ctx, fz_document *doc, int xref) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return 0;
    int result = 0;
    fz_try(ctx) { result = pdf_xref_is_stream(ctx, pdf, xref); }
    fz_catch(ctx) {}
    return result;
}

// Embedded Files
int gomupdf_pdf_embedded_file_count(fz_context *ctx, fz_document *doc) {
    if (!ctx || !doc) return 0;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return 0;
    int count = 0;
    fz_try(ctx) { count = pdf_count_embedded_files(ctx, pdf); }
    fz_catch(ctx) {}
    return count;
}

const char *gomupdf_pdf_embedded_file_name(fz_context *ctx, fz_document *doc, int idx) {
    if (!ctx || !doc) return "";
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return "";
    static __thread char buf[512];
    fz_try(ctx) {
        pdf_obj *fs = pdf_load_embedded_file_n(ctx, pdf, idx);
        if (fs) {
            const char *name = pdf_dict_get_text_string(ctx, fs, PDF_NAME(F));
            if (name) snprintf(buf, sizeof(buf), "%s", name);
            else buf[0] = 0;
            pdf_drop_obj(ctx, fs);
        } else { buf[0] = 0; }
    }
    fz_catch(ctx) { buf[0] = 0; }
    return buf;
}

unsigned char *gomupdf_pdf_embedded_file_get(fz_context *ctx, fz_document *doc, int idx, size_t *out_len) {
    if (!ctx || !doc || !out_len) return NULL;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return NULL;
    unsigned char *result = NULL;
    *out_len = 0;
    fz_try(ctx) {
        fz_buffer *buf = pdf_load_embedded_file_contents_n(ctx, pdf, idx);
        if (buf) {
            *out_len = buf->len;
            result = (unsigned char *)malloc(buf->len);
            if (result) memcpy(result, buf->data, buf->len);
            fz_drop_buffer(ctx, buf);
        }
    }
    fz_catch(ctx) { if (result) { free(result); result = NULL; } *out_len = 0; }
    return result;
}

int gomupdf_pdf_add_embedded_file(fz_context *ctx, fz_document *doc,
    const char *filename, const char *mimetype, const unsigned char *data, size_t len) {
    if (!ctx || !doc || !filename || !data) return -1;
    pdf_document *pdf = pdf_document_from_fz_document(ctx, doc);
    if (!pdf) return -1;
    fz_try(ctx) {
        fz_buffer *buf = fz_new_buffer_from_data(ctx, (unsigned char *)data, len);
        pdf_add_embedded_file(ctx, pdf, filename, mimetype ? mimetype : "application/octet-stream",
            buf, 0, time(NULL));
        fz_drop_buffer(ctx, buf);
    }
    fz_catch(ctx) { return -1; }
    return 0;
}