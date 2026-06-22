package tty_test

import (
	"os"
	"testing"

	"github.com/devlusoft/go-ask/internal/tty"
	"github.com/stretchr/testify/require"
)

func TestReadStdinIfPiped(t *testing.T) {
	t.Run("should return content from pipe", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		defer func() { _ = r.Close() }()
		defer func() { _ = w.Close() }()
		go func() {
			_, _ = w.WriteString("piped content")
			_ = w.Close()
		}()
		got, err := tty.ReadStdinIfPiped(r)
		require.NoError(t, err)
		require.Equal(t, "piped content", got)
	})

	t.Run("should return empty when pipe is empty", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		defer func() { _ = r.Close() }()
		_ = w.Close()
		got, err := tty.ReadStdinIfPiped(r)
		require.NoError(t, err)
		require.Equal(t, "", got)
	})
}
