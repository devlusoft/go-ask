package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/devlusoft/go-ask/internal/config"
	"github.com/urfave/cli/v3"
)

const version = "0.1.0"

func Execute() {
	cmd := &cli.Command{
		Name:      "ask",
		Usage:     "Agente CLI one-shot que investiga con tools read-only",
		Version:   version,
		ErrWriter: os.Stderr,
		Action: func(ctx context.Context, c *cli.Command) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err = cfg.Validate(); err != nil {
				return fmt.Errorf("config validation failed: %w", err)
			}

			masked := cfg.APIKey
			if len(masked) > 8 {
				masked = masked[:8] + "..."
			}

			fmt.Printf("config OK\n")
			fmt.Printf("  api_key: %s\n", masked)
			fmt.Printf("  base_url: %s\n", cfg.BaseURL)
			fmt.Printf("  model: %s\n", cfg.Model)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
