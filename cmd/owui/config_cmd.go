package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/output"
	"gopkg.in/yaml.v3"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	show := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			safe := cfg
			if safe.APIKey != "" {
				safe.APIKey = safe.APIKey[:8] + "..." + safe.APIKey[len(safe.APIKey)-4:]
			}
			b, err := yaml.Marshal(safe)
			if err != nil {
				return err
			}
			fmt.Print(string(b))
			return nil
		},
	}

	init := &cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			def := config.Default()
			url, _ := cmd.Flags().GetString("url")
			if url != "" {
				def.BaseURL = url
			}
			model, _ := cmd.Flags().GetString("model")
			if model != "" {
				def.DefaultModel = model
			}
			key, _ := cmd.Flags().GetString("api-key")
			if key != "" {
				def.APIKey = key
			}
			if err := config.Save(def); err != nil {
				return err
			}
			path, _ := config.Path()
			output.Success("config written to " + path)
			return nil
		},
	}
	init.Flags().String("url", "http://localhost:3000", "Open WebUI server URL")
	init.Flags().String("model", "", "default model")
	init.Flags().String("api-key", "", "API key")

	set := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			switch key {
			case "base_url", "url":
				cfg.BaseURL = value
			case "api_key", "token":
				cfg.APIKey = value
			case "default_model", "model":
				cfg.DefaultModel = value
			case "system_prompt":
				cfg.SystemPrompt = value
			case "stream":
				cfg.Stream = value == "true" || value == "1"
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			output.Success(key + " updated")
			return nil
		},
	}

	cmd.AddCommand(show, init, set)
	return cmd
}