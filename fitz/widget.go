// Package fitz 提供了对 MuPDF 库的高层 Go 封装，用于处理 PDF 和其他文档格式。
// 本文件（widget.go）封装了 PDF 表单控件（Widget）的操作，包括文本框、复选框、单选按钮等。
package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// WidgetType 常量定义了 PDF 表单控件的类型。
const (
	WidgetText     = cgo_bindings.WidgetText     // 文本输入框
	WidgetCheckBox = cgo_bindings.WidgetCheckBox // 复选框
	WidgetRadio    = cgo_bindings.WidgetRadio    // 单选按钮
	WidgetList     = cgo_bindings.WidgetList     // 列表框
	WidgetChoice   = cgo_bindings.WidgetChoice   // 下拉选择框
)

// Widget 表示一个 PDF 表单控件，封装了 MuPDF 的 Widget 指针。
type Widget struct {
	ctx    *cgo_bindings.Context   // MuPDF 上下文
	widget *cgo_bindings.Widget    // MuPDF 控件指针
}

// Type 返回控件的类型。
func (w *Widget) Type() int {
	if w.widget == nil {
		return -1
	}
	return w.widget.Type()
}

// FieldName 返回控件的字段名称。
func (w *Widget) FieldName() string {
	if w.widget == nil {
		return ""
	}
	return w.widget.FieldName()
}

// FieldValue 返回控件的字段值。
func (w *Widget) FieldValue() string {
	if w.widget == nil {
		return ""
	}
	return w.widget.FieldValue()
}

// SetFieldValue 设置控件的字段值。
func (w *Widget) SetFieldValue(value string) error {
	if w.widget == nil {
		return fmt.Errorf("widget is nil")
	}
	return w.widget.SetFieldValue(value)
}

// FieldFlags 返回控件的字段标志位。
func (w *Widget) FieldFlags() int {
	if w.widget == nil {
		return 0
	}
	return w.widget.FieldFlags()
}

// IsChecked 判断复选框控件是否被勾选。
func (w *Widget) IsChecked() bool {
	if w.widget == nil {
		return false
	}
	return w.widget.IsChecked()
}

// Toggle 切换复选框控件的勾选状态。
func (w *Widget) Toggle() error {
	if w.widget == nil {
		return fmt.Errorf("widget is nil")
	}
	return w.widget.Toggle()
}

// String 返回控件的字符串描述信息。
func (w *Widget) String() string {
	if w.widget == nil {
		return "Widget(<nil>)"
	}
	return fmt.Sprintf("Widget(type=%d, name=%q, value=%q)", w.Type(), w.FieldName(), w.FieldValue())
}
