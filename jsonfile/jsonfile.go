// Package jsonfile provides helpers for writing JSON to disk.
// WriteIndent writes a value as pretty-printed JSON to a single file;
// AppendLine appends a single JSON-encoded record followed by a newline (JSONL).
package jsonfile

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// WriteIndent marshals value as indented JSON and writes it to path. Any
// missing parent directories are created with mode 0o755. The file is written
// with mode 0o644 and ends with a single trailing newline.
func WriteIndent(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// AppendLine marshals value as compact JSON and appends one line (followed by
// '\n') to path. Any missing parent directories are created with mode 0o755.
// path must be non-empty.
func AppendLine(path string, value any) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("jsonfile: path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := openAppendFile(path)
	if err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

type appendFile interface {
	io.Writer
	io.Closer
}

var openAppendFile = func(path string) (appendFile, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
