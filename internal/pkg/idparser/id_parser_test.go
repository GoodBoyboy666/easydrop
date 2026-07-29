package idparser

import (
	"strconv"
	"testing"
)

func TestParse(t *testing.T) {
	maxUint := ^uint(0)
	tests := []struct {
		name    string
		value   string
		want    uint
		wantErr bool
	}{
		{name: "valid", value: "42", want: 42},
		{name: "maximum uint", value: strconv.FormatUint(uint64(maxUint), 10), want: maxUint},
		{name: "negative", value: "-1", wantErr: true},
		{name: "not a number", value: "invalid", wantErr: true},
		{name: "out of range", value: valueAboveMaxUint(), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %d", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.value, err)
			}
			if got != tt.want {
				t.Fatalf("Parse(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func valueAboveMaxUint() string {
	if strconv.IntSize == 32 {
		return "4294967296"
	}
	return "18446744073709551616"
}
