package contractcheck

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/egidinas/signalforge/errorsx"
)

// RequireNonBlank returns CodeInvalidArgument if value is empty or whitespace-only.
func RequireNonBlank(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return errorsx.WithCode(fmt.Errorf("%s must not be blank", field), errorsx.CodeInvalidArgument)
	}
	return nil
}

// RequireEnum returns CodeInvalidArgument if value is not one of the allowed values.
func RequireEnum[T ~string](field string, value T, allowed ...T) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return errorsx.WithCode(
		fmt.Errorf("%s %q is not a valid value", field, value),
		errorsx.CodeInvalidArgument,
	)
}

// RequirePositive returns CodeInvalidArgument if value is not greater than zero.
func RequirePositive(field string, value int) error {
	if value <= 0 {
		return errorsx.WithCode(fmt.Errorf("%s must be positive, got %d", field, value), errorsx.CodeInvalidArgument)
	}
	return nil
}

// SafeDisplayString truncates value to maxLen runes and replaces non-printable
// runes with '?'. It always returns valid UTF-8.
func SafeDisplayString(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	var b strings.Builder
	count := 0
	for _, r := range value {
		if count >= maxLen {
			break
		}
		if r == utf8.RuneError || !utf8.ValidRune(r) || unicode.IsControl(r) {
			b.WriteRune('?')
		} else {
			b.WriteRune(r)
		}
		count++
	}
	return b.String()
}

// RejectControlChars returns CodeInvalidArgument if value contains a Unicode
// control character.
func RejectControlChars(field, value string) error {
	for i, r := range value {
		if unicode.IsControl(r) {
			return errorsx.WithCode(
				fmt.Errorf("%s contains control character at byte %d", field, i),
				errorsx.CodeInvalidArgument,
			)
		}
	}
	return nil
}

// HasControlChars checks for control characters in a string, excluding common whitespace.
func HasControlChars(value string) bool {
	for _, r := range value {
		if r < 32 && r != '\n' && r != '\t' && r != '\r' {
			return true
		}
	}
	return false
}
