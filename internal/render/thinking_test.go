package render_test

import (
	"testing"

	"github.com/devlusoft/go-ask/internal/render"
	"github.com/stretchr/testify/require"
)

func TestStripThinking(t *testing.T) {
	t.Run("should strip a single thinking block", func(t *testing.T) {
		input := "<think>some reasoning</think>\nactual answer"
		require.Equal(t, "actual answer", render.StripThinking(input))
	})

	t.Run("should strip multiple thinking blocks", func(t *testing.T) {
		input := "<think>first</think>between<think>second</think>end"
		require.Equal(t, "betweenend", render.StripThinking(input))
	})

	t.Run("should return input unchanged when no thinking block is present", func(t *testing.T) {
		input := "just an answer with no thinking"
		require.Equal(t, "just an answer with no thinking", render.StripThinking(input))
	})

	t.Run("should handle multiline thinking content", func(t *testing.T) {
		input := "<think>\nline 1\nline 2\nline 3\n</think>\nthe answer"
		require.Equal(t, "the answer", render.StripThinking(input))
	})
}
