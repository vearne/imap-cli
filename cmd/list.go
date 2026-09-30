// cmd/list.go
package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/output"
)

var (
	listFrom    string
	listSubject string
	listSince   string
	listBefore  string
	listUnread  bool
	listLimit   int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Search and list messages (envelope summary)",
	RunE: func(cmd *cobra.Command, args []string) error {
		crit := &imap.SearchCriteria{}
		if listFrom != "" {
			crit.Header = append(crit.Header, imap.SearchCriteriaHeaderField{Key: "From", Value: listFrom})
		}
		if listSubject != "" {
			crit.Header = append(crit.Header, imap.SearchCriteriaHeaderField{Key: "Subject", Value: listSubject})
		}
		if listSince != "" {
			tm, err := parseDate(listSince)
			if err != nil {
				return err
			}
			crit.SentSince = tm
		}
		if listBefore != "" {
			tm, err := parseDate(listBefore)
			if err != nil {
				return err
			}
			crit.SentBefore = tm
		}
		if listUnread {
			crit.NotFlag = append(crit.NotFlag, imap.FlagSeen)
		}

		return runWithFolder(func(c *imapclient.Client, folder string) error {
			// SEARCH runs against the selected mailbox — SELECT must come first.
			if _, err := c.Select(folder, nil).Wait(); err != nil {
				return fmt.Errorf("select %s: %w", folder, err)
			}
			uids, err := searchUIDs(c, crit)
			if err != nil {
				return err
			}
			msgs, err := fetchSummaries(c, uids)
			if err != nil {
				return err
			}
			sort.Slice(msgs, func(i, j int) bool { return msgs[i].Date.After(msgs[j].Date) })
			if listLimit > 0 && len(msgs) > listLimit {
				msgs = msgs[:listLimit]
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return output.WriteJSON(out, msgs)
			}
			return output.PrintMessages(out, msgs)
		})
	},
}

func parseDate(s string) (time.Time, error) {
	tm, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD", s)
	}
	return tm, nil
}

func searchUIDs(c *imapclient.Client, crit *imap.SearchCriteria) ([]imap.UID, error) {
	data, err := c.UIDSearch(crit, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return data.AllUIDs(), nil
}

// fetchSummaries fetches envelopes for uids; the mailbox is already selected.
func fetchSummaries(c *imapclient.Client, uids []imap.UID) ([]output.Message, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	// Pass by value: imapwire.NumSetKind panics on *UIDSet (matches value type only).
	set := imap.UIDSet{}
	for _, uid := range uids {
		set.AddNum(uid)
	}
	bufs, err := c.Fetch(set, &imap.FetchOptions{
		UID: true, Envelope: true, Flags: true, InternalDate: true, RFC822Size: true,
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	msgs := make([]output.Message, 0, len(bufs))
	for _, b := range bufs {
		m := output.Message{UID: uint32(b.UID), Size: b.RFC822Size}
		if b.Envelope != nil {
			m.Subject = b.Envelope.Subject
			m.Date = b.Envelope.Date
			if len(b.Envelope.From) > 0 {
				m.From = b.Envelope.From[0].Addr()
			}
		}
		if m.Date.IsZero() {
			m.Date = b.InternalDate
		}
		for _, fl := range b.Flags {
			switch fl {
			case imap.FlagSeen:
				m.Seen = true
			case imap.FlagFlagged:
				m.Flagged = true
			}
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func init() {
	listCmd.Flags().StringVar(&listFrom, "from", "", "From header contains")
	listCmd.Flags().StringVar(&listSubject, "subject", "", "Subject header contains")
	listCmd.Flags().StringVar(&listSince, "since", "", "sent on or after (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listBefore, "before", "", "sent before (YYYY-MM-DD)")
	listCmd.Flags().BoolVar(&listUnread, "unread", false, "only unseen messages")
	listCmd.Flags().IntVar(&listLimit, "limit", 50, "max results after sorting (0 = unlimited)")
}
