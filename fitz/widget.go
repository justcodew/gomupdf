package fitz

import (
	"fmt"

	cgo_bindings "github.com/go-pymupdf/gomupdf/cgo_bindings"
)

// WidgetType constants.
const (
	WidgetText     = cgo_bindings.WidgetText
	WidgetCheckBox = cgo_bindings.WidgetCheckBox
	WidgetRadio    = cgo_bindings.WidgetRadio
	WidgetList     = cgo_bindings.WidgetList
	WidgetChoice   = cgo_bindings.WidgetChoice
)

// Widget represents a PDF form widget.
type Widget struct {
	ctx    *cgo_bindings.Context
	widget *cgo_bindings.Widget
}

// Type returns the widget type.
func (w *Widget) Type() int {
	if w.widget == nil {
		return -1
	}
	return w.widget.Type()
}

// FieldName returns the widget field name.
func (w *Widget) FieldName() string {
	if w.widget == nil {
		return ""
	}
	return w.widget.FieldName()
}

// FieldValue returns the widget field value.
func (w *Widget) FieldValue() string {
	if w.widget == nil {
		return ""
	}
	return w.widget.FieldValue()
}

// SetFieldValue sets the widget field value.
func (w *Widget) SetFieldValue(value string) error {
	if w.widget == nil {
		return fmt.Errorf("widget is nil")
	}
	return w.widget.SetFieldValue(value)
}

// FieldFlags returns the widget field flags.
func (w *Widget) FieldFlags() int {
	if w.widget == nil {
		return 0
	}
	return w.widget.FieldFlags()
}

// IsChecked returns whether a checkbox widget is checked.
func (w *Widget) IsChecked() bool {
	if w.widget == nil {
		return false
	}
	return w.widget.IsChecked()
}

// Toggle toggles a checkbox widget.
func (w *Widget) Toggle() error {
	if w.widget == nil {
		return fmt.Errorf("widget is nil")
	}
	return w.widget.Toggle()
}

// String returns a string representation.
func (w *Widget) String() string {
	if w.widget == nil {
		return "Widget(<nil>)"
	}
	return fmt.Sprintf("Widget(type=%d, name=%q, value=%q)", w.Type(), w.FieldName(), w.FieldValue())
}
