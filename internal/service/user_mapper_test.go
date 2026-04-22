package service

import "testing"

func TestBuildGravatarURLUsesOfficialDefaultBase(t *testing.T) {
	url := buildGravatarURL(" Test@Example.com ", "")
	if url == nil {
		t.Fatal("expected gravatar url, got nil")
	}

	const expected = "https://www.gravatar.com/avatar/55502f40dc8b7c769880b10874abc9d0"
	if *url != expected {
		t.Fatalf("expected %q, got %q", expected, *url)
	}
}

func TestBuildGravatarURLUsesCustomBase(t *testing.T) {
	url := buildGravatarURL("test@example.com", "https://gravatar.example.com/avatar")
	if url == nil {
		t.Fatal("expected gravatar url, got nil")
	}

	const expected = "https://gravatar.example.com/avatar/55502f40dc8b7c769880b10874abc9d0"
	if *url != expected {
		t.Fatalf("expected %q, got %q", expected, *url)
	}
}

func TestBuildGravatarURLReturnsNilForEmptyEmail(t *testing.T) {
	url := buildGravatarURL("   ", "https://gravatar.example.com/avatar")
	if url != nil {
		t.Fatalf("expected nil for empty email, got %q", *url)
	}
}

func TestNormalizeGravatarBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty fallback",
			input:    "",
			expected: "https://www.gravatar.com/avatar/",
		},
		{
			name:     "trim and append slash",
			input:    " https://gravatar.example.com/avatar ",
			expected: "https://gravatar.example.com/avatar/",
		},
		{
			name:     "keep single slash",
			input:    "https://gravatar.example.com/avatar/",
			expected: "https://gravatar.example.com/avatar/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeGravatarBaseURL(tt.input); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
