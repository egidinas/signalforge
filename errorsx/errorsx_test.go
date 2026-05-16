package errorsx_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/egidinas/signalforge/errorsx"
)

func TestWrappedErrorPreservesErrorsIs(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := errorsx.WithCode(sentinel, errorsx.CodeNotFound)
	if !errors.Is(wrapped, sentinel) {
		t.Fatal("errors.Is should find sentinel through WithCode wrapper")
	}
}

func TestCodeOfReturnsOuterCode(t *testing.T) {
	inner := errorsx.WithCode(errors.New("inner"), errorsx.CodeInternal)
	outer := errorsx.WithCode(inner, errorsx.CodeConflict)
	if got := errorsx.CodeOf(outer); got != errorsx.CodeConflict {
		t.Fatalf("CodeOf outer = %q, want %q", got, errorsx.CodeConflict)
	}
}

func TestCodeOfUnknownErrorReturnsInternal(t *testing.T) {
	if got := errorsx.CodeOf(errors.New("bare")); got != errorsx.CodeInternal {
		t.Fatalf("CodeOf bare error = %q, want %q", got, errorsx.CodeInternal)
	}
}

func TestIsCode(t *testing.T) {
	err := errorsx.WithCode(errors.New("not found"), errorsx.CodeNotFound)
	if !errorsx.IsCode(err, errorsx.CodeNotFound) {
		t.Fatal("IsCode should return true for matching code")
	}
	if errorsx.IsCode(err, errorsx.CodeConflict) {
		t.Fatal("IsCode should return false for non-matching code")
	}
}

func TestHTTPStatusMapping(t *testing.T) {
	cases := []struct {
		code errorsx.Code
		want int
	}{
		{errorsx.CodeInvalidArgument, http.StatusBadRequest},
		{errorsx.CodeNotFound, http.StatusNotFound},
		{errorsx.CodeConflict, http.StatusConflict},
		{errorsx.CodeUnauthorized, http.StatusUnauthorized},
		{errorsx.CodeForbidden, http.StatusForbidden},
		{errorsx.CodeDeadline, http.StatusGatewayTimeout},
		{errorsx.CodeUnavailable, http.StatusServiceUnavailable},
		{errorsx.CodeInternal, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		if got := errorsx.HTTPStatus(tc.code); got != tc.want {
			t.Errorf("HTTPStatus(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
}
