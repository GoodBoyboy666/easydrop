package validator

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid simple", input: "user_01"},
		{name: "valid edge min", input: "abc"},
		{name: "valid edge max", input: strings.Repeat("a", 32)},
		{name: "valid underscore prefix", input: "_abc"},
		{name: "valid underscore suffix", input: "abc_"},
		{name: "empty", input: "", wantErr: ErrEmptyUsername},
		{name: "too short", input: "ab", wantErr: ErrUsernameTooShort},
		{name: "too long", input: strings.Repeat("a", 33), wantErr: ErrUsernameTooLong},
		{name: "contains hyphen", input: "ab-c", wantErr: ErrInvalidUsernameFormat},
		{name: "contains space", input: "ab c", wantErr: ErrInvalidUsernameFormat},
		{name: "contains chinese", input: "用户名", wantErr: ErrInvalidUsernameFormat},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUsername(tc.input)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("期望无错误，实际为: %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("期望错误 %v，实际为: %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid letters and numbers", input: "abc12345"},
		{name: "valid symbols", input: "Abc12345!@#"},
		{name: "empty", input: "", wantErr: ErrEmptyPassword},
		{name: "too short", input: "a1b2c3", wantErr: ErrPasswordTooShort},
		{name: "missing letter", input: "12345678", wantErr: ErrPasswordMissingLetter},
		{name: "missing number", input: "abcdefgh", wantErr: ErrPasswordMissingNumber},
		{name: "contains space", input: "abc 12345", wantErr: ErrPasswordContainsSpace},
		{name: "contains tab", input: "abc\t12345", wantErr: ErrPasswordContainsSpace},
		{name: "contains newline", input: "abc\n12345", wantErr: ErrPasswordContainsSpace},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.input)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("期望无错误，实际为: %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("期望错误 %v，实际为: %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid simple", input: "user@example.com"},
		{name: "valid plus tag", input: "user+tag@sub.example.com"},
		{name: "valid long tld", input: "user@example.technology"},
		{name: "empty", input: "", wantErr: ErrEmptyEmail},
		{name: "missing at", input: "user.example.com", wantErr: ErrInvalidEmailFormat},
		{name: "double at", input: "user@@example.com", wantErr: ErrInvalidEmailFormat},
		{name: "quoted local part", input: "\"user\"@example.com", wantErr: ErrInvalidEmailFormat},
		{name: "localhost domain", input: "user@localhost", wantErr: ErrInvalidEmailFormat},
		{name: "invalid domain hyphen", input: "user@-example.com", wantErr: ErrInvalidEmailFormat},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEmail(tc.input)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("期望无错误，实际为: %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("期望错误 %v，实际为: %v", tc.wantErr, err)
			}
		})
	}
}

