package main

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/output"
)

func modelsCmd() *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "models",
		Short: "List and inspect models",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List available models",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			models, err := client.ListModels()
			if err != nil {
				return err
			}

			if filter != "" {
				f := strings.ToLower(filter)
				filtered := make([]api.Model, 0)
				for _, m := range models {
					if strings.Contains(strings.ToLower(m.ID), f) {
						filtered = append(filtered, m)
					}
				}
				models = filtered
			}

			if jsonOut {
				b, _ := json.MarshalIndent(models, "", "  ")
				output.JSON(string(b))
				return nil
			}

			rows := make([][]string, 0, len(models))
			for _, m := range models {
				owner := m.OwnedBy
				if owner == "" {
					owner = "-"
				}
				rows = append(rows, []string{m.ID, owner})
			}
			output.Table([]string{"MODEL", "OWNER"}, rows)
			return nil
		},
	}
	list.Flags().StringVar(&filter, "filter", "", "filter models by name")

	cmd.AddCommand(list)
	return cmd
}