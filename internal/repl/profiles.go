package repl

import (
	"fmt"
	"strings"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
)

// ProfilePick is one entry in the interactive /profile picker.
type ProfilePick struct {
	Name    string
	URL     string
	Model   string
	Current bool
}

// SearchProfiles returns profiles matching a fuzzy query for the active REPL.
func (r *REPL) SearchProfiles(query string) ([]ProfilePick, error) {
	return searchProfiles(query, r.ProfileName())
}

func searchProfiles(query string, active string) ([]ProfilePick, error) {
	file, err := config.ReadFile()
	if err != nil {
		return nil, err
	}
	file.Normalize()
	names := file.ProfileNames()
	if len(names) == 0 {
		return nil, nil
	}

	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]ProfilePick, 0, len(names))
	for _, name := range names {
		if q != "" && !strings.Contains(strings.ToLower(name), q) {
			ps := file.Profiles[name]
			if !strings.Contains(strings.ToLower(ps.BaseURL), q) {
				continue
			}
		}
		ps := file.Profiles[name]
		out = append(out, ProfilePick{
			Name:    name,
			URL:     ps.BaseURL,
			Model:   ps.DefaultModel,
			Current: name == active,
		})
	}
	return out, nil
}

// SwitchProfile changes the active profile and reloads the REPL state.
func (r *REPL) SwitchProfile(name string) SlashResult {
	newCfg, err := config.SwitchActiveProfile(name)
	if err != nil {
		return SlashResult{Err: err}
	}
	if err := newCfg.Validate(); err != nil {
		return SlashResult{Err: fmt.Errorf("profile %q: %w", name, err)}
	}

	r.cfg = newCfg
	r.client = api.New(newCfg)
	r.session = Session{Model: newCfg.DefaultModel}
	if newCfg.SystemPrompt != "" {
		r.session.Messages = append(r.session.Messages, api.Message{
			Role: "system", Content: newCfg.SystemPrompt,
		})
	}
	r.initLocalSession()

	r.models = modelCache{}
	r.chats = chatCache{}
	r.knowledge = knowledgeCache{}
	r.functions = functionCache{}
	r.preloadModels()
	r.preloadChats()
	r.preloadKnowledge()
	r.preloadFunctions()

	host := strings.TrimPrefix(strings.TrimPrefix(newCfg.BaseURL, "https://"), "http://")
	return SlashResult{
		Cleared:        true,
		ReloadMessages: true,
		Output:         fmt.Sprintf("switched to profile %q (%s)", name, host),
	}
}

func (r *REPL) slashProfile(args []string) SlashResult {
	if len(args) == 0 {
		return SlashResult{Output: "type /profile in the input to switch profiles"}
	}
	switch args[0] {
	case "list":
		return SlashResult{Output: r.formatProfilesList()}
	default:
		name := strings.Join(args, " ")
		file, err := config.ReadFile()
		if err != nil {
			return SlashResult{Err: err}
		}
		file.Normalize()
		for _, n := range file.ProfileNames() {
			if n == name || strings.HasPrefix(n, name) {
				return r.SwitchProfile(n)
			}
		}
		return SlashResult{Err: fmt.Errorf("profile %q not found", name)}
	}
}

func (r *REPL) formatProfilesList() string {
	file, err := config.ReadFile()
	if err != nil {
		return "error: " + err.Error()
	}
	file.Normalize()
	names := file.ProfileNames()
	if len(names) == 0 {
		return "no profiles — run: owui config profile add <name>"
	}
	var b strings.Builder
	for _, name := range names {
		marker := " "
		if name == r.ProfileName() {
			marker = "*"
		}
		ps := file.Profiles[name]
		fmt.Fprintf(&b, "%s %s\n", marker, config.ProfileSummary(name, ps))
	}
	return strings.TrimRight(b.String(), "\n")
}