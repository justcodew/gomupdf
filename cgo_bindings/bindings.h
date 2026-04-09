#ifndef GOMUPDF_BINDINGS_H
#define GOMUPDF_BINDINGS_H

#include <mupdf/fitz.h>
#include <mupdf/pdf.h>

// Context management
fz_context *gomupdf_new_context(void);
void gomupdf_drop_context(fz_context *ctx);

// Document operations
fz_document *gomupdf_open_document(fz_context *ctx, const char *filename);
fz_document *gomupdf_open_document_with_stream(fz_context *ctx, const char *type, unsigned char *data, size_t len);
fz_document *gomupdf_new_pdf_document(fz_context *ctx);
void gomupdf_drop_document(fz_context *ctx, fz_document *doc);
int gomupdf_page_count(fz_context *ctx, fz_document *doc);
int gomupdf_is_pdf(fz_context *ctx, fz_document *doc);
const char *gomupdf_document_metadata(fz_context *ctx, fz_document *doc, const char *key);
int gomupdf_needs_password(fz_context *ctx, fz_document *doc);
int gomupdf_authenticate_password(fz_context *ctx, fz_document *doc, const char *password);

// Page operations
fz_page *gomupdf_load_page(fz_context *ctx, fz_document *doc, int number);
fz_rect gomupdf_page_rect(fz_context *ctx, fz_page *page);
int gomupdf_page_rotation(fz_context *ctx, fz_page *page);
int gomupdf_pdf_page_rotation(fz_context *ctx, fz_document *doc, int page_num);
void gomupdf_drop_page(fz_context *ctx, fz_page *page);

// Pixmap operations
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

// Text operations
fz_stext_page *gomupdf_new_stext_page_from_page(fz_context *ctx, fz_page *page);
void gomupdf_drop_stext_page(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_text(fz_context *ctx, fz_stext_page *page);

// Text block iteration
int gomupdf_stext_page_block_count(fz_context *ctx, fz_stext_page *page);
fz_stext_block *gomupdf_stext_page_get_block(fz_context *ctx, fz_stext_page *page, int idx);

// Block type
int gomupdf_stext_block_type(fz_context *ctx, fz_stext_block *block);

// Block bbox
fz_rect gomupdf_stext_block_bbox(fz_context *ctx, fz_stext_block *block);

// Text line iteration
int gomupdf_stext_block_line_count(fz_context *ctx, fz_stext_block *block);
fz_stext_line *gomupdf_stext_block_get_line(fz_context *ctx, fz_stext_block *block, int idx);

// Line bbox and direction
fz_rect gomupdf_stext_line_bbox(fz_context *ctx, fz_stext_line *line);
fz_point gomupdf_stext_line_dir(fz_context *ctx, fz_stext_line *line);

// Text fragment iteration
int gomupdf_stext_line_char_count(fz_context *ctx, fz_stext_line *line);
fz_stext_char *gomupdf_stext_line_first_char(fz_context *ctx, fz_stext_line *line);

// Character iteration
fz_point gomupdf_stext_char_origin(fz_context *ctx, fz_stext_char *ch);
int gomupdf_stext_char_c(fz_context *ctx, fz_stext_char *ch);
float gomupdf_stext_char_size(fz_context *ctx, fz_stext_char *ch);
fz_rect gomupdf_stext_char_bbox(fz_context *ctx, fz_stext_char *ch);
fz_stext_char *gomupdf_stext_char_next(fz_context *ctx, fz_stext_char *ch);
fz_stext_line *gomupdf_stext_block_first_line(fz_context *ctx, fz_stext_block *block);
fz_stext_line *gomupdf_stext_line_next(fz_context *ctx, fz_stext_line *line);

// Image extraction
fz_image *gomupdf_stext_block_get_image(fz_context *ctx, fz_stext_block *block);
int gomupdf_image_width(fz_context *ctx, fz_image *img);
int gomupdf_image_height(fz_context *ctx, fz_image *img);
int gomupdf_image_n(fz_context *ctx, fz_image *img);
int gomupdf_image_bpc(fz_context *ctx, fz_image *img);
const char *gomupdf_image_colorspace_name(fz_context *ctx, fz_image *img);

// Colorspace operations
fz_colorspace *gomupdf_new_colorspace_rgb(fz_context *ctx);
fz_colorspace *gomupdf_new_colorspace_gray(fz_context *ctx);
fz_colorspace *gomupdf_new_colorspace_cmyk(fz_context *ctx);
void gomupdf_drop_colorspace(fz_context *ctx, fz_colorspace *cs);

// Matrix operations
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

// Rect operations
fz_rect gomupdf_make_rect(float x0, float y0, float x1, float y1);
fz_irect gomupdf_make_irect(int x0, int y0, int x1, int y1);
int gomupdf_rect_is_empty(fz_rect r);
int gomupdf_rect_is_infinite(fz_rect r);

// Point operations
fz_point gomupdf_make_point(float x, float y);

// Quad operations
fz_quad gomupdf_make_quad(fz_point ul, fz_point ur, fz_point ll, fz_point lr);
fz_rect gomupdf_quad_rect(fz_quad q);

// Utility operations
const char *gomupdf_version(void);
const char *gomupdf_get_last_error(void);
void gomupdf_clear_error(void);

// PDF save operations
int gomupdf_pdf_save_document(fz_context *ctx, fz_document *doc, const char *filename,
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata);

// PDF write to buffer (returns buffer data and length; caller must free with gomupdf_free)
int gomupdf_pdf_write_document(fz_context *ctx, fz_document *doc,
    unsigned char **out_data, size_t *out_len,
    int do_garbage, int do_clean, int do_compress, int do_compress_images, int do_compress_fonts,
    int do_decompress, int do_linear, int do_ascii, int do_incremental, int do_pretty,
    int do_sanitize, int do_appearance, int do_preserve_metadata);
void gomupdf_free(void *ptr);

// PDF page management
int gomupdf_pdf_insert_page(fz_context *ctx, fz_document *doc, int at, float x0, float y0, float x1, float y1, int rotation);
int gomupdf_pdf_delete_page(fz_context *ctx, fz_document *doc, int number);
int gomupdf_pdf_delete_page_range(fz_context *ctx, fz_document *doc, int start, int end);

// PDF set metadata
int gomupdf_pdf_set_metadata(fz_context *ctx, fz_document *doc, const char *key, const char *value);

// PDF outline/TOC
int gomupdf_pdf_outline_count(fz_context *ctx, fz_document *doc);
// Returns outline entry: title, page, level, uri (all as separate out params)
int gomupdf_pdf_outline_get(fz_context *ctx, fz_document *doc, int idx,
    const char **title, int *page, int *level, const char **uri, int *is_open);

// PDF links
fz_link *gomupdf_page_load_links(fz_context *ctx, fz_page *page);
void gomupdf_drop_link(fz_context *ctx, fz_link *link);
fz_link *gomupdf_link_next(fz_link *link);
fz_rect gomupdf_link_rect(fz_link *link);
const char *gomupdf_link_uri(fz_link *link);
int gomupdf_link_page(fz_context *ctx, fz_document *doc, fz_link *link);

// PDF text search
int gomupdf_search_text(fz_context *ctx, fz_stext_page *page, const char *needle,
    int max_hits, fz_quad *hits);

// PDF permissions
int gomupdf_pdf_permissions(fz_context *ctx, fz_document *doc);

#endif // GOMUPDF_BINDINGS_H
