package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	kjson "github.com/knadh/koanf/parsers/json"
	kyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	ErrMissingAPIKey  = errors.New("config: api_key is required")
	ErrMissingBaseURL = errors.New("config: base_url is required")
	ErrMissingModel   = errors.New("config: model is required")
	ErrInvalidBaseURL = errors.New("config: base_url is not a valid URL")
)

type Config struct {
	APIKey  string `koanf:"api_key"`
	BaseURL string `koanf:"base_url"`
	Model   string `koanf:"model"`
}

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("%w - set it in ~/.config/go-ask/config.json or via ASK_API_KEY", ErrMissingAPIKey)
	}

	if c.BaseURL == "" {
		return fmt.Errorf("%w - set it in ~/.config/go-ask/config.json or via ASK_BASE_URL", ErrMissingBaseURL)
	}

	if _, err := url.Parse(c.BaseURL); err != nil || !strings.HasPrefix(c.BaseURL, "http") {
		return fmt.Errorf("%w: %q", ErrInvalidBaseURL, c.BaseURL)
	}

	if c.Model == "" {
		return fmt.Errorf("%w - set it in ~/.config/go-ask/config.json or via ASK_MODEL", ErrMissingModel)
	}

	return nil
}

func Load() (*Config, error) {
	k := koanf.New(".")

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("config: resolving home dir: %w", err)
	}

	globalDir := filepath.Join(home, ".config", "go-ask")
	if err = loadFile(k, filepath.Join(globalDir, "config.json")); err != nil {
		return nil, fmt.Errorf("config: resolving user config dir: %w", err)
	}
	if err = loadFile(k, filepath.Join(globalDir, "config.json")); err != nil {
		return nil, err
	}
	if err = loadFile(k, filepath.Join(globalDir, "config.yaml")); err != nil {
		return nil, err
	}

	if err = loadFile(k, "go-ask.json"); err != nil {
		return nil, err
	}
	if err = loadFile(k, "go-ask.yaml"); err != nil {
		return nil, err
	}

	if err = k.Load(env.Provider("ASK_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "ASK_"))
	}), nil); err != nil {
		return nil, fmt.Errorf("config: loading env vars: %w", err)
	}

	cfg := &Config{}
	if err = k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshalling: %w", err)
	}
	return cfg, nil
}

func parserFor(path string) (koanf.Parser, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return kyaml.Parser(), nil
	default:
		return kjson.Parser(), nil
	}
}

func loadFile(k *koanf.Koanf, path string) error {
	parser, err := parserFor(path)
	if err != nil {
		return fmt.Errorf("config: resolving parser for %q: %w", path, err)
	}
	if err = k.Load(file.Provider(path), parser); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("config: loading %q: %w", path, err)
	}

	return nil
}
