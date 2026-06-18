package api

import "testing"

func TestParseLoadedChat(t *testing.T) {
	raw := map[string]any{
		"id":    "abc-123",
		"title": "Test Chat",
		"chat": map[string]any{
			"models": []any{"mark-with-a-k"},
			"history": map[string]any{
				"currentId": "b",
				"messages": map[string]any{
					"a": map[string]any{"role": "user", "content": "hey", "parentId": nil},
					"b": map[string]any{"role": "assistant", "content": "hello", "parentId": "a"},
				},
			},
		},
	}

	loaded, err := ParseLoadedChat(raw)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model != "mark-with-a-k" {
		t.Fatalf("model %s", loaded.Model)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("messages %d", len(loaded.Messages))
	}
	if loaded.Messages[0].Content != "hey" || loaded.Messages[1].Content != "hello" {
		t.Fatalf("content %+v", loaded.Messages)
	}
}