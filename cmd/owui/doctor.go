package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/output"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose configuration and server connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			var issues int

			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Printf("Config file: %s\n", path)
			if !config.FileExists() {
				output.Error("config file not found — run `owui setup`")
				issues++
			} else {
				output.Success("config file exists")
			}

			fmt.Printf("Profile:     %s\n", cfg.ProfileName)
			if cfg.BaseURL == "" {
				output.Error("base_url not set")
				issues++
			} else {
				output.Success("base_url: " + cfg.BaseURL)
			}

			if cfg.APIKey == "" {
				output.Error("api_key not set — run `owui auth login` or `owui auth token`")
				issues++
			} else {
				output.Success("api_key: " + cfg.MaskedAPIKey())
			}

			if cfg.DefaultModel == "" {
				output.Info("default_model not set (pick at runtime)")
			} else {
				output.Success("default_model: " + cfg.DefaultModel)
			}

			sessDir, err := config.SessionsDir(cfg.ProfileName)
			if err != nil {
				output.Error("sessions dir: " + err.Error())
				issues++
			} else {
				fmt.Printf("Sessions:    %s\n", sessDir)
				if _, err := os.Stat(sessDir); err != nil {
					output.Error("sessions directory not accessible")
					issues++
				} else {
					output.Success("sessions directory ok")
				}
			}

			if cfg.BaseURL != "" && cfg.APIKey != "" {
				client, err := mustClient()
				if err != nil {
					output.Error("client: " + err.Error())
					issues++
				} else {
					if err := client.Health(); err != nil {
						output.Error("health: " + err.Error())
						issues++
					} else {
						output.Success("server health ok")
					}
					srv, err := client.Config()
					if err != nil {
						output.Error("config endpoint: " + err.Error())
						issues++
					} else {
						output.Success(fmt.Sprintf("connected to %s v%s", srv.Name, srv.Version))
					}
					models, err := client.ListModels()
					if err != nil {
						output.Error("models: " + err.Error())
						issues++
					} else {
						output.Success(fmt.Sprintf("%d models available", len(models)))
						if cfg.DefaultModel != "" {
							found := false
							for _, m := range models {
								if m.ID == cfg.DefaultModel {
									found = true
									break
								}
							}
							if !found {
								output.Error("default_model " + cfg.DefaultModel + " not found on server")
								issues++
							} else {
								output.Success("default_model reachable")
							}
						}
					}
				}
			}

			fmt.Println()
			if issues == 0 {
				output.Success("all checks passed")
				return nil
			}
			return fmt.Errorf("%d issue(s) found", issues)
		},
	}
}