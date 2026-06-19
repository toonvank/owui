package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var reasoningBlock = regexp.MustCompile(`(?s)<details type="reasoning"[^>]*>.*?</details>\s*`)

type LoadedChat struct {
	ID       string
	Title    string
	Model    string
	Messages []Message
}

func ParseLoadedChat(raw map[string]any) (LoadedChat, error) {
	out := LoadedChat{}

	id, _ := raw["id"].(string)
	out.ID = id
	out.Title, _ = raw["title"].(string)

	chatObj, ok := raw["chat"].(map[string]any)
	if !ok {
		return out, fmt.Errorf("chat payload missing")
	}

	if models, ok := chatObj["models"].([]any); ok && len(models) > 0 {
		if m, ok := models[0].(string); ok {
			out.Model = m
		}
	}

	history, _ := chatObj["history"].(map[string]any)
	if history == nil {
		return out, nil
	}

	msgMap, _ := history["messages"].(map[string]any)
	if msgMap == nil {
		return out, nil
	}

	currentID, _ := history["currentId"].(string)
	out.Messages = walkHistory(msgMap, currentID)
	return out, nil
}

func walkHistory(msgMap map[string]any, currentID string) []Message {
	if currentID == "" || msgMap == nil {
		return nil
	}

	type node struct {
		id   string
		data map[string]any
	}
	chain := make([]node, 0)
	seen := make(map[string]bool)
	cid := currentID

	for cid != "" && !seen[cid] {
		raw, ok := msgMap[cid].(map[string]any)
		if !ok {
			break
		}
		seen[cid] = true
		chain = append(chain, node{id: cid, data: raw})
		parent, _ := raw["parentId"].(string)
		cid = parent
	}

	// reverse to chronological order
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	out := make([]Message, 0, len(chain))
	for _, n := range chain {
		role, _ := n.data["role"].(string)
		if role == "" {
			continue
		}
		content := extractMessageContent(n.data["content"])
		if content == "" && role != "system" {
			continue
		}
		out = append(out, Message{Role: role, Content: content})
	}
	return out
}

func extractMessageContent(v any) string {
	switch c := v.(type) {
	case string:
		return cleanContent(c)
	case []any:
		var parts []string
		for _, item := range c {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				} else if t, ok := m["content"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return cleanContent(strings.Join(parts, "\n"))
	default:
		return ""
	}
}

func cleanContent(s string) string {
	s = reasoningBlock.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func (c *Client) LoadChat(id string) (LoadedChat, error) {
	raw, err := c.GetChat(id)
	if err != nil {
		// try prefix resolve via list
		if len(id) < 36 {
			resolved, rerr := c.resolveChatID(id)
			if rerr == nil {
				raw, err = c.GetChat(resolved)
				id = resolved
			}
		}
		if err != nil {
			return LoadedChat{}, err
		}
	}
	parsed, err := ParseLoadedChat(raw)
	if err != nil {
		return LoadedChat{}, err
	}
	if parsed.ID == "" {
		parsed.ID = id
	}
	return parsed, nil
}

type chatFormData struct {
	Chat map[string]any `json:"chat"`
}

type chatMutationResponse struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Pinned *bool  `json:"pinned"`
}

func (c *Client) ListPinnedChats() ([]ChatSummary, error) {
	var chats []ChatSummary
	err := c.request(http.MethodGet, "/api/v1/chats/pinned", nil, &chats)
	if err != nil {
		return nil, err
	}
	for i := range chats {
		chats[i].Pinned = true
	}
	return chats, nil
}

func (c *Client) UpdateChatTitle(id, title string) error {
	raw, err := c.GetChat(id)
	if err != nil {
		return c.postChatUpdate(id, map[string]any{"title": title})
	}
	chatObj, _ := raw["chat"].(map[string]any)
	if chatObj == nil {
		chatObj = map[string]any{}
	}
	chatObj["title"] = title
	return c.postChatUpdate(id, chatObj)
}

func (c *Client) postChatUpdate(id string, chat map[string]any) error {
	var resp chatMutationResponse
	return c.request(http.MethodPost, "/api/v1/chats/"+id, chatFormData{Chat: chat}, &resp)
}

func (c *Client) ToggleChatPinned(id string) (bool, error) {
	var resp chatMutationResponse
	if err := c.request(http.MethodPost, "/api/v1/chats/"+id+"/pin", nil, &resp); err != nil {
		return false, err
	}
	if resp.Pinned != nil {
		return *resp.Pinned, nil
	}
	return true, nil
}

func (c *Client) resolveChatID(prefix string) (string, error) {
	chats, err := c.ListChats(1)
	if err != nil {
		return "", err
	}
	prefix = strings.ToLower(prefix)
	var matches []string
	for _, ch := range chats {
		if strings.HasPrefix(strings.ToLower(ch.ID), prefix) {
			matches = append(matches, ch.ID)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no chat matching %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%d chats match %q — be more specific", len(matches), prefix)
	}
}