package repl

// ModelPick is one entry in the interactive /model picker.
type ModelPick struct {
	ID      string
	Name    string
	Kind    string
	Current bool
}

// SearchModels returns models matching a fuzzy query for the interactive picker.
// An empty query returns all models (sorted by relevance).
func (r *REPL) SearchModels(query string, limit int) []ModelPick {
	if limit <= 0 {
		limit = 30
	}
	matches := r.matchModels(query, limit)
	out := make([]ModelPick, 0, len(matches))
	for _, m := range matches {
		kind := "model"
		if m.entry.Custom {
			kind = "custom"
		}
		if m.entry.Name != "" && m.entry.Name != m.entry.ID {
			kind += " · " + m.entry.Name
		}
		out = append(out, ModelPick{
			ID:      m.entry.ID,
			Name:    m.entry.Name,
			Kind:    kind,
			Current: m.entry.ID == r.session.Model,
		})
	}
	return out
}

// SetModelID switches the active model without clearing conversation history.
func (r *REPL) SetModelID(id string) {
	r.session.Model = id
	r.persistSession()
}

// ModelsReady reports whether the model list has been fetched.
func (r *REPL) ModelsReady() bool {
	return r.models.ready()
}