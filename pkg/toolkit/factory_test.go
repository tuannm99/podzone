package toolkit

import (
	"context"
	"testing"
)

func TestRegistry_Register_Use_Get_Lookup(t *testing.T) {
	type cfg string

	// factory A: length of string cfg + 1
	fA := func(_ context.Context, c cfg) (int, error) { return len(c) + 1, nil }
	// factory B: +2
	fB := func(_ context.Context, c cfg) (int, error) { return len(c) + 2, nil }

	reg := NewRegistry[int, cfg]("A")
	reg.Register("A", fA)
	reg.Register("B", fB)

	// default A
	f := reg.Get()
	v, _ := f(context.Background(), "abc")
	if v != 4 { // 3 + 1
		t.Fatalf("want 4, got %d", v)
	}

	reg.Use("B")
	f = reg.Get()
	v, _ = f(context.Background(), "abc")
	if v != 5 {
		t.Fatalf("want 5, got %d", v)
	}

	if _, ok := reg.Lookup("A"); !ok {
		t.Fatalf("lookup A failed")
	}
	if _, ok := reg.Lookup("not-exist"); ok {
		t.Fatalf("lookup returned ok for non-existing id")
	}
}
