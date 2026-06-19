package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toonvank/owui/internal/api"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &Store{dir: dir} // profile-scoped dir

	orig := Saved{
		ID:    "test-session-1",
		Title: "hello",
		Model: "smart:small",
		Messages: []api.Message{
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hey"},
		},
	}
	if err := s.Save(orig); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load("test-session-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != orig.Title || len(got.Messages) != 2 {
		t.Fatalf("unexpected loaded session: %+v", got)
	}

	path := filepath.Join(dir, "test-session-1.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 perms, got %o", info.Mode().Perm())
	}
}