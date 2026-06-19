package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/repl"
	"github.com/toonvank/owui/internal/tui"
	"golang.org/x/term"
)

type chatJSONResult struct {
	Model   string `json:"model"`
	Content string `json:"content"`
	ChatID  string `json:"chat_id,omitempty"`
}

func chatCmd() *cobra.Command {
	var (
		model      string
		stream     bool
		system     string
		noStream   bool
		collection string
		fileID     string
		chatID     string
		resumeID   string
	)

	cmd := &cobra.Command{
		Use:   "chat [prompt]",
		Short: "Send a chat message or start interactive chat",
		Long:  "One-shot: owui chat \"your question\"\nPipe:      cat log.txt | owui chat\nInteractive: owui chat",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			localCfg := cfg
			if model != "" {
				localCfg.DefaultModel = model
			}
			if cmd.Flags().Changed("stream") {
				localCfg.Stream = stream
			}
			if noStream {
				localCfg.Stream = false
			}
			if system != "" {
				localCfg.SystemPrompt = system
			}

			prompt, err := readChatPrompt(args)
			if err != nil {
				return err
			}

			resume := resumeID
			if resume == "" {
				resume = chatID
			}

			if prompt == "" {
				if !quietMode {
					fmt.Fprintf(os.Stderr, "Model: %s\n", localCfg.DefaultModel)
				}
				r := repl.New(client, localCfg)
				if resume != "" {
					if result := r.ResumeChatByID(resume); result.Err != nil {
						return result.Err
					}
				}
				if repl.IsInteractive() {
					return tui.Run(r)
				}
				return r.RunBasic()
			}

			messages, resolvedChatID := buildChatMessages(client, localCfg.SystemPrompt, resume, prompt)
			opts := chatOpts(collection, fileID, resolvedChatID)

			if localCfg.Stream {
				var reply string
				reply, err = client.ChatWithOptions(messages, localCfg.DefaultModel, true, opts, func(delta string) error {
					if jsonOut {
						return nil
					}
					fmt.Print(delta)
					return nil
				})
				if err != nil {
					return err
				}
				if jsonOut {
					return printChatJSON(localCfg.DefaultModel, reply, resolvedChatID)
				}
				fmt.Println()
				return nil
			}

			reply, err := client.ChatWithOptions(messages, localCfg.DefaultModel, false, opts, nil)
			if err != nil {
				return err
			}
			if jsonOut {
				return printChatJSON(localCfg.DefaultModel, reply, resolvedChatID)
			}
			fmt.Println(reply)
			return nil
		},
	}

	cmd.Flags().StringVarP(&model, "model", "m", "", "model to use")
	cmd.Flags().BoolVar(&stream, "stream", false, "stream response (falls back if unsupported)")
	cmd.Flags().BoolVar(&noStream, "no-stream", false, "disable streaming")
	cmd.Flags().StringVar(&system, "system", "", "system prompt override")
	cmd.Flags().StringVar(&collection, "collection", "", "knowledge collection ID for RAG")
	cmd.Flags().StringVar(&fileID, "file", "", "file ID for RAG context")
	cmd.Flags().StringVar(&chatID, "chat-id", "", "server chat ID to continue")
	cmd.Flags().StringVar(&resumeID, "resume", "", "resume server chat by ID prefix")

	return cmd
}

func readChatPrompt(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return "", nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func buildChatMessages(client *api.Client, systemPrompt, resumeID, prompt string) ([]api.Message, string) {
	messages := []api.Message{}
	resolvedChatID := ""
	if resumeID != "" {
		loaded, err := client.LoadChat(resumeID)
		if err == nil {
			resolvedChatID = loaded.ID
			if len(loaded.Messages) > 0 {
				messages = append(messages, loaded.Messages...)
			}
		}
	} else if systemPrompt != "" {
		messages = append(messages, api.Message{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, api.Message{Role: "user", Content: prompt})
	return messages, resolvedChatID
}

func chatOpts(collection, fileID, chatID string) *api.ChatOptions {
	if collection == "" && fileID == "" && chatID == "" {
		return nil
	}
	opts := &api.ChatOptions{}
	if chatID != "" {
		opts.ChatID = chatID
	}
	if collection != "" {
		opts.Collection = collection
	}
	if fileID != "" {
		opts.FileIDs = []string{fileID}
	}
	return opts
}

func printChatJSON(model, content, chatID string) error {
	out := chatJSONResult{
		Model:   model,
		Content: content,
		ChatID:  chatID,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}