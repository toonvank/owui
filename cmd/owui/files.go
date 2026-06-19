package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/files"
	"github.com/toonvank/owui/internal/output"
)

func filesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Upload and manage RAG files",
	}

	upload := &cobra.Command{
		Use:   "upload <path>",
		Short: "Upload a file for RAG",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			resp, err := client.UploadFile(args[0])
			if err != nil {
				return err
			}

			wait, _ := cmd.Flags().GetBool("wait")
			if wait {
				if err := files.WaitForProcessing(client, resp.ID, files.DefaultWaitTimeout); err != nil {
					return err
				}
				output.Success("file processing completed")
			}

			if jsonOut {
				b, _ := json.MarshalIndent(resp, "", "  ")
				output.JSON(string(b))
				return nil
			}
			fmt.Println(resp.ID)
			return nil
		},
	}
	upload.Flags().Bool("wait", true, "wait for file processing to complete")

	status := &cobra.Command{
		Use:   "status <id>",
		Short: "Check file processing status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			st, err := client.FileStatus(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				b, _ := json.MarshalIndent(st, "", "  ")
				output.JSON(string(b))
				return nil
			}
			fmt.Println(st.Status)
			return nil
		},
	}

	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an uploaded file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}
			if err := client.DeleteFile(args[0]); err != nil {
				return err
			}
			output.Success("file deleted")
			return nil
		},
	}

	cmd.AddCommand(upload, status, del)
	return cmd
}