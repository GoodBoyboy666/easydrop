package service

import "testing"

func TestNormalizePostListOrder(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "default", input: "", want: "pin desc, created_at desc"},
		{name: "created_at_desc", input: "created_at_desc", want: "pin desc, created_at desc"},
		{name: "created_at_asc", input: "created_at_asc", want: "pin desc, created_at asc"},
		{name: "invalid", input: "foo", want: "pin desc, created_at desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePostListOrder(tt.input); got != tt.want {
				t.Fatalf("normalizePostListOrder(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeServiceListPageSize(t *testing.T) {
	page, size := normalizeServiceListPageSize(0, 1000)
	if page != 1 || size != 100 {
		t.Fatalf("expected normalized page/size 1/100, got %d/%d", page, size)
	}
}

func TestNormalizeTagListOrder(t *testing.T) {
	if got := normalizeTagListOrder("HOT"); got != "hot_desc" {
		t.Fatalf("expected hot_desc, got %q", got)
	}
}

func TestNormalizeSettingListOrder(t *testing.T) {
	if got := normalizeSettingListOrder("key_desc"); got != "key desc" {
		t.Fatalf("expected key desc, got %q", got)
	}
	if got := normalizeSettingListOrder("invalid"); got != "key asc" {
		t.Fatalf("expected key asc fallback, got %q", got)
	}
}
