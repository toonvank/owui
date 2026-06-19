package repl

import (
	"fmt"
	"sort"
	"strings"

	"github.com/toonvank/owui/internal/api"
)

func (r *REPL) effectiveFilterIDs() []string {
	if r.session.FiltersCustomized {
		return append([]string(nil), r.session.ActiveFilterIDs...)
	}
	model, _ := r.client.ModelByID(r.session.Model)
	extra := append([]string(nil), r.cfg.FilterIDs...)
	return api.FilterIDsFromModel(model, extra)
}

func (r *REPL) effectiveToolIDs() []string {
	if r.session.ToolsCustomized {
		return append([]string(nil), r.session.ActiveToolIDs...)
	}
	return append([]string(nil), r.cfg.ToolIDs...)
}

func (r *REPL) isFilterEnabled(id string) bool {
	for _, x := range r.effectiveFilterIDs() {
		if x == id {
			return true
		}
	}
	return false
}

func (r *REPL) isToolEnabled(id string) bool {
	for _, x := range r.effectiveToolIDs() {
		if x == id {
			return true
		}
	}
	return false
}

func (r *REPL) ToggleFilter(id string) bool {
	current := r.effectiveFilterIDs()
	next, enabled := toggleID(current, id)
	r.session.ActiveFilterIDs = next
	r.session.FiltersCustomized = true
	r.persistSession()
	return enabled
}

func (r *REPL) ToggleTool(id string) bool {
	current := r.effectiveToolIDs()
	next, enabled := toggleID(current, id)
	r.session.ActiveToolIDs = next
	r.session.ToolsCustomized = true
	r.persistSession()
	return enabled
}

func toggleID(ids []string, id string) ([]string, bool) {
	found := false
	out := make([]string, 0, len(ids))
	for _, x := range ids {
		if x == id {
			found = true
			continue
		}
		out = append(out, x)
	}
	if !found {
		out = append(out, id)
		return out, true
	}
	return out, false
}

func (r *REPL) ResetFilters() {
	r.session.ActiveFilterIDs = nil
	r.session.FiltersCustomized = false
	r.persistSession()
}

func (r *REPL) ResetTools() {
	r.session.ActiveToolIDs = nil
	r.session.ToolsCustomized = false
	r.persistSession()
}

func (r *REPL) IntegrationsLabel() string {
	var parts []string
	if r.session.FiltersCustomized {
		parts = append(parts, fmt.Sprintf("%d filter(s)", len(r.session.ActiveFilterIDs)))
	} else if r.client != nil {
		if n := len(r.effectiveFilterIDs()); n > 0 {
			parts = append(parts, fmt.Sprintf("%d filter(s)", n))
		}
	}
	if r.session.ToolsCustomized {
		parts = append(parts, fmt.Sprintf("%d tool(s)", len(r.session.ActiveToolIDs)))
	} else if r.client != nil {
		if n := len(r.effectiveToolIDs()); n > 0 {
			parts = append(parts, fmt.Sprintf("%d tool(s)", n))
		}
	}
	return strings.Join(parts, " · ")
}

func (r *REPL) slashFilters(args []string) SlashResult {
	if len(args) == 0 {
		return SlashResult{Output: "type /filters in the input to toggle filters"}
	}
	switch args[0] {
	case "list":
		return SlashResult{Output: r.formatFiltersList()}
	case "clear", "reset":
		r.ResetFilters()
		return SlashResult{Output: "filters reset to model defaults"}
	case "toggle":
		if len(args) < 2 {
			return SlashResult{Err: fmt.Errorf("usage: /filters toggle <id>")}
		}
		on := r.ToggleFilter(args[1])
		state := "off"
		if on {
			state = "on"
		}
		return SlashResult{Output: fmt.Sprintf("filter %s %s", shortID(args[1]), state)}
	case "info":
		return SlashResult{Output: r.formatModelIntegrations()}
	default:
		return SlashResult{Output: r.formatFiltersList()}
	}
}

func (r *REPL) slashTools(args []string) SlashResult {
	if len(args) == 0 {
		return SlashResult{Output: "type /tools in the input to toggle tools"}
	}
	switch args[0] {
	case "list":
		return SlashResult{Output: r.formatToolsList()}
	case "clear", "reset":
		r.ResetTools()
		return SlashResult{Output: "tools reset to config defaults"}
	case "toggle":
		if len(args) < 2 {
			return SlashResult{Err: fmt.Errorf("usage: /tools toggle <id>")}
		}
		on := r.ToggleTool(args[1])
		state := "off"
		if on {
			state = "on"
		}
		return SlashResult{Output: fmt.Sprintf("tool %s %s", shortID(args[1]), state)}
	default:
		return SlashResult{Output: r.formatToolsList()}
	}
}

func (r *REPL) formatFiltersList() string {
	return r.formatFunctionList("filter", r.effectiveFilterIDs(), r.session.FiltersCustomized)
}

func (r *REPL) formatToolsList() string {
	return r.formatFunctionList("tool", r.effectiveToolIDs(), r.session.ToolsCustomized)
}

func (r *REPL) formatFunctionList(fnType string, active []string, customized bool) string {
	r.ensureFunctions()
	entries, err := r.functions.list()
	if err != nil {
		return "error: " + err.Error()
	}
	activeSet := make(map[string]bool, len(active))
	for _, id := range active {
		activeSet[id] = true
	}

	mode := "model defaults"
	if customized {
		mode = "custom"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%ss (%s):\n", fnType, mode)
	shown := 0
	for _, fn := range entries {
		if fn.Type != fnType {
			continue
		}
		state := "off"
		if activeSet[fn.ID] {
			state = "on"
		}
		scope := "model"
		if fn.IsGlobal {
			scope = "global"
		}
		fmt.Fprintf(&b, "  [%s/%s] %s (%s)\n", state, scope, fn.Name, shortID(fn.ID))
		shown++
		if shown >= 20 {
			fmt.Fprintf(&b, "showing %d — type /%ss to toggle interactively", shown, fnType)
			break
		}
	}
	if shown == 0 {
		return fmt.Sprintf("no %ss found", fnType)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (r *REPL) formatModelIntegrations() string {
	model, err := r.client.ModelByID(r.session.Model)
	if err != nil {
		return "error: " + err.Error()
	}
	if model.ID == "" {
		return fmt.Sprintf("model %s not found", r.session.Model)
	}
	meta := model.Meta()
	features := api.FeaturesFromModel(model)
	var b strings.Builder
	fmt.Fprintf(&b, "Model: %s\n", model.ID)
	fmt.Fprintf(&b, "Provider: %s\n", model.ModelKindLabel())
	if tags := model.CapabilityTags(); len(tags) > 0 {
		fmt.Fprintf(&b, "Capabilities: %s\n", strings.Join(tags, ", "))
	}
	fmt.Fprintf(&b, "Default filters: %v\n", meta.DefaultFilterIDs)
	fmt.Fprintf(&b, "Active filters: %v\n", r.effectiveFilterIDs())
	fmt.Fprintf(&b, "Active tools: %v\n", r.effectiveToolIDs())
	if len(features) > 0 {
		keys := make([]string, 0, len(features))
		for k := range features {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Fprintf(&b, "Auto features: %s\n", strings.Join(keys, ", "))
	} else {
		b.WriteString("Auto features: (none)\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

