package repl

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
)

func writeClipboard(text string) error {
	if err := clipboard.WriteAll(text); err == nil {
		return nil
	}
	return tryClipboardCLI(text)
}

func tryClipboardCLI(text string) error {
	for _, spec := range []struct {
		cmd  string
		args []string
	}{
		{"wl-copy", nil},
		{"xclip", []string{"-selection", "clipboard"}},
		{"xsel", []string{"--clipboard", "--input"}},
	} {
		if _, err := exec.LookPath(spec.cmd); err != nil {
			continue
		}
		args := append(spec.args, "-")
		c := exec.Command(spec.cmd, args...)
		c.Stdin = strings.NewReader(text)
		if err := c.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("clipboard unavailable (install wl-clipboard, xclip, or xsel)")
}