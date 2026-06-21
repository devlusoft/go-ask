package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	t.Run("should return entire file contents", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "hello.txt")
		require.NoError(t, os.WriteFile(path, []byte("line1\nline2\nline3"), 0644))

		tool := tools.NewReadFile()
		got, err := tool.Execute(context.Background(), []byte(`{"path":"`+path+`"}`))
		require.NoError(t, err)
		require.Equal(t, "line1\nline2\nline3", got)
	})

	t.Run("should return lines starting from offset", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "lines.txt")
		require.NoError(t, os.WriteFile(path, []byte("a\nb\nc\nd\ne"), 0644))

		tool := tools.NewReadFile()
		got, err := tool.Execute(context.Background(), []byte(`{"path":"`+path+`","offset":2}`))
		require.NoError(t, err)
		require.Equal(t, "c\nd\ne", got)
	})

	t.Run("should return at most limit lines", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "lines.txt")
		require.NoError(t, os.WriteFile(path, []byte("a\nb\nc\nd\ne"), 0644))

		tool := tools.NewReadFile()
		got, err := tool.Execute(context.Background(), []byte(`{"path":"`+path+`","offset":1,"limit":2}`))
		require.NoError(t, err)
		require.Equal(t, "b\nc", got)
	})

	t.Run("should return error when file does not exist", func(t *testing.T) {
		tool := tools.NewReadFile()
		_, err := tool.Execute(context.Background(), []byte(`{"path":"/nonexistent/path/file.txt"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "read_file")
	})

	t.Run("should return error when path is empty", func(t *testing.T) {
		tool := tools.NewReadFile()
		_, err := tool.Execute(context.Background(), []byte(`{"path":""}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "path is required")
	})

	t.Run("should truncate output when file exceeds max size", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "big.txt")
		require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("x", 200*1024)), 0644))

		tool := tools.NewReadFile()
		got, err := tool.Execute(context.Background(), []byte(`{"path":"`+path+`"}`))
		require.NoError(t, err)
		require.Contains(t, got, "... (truncated")
	})
}
