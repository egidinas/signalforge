package jsonfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteIndentCreatesFileAndDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "out.json")

	value := map[string]any{"key": "value", "n": 42}
	if err := WriteIndent(path, value); err != nil {
		t.Fatalf("WriteIndent: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty file")
	}
	if data[len(data)-1] != '\n' {
		t.Errorf("expected trailing newline")
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("round-trip mismatch: %v", got)
	}
}

func TestAppendLineCreatesAndAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")

	records := []map[string]any{{"id": 1}, {"id": 2}, {"id": 3}}
	for _, r := range records {
		if err := AppendLine(path, r); err != nil {
			t.Fatalf("AppendLine: %v", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := splitLines(string(data))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var got map[string]any
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d: %v", i, err)
		}
	}
}

func TestAppendLineRejectsEmptyPath(t *testing.T) {
	if err := AppendLine("", map[string]any{}); err == nil {
		t.Fatal("expected error for empty path")
	}
	if err := AppendLine("   ", map[string]any{}); err == nil {
		t.Fatal("expected error for blank path")
	}
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range splitOnNewline(s) {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func splitOnNewline(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
