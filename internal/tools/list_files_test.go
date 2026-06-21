package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestListFiles(t *testing.T) {
	t.Run("should return files in current directory", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644))

		tool := tools.NewListFiles()
		got, err := tool.Execute(context.Background(), []byte(`{"path":"`+dir+`"}`))
		require.NoError(t, err)
		require.Contains(t, got, "a.txt")
		require.Contains(t, got, "b.txt")
	})

	t.Run("should return files recursively when recursive=true", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "deep.txt"), []byte("x"), 0644))

		tool := tools.NewListFiles()
		got, err := tool.Execute(context.Background(), []byte(`{"path":"`+dir+`","recursive":true}`))
		require.NoError(t, err)
		require.Contains(t, got, "deep.txt")
	})

	t.Run("should return error when dir does not exist", func(t *testing.T) {
		tool := tools.NewListFiles()
		_, err := tool.Execute(context.Background(), []byte(`{"path":"/nonexistent"}`))
		require.Error(t, err)
	})
}
