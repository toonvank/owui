package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/output"
)

func pullCmd() *cobra.Command {
	var progress bool

	cmd := &cobra.Command{
		Use:   "pull <model>",
		Short: "Pull an Ollama model via Open WebUI proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			name := args[0]
			if !quietMode {
				output.Info("pulling " + name + "...")
			}

			var onStatus func(string)
			if progress {
				onStatus = func(status string) {
					fmt.Fprintf(os.Stderr, "\r→ %s", status)
				}
			}

			if err := client.PullModelWithProgress(name, onStatus); err != nil {
				if progress {
					fmt.Fprintln(os.Stderr)
				}
				return err
			}
			if progress {
				fmt.Fprintln(os.Stderr)
			}
			output.Success("pull complete")
			return nil
		},
	}

	cmd.Flags().BoolVar(&progress, "progress", false, "show pull progress on stderr")
	return cmd
}