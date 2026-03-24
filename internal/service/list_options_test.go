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
