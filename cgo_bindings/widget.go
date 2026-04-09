package cgo

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/mupdf/lib -lmupdf -lmupdfcpp
#cgo CFLAGS: -I/opt/homebrew/opt/mupdf/include

#include "bindings.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Widget represents a PDF form widget (wraps pdf_annot as widget).
type Widget struct {
	ctx    *Context
	widget *C.pdf_annot
	doc    *C.fz_document
}

// WidgetType constants.
const (
	WidgetText     = 0
	WidgetCheckBox = 1
	WidgetRadio    = 2
	WidgetList     = 3
	WidgetChoice   = 4
)

// FirstWidget returns the first widget on a page.
func FirstWidget(ctx *Context, doc *C.fz_document, page *C.fz_page) *Widget {
	if ctx == nil || doc == nil || page == nil {
		return nil
	}
	var w *C.pdf_annot
	ctx.WithLock(func() {
		w = C.gomupdf_pdf_first_widget(ctx.ctx, doc, page)
	})
	if w == nil {
		return nil
	}
	return &Widget{ctx: ctx, widget: w, doc: doc}
}

// Next returns the next widget.
func (w *Widget) Next() *Widget {
	if w.widget == nil || w.ctx == nil {
		return nil
	}
	var next *C.pdf_annot
	w.ctx.WithLock(func() {
		next = C.gomupdf_pdf_next_widget(w.ctx.ctx, w.widget)
	})
	if next == nil {
		return nil
	}
	return &Widget{ctx: w.ctx, widget: next, doc: w.doc}
}

// Type returns the widget type.
func (w *Widget) Type() int {
	if w.widget == nil || w.ctx == nil {
		return -1
	}
	var typ C.int
	w.ctx.WithLock(func() {
		typ = C.gomupdf_pdf_widget_type(w.ctx.ctx, w.widget)
	})
	return int(typ)
}

// FieldName returns the widget field name.
func (w *Widget) FieldName() string {
	if w.widget == nil || w.ctx == nil {
		return ""
	}
	var s *C.char
	w.ctx.WithLock(func() {
		s = C.gomupdf_pdf_widget_field_name(w.ctx.ctx, w.widget)
	})
	return C.GoString(s)
}

// FieldValue returns the widget field value.
func (w *Widget) FieldValue() string {
	if w.widget == nil || w.ctx == nil {
		return ""
	}
	var s *C.char
	w.ctx.WithLock(func() {
		s = C.gomupdf_pdf_widget_field_value(w.ctx.ctx, w.widget)
	})
	return C.GoString(s)
}

// SetFieldValue sets the widget field value.
func (w *Widget) SetFieldValue(value string) error {
	if w.widget == nil || w.ctx == nil {
		return errors.New("widget is nil")
	}
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))
	var rc C.int
	w.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_widget_set_field_value(w.ctx.ctx, w.doc, w.widget, cvalue)
	})
	if rc != 0 {
		return fmt.Errorf("failed to set field value: %s", GetLastError())
	}
	return nil
}

// FieldFlags returns the widget field flags.
func (w *Widget) FieldFlags() int {
	if w.widget == nil || w.ctx == nil {
		return 0
	}
	var flags C.int
	w.ctx.WithLock(func() {
		flags = C.gomupdf_pdf_widget_field_flags(w.ctx.ctx, w.widget)
	})
	return int(flags)
}

// SetFieldFlags sets the widget field flags.
func (w *Widget) SetFieldFlags(flags int) error {
	if w.widget == nil || w.ctx == nil {
		return errors.New("widget is nil")
	}
	var rc C.int
	w.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_widget_set_field_flags(w.ctx.ctx, w.widget, C.int(flags))
	})
	if rc != 0 {
		return fmt.Errorf("failed to set field flags: %s", GetLastError())
	}
	return nil
}

// IsChecked returns whether a checkbox widget is checked.
func (w *Widget) IsChecked() bool {
	if w.widget == nil || w.ctx == nil {
		return false
	}
	var checked C.int
	w.ctx.WithLock(func() {
		checked = C.gomupdf_pdf_widget_is_checked(w.ctx.ctx, w.widget)
	})
	return checked != 0
}

// Toggle toggles a checkbox widget.
func (w *Widget) Toggle() error {
	if w.widget == nil || w.ctx == nil {
		return errors.New("widget is nil")
	}
	var rc C.int
	w.ctx.WithLock(func() {
		rc = C.gomupdf_pdf_widget_toggle(w.ctx.ctx, w.widget)
	})
	if rc != 0 {
		return fmt.Errorf("failed to toggle widget: %s", GetLastError())
	}
	return nil
}
