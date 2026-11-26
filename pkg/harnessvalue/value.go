// Package harnessvalue provides bounded, immutable values for harness policies.
// It performs lexical validation only; filesystem admission belongs to callers.
package harnessvalue

import (
	"fmt"
	"path"
	"strings"
)

// The following constants define the closed contract vocabulary and bounds.
const (
	MaxTextBytes = 1 << 20
	MaxPathBytes = 4096
	MaxItems     = 1024
)

// RelativePath cannot identify a root, escape it, or use platform-specific drives.
type RelativePath struct{ value string }

// NewRelativePath validates a canonical relative lexical path.
func NewRelativePath(value string) (RelativePath, error) {
	if value == "" || len(value) > MaxPathBytes || value == "." || path.IsAbs(value) || path.Clean(value) != value || strings.ContainsAny(value, "\\:\x00\r\n") {
		return RelativePath{}, fmt.Errorf("harness_path_invalid")
	}
	for _, part := range strings.Split(value, "/") {
		if part == ".." || part == "." || part == "" {
			return RelativePath{}, fmt.Errorf("harness_path_invalid")
		}
	}
	return RelativePath{value}, nil
}

// String returns the validated relative spelling.
func (p RelativePath) String() string { return p.value }

// Valid reports whether the value has a valid lexical path.
func (p RelativePath) Valid() bool { _, err := NewRelativePath(p.value); return err == nil }

// Bytes owns its storage. Neither construction nor access shares mutable data.
type Bytes struct{ value []byte }

// NewBytes bounds and copies bytes into independent storage.
func NewBytes(value []byte) (Bytes, error) {
	if len(value) > MaxTextBytes {
		return Bytes{}, fmt.Errorf("harness_value_too_large")
	}
	return Bytes{append([]byte(nil), value...)}, nil
}

// Copy returns independent byte storage.
func (b Bytes) Copy() []byte { return append([]byte(nil), b.value...) }

// Len returns the stored byte count.
func (b Bytes) Len() int { return len(b.value) }

// Text rejects control bytes and oversized scalar values.
func Text(value string) error {
	if len(value) > MaxPathBytes || strings.ContainsAny(value, "\x00\r\n") {
		return fmt.Errorf("harness_text_invalid")
	}
	return nil
}
