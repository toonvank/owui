package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	EnvBaseURL = "OWUI_BASE_URL"
	EnvAPIKey  = "OWUI_API_KEY"
	EnvModel   = "OWUI_MODEL"
)

type Config struct {
	BaseURL             string   `yaml:"base_url"`
	APIKey              string   `yaml:"api_key"`
	DefaultModel        string   `yaml:"default_model"`
	Stream              bool     `yaml:"stream"`
	SystemPrompt        string   `yaml:"system_prompt"`
	TimeoutSec          int      `yaml:"timeout_sec"`
	InsecureTLS         bool     `yaml:"insecure_tls"`
	CustomHeader        string   `yaml:"custom_api_key_header"`
	ApplyModelFeatures  *bool    `yaml:"apply_model_features"`
	FilterIDs           []string `yaml:"filter_ids"`
	ToolIDs             []string `yaml:"tool_ids"`
}

func (c Config) ShouldApplyModelFeatures() bool {
	if c.ApplyModelFeatures == nil {
		return true
	}
	return *c.ApplyModelFeatures
}

func Default() Config {
	return Config{
		Stream:     true,
		TimeoutSec: 300,
	}
}

// FileExists reports whether a config file has been written to disk.
func FileExists() bool {
	path, err := Path()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "owui"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func SessionsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = applyEnv(cfg)
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	cfg = applyEnv(cfg)
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 300
	}
	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path, err := Path()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func applyEnv(cfg Config) Config {
	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv(EnvModel); v != "" {
		cfg.DefaultModel = v
	}
	return cfg
}

func (c Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("base_url is required (set via config or OWUI_BASE_URL)")
	}
	if c.APIKey == "" {
		return errors.New("api_key is required (run `owui auth login` or set OWUI_API_KEY)")
	}
	return nil
}