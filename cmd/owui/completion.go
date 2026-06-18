package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		root := cmd.Root()
		switch args[0] {
		case "bash":
			return root.GenBashCompletion(os.Stdout)
		case "zsh":
			return root.GenZshCompletion(os.Stdout)
		case "fish":
			return root.GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	}

	return cmd
}