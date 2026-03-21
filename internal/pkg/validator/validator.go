package validator

import (
	"errors"
	"regexp"
	"unicode"
)

const (
	usernameMinLength = 3
	usernameMaxLength = 32
	passwordMinLength = 8
)

var (
	ErrEmptyUsername         = errors.New("用户名不能为空")
	ErrUsernameTooShort      = errors.New("用户名长度不能小于 3")
	ErrUsernameTooLong       = errors.New("用户名长度不能超过 32")
	ErrInvalidUsernameFormat = errors.New("用户名只能包含字母、数字和下划线")

	ErrEmptyPassword         = errors.New("密码不能为空")
	ErrPasswordTooShort      = errors.New("密码长度不能小于 8")
	ErrPasswordContainsSpace = errors.New("密码不能包含空白字符")
	ErrPasswordMissingLetter = errors.New("密码必须至少包含一个字母")
	ErrPasswordMissingNumber = errors.New("密码必须至少包含一个数字")

	ErrEmptyEmail         = errors.New("邮箱不能为空")
	ErrInvalidEmailFormat = errors.New("邮箱格式不合法")
)

var (
	usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	// 常见邮箱格式：支持子域名与长 TLD，不支持本地部分引号等 RFC 特例。
	emailPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._%+-]{0,62}[A-Za-z0-9])?@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)+$`)
)

func ValidateUsername(username string) error {
	if username == "" {
		return ErrEmptyUsername
	}
	if len(username) < usernameMinLength {
		return ErrUsernameTooShort
	}
	if len(username) > usernameMaxLength {
		return ErrUsernameTooLong
	}
	if !usernamePattern.MatchString(username) {
		return ErrInvalidUsernameFormat
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < passwordMinLength {
		return ErrPasswordTooShort
	}

	hasLetter := false
	hasNumber := false
	for _, r := range password {
		if unicode.IsSpace(r) {
			return ErrPasswordContainsSpace
		}
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasNumber = true
		}
	}

	if !hasLetter {
		return ErrPasswordMissingLetter
	}
	if !hasNumber {
		return ErrPasswordMissingNumber
	}

	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmptyEmail
	}
	if !emailPattern.MatchString(email) {
		return ErrInvalidEmailFormat
	}

	return nil
}
