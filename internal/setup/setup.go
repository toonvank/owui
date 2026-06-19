package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/config"
	"github.com/toonvank/owui/internal/output"
	"golang.org/x/term"
)

// Needs reports whether the user must run setup before chatting.
func Needs(cfg config.Config) bool {
	return strings.TrimSpace(cfg.BaseURL) == "" || strings.TrimSpace(cfg.APIKey) == ""
}

// RunWizard interactively configures server URL, auth, and default model.
func RunWizard(cfg *config.Config, reconfigure bool) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	output.Success("owui setup")
	if reconfigure {
		output.Info("Update your Open WebUI connection")
	} else {
		output.Info("First-time setup — connect to your Open WebUI server")
	}
	fmt.Println()

	urlDefault := strings.TrimSpace(cfg.BaseURL)
	if urlDefault == "" {
		urlDefault = "http://localhost:3000"
	}
	fmt.Printf("Open WebUI server URL [%s]: ", urlDefault)
	urlLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	url := strings.TrimSpace(urlLine)
	if url == "" {
		url = urlDefault
	}
	url = strings.TrimRight(url, "/")
	cfg.BaseURL = url

	fmt.Println()
	fmt.Println("How do you want to authenticate?")
	fmt.Println("  1) API key (recommended — create in Open WebUI → Settings → Account)")
	fmt.Println("  2) Email + password")
	fmt.Print("Choice [1]: ")
	choiceLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	choice := strings.TrimSpace(choiceLine)
	if choice == "" || choice == "1" {
		if err := promptAPIKey(reader, cfg); err != nil {
			return err
		}
	} else {
		if err := promptLogin(reader, cfg); err != nil {
			return err
		}
	}

	client := api.New(*cfg)
	fmt.Println()
	output.Info("Testing connection…")
	srv, err := client.Config()
	if err != nil {
		return fmt.Errorf("could not reach %s: %w", cfg.BaseURL, err)
	}
	output.Success(fmt.Sprintf("Connected to %s v%s", srv.Name, srv.Version))

	models, err := client.ListModels()
	if err != nil {
		output.Info("Could not list models — you can set one later with /model")
	} else if len(models) > 0 {
		fmt.Println()
		modelDefault := cfg.DefaultModel
		if modelDefault == "" {
			modelDefault = models[0].ID
		}
		fmt.Printf("Default model [%s]: ", modelDefault)
		modelLine, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		model := strings.TrimSpace(modelLine)
		if model == "" {
			model = modelDefault
		}
		cfg.DefaultModel = model
	}

	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "llama3.2"
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 300
	}
	cfg.Stream = true

	if cfg.ProfileName == "" {
		cfg.ProfileName = config.DefaultProfile
	}
	if err := config.Save(*cfg); err != nil {
		return err
	}
	path, _ := config.Path()
	fmt.Println()
	output.Success("Setup complete")
	output.Info("Config saved to " + path)
	output.Info("Change server later: owui config set url <url>")
	output.Info("Re-run setup anytime: owui setup")
	fmt.Println()
	return nil
}

func promptAPIKey(reader *bufio.Reader, cfg *config.Config) error {
	fmt.Println()
	if cfg.APIKey != "" {
		fmt.Print("API key [keep existing, press Enter]: ")
	} else {
		fmt.Print("API key: ")
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	key := strings.TrimSpace(line)
	if key != "" {
		cfg.APIKey = key
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}

func promptLogin(reader *bufio.Reader, cfg *config.Config) error {
	fmt.Println()
	fmt.Print("Email: ")
	emailLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	email := strings.TrimSpace(emailLine)
	if email == "" {
		return fmt.Errorf("email is required")
	}

	fmt.Print("Password: ")
	pw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return err
	}
	password := string(pw)
	if password == "" {
		return fmt.Errorf("password is required")
	}

	client := api.New(*cfg)
	auth, err := client.SignIn(email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	cfg.APIKey = auth.Token
	output.Success(fmt.Sprintf("Logged in as %s", auth.Name))
	return nil
}