// cmd/delete.go
package cmd

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/uidset"
)

// newDeleteCmd builds a fresh delete command per root command so flag values
// do not leak between Execute calls (see newReadCmd).
func newDeleteCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "delete <uid-set>",
		Short: "Permanently delete messages (\\Deleted + EXPUNGE, unrecoverable)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			set, err := uidset.Parse(args[0])
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "dry-run: would delete %d message(s): UIDs %s in %s\n", countUIDs(set), set.String(), folder)
				return nil
			}
			return runWithFolder(func(c *imapclient.Client, folderName string) error {
				if _, err := c.Select(folderName, nil).Wait(); err != nil {
					return fmt.Errorf("select %s: %w", folderName, err)
				}
				del := &imap.StoreFlags{Op: imap.StoreFlagsAdd, Flags: []imap.Flag{imap.FlagDeleted}, Silent: true}
				if err := c.Store(set, del, nil).Close(); err != nil {
					return fmt.Errorf("mark deleted: %w", err)
				}
				if c.Caps().Has(imap.CapUIDPlus) {
					if _, err := c.UIDExpunge(set).Collect(); err != nil {
						return fmt.Errorf("uid expunge: %w", err)
					}
				} else {
					// Expunges ALL \Deleted messages in the folder, not only `set`.
					fmt.Fprintln(cmd.ErrOrStderr(), "note: server lacks UIDPLUS; expunging every \\Deleted message in this folder")
					if _, err := c.Expunge().Collect(); err != nil {
						return fmt.Errorf("expunge: %w", err)
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "deleted %d message(s) from %s\n", countUIDs(set), folderName)
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be deleted without touching the mailbox")
	return cmd
}

func countUIDs(set imap.UIDSet) int {
	if uids, ok := set.Nums(); ok {
		return len(uids)
	}
	// Dynamic sets (containing "*") are not producible by uidset.Parse;
	// fall back to summing range widths.
	total := 0
	for _, r := range set {
		total += int(r.Stop) - int(r.Start) + 1
	}
	return total
}
