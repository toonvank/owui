package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/output"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check server health and version",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			if err := client.Health(); err != nil {
				return err
			}

			srv, err := client.Config()
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(srv, "", "  ")
				output.JSON(string(b))
				return nil
			}

			output.Table([]string{"FIELD", "VALUE"}, [][]string{
				{"Server", srv.Name},
				{"Version", srv.Version},
				{"URL", cfg.BaseURL},
				{"Status", "healthy"},
			})
			return nil
		},
	}
}