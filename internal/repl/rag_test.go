package repl

import (
	"strings"
	"testing"

	"github.com/toonvank/owui/internal/session"
)

func TestRAGContextLabel(t *testing.T) {
	r := &REPL{
		session: Session{
			CollectionName: "Docs",
			AttachedFiles: []session.AttachedFile{
				{ID: "abc", Name: "readme.md"},
				{ID: "def", Name: "notes.txt"},
			},
		},
	}
	label := r.RAGContextLabel()
	if !strings.Contains(label, "kb:Docs") {
		t.Fatalf("expected kb name in label, got %q", label)
	}
	if !strings.Contains(label, "2 file(s)") {
		t.Fatalf("expected file count in label, got %q", label)
	}
}

func TestChatOptionsIncludesRAG(t *testing.T) {
	r := &REPL{
		session: Session{
			ChatID:       "chat-1",
			CollectionID: "kb-1",
			FileIDs:      []string{"f1", "f2"},
		},
	}
	opts := r.chatOptions()
	if opts == nil {
		t.Fatal("expected chat options")
	}
	if opts.Collection != "kb-1" {
		t.Fatalf("collection = %q", opts.Collection)
	}
	if len(opts.FileIDs) != 2 {
		t.Fatalf("file ids = %v", opts.FileIDs)
	}
}

func TestClearAttachedFiles(t *testing.T) {
	r := &REPL{
		session: Session{
			AttachedFiles: []session.AttachedFile{{ID: "f1", Name: "a.txt"}},
			FileIDs:       []string{"f1"},
		},
	}
	r.clearAttachedFiles()
	if len(r.session.AttachedFiles) != 0 || len(r.session.FileIDs) != 0 {
		t.Fatalf("expected cleared attachments, got files=%v ids=%v", r.session.AttachedFiles, r.session.FileIDs)
	}
}