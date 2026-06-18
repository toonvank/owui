package main

import (
	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/output"
)

func pullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <model>",
		Short: "Pull an Ollama model via Open WebUI proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			output.Info("pulling " + args[0] + "...")
			if err := client.PullModel(args[0]); err != nil {
				return err
			}
			output.Success("pull complete")
			return nil
		},
	}
}