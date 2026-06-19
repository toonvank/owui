package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
			file, err := config.ReadFile()
			if err != nil {
				return err
			}
			file.Normalize()
			if file.Profiles != nil {
				for name, ps := range file.Profiles {
					if ps.APIKey != "" {
						masked := ps.APIKey
						if len(masked) > 12 {
							masked = masked[:8] + "..." + masked[len(masked)-4:]
						}
						ps.APIKey = masked
						file.Profiles[name] = ps
					}
				}
			}
			b, err := yaml.Marshal(file)
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
			def.ProfileName = config.DefaultProfile
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
			case "timeout_sec", "timeout":
				n, err := strconv.Atoi(value)
				if err != nil || n <= 0 {
					return fmt.Errorf("timeout_sec must be a positive integer")
				}
				cfg.TimeoutSec = n
			case "insecure_tls":
				cfg.InsecureTLS = value == "true" || value == "1"
			case "custom_api_key_header", "custom_header":
				cfg.CustomHeader = value
			case "apply_model_features":
				b := value == "true" || value == "1"
				cfg.ApplyModelFeatures = &b
			case "filter_ids":
				cfg.FilterIDs = splitCSV(value)
			case "tool_ids":
				cfg.ToolIDs = splitCSV(value)
			case "theme":
				cfg.Theme = value
			case "vim_keys":
				cfg.VimKeys = value == "true" || value == "1"
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

	path := &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Println(p)
			return nil
		},
	}

	edit := &cobra.Command{
		Use:   "edit",
		Short: "Open the config file in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			if _, err := os.Stat(p); os.IsNotExist(err) {
				if err := config.Save(config.Default()); err != nil {
					return err
				}
			}
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			parts := strings.Fields(editor)
			c := exec.Command(parts[0], append(parts[1:], p)...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	profile := &cobra.Command{
		Use:   "profile",
		Short: "Manage named config profiles",
	}

	profileList := &cobra.Command{
		Use:   "list",
		Short: "List configured profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := config.ReadFile()
			if err != nil {
				return err
			}
			file.Normalize()
			names := file.ProfileNames()
			if len(names) == 0 {
				output.Info("no profiles configured")
				return nil
			}
			rows := make([][]string, 0, len(names))
			for _, name := range names {
				active := ""
				if name == cfg.ProfileName || name == file.ActiveProfile {
					active = "*"
				}
				ps := file.Profiles[name]
				rows = append(rows, []string{active, name, ps.BaseURL, ps.DefaultModel})
			}
			output.Table([]string{"", "PROFILE", "URL", "MODEL"}, rows)
			return nil
		},
	}

	profileUse := &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			newCfg, err := config.SwitchActiveProfile(args[0])
			if err != nil {
				return err
			}
			cfg = newCfg
			output.Success("active profile: " + args[0])
			return nil
		},
	}

	profileAdd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a profile (copies current settings by default)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			settings := cfg.ToProfileSettings()
			url, _ := cmd.Flags().GetString("url")
			if url != "" {
				settings.BaseURL = url
			}
			key, _ := cmd.Flags().GetString("api-key")
			if key != "" {
				settings.APIKey = key
			}
			model, _ := cmd.Flags().GetString("model")
			if model != "" {
				settings.DefaultModel = model
			}
			fromCurrent, _ := cmd.Flags().GetBool("from-current")
			if !fromCurrent && url == "" && key == "" {
				// still copy current as base when adding
			}
			if err := config.AddProfile(args[0], settings); err != nil {
				return err
			}
			output.Success("profile " + args[0] + " added")
			return nil
		},
	}
	profileAdd.Flags().String("url", "", "server URL")
	profileAdd.Flags().String("api-key", "", "API key")
	profileAdd.Flags().String("model", "", "default model")
	profileAdd.Flags().Bool("from-current", true, "copy settings from current profile")

	profileRemove := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.RemoveProfile(args[0]); err != nil {
				return err
			}
			output.Success("profile " + args[0] + " removed")
			return nil
		},
	}

	profile.AddCommand(profileList, profileUse, profileAdd, profileRemove)
	cmd.AddCommand(show, init, set, path, edit, profile)
	return cmd
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}