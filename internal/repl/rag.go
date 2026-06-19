package repl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/files"
	"github.com/toonvank/owui/internal/session"
)

func (r *REPL) chatOptions() *api.ChatOptions {
	opts := &api.ChatOptions{ChatID: r.session.ChatID}
	has := r.session.ChatID != ""

	if r.session.CollectionID != "" {
		opts.Collection = r.session.CollectionID
		has = true
	}
	if len(r.session.FileIDs) > 0 {
		opts.FileIDs = append([]string(nil), r.session.FileIDs...)
		has = true
	}
	if r.session.FiltersCustomized {
		opts.FilterIDs = append([]string(nil), r.session.ActiveFilterIDs...)
		opts.ExplicitFilters = true
		has = true
	}
	if r.session.ToolsCustomized {
		opts.ToolIDs = append([]string(nil), r.session.ActiveToolIDs...)
		opts.ExplicitTools = true
		has = true
	}
	if !has {
		return nil
	}
	return opts
}

func (r *REPL) slashFile(args []string, rawLine string) SlashResult {
	if len(args) == 0 {
		return SlashResult{Output: r.formatAttachedFiles()}
	}
	switch args[0] {
	case "upload":
		path := fileUploadPath(args, rawLine)
		if path == "" {
			return SlashResult{Err: fmt.Errorf("usage: /file upload <path>")}
		}
		return r.slashFileUpload(path)
	case "list":
		return SlashResult{Output: r.formatAttachedFiles()}
	case "status":
		if len(args) < 2 {
			return SlashResult{Output: r.formatFileStatuses()}
		}
		st, err := r.client.FileStatus(args[1])
		if err != nil {
			return SlashResult{Err: err}
		}
		return SlashResult{Output: fmt.Sprintf("%s: %s", shortID(args[1]), st.Status)}
	case "clear":
		r.clearAttachedFiles()
		return SlashResult{Output: "detached all files"}
	case "remove", "rm":
		if len(args) < 2 {
			return SlashResult{Err: fmt.Errorf("usage: /file remove <id>")}
		}
		if !r.removeAttachedFile(args[1]) {
			return SlashResult{Err: fmt.Errorf("no attached file matching %q", args[1])}
		}
		r.persistSession()
		return SlashResult{Output: "file detached"}
	default:
		return SlashResult{Err: fmt.Errorf("unknown /file subcommand: %s", args[0])}
	}
}

func fileUploadPath(args []string, rawLine string) string {
	const prefix = "/file upload "
	if strings.HasPrefix(rawLine, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(rawLine, prefix))
	}
	if len(args) >= 2 {
		return args[1]
	}
	return ""
}

func (r *REPL) slashFileUpload(path string) SlashResult {
	path = expandHome(path)
	if _, err := os.Stat(path); err != nil {
		return SlashResult{Err: fmt.Errorf("file not found: %s", path)}
	}

	resp, err := r.client.UploadFile(path)
	if err != nil {
		return SlashResult{Err: err}
	}
	if err := files.WaitForProcessing(r.client, resp.ID, files.DefaultWaitTimeout); err != nil {
		return SlashResult{Err: err}
	}

	name := resp.Name
	if name == "" {
		name = filepath.Base(path)
	}
	r.attachFile(resp.ID, name)
	return SlashResult{Output: fmt.Sprintf("attached %s (%s)", name, shortID(resp.ID))}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func (r *REPL) attachFile(id, name string) {
	for _, f := range r.session.AttachedFiles {
		if f.ID == id {
			r.persistSession()
			return
		}
	}
	r.session.AttachedFiles = append(r.session.AttachedFiles, session.AttachedFile{ID: id, Name: name})
	r.syncFileIDs()
	r.persistSession()
}

func (r *REPL) removeAttachedFile(idPrefix string) bool {
	idPrefix = strings.ToLower(idPrefix)
	out := r.session.AttachedFiles[:0]
	removed := false
	for _, f := range r.session.AttachedFiles {
		if !removed && (f.ID == idPrefix || strings.HasPrefix(strings.ToLower(f.ID), idPrefix)) {
			removed = true
			continue
		}
		out = append(out, f)
	}
	if !removed {
		return false
	}
	r.session.AttachedFiles = out
	r.syncFileIDs()
	return true
}

func (r *REPL) clearAttachedFiles() {
	r.session.AttachedFiles = nil
	r.session.FileIDs = nil
	r.persistSession()
}

func (r *REPL) syncFileIDs() {
	r.session.FileIDs = make([]string, len(r.session.AttachedFiles))
	for i, f := range r.session.AttachedFiles {
		r.session.FileIDs[i] = f.ID
	}
}

func (r *REPL) formatAttachedFiles() string {
	if len(r.session.AttachedFiles) == 0 {
		return "no files attached — /file upload <path>"
	}
	var b strings.Builder
	for _, f := range r.session.AttachedFiles {
		fmt.Fprintf(&b, "  %s  %s\n", shortID(f.ID), f.Name)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (r *REPL) formatFileStatuses() string {
	if len(r.session.AttachedFiles) == 0 {
		return "no files attached"
	}
	var b strings.Builder
	for _, f := range r.session.AttachedFiles {
		st, err := r.client.FileStatus(f.ID)
		status := "unknown"
		if err == nil {
			status = st.Status
		} else {
			status = "error: " + err.Error()
		}
		fmt.Fprintf(&b, "  %s  %s  %s\n", shortID(f.ID), f.Name, status)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (r *REPL) slashKnowledge(args []string) SlashResult {
	if len(args) == 0 {
		return SlashResult{Output: "type /knowledge in the input to pick a collection"}
	}
	switch args[0] {
	case "list":
		return SlashResult{Output: r.formatKnowledgeList(args[1:])}
	case "clear":
		r.clearKnowledgeCollection()
		return SlashResult{Output: "knowledge collection detached"}
	default:
		query := strings.Join(args, " ")
		if item := r.bestKnowledgeMatch(query); item.ID != "" {
			r.setKnowledgeCollection(item.ID, item.Name)
			return SlashResult{Output: fmt.Sprintf("using knowledge collection %q", item.Name)}
		}
		return SlashResult{Err: fmt.Errorf("no knowledge collection matching %q", query)}
	}
}

func (r *REPL) setKnowledgeCollection(id, name string) {
	r.session.CollectionID = id
	r.session.CollectionName = name
	r.persistSession()
}

func (r *REPL) clearKnowledgeCollection() {
	r.session.CollectionID = ""
	r.session.CollectionName = ""
	r.persistSession()
}

func (r *REPL) formatKnowledgeList(args []string) string {
	items, err := r.client.ListKnowledge()
	if err != nil {
		return "error: " + err.Error()
	}
	filter := strings.ToLower(strings.Join(args, " "))
	limit := 20
	shown := 0
	var b strings.Builder
	for _, item := range items {
		if filter != "" {
			name := strings.ToLower(item.Name)
			id := strings.ToLower(item.ID)
			if !strings.Contains(name, filter) && !strings.HasPrefix(id, filter) {
				continue
			}
		}
		marker := " "
		if item.ID == r.session.CollectionID {
			marker = "*"
		}
		desc := item.Description
		if len(desc) > 35 {
			desc = desc[:32] + "..."
		}
		fmt.Fprintf(&b, "%s %s  %s", marker, shortID(item.ID), item.Name)
		if desc != "" {
			fmt.Fprintf(&b, "  (%s)", desc)
		}
		b.WriteString("\n")
		shown++
		if shown >= limit {
			break
		}
	}
	if shown == 0 {
		return "no knowledge collections found"
	}
	return strings.TrimRight(b.String(), "\n")
}

// RAGContextLabel returns a short summary of active RAG context for the TUI header.
func (r *REPL) RAGContextLabel() string {
	var parts []string
	if r.session.CollectionName != "" {
		name := r.session.CollectionName
		if len(name) > 18 {
			name = name[:15] + "..."
		}
		parts = append(parts, "kb:"+name)
	} else if r.session.CollectionID != "" {
		parts = append(parts, "kb:"+shortID(r.session.CollectionID))
	}
	if n := len(r.session.AttachedFiles); n > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s)", n))
	}
	if label := r.IntegrationsLabel(); label != "" {
		parts = append(parts, label)
	}
	return strings.Join(parts, " · ")
}

