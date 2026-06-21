package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const maxOutputBytes = 100 * 1024

type ReadFile struct{}

func NewReadFile() Tool {
	return &ReadFile{}
}

func (t *ReadFile) Name() string { return "read_file" }

func (t *ReadFile) Description() string {
	return "Reads a file's contents. Use offset and limit to read large files in chunks."
}

func (t *ReadFile) Parameters() json.RawMessage {
	return json.RawMessage(`{
	"type": "object",
	"properties": {
		"path": {"type": "string", "description": "Path to the file, relative to CWD"},
		"offset": {"type": "integer", "description": "Line offset to start from (0-indexed)"},
		"limit": {"type": "integer", "description": "Maximum number of lines to read"}
	},
	"required": ["path"]
}`)
}

type readFileArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func (t *ReadFile) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var a readFileArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return "", fmt.Errorf("read_file: parsing args: %w", err)
	}
	if a.Path == "" {
		return "", fmt.Errorf("read_file: path is required")
	}

	if a.Offset == 0 && a.Limit == 0 {
		data, err := os.ReadFile(a.Path)
		if err != nil {
			return "", fmt.Errorf("read_file: %w", err)
		}
		return truncate(string(data), len(data)), nil
	}

	f, err := os.Open(a.Path)
	if err != nil {
		return "", fmt.Errorf("read_file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var lines []string
	lineNum := 0
	for scanner.Scan() {
		if lineNum < a.Offset {
			lineNum++
			continue
		}
		if a.Limit > 0 && len(lines) >= a.Limit {
			break
		}
		lines = append(lines, scanner.Text())
		lineNum++
	}
	if err = scanner.Err(); err != nil {
		return "", fmt.Errorf("read_file: scanning: %w", err)
	}
	return truncate(strings.Join(lines, "\n"), -1), nil
}

func truncate(s string, originalBytes int) string {
	if len(s) <= maxOutputBytes {
		return s
	}
	if originalBytes < 0 {
		originalBytes = len(s)
	}
	return s[:maxOutputBytes] + fmt.Sprintf("\n... (truncated, original size: %d bytes)", originalBytes)
}
