package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/output"
)

func chatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chats",
		Short: "Manage chat history on the server",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List saved chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			page, _ := cmd.Flags().GetInt("page")
			chats, err := client.ListChats(page)
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(chats, "", "  ")
				output.JSON(string(b))
				return nil
			}

			rows := make([][]string, 0, len(chats))
			for _, c := range chats {
				ts := "-"
				if c.UpdatedAt > 0 {
					ts = time.Unix(c.UpdatedAt, 0).Format("2006-01-02 15:04")
				}
				title := c.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}
				rows = append(rows, []string{c.ID[:8], title, ts})
			}
			output.Table([]string{"ID", "TITLE", "UPDATED"}, rows)
			return nil
		},
	}
	list.Flags().Int("page", 1, "page number")

	show := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a chat by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			chat, err := client.GetChat(args[0])
			if err != nil {
				return err
			}

			if jsonOut {
				b, _ := json.MarshalIndent(chat, "", "  ")
				output.JSON(string(b))
				return nil
			}

			b, _ := json.MarshalIndent(chat, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	load := &cobra.Command{
		Use:   "load <id>",
		Short: "Load a chat into interactive mode context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			loaded, err := client.LoadChat(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				b, _ := json.MarshalIndent(loaded, "", "  ")
				output.JSON(string(b))
				return nil
			}
			output.Success(fmt.Sprintf("loaded %q — %d messages, model %s", loaded.Title, len(loaded.Messages), loaded.Model))
			for _, m := range loaded.Messages {
				fmt.Printf("[%s] %s\n", m.Role, truncate(m.Content, 120))
			}
			output.Info("use `owui` then /resume " + loaded.ID[:8] + " to resume in REPL")
			return nil
		},
	}

	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a chat",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			if err := client.DeleteChat(args[0]); err != nil {
				return err
			}
			output.Success("chat deleted")
			return nil
		},
	}

	cmd.AddCommand(list, load, show, del)
	return cmd
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}