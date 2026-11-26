package harnessvalue

import (
	"strings"
	"testing"
)

// Valid lexical paths survive unchanged; traversal, ambiguous platform forms,
// controls and zero values cannot become a transaction destination.
func TestRelativePath(t *testing.T) {
	for _, value := range []string{"instructions/root.md", "config/native", "skills/sample/SKILL.md"} {
		p, err := NewRelativePath(value)
		if err != nil || p.String() != value || !p.Valid() {
			t.Fatalf("valid path: %v", err)
		}
	}
	for _, value := range []string{"", ".", "..", "../outside", "x/../../outside", "/absolute", "a//b", "a/./b", "a/../b", "a/", "C:/native", "C:\\native", "a\x00b", "a\nb", strings.Repeat("x", MaxPathBytes+1)} {
		if _, err := NewRelativePath(value); err == nil {
			t.Fatal("unsafe path accepted")
		}
	}
	if (RelativePath{}).Valid() {
		t.Fatal("zero path accepted")
	}
}

// Owned bytes resist caller/accessor mutation and enforce the exact size bound.
func TestBytes(t *testing.T) {
	input := []byte("fixed")
	value, err := NewBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 'X'
	output := value.Copy()
	output[0] = 'Y'
	if string(value.Copy()) != "fixed" {
		t.Fatal("aliased bytes")
	}
	if _, err := NewBytes(make([]byte, MaxTextBytes)); err != nil {
		t.Fatal(err)
	}
	if _, err := NewBytes(make([]byte, MaxTextBytes+1)); err == nil {
		t.Fatal("oversized bytes accepted")
	}
}
