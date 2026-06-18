package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/toonvank/owui/internal/api"
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
				if err := waitForFile(client, resp.ID); err != nil {
					return err
				}
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

func waitForFile(client *api.Client, id string) error {
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		st, err := client.FileStatus(id)
		if err != nil {
			return err
		}
		switch st.Status {
		case "completed":
			output.Success("file processing completed")
			return nil
		case "failed":
			return fmt.Errorf("file processing failed: %s", st.Error)
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for file processing")
}