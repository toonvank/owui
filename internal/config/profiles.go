package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EnvProfile     = "OWUI_PROFILE"
	DefaultProfile = "default"
)

// ProfileSettings holds connection settings for one named profile.
type ProfileSettings struct {
	BaseURL            string   `yaml:"base_url,omitempty"`
	APIKey             string   `yaml:"api_key,omitempty"`
	DefaultModel       string   `yaml:"default_model,omitempty"`
	Stream             *bool    `yaml:"stream,omitempty"`
	SystemPrompt       string   `yaml:"system_prompt,omitempty"`
	TimeoutSec         int      `yaml:"timeout_sec,omitempty"`
	InsecureTLS        bool     `yaml:"insecure_tls,omitempty"`
	CustomHeader       string   `yaml:"custom_api_key_header,omitempty"`
	ApplyModelFeatures *bool    `yaml:"apply_model_features,omitempty"`
	FilterIDs          []string `yaml:"filter_ids,omitempty"`
	ToolIDs            []string `yaml:"tool_ids,omitempty"`
	Theme              string   `yaml:"theme,omitempty"`
	VimKeys            bool     `yaml:"vim_keys,omitempty"`
}

// File is the on-disk config.yaml structure.
type File struct {
	ActiveProfile string                     `yaml:"profile,omitempty"`
	Profiles      map[string]ProfileSettings `yaml:"profiles,omitempty"`
	ProfileSettings `yaml:",inline"`
}

func (ps ProfileSettings) toConfig(name string) Config {
	cfg := Default()
	cfg.ProfileName = name
	if ps.BaseURL != "" {
		cfg.BaseURL = ps.BaseURL
	}
	if ps.APIKey != "" {
		cfg.APIKey = ps.APIKey
	}
	if ps.DefaultModel != "" {
		cfg.DefaultModel = ps.DefaultModel
	}
	if ps.Stream != nil {
		cfg.Stream = *ps.Stream
	}
	if ps.SystemPrompt != "" {
		cfg.SystemPrompt = ps.SystemPrompt
	}
	if ps.TimeoutSec > 0 {
		cfg.TimeoutSec = ps.TimeoutSec
	}
	cfg.InsecureTLS = ps.InsecureTLS
	if ps.CustomHeader != "" {
		cfg.CustomHeader = ps.CustomHeader
	}
	if ps.ApplyModelFeatures != nil {
		cfg.ApplyModelFeatures = ps.ApplyModelFeatures
	}
	if len(ps.FilterIDs) > 0 {
		cfg.FilterIDs = append([]string(nil), ps.FilterIDs...)
	}
	if len(ps.ToolIDs) > 0 {
		cfg.ToolIDs = append([]string(nil), ps.ToolIDs...)
	}
	if ps.Theme != "" {
		cfg.Theme = ps.Theme
	}
	cfg.VimKeys = ps.VimKeys
	return cfg
}

// ToProfileSettings converts a resolved config to storable profile settings.
func (c Config) ToProfileSettings() ProfileSettings {
	stream := c.Stream
	return ProfileSettings{
		BaseURL:            c.BaseURL,
		APIKey:             c.APIKey,
		DefaultModel:       c.DefaultModel,
		Stream:             &stream,
		SystemPrompt:       c.SystemPrompt,
		TimeoutSec:         c.TimeoutSec,
		InsecureTLS:        c.InsecureTLS,
		CustomHeader:       c.CustomHeader,
		ApplyModelFeatures: c.ApplyModelFeatures,
		FilterIDs:          append([]string(nil), c.FilterIDs...),
		ToolIDs:            append([]string(nil), c.ToolIDs...),
		Theme:              c.Theme,
		VimKeys:            c.VimKeys,
	}
}

// ReadFile loads the raw config file from disk.
func ReadFile() (File, error) {
	var file File
	path, err := Path()
	if err != nil {
		return file, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return file, nil
		}
		return file, err
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return file, fmt.Errorf("parse config: %w", err)
	}
	return file, nil
}

// WriteFile persists the config file.
func WriteFile(file File) error {
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
	data, err := yaml.Marshal(file)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Normalize migrates legacy flat config into the profiles map.
func (f *File) Normalize() {
	f.normalize()
}

func (f *File) normalize() {
	if f.Profiles == nil {
		f.Profiles = make(map[string]ProfileSettings)
	}
	if len(f.Profiles) == 0 && f.BaseURL != "" {
		f.Profiles[DefaultProfile] = f.ProfileSettings
		f.ProfileSettings = ProfileSettings{}
		f.ActiveProfile = DefaultProfile
	}
	if f.ActiveProfile == "" {
		f.ActiveProfile = DefaultProfile
	}
}

// Resolve returns the effective Config for a profile name.
func (f *File) Resolve(name string) (Config, error) {
	f.normalize()
	if name == "" {
		name = f.ActiveProfile
	}
	if name == "" {
		name = DefaultProfile
	}
	if ps, ok := f.Profiles[name]; ok {
		cfg := ps.toConfig(name)
		cfg = applyEnv(cfg)
		if cfg.TimeoutSec <= 0 {
			cfg.TimeoutSec = 300
		}
		return cfg, nil
	}
	if len(f.Profiles) == 0 {
		cfg := f.ProfileSettings.toConfig(name)
		cfg = applyEnv(cfg)
		if cfg.TimeoutSec <= 0 {
			cfg.TimeoutSec = 300
		}
		return cfg, nil
	}
	return Config{}, fmt.Errorf("profile %q not found", name)
}

// ProfileNames returns sorted profile names.
func (f *File) ProfileNames() []string {
	f.normalize()
	names := make([]string, 0, len(f.Profiles))
	for name := range f.Profiles {
		names = append(names, name)
	}
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
	return names
}

// LoadWithProfile loads config for the given profile (empty = env + file active).
func LoadWithProfile(profile string) (Config, error) {
	file, err := ReadFile()
	if err != nil {
		return Config{}, err
	}
	if profile == "" {
		profile = os.Getenv(EnvProfile)
	}
	cfg, err := file.Resolve(profile)
	if err != nil {
		return cfg, err
	}
	if cfg.ProfileName == "" {
		cfg.ProfileName = DefaultProfile
	}
	return cfg, nil
}

// SwitchActiveProfile sets the active profile and returns its config.
func SwitchActiveProfile(name string) (Config, error) {
	file, err := ReadFile()
	if err != nil {
		return Config{}, err
	}
	file.normalize()
	if _, ok := file.Profiles[name]; !ok {
		return Config{}, fmt.Errorf("profile %q not found", name)
	}
	file.ActiveProfile = name
	if err := WriteFile(file); err != nil {
		return Config{}, err
	}
	return file.Resolve(name)
}

// AddProfile creates or updates a named profile.
func AddProfile(name string, settings ProfileSettings) error {
	if err := validateProfileName(name); err != nil {
		return err
	}
	file, err := ReadFile()
	if err != nil {
		return err
	}
	file.normalize()
	file.Profiles[name] = settings
	if file.ActiveProfile == "" {
		file.ActiveProfile = name
	}
	return WriteFile(file)
}

// RemoveProfile deletes a profile (not the active one).
func RemoveProfile(name string) error {
	file, err := ReadFile()
	if err != nil {
		return err
	}
	file.normalize()
	if name == file.ActiveProfile {
		return fmt.Errorf("cannot remove active profile %q — switch first", name)
	}
	if _, ok := file.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(file.Profiles, name)
	return WriteFile(file)
}

// SanitizeProfileName makes a profile name safe for directory names.
func SanitizeProfileName(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
}

func validateProfileName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("profile name is required")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("invalid profile name: %q", name)
	}
	return nil
}

func legacySessionsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

func dirHasSessionJSON(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			return true
		}
	}
	return false
}