// cmd/root.go
package cmd

import (
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/client"
	"github.com/vearne/imap-cli/internal/config"
)

var jsonOut bool

// NewRootCmd builds a fresh root command with all subcommands registered.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "imap-cli",
		Short: "IMAP mailbox operations for humans and LLMs",
		Long: `imap-cli — list, read, search and batch-manage IMAP mail.

Credentials come from environment variables:
  IMAP_HOST       e.g. imap.gmail.com        (required)
  IMAP_PORT       default 993
  IMAP_USERNAME                              (required)
  IMAP_PASSWORD   Gmail: app password; QQ/163: authorization code (required)
  IMAP_TLS        tls (default) | starttls | none

Use --json for machine-readable output.`,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output JSON instead of text")
	rootCmd.AddCommand(foldersCmd)
	return rootCmd
}

func Execute() error { return NewRootCmd().Execute() }

// runWithConfig loads env config and runs fn with a logged-in client.
func runWithConfig(fn func(c *imapclient.Client) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return client.Run(&cfg, fn)
}
