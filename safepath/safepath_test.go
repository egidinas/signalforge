package safepath

import (
	"path/filepath"
	"testing"
)

func TestResolveUnderRootAcceptsRelativePath(t *testing.T) {
	root := t.TempDir()

	got, err := ResolveUnderRoot(root, "fixtures/example.json")
	if err != nil {
		t.Fatalf("ResolveUnderRoot returned error: %v", err)
	}

	want := filepath.Join(root, "fixtures", "example.json")
	if got != want {
		t.Fatalf("ResolveUnderRoot() = %q, want %q", got, want)
	}
}

func TestResolveUnderRootRejectsEmptyRoot(t *testing.T) {
	if _, err := ResolveUnderRoot("", "fixtures/example.json"); err == nil {
		t.Fatal("ResolveUnderRoot accepted an empty root")
	}
}

func TestResolveUnderRootRejectsUnsafePaths(t *testing.T) {
	root := t.TempDir()
	tests := []string{
		"",
		"   ",
		".",
		"/tmp/example.json",
		"../example.json",
		"fixtures/../example.json",
		"fixtures/../../example.json",
		`C:\tmp\example.json`,
		`\\server\share\example.json`,
		"fixtures/\x00evil.json",
		"sub\x00/../etc/passwd",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := ResolveUnderRoot(root, input); err == nil {
				t.Fatalf("ResolveUnderRoot accepted unsafe path %q", input)
			}
		})
	}
}

func TestResolveUnderRootRejectsEscapingPath(t *testing.T) {
	root := t.TempDir()

	// Contains ".." segment so rejected by syntactic guard.
	if _, err := ResolveUnderRoot(root, "fixtures/../../outside.json"); err == nil {
		t.Fatal("ResolveUnderRoot accepted a path escaping the root")
	}
}

func TestResolveUnderRootRejectsWindowsUNCPrefix(t *testing.T) {
	root := t.TempDir()
	cases := []string{"//etc/passwd", `\\server\share`}
	for _, c := range cases {
		if _, err := ResolveUnderRoot(root, c); err == nil {
			t.Fatalf("ResolveUnderRoot accepted UNC-style path %q", c)
		}
	}
}
