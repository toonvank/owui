package repl

import (
	"fmt"
	"strings"

	"github.com/toonvank/owui/internal/api"
)

// SlashResult is the outcome of a slash command executed from the TUI.
type SlashResult struct {
	Quit           bool
	Output         string
	Err            error
	Cleared        bool
	ReloadMessages bool
	ModelSet       string
}

// GetSuggestions returns autocomplete entries for the current input line.
func (r *REPL) GetSuggestions(line string) []Suggestion {
	if strings.HasPrefix(line, "/") {
		return r.completeSlash(line)
	}
	if strings.HasPrefix(line, "@") {
		return r.modelSuggests(line[1:], true)
	}
	return nil
}

// ChatUserMessage sends a user message and returns the assistant reply.
func (r *REPL) ChatUserMessage(prompt string, onDelta func(string)) (string, error) {
	r.session.Messages = append(r.session.Messages, api.Message{Role: "user", Content: prompt})

	opts := &api.ChatOptions{ChatID: r.session.ChatID}
	var streamFn func(string) error
	if onDelta != nil {
		streamFn = func(delta string) error {
			onDelta(delta)
			return nil
		}
	}

	reply, err := r.client.ChatWithOptions(r.session.Messages, r.session.Model, r.cfg.Stream, opts, streamFn)
	if err != nil {
		return "", err
	}
	r.session.Messages = append(r.session.Messages, api.Message{Role: "assistant", Content: reply})
	r.persistSession()
	return reply, nil
}

// ParseAtPrefix handles @model or @model message patterns.
// Returns the remaining prompt (if any) and the model id when switched.
func (r *REPL) ParseAtPrefix(line string) (prompt string, modelChanged string, err error) {
	if !strings.HasPrefix(line, "@") {
		return line, "", nil
	}
	rest := strings.TrimPrefix(line, "@")
	parts := strings.SplitN(rest, " ", 2)
	query := parts[0]
	if query == "" {
		return "", "", nil
	}

	id := r.bestModelMatch(query)
	if id == "" {
		return "", "", fmt.Errorf("no model matching %s", query)
	}
	r.session.Model = id
	modelChanged = id
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1]), modelChanged, nil
	}
	return "", modelChanged, nil
}

// StreamEnabled reports whether responses stream token-by-token.
func (r *REPL) StreamEnabled() bool {
	return r.cfg.Stream
}

// ShortcutBarLine returns the footer hint line for the TUI.
func (r *REPL) ShortcutBarLine() string {
	return r.shortcutBarLine()
}

// BaseURL returns the configured server URL.
func (r *REPL) BaseURL() string {
	return r.cfg.BaseURL
}

// CurrentModel returns the active model id.
func (r *REPL) CurrentModel() string {
	return r.session.Model
}

// SessionMessages returns the current conversation messages.
func (r *REPL) SessionMessages() []api.Message {
	return r.session.Messages
}

// RunSlashCommand executes a slash command line and returns structured output.
func (r *REPL) RunSlashCommand(line string) SlashResult {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return SlashResult{}
	}
	cmd := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	switch cmd {
	case "quit", "exit", "q":
		return SlashResult{Quit: true}
	case "help", "h", "?":
		return SlashResult{Output: helpText()}
	case "clear":
		r.newLocalSession()
		return SlashResult{Cleared: true, Output: "conversation cleared"}
	case "session":
		return r.slashSession(args)
	case "model":
		if len(args) == 0 {
			return SlashResult{Output: "current model: " + r.session.Model}
		}
		query := strings.Join(args, " ")
		if id := r.bestModelMatch(query); id != "" {
			r.session.Model = id
			return SlashResult{ModelSet: id, Output: "model set to " + id}
		}
		r.session.Model = query
		return SlashResult{ModelSet: query, Output: "model set to " + query}
	case "models":
		return SlashResult{Output: r.formatModels(args)}
	case "stream":
		r.cfg.Stream = !r.cfg.Stream
		return SlashResult{Output: fmt.Sprintf("streaming %v", r.cfg.Stream)}
	case "resume", "load":
		if len(args) == 0 {
			return SlashResult{Err: fmt.Errorf("usage: /resume <chat-id>")}
		}
		return r.slashLoad(strings.Join(args, " "))
	case "chats":
		if len(args) == 0 {
			return SlashResult{Output: "type /chats in the input to browse and resume"}
		}
		return SlashResult{Output: r.formatChats(args)}
	case "server":
		return SlashResult{Output: r.formatServerInfo()}
	case "setup":
		return SlashResult{Output: "Exit owui and run: owui setup\nChange URL only: owui config set url <url>"}
	case "filters", "functions":
		return SlashResult{Output: r.formatFilters()}
	default:
		return SlashResult{Err: fmt.Errorf("unknown command: /%s", cmd)}
	}
}

func (r *REPL) slashSession(args []string) SlashResult {
	if len(args) == 0 {
		return SlashResult{Output: fmt.Sprintf("local session %s (%d messages)", r.session.LocalID, len(r.session.Messages))}
	}
	switch args[0] {
	case "new":
		r.newLocalSession()
		return SlashResult{Cleared: true, Output: "started new local session " + shortID(r.session.LocalID)}
	case "list":
		return SlashResult{Output: r.formatLocalSessions()}
	case "load":
		if len(args) < 2 {
			return SlashResult{Err: fmt.Errorf("usage: /session load <id>")}
		}
		if err := r.loadLocalSession(args[1]); err != nil {
			return SlashResult{Err: err}
		}
		return SlashResult{Output: fmt.Sprintf("loaded local session %s (%d messages)", shortID(r.session.LocalID), len(r.session.Messages)), ReloadMessages: true}
	case "save":
		r.persistSession()
		return SlashResult{Output: "session saved"}
	default:
		return SlashResult{Err: fmt.Errorf("unknown /session subcommand: %s", args[0])}
	}
}

func (r *REPL) slashLoad(id string) SlashResult {
	loaded, err := r.client.LoadChat(id)
	if err != nil {
		return SlashResult{Err: err}
	}
	r.session.ChatID = loaded.ID
	r.session.Title = loaded.Title
	r.session.LocalTitle = loaded.Title
	r.session.Messages = loaded.Messages
	if loaded.Model != "" {
		r.session.Model = loaded.Model
	}
	r.persistSession()
	title := loaded.Title
	if title == "" {
		title = loaded.ID[:8]
	}
	return SlashResult{
		Output:         fmt.Sprintf("resumed %q (%d messages, model %s)", title, len(loaded.Messages), r.session.Model),
		ReloadMessages: true,
	}
}

func (r *REPL) formatModels(args []string) string {
	models, err := r.client.ListModels()
	if err != nil {
		return "error: " + err.Error()
	}
	if len(args) > 0 {
		filter := strings.ToLower(strings.Join(args, " "))
		filtered := make([]api.Model, 0)
		for _, m := range models {
			if strings.Contains(strings.ToLower(m.ID), filter) {
				filtered = append(filtered, m)
			}
		}
		models = filtered
	}
	limit := 30
	note := ""
	if len(models) > limit {
		models = models[:limit]
		note = fmt.Sprintf("\nshowing first %d models (use /models <filter>)", limit)
	}
	var b strings.Builder
	for _, m := range models {
		b.WriteString("  ")
		b.WriteString(m.ID)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + note
}

func (r *REPL) formatServerInfo() string {
	return fmt.Sprintf("server: %s\nmodel:   %s\n\nchange server: owui config set url <url>\nreconfigure:    owui setup", r.cfg.BaseURL, r.session.Model)
}

func (r *REPL) formatChats(args []string) string {
	chats, err := r.client.ListChats(1)
	if err != nil {
		return "error: " + err.Error()
	}
	filter := strings.ToLower(strings.Join(args, " "))
	limit := 15
	shown := 0
	var b strings.Builder
	for _, ch := range chats {
		if filter != "" {
			title := strings.ToLower(ch.Title)
			id := strings.ToLower(ch.ID)
			if !strings.Contains(title, filter) && !strings.HasPrefix(id, filter) {
				continue
			}
		}
		title := ch.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}
		fmt.Fprintf(&b, "  %s  %s\n", ch.ID[:8], title)
		shown++
		if shown >= limit {
			fmt.Fprintf(&b, "showing %d chats — /resume <id> to resume", limit)
			break
		}
	}
	if shown == 0 {
		return "no chats found"
	}
	return strings.TrimRight(b.String(), "\n")
}

func (r *REPL) formatFilters() string {
	fns, err := r.client.ListFunctions()
	if err != nil {
		return "error: " + err.Error()
	}
	model, _ := r.client.ModelByID(r.session.Model)
	meta := model.Meta()
	var b strings.Builder
	fmt.Fprintf(&b, "model %s — default filters: %v\n", r.session.Model, meta.DefaultFilterIDs)
	if features := api.FeaturesFromModel(model); len(features) > 0 {
		keys := make([]string, 0, len(features))
		for k := range features {
			keys = append(keys, k)
		}
		fmt.Fprintf(&b, "auto features: %s\n", strings.Join(keys, ", "))
	}
	for _, fn := range fns {
		if fn.Type != "filter" {
			continue
		}
		state := "off"
		if fn.IsActive {
			state = "on"
		}
		scope := "model"
		if fn.IsGlobal {
			scope = "global"
		}
		fmt.Fprintf(&b, "  [%s/%s] %s (%s)\n", state, scope, fn.Name, fn.ID)
	}
	return strings.TrimRight(b.String(), "\n")
}