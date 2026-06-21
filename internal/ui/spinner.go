package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type statusLine struct {
	mu sync.Mutex
}

func ShowSpinner(initial string, fn func(setStatus func(string)) (string, error)) (string, error) {
	sl := &statusLine{}

	current := initial
	setStatus := func(msg string) {
		sl.mu.Lock()
		defer sl.mu.Unlock()
		current = msg
	}

	done := make(chan struct{})
	var result string
	var err error

	go func() {
		result, err = fn(setStatus)
		close(done)
	}()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-done:
			_, _ = fmt.Fprintf(os.Stderr, "\r\033[K")
			return result, err
		case <-ticker.C:
			sl.mu.Lock()
			display := current
			sl.mu.Unlock()
			_, _ = fmt.Fprintf(os.Stderr, "\r\033[K%s %s", spinnerFrames[i%len(spinnerFrames)], display)
			i++
		}
	}
}
