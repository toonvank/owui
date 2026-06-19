package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/toonvank/owui/internal/config"
)

func TestToggleChatPinned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/chats/chat-1/pin" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "chat-1", "pinned": true})
	}))
	defer srv.Close()

	client := New(config.Config{BaseURL: srv.URL})
	pinned, err := client.ToggleChatPinned("chat-1")
	if err != nil {
		t.Fatal(err)
	}
	if !pinned {
		t.Fatal("expected pinned")
	}
}

func TestUpdateChatTitle(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/chats/chat-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "chat-1",
				"title": "old",
				"chat":  map[string]any{"title": "old", "models": []any{"m1"}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/chats/chat-1":
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "chat-1", "title": "hello world"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := New(config.Config{BaseURL: srv.URL})
	if err := client.UpdateChatTitle("chat-1", "hello world"); err != nil {
		t.Fatal(err)
	}
	chat, ok := gotBody["chat"].(map[string]any)
	if !ok || chat["title"] != "hello world" {
		t.Fatalf("unexpected body: %+v", gotBody)
	}
}

func TestListPinnedChats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/chats/pinned" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "p1", "title": "Pinned chat", "updated_at": 100, "created_at": 50},
		})
	}))
	defer srv.Close()

	client := New(config.Config{BaseURL: srv.URL})
	chats, err := client.ListPinnedChats()
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 1 || !chats[0].Pinned || chats[0].ID != "p1" {
		t.Fatalf("unexpected chats: %+v", chats)
	}
}