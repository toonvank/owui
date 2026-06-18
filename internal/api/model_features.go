package api

import (
	"encoding/json"
	"net/http"
)

type ModelMeta struct {
	FilterIDs        []string          `json:"filterIds"`
	DefaultFilterIDs []string          `json:"defaultFilterIds"`
	Capabilities     map[string]bool   `json:"capabilities"`
}

type ModelInfo struct {
	Meta ModelMeta `json:"meta"`
}

func (m Model) Meta() ModelMeta {
	if m.Info == nil {
		return ModelMeta{}
	}
	switch info := m.Info.(type) {
	case map[string]any:
		b, _ := json.Marshal(info)
		var parsed ModelInfo
		if json.Unmarshal(b, &parsed) == nil {
			return parsed.Meta
		}
	}
	return ModelMeta{}
}

// FeaturesFromModel maps Open WebUI model capabilities to the features block
// the server expects in /api/chat/completions (same as the web UI toggles).
func FeaturesFromModel(m Model) map[string]any {
	meta := m.Meta()
	if len(meta.Capabilities) == 0 {
		return nil
	}

	features := make(map[string]any)
	capMap := map[string]string{
		"web_search":        "web_search",
		"image_generation":  "image_generation",
		"code_interpreter":  "code_interpreter",
		"memory":            "memory",
		"file_context":      "file_context",
		"vision":            "vision",
		"file_upload":       "file_upload",
		"builtin_tools":     "builtin_tools",
		"citations":         "citations",
		"status_updates":    "status_updates",
		"terminal":          "terminal",
	}

	for cap, key := range capMap {
		if meta.Capabilities[cap] {
			features[key] = true
		}
	}
	if len(features) == 0 {
		return nil
	}
	return features
}

func FilterIDsFromModel(m Model, extra []string) []string {
	meta := m.Meta()
	ids := make([]string, 0, len(meta.DefaultFilterIDs)+len(extra))
	seen := make(map[string]bool)

	for _, id := range meta.DefaultFilterIDs {
		if id != "" && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	for _, id := range extra {
		if id != "" && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids
}

func (c *Client) ModelByID(id string) (Model, error) {
	models, err := c.ListModels()
	if err != nil {
		return Model{}, err
	}
	for _, m := range models {
		if m.ID == id {
			return m, nil
		}
	}
	return Model{}, nil
}

type OWUIFunction struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsActive bool   `json:"is_active"`
	IsGlobal bool   `json:"is_global"`
}

func (c *Client) ListFunctions() ([]OWUIFunction, error) {
	var fns []OWUIFunction
	err := c.request(http.MethodGet, "/api/v1/functions/", nil, &fns)
	return fns, err
}