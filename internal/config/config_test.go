package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devlusoft/go-ask/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("should load config from global config file", func(t *testing.T) {
		configDir := setupTestHome(t)
		writeConfigFile(t, filepath.Join(configDir, "config.json"), `{
          "api_key": "global-key",
          "base_url": "https://api.example.com",
          "model": "global-model"
      }`)

		cfg, err := config.Load()
		require.NoError(t, err)
		require.Equal(t, "global-key", cfg.APIKey)
		require.Equal(t, "https://api.example.com", cfg.BaseURL)
		require.Equal(t, "global-model", cfg.Model)
	})

	t.Run("should load config from local config file", func(t *testing.T) {
		_ = setupTestHome(t)
		cwd := t.TempDir()
		withTestCWD(t, cwd)
		writeConfigFile(t, filepath.Join(cwd, "go-ask.json"), `{
          "api_key": "local-key",
          "base_url": "https://api.example.com",
          "model": "local-model"
      }`)

		cfg, err := config.Load()
		require.NoError(t, err)
		require.Equal(t, "local-key", cfg.APIKey)
		require.Equal(t, "local-model", cfg.Model)
	})

	t.Run("should let local config override global", func(t *testing.T) {
		configDir := setupTestHome(t)
		writeConfigFile(t, filepath.Join(configDir, "config.json"), `{
          "api_key": "global-key",
          "base_url": "https://global.example.com",
          "model": "global-model"
      }`)
		cwd := t.TempDir()
		withTestCWD(t, cwd)
		writeConfigFile(t, filepath.Join(cwd, "go-ask.yaml"), `
  api_key: local-key
  base_url: https://local.example.com
  model: local-model
  `)

		cfg, err := config.Load()
		require.NoError(t, err)
		require.Equal(t, "local-key", cfg.APIKey)
		require.Equal(t, "https://local.example.com", cfg.BaseURL)
		require.Equal(t, "local-model", cfg.Model)
	})

	t.Run("should let env vars override file config", func(t *testing.T) {
		configDir := setupTestHome(t)
		writeConfigFile(t, filepath.Join(configDir, "config.json"), `{
          "api_key": "file-key",
          "base_url": "https://file.example.com",
          "model": "file-model"
      }`)
		t.Setenv("ASK_API_KEY", "env-key")
		t.Setenv("ASK_BASE_URL", "https://env.example.com")
		t.Setenv("ASK_MODEL", "env-model")

		cfg, err := config.Load()
		require.NoError(t, err)
		require.Equal(t, "env-key", cfg.APIKey)
		require.Equal(t, "https://env.example.com", cfg.BaseURL)
		require.Equal(t, "env-model", cfg.Model)
	})

	t.Run("should return empty config when no files exist", func(t *testing.T) {
		_ = setupTestHome(t)
		cwd := t.TempDir()
		withTestCWD(t, cwd)

		cfg, err := config.Load()
		require.NoError(t, err)
		require.Equal(t, &config.Config{}, cfg)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("should accept valid config", func(t *testing.T) {
		cfg := &config.Config{
			APIKey:  "key",
			BaseURL: "https://api.example.com",
			Model:   "model",
		}
		require.NoError(t, cfg.Validate())
	})

	t.Run("should return ErrMissingAPIKey when api_key is empty", func(t *testing.T) {
		cfg := &config.Config{
			BaseURL: "https://api.example.com",
			Model:   "model",
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrMissingAPIKey)
	})

	t.Run("should return ErrMissingBaseURL when base_url is empty", func(t *testing.T) {
		cfg := &config.Config{
			APIKey: "key",
			Model:  "model",
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrMissingBaseURL)
	})

	t.Run("should return ErrMissingModel when model is empty", func(t *testing.T) {
		cfg := &config.Config{
			APIKey:  "key",
			BaseURL: "https://api.example.com",
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrMissingModel)
	})

	t.Run("should return ErrInvalidBaseURL when base_url is not a URL", func(t *testing.T) {
		cfg := &config.Config{
			APIKey:  "key",
			BaseURL: "not-a-url",
			Model:   "model",
		}
		require.ErrorIs(t, cfg.Validate(), config.ErrInvalidBaseURL)
	})
}

func setupTestHome(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)

	return filepath.Join(configDir, "go-ask")
}

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err)
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}

func withTestCWD(t *testing.T, dir string) {
	t.Helper()

	original, err := os.Getwd()
	require.NoError(t, err, "failed to get current working directory")

	t.Cleanup(func() {
		if err = os.Chdir(original); err != nil {
			t.Errorf("failed to restore CWD to %q: %v", original, err)
		}
	})

	require.NoError(t, os.Chdir(dir), "failed to change CWD to", dir)
}
