package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/devlusoft/go-ask/internal/config"
	"github.com/devlusoft/go-ask/internal/provider"
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
			args := c.Args()
			if args.Len() == 0 {
				return cli.ShowAppHelp(c)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err = cfg.Validate(); err != nil {
				return fmt.Errorf("config validation failed: %w", err)
			}

			p := provider.New(cfg.BaseURL, cfg.APIKey, cfg.Model)
			resp, err := p.Chat(ctx, provider.Message{
				Role:    provider.RoleUser,
				Content: args.First(),
			})
			if err != nil {
				var apiErr *provider.Error
				if errors.As(err, &apiErr) {
					return fmt.Errorf("provider error (status %d): %s\nbody: %s", apiErr.StatusCode, apiErr.Message, apiErr.Body)
				}
				return fmt.Errorf("calling LLM: %w", err)
			}

			fmt.Println(resp)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
