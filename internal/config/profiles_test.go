package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	cfg := Default()
	cfg.BaseURL = "http://localhost:3000"
	cfg.APIKey = "sk-test-key"
	cfg.DefaultModel = "llama3.2"
	cfg.ProfileName = DefaultProfile

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.BaseURL != cfg.BaseURL || loaded.APIKey != cfg.APIKey {
		t.Fatalf("load mismatch: %+v", loaded)
	}
	if loaded.ProfileName != DefaultProfile {
		t.Fatalf("profile name = %q", loaded.ProfileName)
	}
}

func TestMultipleProfiles(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	home, _ := DefaultProfileSettings(t)
	_ = home
	if err := AddProfile("work", ProfileSettings{
		BaseURL:      "https://work.example.com",
		APIKey:       "sk-work",
		DefaultModel: "gpt-4",
	}); err != nil {
		t.Fatal(err)
	}

	work, err := LoadWithProfile("work")
	if err != nil {
		t.Fatal(err)
	}
	if work.BaseURL != "https://work.example.com" {
		t.Fatalf("work url = %q", work.BaseURL)
	}

	switched, err := SwitchActiveProfile("work")
	if err != nil {
		t.Fatal(err)
	}
	if switched.ProfileName != "work" {
		t.Fatalf("switched profile = %q", switched.ProfileName)
	}
}

func TestSessionsDirPerProfile(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	d1, err := SessionsDir("home")
	if err != nil {
		t.Fatal(err)
	}
	d2, err := SessionsDir("work")
	if err != nil {
		t.Fatal(err)
	}
	if d1 == d2 {
		t.Fatalf("expected different session dirs: %s", d1)
	}
	if filepath.Base(d1) != "home" {
		t.Fatalf("got %s", d1)
	}
}

func DefaultProfileSettings(t *testing.T) (ProfileSettings, error) {
	t.Helper()
	cfg := Default()
	cfg.BaseURL = "http://localhost:3000"
	cfg.APIKey = "sk-home"
	cfg.ProfileName = DefaultProfile
	if err := Save(cfg); err != nil {
		return ProfileSettings{}, err
	}
	return cfg.ToProfileSettings(), nil
}