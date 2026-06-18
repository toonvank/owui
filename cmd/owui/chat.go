package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
	"github.com/toonvank/owui/internal/repl"
	"github.com/toonvank/owui/internal/tui"
)

func chatCmd() *cobra.Command {
	var (
		model      string
		stream     bool
		system     string
		noStream   bool
		collection string
		fileID     string
	)

	cmd := &cobra.Command{
		Use:   "chat [prompt]",
		Short: "Send a chat message or start interactive chat",
		Long:  "One-shot: owui chat \"your question\"\nInteractive: owui chat",
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

			if len(args) == 0 {
				if !quietMode {
					fmt.Fprintf(os.Stderr, "Model: %s\n", localCfg.DefaultModel)
				}
				r := repl.New(client, localCfg)
				if repl.IsInteractive() {
					return tui.Run(r)
				}
				return r.RunBasic()
			}

			messages := []api.Message{{Role: "user", Content: args[0]}}
			if localCfg.SystemPrompt != "" {
				messages = append([]api.Message{{Role: "system", Content: localCfg.SystemPrompt}}, messages...)
			}

			opts := chatOpts(collection, fileID)

			if localCfg.Stream {
				_, err = client.ChatWithOptions(messages, localCfg.DefaultModel, true, opts, func(delta string) error {
					fmt.Print(delta)
					return nil
				})
				fmt.Println()
				return err
			}

			reply, err := client.ChatWithOptions(messages, localCfg.DefaultModel, false, opts, nil)
			if err != nil {
				return err
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

	return cmd
}

func chatOpts(collection, fileID string) *api.ChatOptions {
	if collection == "" && fileID == "" {
		return nil
	}
	opts := &api.ChatOptions{Collection: collection}
	if fileID != "" {
		opts.FileIDs = []string{fileID}
	}
	return opts
}

