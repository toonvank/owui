package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/output"
	"golang.org/x/term"
)

type Session struct {
	Model      string
	Title      string
	Messages   []api.Message
	ChatID     string
	LocalID    string
	LocalTitle string
}

type REPL struct {
	client  *api.Client
	cfg     config.Config
	session Session
	models  modelCache
	chats   chatCache
}

func New(client *api.Client, cfg config.Config) *REPL {
	s := Session{Model: cfg.DefaultModel}
	if cfg.SystemPrompt != "" {
		s.Messages = append(s.Messages, api.Message{Role: "system", Content: cfg.SystemPrompt})
	}
	r := &REPL{client: client, cfg: cfg, session: s}
	r.initLocalSession()
	r.preloadModels()
	r.preloadChats()
	return r
}

func (r *REPL) Run() error {
	return r.RunBasic()
}

// IsInteractive reports whether stdin and stdout are both terminals.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// RunBasic runs line-oriented REPL for piped or non-TTY stdin.
func (r *REPL) RunBasic() error {
	output.Info(fmt.Sprintf("Connected to %s", r.cfg.BaseURL))
	output.Info(fmt.Sprintf("Model: %s", r.session.Model))
	r.printShortcutBar()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n› ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			return nil
		}
		r.handleLine(strings.TrimSpace(line))
	}
}

func (r *REPL) handleLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	if line == "?" {
		fmt.Println(ShortcutsPanel())
		r.printShortcutBar()
		return
	}

	if strings.HasPrefix(line, "/") {
		result := r.RunSlashCommand(line)
		if result.Quit {
			os.Exit(0)
		}
		if result.Err != nil {
			output.Error(result.Err.Error())
		} else if result.Output != "" {
			fmt.Println(result.Output)
		}
		if result.Cleared {
			// already cleared in RunSlashCommand
		}
		r.printShortcutBar()
		return
	}

	prompt, modelChanged, err := r.ParseAtPrefix(line)
	if err != nil {
		output.Error(err.Error())
		r.printShortcutBar()
		return
	}
	if modelChanged != "" {
		output.Success("model set to " + modelChanged)
		fmt.Println()
	}
	if prompt == "" {
		r.printShortcutBar()
		return
	}

	if err := r.send(prompt); err != nil {
		output.Error(err.Error())
	}
}

func (r *REPL) send(prompt string) error {
	fmt.Print("\n")
	var onDelta func(string)
	if r.cfg.Stream {
		onDelta = func(delta string) {
			fmt.Print(delta)
		}
	}
	reply, err := r.ChatUserMessage(prompt, onDelta)
	if r.cfg.Stream {
		fmt.Println()
	} else {
		fmt.Println(reply)
	}
	if err != nil {
		r.printShortcutBar()
		return err
	}
	r.printShortcutBar()
	return nil
}