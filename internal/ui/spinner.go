package ui

import (
	"fmt"
	"os"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func ShowSpinner(message string, fn func() (string, error)) (string, error) {
	done := make(chan struct{})
	var result string
	var err error

	go func() {
		result, err = fn()
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
			_, _ = fmt.Fprintf(os.Stderr, "\r\033[K%s %s", spinnerFrames[i%len(spinnerFrames)], message)
			i++
		}
	}
}
