package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/output"
)

func knowledgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "knowledge",
		Aliases: []string{"kb"},
		Short:   "Manage knowledge collections",
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List knowledge collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			items, err := client.ListKnowledge()
			if err != nil {
				return err
			}
			if jsonOut {
				b, _ := json.MarshalIndent(items, "", "  ")
				output.JSON(string(b))
				return nil
			}
			rows := make([][]string, 0, len(items))
			for _, k := range items {
				desc := k.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				rows = append(rows, []string{k.ID[:8], k.Name, desc})
			}
			output.Table([]string{"ID", "NAME", "DESCRIPTION"}, rows)
			return nil
		},
	}

	create := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a knowledge collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			desc, _ := cmd.Flags().GetString("description")
			item, err := client.CreateKnowledge(args[0], desc)
			if err != nil {
				return err
			}
			if jsonOut {
				b, _ := json.MarshalIndent(item, "", "  ")
				output.JSON(string(b))
				return nil
			}
			output.Success("created " + item.ID)
			return nil
		},
	}
	create.Flags().String("description", "", "collection description")

	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a knowledge collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			if err := client.DeleteKnowledge(args[0]); err != nil {
				return err
			}
			output.Success("knowledge collection deleted")
			return nil
		},
	}

	addFile := &cobra.Command{
		Use:   "add-file <collection-id> <file-id>",
		Short: "Add a file to a knowledge collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			if err := client.AddFileToKnowledge(args[0], args[1]); err != nil {
				return err
			}
			output.Success("file added to collection")
			return nil
		},
	}

	cmd.AddCommand(list, create, del, addFile)
	return cmd
}