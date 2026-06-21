package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestGrep(t *testing.T) {
	t.Run("should find matches in file:line:content format", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "x.txt")
		require.NoError(t, os.WriteFile(path, []byte("alpha\nbeta\nalpha again"), 0644))

		tool := tools.NewGrep()
		got, err := tool.Execute(context.Background(), []byte(`{"pattern":"alpha","path":"`+dir+`"}`))
		require.NoError(t, err)
		require.Contains(t, got, "x.txt:1:alpha")
		require.Contains(t, got, "x.txt:3:alpha again")
	})

	t.Run("should filter by include pattern", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "x.go"), []byte("func A() {}"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "x.txt"), []byte("func B() {}"), 0644))

		tool := tools.NewGrep()
		got, err := tool.Execute(context.Background(), []byte(`{"pattern":"func","include":"*.go","path":"`+dir+`"}`))
		require.NoError(t, err)
		require.Contains(t, got, "x.go")
		require.NotContains(t, got, "x.txt")
	})

	t.Run("should return error when pattern is invalid regex", func(t *testing.T) {
		tool := tools.NewGrep()
		_, err := tool.Execute(context.Background(), []byte(`{"pattern":"[invalid"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid pattern")
	})
}
