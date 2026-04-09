package cgo

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/mupdf/lib -lmupdf -lmupdfcpp
#cgo CFLAGS: -I/opt/homebrew/opt/mupdf/include

#include "bindings.h"
#include <stdlib.h>
*/
import "C"
import (
	"sync"
)

// Context manages the MuPDF context.
type Context struct {
	ctx *C.fz_context
	mu  sync.Mutex
}

// New creates a new MuPDF context.
func New() *Context {
	c := C.gomupdf_new_context()
	if c == nil {
		panic("failed to create MuPDF context")
	}
	ctx := &Context{ctx: c}
	// Note: No SetFinalizer - caller must explicitly call Destroy()
	// This avoids crashes when GC runs Destroy on a corrupted context
	return ctx
}

// NewContext creates a new MuPDF context (alias for New for clarity).
func NewContext() *Context {
	return New()
}

// Lock acquires the context lock for thread-safe operations.
func (c *Context) Lock() {
	c.mu.Lock()
}

// Unlock releases the context lock.
func (c *Context) Unlock() {
	c.mu.Unlock()
}

// Destroy releases the MuPDF context.
func (c *Context) Destroy() {
	if c.ctx != nil {
		C.gomupdf_drop_context(c.ctx)
		c.ctx = nil
	}
}

// WithLock executes a function while holding the context lock.
func (c *Context) WithLock(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn()
}

// Version returns the MuPDF version string.
func Version() string {
	return C.GoString(C.gomupdf_version())
}

// GetLastError returns the last MuPDF error message.
func GetLastError() string {
	return C.GoString(C.gomupdf_get_last_error())
}

// ClearError clears the last MuPDF error message.
func ClearError() {
	C.gomupdf_clear_error()
}
