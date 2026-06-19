package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	EnvBaseURL = "OWUI_BASE_URL"
	EnvAPIKey  = "OWUI_API_KEY"
	EnvModel   = "OWUI_MODEL"
)

type Config struct {
	ProfileName        string   `yaml:"-"`
	BaseURL            string   `yaml:"base_url"`
	APIKey             string   `yaml:"api_key"`
	DefaultModel       string   `yaml:"default_model"`
	Stream             bool     `yaml:"stream"`
	SystemPrompt       string   `yaml:"system_prompt"`
	TimeoutSec         int      `yaml:"timeout_sec"`
	InsecureTLS        bool     `yaml:"insecure_tls"`
	CustomHeader       string   `yaml:"custom_api_key_header"`
	ApplyModelFeatures *bool    `yaml:"apply_model_features"`
	FilterIDs          []string `yaml:"filter_ids"`
	ToolIDs            []string `yaml:"tool_ids"`
	Theme              string   `yaml:"theme,omitempty"`
	VimKeys            bool     `yaml:"vim_keys,omitempty"`
}

func (c Config) ShouldApplyModelFeatures() bool {
	if c.ApplyModelFeatures == nil {
		return true
	}
	return *c.ApplyModelFeatures
}

func Default() Config {
	return Config{
		ProfileName: DefaultProfile,
		Stream:      true,
		TimeoutSec:  300,
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

// SessionsDir returns the session storage directory for a profile.
func SessionsDir(profile string) (string, error) {
	if profile == "" {
		profile = DefaultProfile
	}
	base, err := Dir()
	if err != nil {
		return "", err
	}
	safe := SanitizeProfileName(profile)
	profDir := filepath.Join(base, "sessions", safe)

	// Backward compat: flat sessions/ dir used before profiles.
	if profile == DefaultProfile {
		legacy, err := legacySessionsDir()
		if err == nil && dirHasSessionJSON(legacy) && !dirHasSessionJSON(profDir) {
			return legacy, nil
		}
	}
	if err := os.MkdirAll(profDir, 0o700); err != nil {
		return "", err
	}
	return profDir, nil
}

func Load() (Config, error) {
	return LoadWithProfile("")
}

func Save(cfg Config) error {
	file, err := ReadFile()
	if err != nil {
		return err
	}
	file.normalize()

	name := cfg.ProfileName
	if name == "" {
		name = file.ActiveProfile
	}
	if name == "" {
		name = DefaultProfile
	}
	cfg.ProfileName = name
	file.ActiveProfile = name
	file.Profiles[name] = cfg.ToProfileSettings()
	file.ProfileSettings = ProfileSettings{}
	return WriteFile(file)
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

// MaskedAPIKey returns a redacted API key for display.
func (c Config) MaskedAPIKey() string {
	if c.APIKey == "" {
		return ""
	}
	if len(c.APIKey) <= 12 {
		return "..."
	}
	return c.APIKey[:8] + "..." + c.APIKey[len(c.APIKey)-4:]
}

// ProfileSummary returns a one-line description of a profile.
func ProfileSummary(name string, ps ProfileSettings) string {
	url := ps.BaseURL
	if url == "" {
		url = "(not configured)"
	}
	model := ps.DefaultModel
	if model == "" {
		model = "-"
	}
	return fmt.Sprintf("%s  %s  model:%s", name, url, model)
}