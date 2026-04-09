// Package cgo 提供对 MuPDF C 库的 CGO 绑定，封装 PDF 文档的底层操作。
//
// 本文件（widget.go）包含 PDF 表单控件（Widget）的 CGO 封装函数，涵盖：
//   - 控件遍历（FirstWidget、Next）
//   - 控件类型判断（Type）
//   - 字段名与值读写（FieldName、FieldValue、SetFieldValue）
//   - 字段标志读写（FieldFlags、SetFieldFlags）
//   - 复选框操作（IsChecked、Toggle）
//
// CGO 模式说明：
//   - Widget 内部包装 C.pdf_annot 指针（MuPDF 将表单控件视为一种特殊标注）
//   - 所有 MuPDF 调用均在 ctx.WithLock 回调中执行，保证线程安全
//   - Go 字符串通过 C.CString 转换，通过 defer C.free(unsafe.Pointer(...)) 释放
//   - C 端返回的字符串通过 C.GoString 转换为 Go string
//   - C.int 通过 int() 转换为 Go int
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

// Widget 表示 PDF 表单中的一个控件（内部包装 C.pdf_annot，MuPDF 将控件视为特殊标注）。
type Widget struct {
	ctx    *Context         // MuPDF 上下文
	widget *C.pdf_annot    // C 端 pdf_annot 指针（作为控件使用）
	doc    *C.fz_document  // 所属文档指针
}

// WidgetType 常量定义 PDF 表单控件的类型。
const (
	WidgetText     = 0 // 文本输入框
	WidgetCheckBox = 1 // 复选框
	WidgetRadio    = 2 // 单选按钮
	WidgetList     = 3 // 列表框
	WidgetChoice   = 4 // 下拉选择框
)

// FirstWidget 返回页面上的第一个表单控件。如果没有控件则返回 nil。
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

// Next 返回当前控件的下一个控件（遍历链表）。如果没有更多控件则返回 nil。
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

// Type 返回控件的类型（对应 WidgetType 常量）。
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

// FieldName 返回控件的字段名称。
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

// FieldValue 返回控件的字段值。
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

// SetFieldValue 设置控件的字段值。
// Go 字符串通过 C.CString 转换，使用后通过 defer C.free(unsafe.Pointer(...)) 释放。
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

// FieldFlags 返回控件的字段标志位。
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

// SetFieldFlags 设置控件的字段标志位。Go int 通过 C.int() 转换传入 C 端。
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

// IsChecked 返回复选框控件是否被勾选。C.int 通过 != 0 转换为 Go bool。
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

// Toggle 切换复选框控件的勾选状态。
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
