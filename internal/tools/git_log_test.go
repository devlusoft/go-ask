package tools_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/stretchr/testify/require"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "initial commit")
	runGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "second commit")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func TestGitLog(t *testing.T) {
	t.Run("should return recent commits in a git repo", func(t *testing.T) {
		repo := initGitRepo(t)
		t.Chdir(repo)

		tool := tools.NewGitLog()
		got, err := tool.Execute(context.Background(), []byte(`{"n":5}`))
		require.NoError(t, err)
		require.Contains(t, got, "initial commit")
		require.Contains(t, got, "second commit")
	})

	t.Run("should respect the n parameter", func(t *testing.T) {
		repo := initGitRepo(t)
		t.Chdir(repo)

		tool := tools.NewGitLog()
		got, err := tool.Execute(context.Background(), []byte(`{"n":1}`))
		require.NoError(t, err)
		require.Contains(t, got, "second commit")
		require.NotContains(t, got, "initial commit")
	})

	t.Run("should return error when not in a git repo", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		tool := tools.NewGitLog()
		_, err := tool.Execute(context.Background(), []byte(`{}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "git_log")
	})
}
