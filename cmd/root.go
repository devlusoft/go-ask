package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/devlusoft/go-ask/internal/agent"
	"github.com/devlusoft/go-ask/internal/config"
	"github.com/devlusoft/go-ask/internal/provider"
	"github.com/devlusoft/go-ask/internal/render"
	"github.com/devlusoft/go-ask/internal/tools"
	"github.com/devlusoft/go-ask/internal/ui"
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

			r := tools.NewRegistry()
			mustRegister(r, tools.NewReadFile())
			mustRegister(r, tools.NewListFiles())
			mustRegister(r, tools.NewGrep())
			mustRegister(r, tools.NewGlob())
			mustRegister(r, tools.NewGitLog())
			mustRegister(r, tools.NewGitDiff())
			mustRegister(r, tools.NewWebSearch())
			mustRegister(r, tools.NewFetchURL())

			a := agent.New(p, r)

			resp, err := ui.ShowSpinner("Pensando...", func(setStatus func(string)) (string, error) {
				a.OnToolCall = func(name, args string) {
					setStatus(name + "(" + formatArgs(args) + ")")
				}
				return a.Run(ctx, args.First())
			})
			if err != nil {
				var apiErr *provider.Error
				if errors.As(err, &apiErr) {
					return fmt.Errorf("provider error (status %d): %s\nbody: %s", apiErr.StatusCode, apiErr.Message, apiErr.Body)
				}
				return fmt.Errorf("running agent: %w", err)
			}

			out := render.StripThinking(resp)
			if rendered, rErr := render.Markdown(out); rErr == nil {
				out = rendered
			}

			fmt.Println(out)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}

func mustRegister(r *tools.Registry, t tools.Tool) {
	if err := r.Register(t); err != nil {
		panic(fmt.Sprintf("cmd: registering tool: %v", err))
	}
}

func formatArgs(argsJSON string) string {
	if argsJSON == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &m); err != nil {
		return argsJSON
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, formatKV(k, v))
	}
	return strings.Join(parts, ", ")
}

func formatKV(k string, v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%s=%q", k, val)
	case bool:
		return fmt.Sprintf("%s=%t", k, val)
	default:
		return fmt.Sprintf("%s=%v", k, val)
	}
}
