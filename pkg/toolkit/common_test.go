package toolkit

import (
	"testing"
)

func TestGetEnv_String(t *testing.T) {
	t.Setenv("TK_STR", "hello")
	got := GetEnv("TK_STR", "fallback")
	if got != "hello" {
		t.Fatalf("want hello, got %q", got)
	}
}

func TestGetEnv_StringFallback(t *testing.T) {
	got := GetEnv("TK_MISSING", "fallback")
	if got != "fallback" {
		t.Fatalf("want fallback, got %q", got)
	}
}

func TestGetEnv_IntOK(t *testing.T) {
	t.Setenv("TK_INT", "42")
	got := GetEnv("TK_INT", 7)
	if got != 42 {
		t.Fatalf("want 42, got %d", got)
	}
}

func TestGetEnv_IntBad_UseFallback(t *testing.T) {
	t.Setenv("TK_INT_BAD", "notanint")
	got := GetEnv("TK_INT_BAD", 99)
	if got != 99 {
		t.Fatalf("want 99 fallback, got %d", got)
	}
}

func TestGetEnv_BoolOK_True(t *testing.T) {
	t.Setenv("TK_BOOL", "true")
	got := GetEnv("TK_BOOL", false)
	if got != true {
		t.Fatalf("want true, got %v", got)
	}
}

func TestGetEnv_BoolBad_UseFallback(t *testing.T) {
	t.Setenv("TK_BOOL_BAD", "???")
	got := GetEnv("TK_BOOL_BAD", true)
	if got != true {
		t.Fatalf("want true fallback, got %v", got)
	}
}

// -------- MapStruct ----------

type srcOK struct {
	A int
	B string
}
type dstOK struct {
	A int
}

type srcBad struct {
	C chan int // json.Marshal fail with channel
}

func TestMapStruct_OK(t *testing.T) {
	s := srcOK{A: 7, B: "hi"}
	out := MapStruct[srcOK, dstOK](s)
	if out == nil {
		t.Fatalf("MapStruct returned nil")
	}
	if out.A != 7 {
		t.Fatalf("want A=7, got %d", out.A)
	}
}

func TestMapStruct_UnsupportedField_ReturnNil(t *testing.T) {
	s := srcBad{C: make(chan int)}
	out := MapStruct[srcBad, dstOK](s)
	if out != nil {
		t.Fatalf("expected nil due to marshal error")
	}
}
