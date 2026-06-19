package main

import (
	"encoding/json"
	"fmt"
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
				owner := m.ModelKindLabel()
				caps := strings.Join(m.CapabilityTags(), ", ")
				if caps == "" {
					caps = "-"
				}
				rows = append(rows, []string{m.ID, owner, caps})
			}
			output.Table([]string{"MODEL", "PROVIDER", "CAPABILITIES"}, rows)
			return nil
		},
	}
	list.Flags().StringVar(&filter, "filter", "", "filter models by name")

	info := &cobra.Command{
		Use:   "info <id>",
		Short: "Show model capabilities, filters, and features",
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
			if jsonOut {
				meta := model.Meta()
				payload := map[string]any{
					"id":           model.ID,
					"name":         model.Name,
					"provider":     model.ModelKindLabel(),
					"capabilities": model.CapabilityTags(),
					"meta":         meta,
					"features":     api.FeaturesFromModel(model),
				}
				b, _ := json.MarshalIndent(payload, "", "  ")
				output.JSON(string(b))
				return nil
			}
			meta := model.Meta()
			features := api.FeaturesFromModel(model)
			fmt.Printf("Model: %s\n", model.ID)
			fmt.Printf("Provider: %s\n", model.ModelKindLabel())
			if tags := model.CapabilityTags(); len(tags) > 0 {
				fmt.Printf("Capabilities: %s\n", strings.Join(tags, ", "))
			}
			fmt.Printf("Default filter IDs: %v\n", meta.DefaultFilterIDs)
			fmt.Printf("Filter IDs: %v\n", meta.FilterIDs)
			if len(features) > 0 {
				fmt.Printf("Auto features: %v\n", features)
			} else {
				fmt.Println("Auto features: (none)")
			}
			return nil
		},
	}

	ollama := &cobra.Command{
		Use:   "ollama",
		Short: "Ollama-specific model commands",
	}
	ollamaList := &cobra.Command{
		Use:   "list",
		Short: "List models available via Ollama",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			models, err := client.ListOllamaModels()
			if err != nil {
				return err
			}
			if jsonOut {
				b, _ := json.MarshalIndent(models, "", "  ")
				output.JSON(string(b))
				return nil
			}
			for _, m := range models {
				name, _ := m["name"].(string)
				size, _ := m["size"].(float64)
				if name == "" {
					continue
				}
				if size > 0 {
					fmt.Printf("%s  (%.1f GB)\n", name, size/1e9)
				} else {
					fmt.Println(name)
				}
			}
			return nil
		},
	}
	ollama.AddCommand(ollamaList)

	cmd.AddCommand(list, info, ollama)
	return cmd
}