// cmd/mark.go
package cmd

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/uidset"
)

// newMarkCmd builds a fresh mark command per root command so flag values do
// not leak between Execute calls (see newReadCmd).
func newMarkCmd() *cobra.Command {
	markCmd := &cobra.Command{
		Use:   "mark <uid-set> read|unread|flagged|unflagged",
		Short: "Set or clear message flags",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			set, err := uidset.Parse(args[0])
			if err != nil {
				return err
			}
			var op imap.StoreFlagsOp
			var flags []imap.Flag
			switch args[1] {
			case "read":
				op, flags = imap.StoreFlagsAdd, []imap.Flag{imap.FlagSeen}
			case "unread":
				op, flags = imap.StoreFlagsDel, []imap.Flag{imap.FlagSeen}
			case "flagged":
				op, flags = imap.StoreFlagsAdd, []imap.Flag{imap.FlagFlagged}
			case "unflagged":
				op, flags = imap.StoreFlagsDel, []imap.Flag{imap.FlagFlagged}
			default:
				return fmt.Errorf("unknown state %q: use read|unread|flagged|unflagged", args[1])
			}
			return runWithFolder(func(c *imapclient.Client, folder string) error {
				if _, err := c.Select(folder, nil).Wait(); err != nil {
					return fmt.Errorf("select %s: %w", folder, err)
				}
				storeCmd := c.Store(set, &imap.StoreFlags{Op: op, Flags: flags, Silent: true}, nil)
				return storeCmd.Close() // Store returns FetchCommand; must be consumed
			})
		},
	}
	return markCmd
}
