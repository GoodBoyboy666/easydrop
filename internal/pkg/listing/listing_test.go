package listing

import (
	"math"
	"testing"
)

func TestBoundsNormalizePageSize(t *testing.T) {
	bounds := Bounds{DefaultPage: 1, DefaultSize: 20, MaxSize: 100}

	page, size := bounds.NormalizePageSize(0, 0)
	if page != 1 || size != 20 {
		t.Fatalf("expected default page/size 1/20, got %d/%d", page, size)
	}

	page, size = bounds.NormalizePageSize(3, 200)
	if page != 3 || size != 100 {
		t.Fatalf("expected normalized page/size 3/100, got %d/%d", page, size)
	}
}

func TestBoundsNormalizeLimitOffset(t *testing.T) {
	bounds := Bounds{DefaultSize: 20, MaxSize: 100}

	limit, offset := bounds.NormalizeLimitOffset(0, -1)
	if limit != 20 || offset != 0 {
		t.Fatalf("expected normalized limit/offset 20/0, got %d/%d", limit, offset)
	}

	limit, offset = bounds.NormalizeLimitOffset(200, 15)
	if limit != 100 || offset != 15 {
		t.Fatalf("expected normalized limit/offset 100/15, got %d/%d", limit, offset)
	}
}

func TestBoundsOffsetFromPage(t *testing.T) {
	bounds := Bounds{DefaultPage: 1, DefaultSize: 20, MaxSize: 100}

	if got := bounds.OffsetFromPage(3, 10); got != 20 {
		t.Fatalf("expected offset 20, got %d", got)
	}

	if got := bounds.OffsetFromPage(math.MaxInt, 100); got != math.MaxInt {
		t.Fatalf("expected saturated max int offset, got %d", got)
	}
}

func TestSQLOrderMapResolve(t *testing.T) {
	orders := SQLOrderMap{
		"created_at_asc":  "created_at asc",
		"created_at_desc": "created_at desc",
	}

	if got := orders.Resolve(" CREATED_AT_ASC ", "created_at_desc"); got != "created_at asc" {
		t.Fatalf("expected mapped asc order, got %q", got)
	}

	if got := orders.Resolve("invalid", "created_at_desc"); got != "created_at desc" {
		t.Fatalf("expected default mapped desc order, got %q", got)
	}
}
