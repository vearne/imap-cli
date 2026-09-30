// cmd/folders.go
package cmd

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/output"
)

var folderCounts bool

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "List all mailboxes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithConfig(func(c *imapclient.Client) error {
			list, err := c.List("", "*", nil).Collect()
			if err != nil {
				return fmt.Errorf("list folders: %w", err)
			}
			folders := make([]output.Folder, 0, len(list))
			for _, ld := range list {
				f := output.Folder{Name: ld.Mailbox, Delim: string(ld.Delim)}
				if folderCounts {
					status, err := c.Status(ld.Mailbox, &imap.StatusOptions{NumMessages: true, NumUnseen: true}).Wait()
					if err != nil {
						return fmt.Errorf("status %s: %w", ld.Mailbox, err)
					}
					f.Messages, f.Unseen = status.NumMessages, status.NumUnseen
				}
				folders = append(folders, f)
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return output.WriteJSON(out, folders)
			}
			return output.PrintFolders(out, folders)
		})
	},
}

func init() {
	foldersCmd.Flags().BoolVar(&folderCounts, "counts", false, "also query message and unread counts (slower)")
}
