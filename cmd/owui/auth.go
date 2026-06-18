package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/output"
	"golang.org/x/term"
)

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and API key management",
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Login with email/password and save API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			if email == "" {
				fmt.Print("Email: ")
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				email = strings.TrimSpace(line)
			}

			password, _ := cmd.Flags().GetString("password")
			if password == "" {
				fmt.Print("Password: ")
				b, err := term.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					return err
				}
				password = string(b)
			}

			baseURL, _ := cmd.Flags().GetString("url")
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}

			client := api.New(cfg)
			auth, err := client.SignIn(email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cfg.APIKey = auth.Token
			if err := config.Save(cfg); err != nil {
				return err
			}

			output.Success(fmt.Sprintf("logged in as %s (%s)", auth.Name, auth.Email))
			output.Info("token saved to config (use Settings > Account to create a persistent API key)")
			return nil
		},
	}
	login.Flags().String("email", "", "account email")
	login.Flags().String("password", "", "account password")
	login.Flags().String("url", "", "server base URL")

	token := &cobra.Command{
		Use:   "token <api-key>",
		Short: "Save an API key to config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.APIKey = args[0]
			if err := config.Save(cfg); err != nil {
				return err
			}
			output.Success("API key saved")
			return nil
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Printf("Config: %s\n", path)
			fmt.Printf("Server: %s\n", cfg.BaseURL)
			if cfg.APIKey == "" {
				fmt.Println("Auth:   not configured")
				return nil
			}
			masked := cfg.APIKey
			if len(masked) > 12 {
				masked = masked[:8] + "..." + masked[len(masked)-4:]
			}
			fmt.Printf("Auth:   configured (%s)\n", masked)

			client, err := mustClient()
			if err != nil {
				return err
			}
			srv, err := client.Config()
			if err != nil {
				return err
			}
			output.Success(fmt.Sprintf("connected to %s v%s", srv.Name, srv.Version))
			return nil
		},
	}

	logout := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.APIKey = ""
			if err := config.Save(cfg); err != nil {
				return err
			}
			output.Success("credentials removed")
			return nil
		},
	}

	cmd.AddCommand(login, token, status, logout)
	return cmd
}