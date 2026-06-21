package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func TestGitDiff(t *testing.T) {
	t.Run("should return diff of unstaged changes", func(t *testing.T) {
		repo := initGitRepo(t)
		t.Chdir(repo)

		path := filepath.Join(repo, "x.txt")
		require.NoError(t, os.WriteFile(path, []byte("hello\n"), 0644))
		runGit(t, repo, "add", "x.txt")
		runGit(t, repo, "commit", "-m", "add x")
		require.NoError(t, os.WriteFile(path, []byte("hello world\n"), 0644))

		tool := tools.NewGitDiff()
		got, err := tool.Execute(context.Background(), []byte(`{}`))
		require.NoError(t, err)
		require.Contains(t, got, "hello world")
		require.Contains(t, got, "+")
	})

	t.Run("should return empty string when no diff", func(t *testing.T) {
		repo := initGitRepo(t)
		t.Chdir(repo)

		tool := tools.NewGitDiff()
		got, err := tool.Execute(context.Background(), []byte(`{}`))
		require.NoError(t, err)
		require.Equal(t, "", got)
	})

	t.Run("should return error when not in a git repo", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		tool := tools.NewGitDiff()
		_, err := tool.Execute(context.Background(), []byte(`{}`))
		require.Error(t, err)
	})
}
