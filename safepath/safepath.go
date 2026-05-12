// Package safepath resolves user-supplied relative paths under a fixed root.
package safepath

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	ErrEmptyRoot    = errors.New("root must not be empty")
	ErrUnsafePath   = errors.New("path must be a non-empty relative path below root")
	ErrEscapingRoot = errors.New("resolved path escapes root")
)

// ResolveUnderRoot returns userPath resolved below root.
//
// userPath must be a non-empty relative path. Traversal segments (..) are
// rejected outright, even when the cleaned path would remain under root -
// traversal in user-supplied paths is treated as an intent signal, not just
// an outcome check. This is lexical path containment; callers that dereference
// paths through untrusted directories must still account for symlinks.
func ResolveUnderRoot(root, userPath string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", ErrEmptyRoot
	}
	if isUnsafeUserPath(userPath) {
		return "", ErrUnsafePath
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	candidate := filepath.Join(rootAbs, filepath.FromSlash(userPath))
	rel, err := filepath.Rel(rootAbs, candidate)
	if err != nil {
		return "", fmt.Errorf("compare resolved path to root: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", ErrEscapingRoot
	}

	return candidate, nil
}

func isUnsafeUserPath(userPath string) bool {
	if strings.TrimSpace(userPath) == "" {
		return true
	}
	if strings.Contains(userPath, "\x00") {
		return true
	}
	if filepath.IsAbs(userPath) || userPath == "." {
		return true
	}
	if looksLikeWindowsAbsolute(userPath) {
		return true
	}

	normalized := strings.ReplaceAll(userPath, "\\", "/")
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return true
		}
	}

	return false
}

func looksLikeWindowsAbsolute(userPath string) bool {
	if strings.HasPrefix(userPath, `\\`) || strings.HasPrefix(userPath, `//`) {
		return true
	}
	if len(userPath) >= 3 && userPath[1] == ':' && (userPath[2] == '\\' || userPath[2] == '/') {
		drive := userPath[0]
		return (drive >= 'a' && drive <= 'z') || (drive >= 'A' && drive <= 'Z')
	}
	return false
}
