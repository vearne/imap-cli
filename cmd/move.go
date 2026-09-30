// cmd/move.go
package cmd

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/uidset"
)

// newMoveCmd builds a fresh move command per root command so flag values do
// not leak between Execute calls (see newReadCmd).
func newMoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move <uid-set> <dest-folder>",
		Short: "Move messages to another mailbox",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			set, err := uidset.Parse(args[0])
			if err != nil {
				return err
			}
			return runWithFolder(func(c *imapclient.Client, folder string) error {
				if _, err := c.Select(folder, nil).Wait(); err != nil {
					return fmt.Errorf("select %s: %w", folder, err)
				}
				return moveUIDs(c, folder, args[1], set)
			})
		},
	}
}

func moveUIDs(c *imapclient.Client, folder, dest string, set imap.UIDSet) error {
	if len(set) == 0 {
		return fmt.Errorf("empty UID set")
	}
	if c.Caps().Has(imap.CapMove) {
		if _, err := c.Move(set, dest).Wait(); err != nil {
			return fmt.Errorf("move to %s: %w", dest, err)
		}
		return nil
	}
	// Fallback: COPY + \Deleted + EXPUNGE. Note: EXPUNGE removes ALL \Deleted
	// messages in the folder, which is safe in this stateless CLI (single command
	// owns the whole session) but worth knowing.
	if _, err := c.Copy(set, dest).Wait(); err != nil {
		return fmt.Errorf("copy to %s: %w", dest, err)
	}
	del := &imap.StoreFlags{Op: imap.StoreFlagsAdd, Flags: []imap.Flag{imap.FlagDeleted}, Silent: true}
	if err := c.Store(set, del, nil).Close(); err != nil {
		return fmt.Errorf("mark deleted: %w", err)
	}
	if _, err := c.Expunge().Collect(); err != nil {
		return fmt.Errorf("expunge: %w", err)
	}
	return nil
}
