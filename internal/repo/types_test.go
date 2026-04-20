package repo

import "testing"

func TestNormalizeListOptions(t *testing.T) {
	opts := normalizeListOptions(ListOptions{
		Limit:  0,
		Offset: -5,
		Order:  "",
	}, "created_at desc")

	if opts.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", opts.Limit)
	}
	if opts.Offset != 0 {
		t.Fatalf("expected offset 0, got %d", opts.Offset)
	}
	if opts.Order != "created_at desc" {
		t.Fatalf("expected default order, got %q", opts.Order)
	}
}

func TestNormalizeListOptionsCapsLimit(t *testing.T) {
	opts := normalizeListOptions(ListOptions{
		Limit:  200,
		Offset: 10,
		Order:  "KEY ASC",
	}, "created_at desc")

	if opts.Limit != 100 {
		t.Fatalf("expected capped limit 100, got %d", opts.Limit)
	}
	if opts.Offset != 10 {
		t.Fatalf("expected offset 10, got %d", opts.Offset)
	}
	if opts.Order != "key asc" {
		t.Fatalf("expected normalized order key asc, got %q", opts.Order)
	}
}
