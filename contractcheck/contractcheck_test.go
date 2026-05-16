package contractcheck_test

import (
	"strings"
	"testing"

	"github.com/egidinas/signalforge/contractcheck"
	"github.com/egidinas/signalforge/errorsx"
)

func TestRequireNonBlank(t *testing.T) {
	if err := contractcheck.RequireNonBlank("name", "hello"); err != nil {
		t.Fatalf("non-blank value: unexpected error: %v", err)
	}
	for _, bad := range []string{"", "  ", "\t"} {
		err := contractcheck.RequireNonBlank("name", bad)
		if err == nil {
			t.Fatalf("blank value %q: expected error, got nil", bad)
		}
		if !errorsx.IsCode(err, errorsx.CodeInvalidArgument) {
			t.Fatalf("blank value: expected CodeInvalidArgument, got %v", errorsx.CodeOf(err))
		}
		if !strings.Contains(err.Error(), "name") {
			t.Fatalf("error should mention field name: %v", err)
		}
	}
}

func TestRequireEnum(t *testing.T) {
	type Mode string
	const (
		ModeA Mode = "a"
		ModeB Mode = "b"
	)
	if err := contractcheck.RequireEnum("mode", ModeA, ModeA, ModeB); err != nil {
		t.Fatalf("valid enum: unexpected error: %v", err)
	}
	err := contractcheck.RequireEnum("mode", Mode("c"), ModeA, ModeB)
	if err == nil {
		t.Fatal("invalid enum: expected error, got nil")
	}
	if !errorsx.IsCode(err, errorsx.CodeInvalidArgument) {
		t.Fatalf("invalid enum: expected CodeInvalidArgument, got %v", errorsx.CodeOf(err))
	}
}

func TestRequirePositive(t *testing.T) {
	if err := contractcheck.RequirePositive("count", 1); err != nil {
		t.Fatalf("positive value: unexpected error: %v", err)
	}
	for _, bad := range []int{0, -1, -100} {
		err := contractcheck.RequirePositive("count", bad)
		if err == nil {
			t.Fatalf("value %d: expected error, got nil", bad)
		}
		if !errorsx.IsCode(err, errorsx.CodeInvalidArgument) {
			t.Fatalf("value %d: expected CodeInvalidArgument", bad)
		}
	}
}

func TestSafeDisplayString(t *testing.T) {
	if got := contractcheck.SafeDisplayString("hello", 3); got != "hel" {
		t.Fatalf("truncate: got %q, want %q", got, "hel")
	}
	if got := contractcheck.SafeDisplayString("ab\x00cd", 10); got != "ab?cd" {
		t.Fatalf("control char replacement: got %q, want %q", got, "ab?cd")
	}
	if got := contractcheck.SafeDisplayString("", 5); got != "" {
		t.Fatalf("empty: got %q", got)
	}
	if got := contractcheck.SafeDisplayString("hello", 0); got != "" {
		t.Fatalf("maxLen=0: got %q", got)
	}
}

func TestRejectControlChars(t *testing.T) {
	if err := contractcheck.RejectControlChars("id", "valid-id_123"); err != nil {
		t.Fatalf("clean string: unexpected error: %v", err)
	}
	for _, bad := range []string{"\x00", "foo\nbar", "tab\there"} {
		err := contractcheck.RejectControlChars("id", bad)
		if err == nil {
			t.Fatalf("control char in %q: expected error, got nil", bad)
		}
		if !errorsx.IsCode(err, errorsx.CodeInvalidArgument) {
			t.Fatalf("control char: expected CodeInvalidArgument, got %v", errorsx.CodeOf(err))
		}
		if !strings.Contains(err.Error(), "id") {
			t.Fatalf("error should mention field name: %v", err)
		}
	}
}

func TestHasControlChars(t *testing.T) {
	for _, clean := range []string{"valid-id_123", "line\nwith\ttabs\rand carriage returns"} {
		if contractcheck.HasControlChars(clean) {
			t.Fatalf("clean string %q: expected no disallowed control chars", clean)
		}
	}
	for _, bad := range []string{"\x00", "\x7f", "\u0085"} {
		if !contractcheck.HasControlChars(bad) {
			t.Fatalf("control char %q: expected detection", bad)
		}
	}
}
