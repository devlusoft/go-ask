package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestGlob(t *testing.T) {
	t.Run("should find files matching simple pattern", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "c.go"), []byte("c"), 0644))

		tool := tools.NewGlob()
		got, err := tool.Execute(context.Background(), []byte(`{"pattern":"`+filepath.Join(dir, "*.txt")+`"}`))
		require.NoError(t, err)
		require.Contains(t, got, "a.txt")
		require.Contains(t, got, "b.txt")
		require.NotContains(t, got, "c.go")
	})

	t.Run("should support recursive ** pattern", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub", "deeper"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "deeper", "deep.txt"), []byte("x"), 0644))

		tool := tools.NewGlob()
		got, err := tool.Execute(context.Background(), []byte(`{"pattern":"`+filepath.Join(dir, "**", "*.txt")+`"}`))
		require.NoError(t, err)
		require.Contains(t, got, "deep.txt")
	})

	t.Run("should return error when pattern is empty", func(t *testing.T) {
		tool := tools.NewGlob()
		_, err := tool.Execute(context.Background(), []byte(`{"pattern":""}`))
		require.Error(t, err)
	})
}
