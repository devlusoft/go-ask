package ui_test

import (
	"errors"
	"testing"
	"time"

	"github.com/devlusoft/go-ask/internal/ui"
)

func TestShowSpinner(t *testing.T) {
	t.Run("should call fn and return its result", func(t *testing.T) {
		got, err := ui.ShowSpinner("loading", func() (string, error) {
			return "hello", nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})

	t.Run("should propagate error from fn", func(t *testing.T) {
		wantErr := errors.New("boom")
		_, err := ui.ShowSpinner("loading", func() (string, error) {
			return "", wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Errorf("got error %v, want %v", err, wantErr)
		}
	})

	t.Run("should return immediately when fn is fast", func(t *testing.T) {
		start := time.Now()
		_, _ = ui.ShowSpinner("loading", func() (string, error) {
			return "ok", nil
		})
		elapsed := time.Since(start)
		if elapsed > 200*time.Millisecond {
			t.Errorf("took %v, expected near-instant return for fast fn", elapsed)
		}
	})
}
