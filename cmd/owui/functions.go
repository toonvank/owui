package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/output"
)

func functionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "functions",
		Aliases: []string{"filters"},
		Short:   "List Open WebUI functions and filters",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List installed functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			fns, err := client.ListFunctions()
			if err != nil {
				return err
			}
			if jsonOut {
				b, _ := json.MarshalIndent(fns, "", "  ")
				output.JSON(string(b))
				return nil
			}
			rows := make([][]string, 0, len(fns))
			for _, fn := range fns {
				scope := "model"
				if fn.IsGlobal {
					scope = "global"
				}
				active := "off"
				if fn.IsActive {
					active = "on"
				}
				rows = append(rows, []string{fn.ID, fn.Type, fn.Name, scope, active})
			}
			output.Table([]string{"ID", "TYPE", "NAME", "SCOPE", "ACTIVE"}, rows)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "model <id>",
		Short: "Show features and filters for a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			model, err := client.ModelByID(args[0])
			if err != nil {
				return err
			}
			if model.ID == "" {
				return fmt.Errorf("model not found: %s", args[0])
			}
			meta := model.Meta()
			features := api.FeaturesFromModel(model)
			fmt.Printf("Model: %s\n", model.ID)
			fmt.Printf("Default filter IDs: %v\n", meta.DefaultFilterIDs)
			fmt.Printf("Filter IDs: %v\n", meta.FilterIDs)
			if len(features) > 0 {
				fmt.Printf("Auto features: %v\n", features)
			} else {
				fmt.Println("Auto features: (none)")
			}
			return nil
		},
	})

	return cmd
}