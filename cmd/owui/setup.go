package main

import (
	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/setup"
)

func setupCmd() *cobra.Command {
	var reconfigure bool
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard for server URL and authentication",
		Long: `Walk through connecting owui to your Open WebUI instance.

Run this on first use, or any time you want to change server or credentials.
You can also change just the URL later:

  owui config set url https://your-server.example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.RunWizard(&cfg, reconfigure)
		},
	}
	cmd.Flags().BoolVar(&reconfigure, "reconfigure", false, "re-run setup to change server or credentials")
	return cmd
}