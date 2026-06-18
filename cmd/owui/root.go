package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/output"
	"github.com/toonvank/owui/internal/repl"
	"github.com/toonvank/owui/internal/setup"
	"github.com/toonvank/owui/internal/tui"
)

var (
	version   = "0.1.0"
	cfg       config.Config
	jsonOut   bool
	quietMode bool
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "owui",
		Short:         "Professional CLI for Open WebUI",
		Long:          "Connect to your Open WebUI server for chat, models, knowledge, and more.\n\nRun without arguments to start interactive mode.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.Load()
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !quietMode {
				output.Banner(version)
			}
			if setup.Needs(cfg) {
				if repl.IsInteractive() {
					output.Info("No configuration found — starting setup wizard")
					if err := setup.RunWizard(&cfg, false); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("not configured — run `owui setup` (or set OWUI_BASE_URL and OWUI_API_KEY)")
				}
			}
			client, err := mustClient()
			if err != nil {
				return err
			}
			r := repl.New(client, cfg)
			if repl.IsInteractive() {
				return tui.Run(r)
			}
			return r.RunBasic()
		},
	}

	root.PersistentFlags().BoolVar(&jsonOut, "json", false, "output raw JSON")
	root.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "suppress non-essential output")

	root.AddCommand(
		chatCmd(),
		modelsCmd(),
		chatsCmd(),
		knowledgeCmd(),
		filesCmd(),
		authCmd(),
		configCmd(),
		setupCmd(),
		pullCmd(),
		statusCmd(),
		functionsCmd(),
		completionCmd(),
	)

	root.Version = version
	return root
}

func mustClient() (*api.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return api.New(cfg), nil
}