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

// Phase 2: Text output formats
char *gomupdf_stext_page_to_html(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_to_xml(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_to_xhtml(fz_context *ctx, fz_stext_page *page);
char *gomupdf_stext_page_to_json(fz_context *ctx, fz_stext_page *page);
const char *gomupdf_stext_char_font(fz_context *ctx, fz_stext_char *ch);
int gomupdf_stext_char_flags(fz_context *ctx, fz_stext_char *ch);

// Phase 3: Image processing
fz_pixmap *gomupdf_image_get_pixmap(fz_context *ctx, fz_image *img);
int gomupdf_pixmap_to_png_bytes(fz_context *ctx, fz_pixmap *pix, unsigned char **out_data, size_t *out_len);
int gomupdf_pixmap_to_jpeg_bytes(fz_context *ctx, fz_pixmap *pix, int quality, unsigned char **out_data, size_t *out_len);
int gomupdf_pixmap_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y);
void gomupdf_pixmap_set_pixel(fz_context *ctx, fz_pixmap *pix, int x, int y, unsigned int val);
void gomupdf_pixmap_clear_with(fz_context *ctx, fz_pixmap *pix, int value);
void gomupdf_pixmap_invert(fz_context *ctx, fz_pixmap *pix);
void gomupdf_pixmap_gamma(fz_context *ctx, fz_pixmap *pix, float gamma);
void gomupdf_pixmap_tint(fz_context *ctx, fz_pixmap *pix, int black, int white);

// Phase 4: Annotation system
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

// Phase 5: Link operations
int gomupdf_pdf_create_link(fz_context *ctx, fz_document *doc, fz_page *page,
    float x0, float y0, float x1, float y1, const char *uri, int page_num);
int gomupdf_pdf_delete_link(fz_context *ctx, fz_document *doc, fz_page *page, fz_link *link);

// Phase 6: Shape drawing
fz_buffer *gomupdf_pdf_page_write_begin(fz_context *ctx, fz_document *doc, fz_page *page);
int gomupdf_pdf_page_write_end(fz_context *ctx, fz_document *doc, fz_page *page, fz_buffer *contents);

// Phase 7: Font operations
fz_font *gomupdf_new_font_from_file(fz_context *ctx, const char *filename, int index);
fz_font *gomupdf_new_font_from_buffer(fz_context *ctx, const char *data, size_t len, int index);
void gomupdf_drop_font(fz_context *ctx, fz_font *font);
const char *gomupdf_font_name(fz_context *ctx, fz_font *font);
float gomupdf_font_ascender(fz_context *ctx, fz_font *font);
float gomupdf_font_descender(fz_context *ctx, fz_font *font);
float gomupdf_measure_text(fz_context *ctx, fz_font *font, const char *text, float size);
float gomupdf_font_glyph_advance(fz_context *ctx, fz_font *font, int glyph, float size);

// Phase 8: Widget/Form system
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

// Phase 9: Advanced features
fz_display_list *gomupdf_new_display_list(fz_context *ctx, fz_rect bounds);
void gomupdf_drop_display_list(fz_context *ctx, fz_display_list *list);
fz_display_list *gomupdf_run_page_to_list(fz_context *ctx, fz_page *page, fz_display_list *list, fz_matrix ctm);
fz_pixmap *gomupdf_display_list_get_pixmap(fz_context *ctx, fz_display_list *list, fz_matrix ctm, int alpha);

fz_rect gomupdf_pdf_page_cropbox(fz_context *ctx, fz_document *doc, int page_num);
int gomupdf_pdf_set_page_cropbox(fz_context *ctx, fz_document *doc, int page_num, float x0, float y0, float x1, float y1);
fz_rect gomupdf_pdf_page_mediabox(fz_context *ctx, fz_document *doc, int page_num);
int gomupdf_pdf_set_page_mediabox(fz_context *ctx, fz_document *doc, int page_num, float x0, float y0, float x1, float y1);
int gomupdf_pdf_set_page_rotation(fz_context *ctx, fz_document *doc, int page_num, int rotation);

int gomupdf_pdf_xref_length(fz_context *ctx, fz_document *doc);
const char *gomupdf_pdf_xref_get_key(fz_context *ctx, fz_document *doc, int xref, const char *key);
int gomupdf_pdf_xref_is_stream(fz_context *ctx, fz_document *doc, int xref);

int gomupdf_pdf_embedded_file_count(fz_context *ctx, fz_document *doc);
const char *gomupdf_pdf_embedded_file_name(fz_context *ctx, fz_document *doc, int idx);
unsigned char *gomupdf_pdf_embedded_file_get(fz_context *ctx, fz_document *doc, int idx, size_t *out_len);
int gomupdf_pdf_add_embedded_file(fz_context *ctx, fz_document *doc,
    const char *filename, const char *mimetype, const unsigned char *data, size_t len);

#endif // GOMUPDF_BINDINGS_H
