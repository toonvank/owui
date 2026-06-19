package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
)

// AttachedFile is an uploaded file linked to a local session for RAG.
type AttachedFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Saved struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Model          string         `json:"model"`
	Messages       []api.Message  `json:"messages"`
	ChatID         string         `json:"chat_id,omitempty"`
	CollectionID   string         `json:"collection_id,omitempty"`
	CollectionName string         `json:"collection_name,omitempty"`
	AttachedFiles     []AttachedFile `json:"attached_files,omitempty"`
	ActiveFilterIDs   []string       `json:"active_filter_ids,omitempty"`
	FiltersCustomized bool           `json:"filters_customized,omitempty"`
	ActiveToolIDs     []string       `json:"active_tool_ids,omitempty"`
	ToolsCustomized   bool           `json:"tools_customized,omitempty"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type Store struct {
	dir string
}

func NewStore(profile string) (*Store, error) {
	dir, err := config.SessionsDir(profile)
	if err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(id string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, id)
	return filepath.Join(s.dir, safe+".json")
}

func (s *Store) Save(sess Saved) error {
	if sess.ID == "" {
		return errors.New("session id required")
	}
	sess.UpdatedAt = time.Now().UTC()
	if sess.Title == "" {
		sess.Title = DeriveTitle(sess.Messages)
	}
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(sess.ID), data, 0o600)
}

func (s *Store) Load(id string) (Saved, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		return Saved{}, err
	}
	var sess Saved
	if err := json.Unmarshal(data, &sess); err != nil {
		return Saved{}, fmt.Errorf("parse session: %w", err)
	}
	return sess, nil
}

func (s *Store) Delete(id string) error {
	return os.Remove(s.path(id))
}

func (s *Store) List() ([]Saved, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []Saved
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var sess Saved
		if json.Unmarshal(data, &sess) != nil {
			continue
		}
		out = append(out, sess)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (s *Store) Latest() (Saved, error) {
	all, err := s.List()
	if err != nil {
		return Saved{}, err
	}
	if len(all) == 0 {
		return Saved{}, os.ErrNotExist
	}
	return all[0], nil
}

// DeriveTitle builds a short title from the latest non-empty user message.
func DeriveTitle(messages []api.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" && strings.TrimSpace(messages[i].Content) != "" {
			t := strings.TrimSpace(messages[i].Content)
			if len(t) > 48 {
				return t[:45] + "..."
			}
			return t
		}
	}
	return "untitled"
}

func NewID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}