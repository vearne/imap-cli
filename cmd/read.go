// cmd/read.go
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/spf13/cobra"
	"github.com/vearne/imap-cli/internal/output"
	"github.com/vearne/imap-cli/internal/uidset"

	// Registers all x/text charsets (incl. GBK) with go-message.
	_ "github.com/emersion/go-message/charset"
)

type attachment struct {
	Filename string `json:"filename"`
	Size     int    `json:"size"`
}

type ReadResult struct {
	UID         uint32       `json:"uid"`
	From        string       `json:"from"`
	To          string       `json:"to"`
	Cc          string       `json:"cc"`
	Subject     string       `json:"subject"`
	Date        string       `json:"date"` // RFC1123Z formatted
	TextBody    string       `json:"text_body"`
	HTMLBody    string       `json:"html_body,omitempty"` // only when no text/plain part exists
	Attachments []attachment `json:"attachments"`
}

// newReadCmd builds a fresh read command per root command so flag values do
// not leak between Execute calls (pflag keeps parsed values across parses).
func newReadCmd() *cobra.Command {
	readCmd := &cobra.Command{
		Use:   "read <uid>",
		Short: "Read one message: headers, body, attachment list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			readRaw, _ := cmd.Flags().GetBool("raw")
			set, err := uidset.Parse(args[0])
			if err != nil {
				return err
			}
			if len(set) != 1 {
				return fmt.Errorf("read takes exactly one UID, got %q", args[0])
			}
			return runWithFolder(func(c *imapclient.Client, folder string) error {
				if _, err := c.Select(folder, nil).Wait(); err != nil {
					return fmt.Errorf("select %s: %w", folder, err)
				}
				sec := imap.FetchItemBodySection{Peek: true}
				bufs, err := c.Fetch(set, &imap.FetchOptions{UID: true, Envelope: true, BodySection: []*imap.FetchItemBodySection{&sec}}).Collect()
				if err != nil {
					return fmt.Errorf("fetch: %w", err)
				}
				if len(bufs) == 0 {
					return fmt.Errorf("UID %s not found in %s (hint: run `list` first; UIDs are stable but mailboxes can be recreated)", args[0], folder)
				}
				b := bufs[0]
				raw := b.FindBodySection(&sec)

				if readRaw {
					_, err := cmd.OutOrStdout().Write(raw)
					return err
				}

				res := ReadResult{UID: uint32(b.UID)}
				if b.Envelope != nil {
					res.Subject = b.Envelope.Subject
					if len(b.Envelope.From) > 0 {
						res.From = b.Envelope.From[0].Addr()
					}
					if !b.Envelope.Date.IsZero() {
						res.Date = b.Envelope.Date.Format(time.RFC1123Z)
					}
					res.To = joinAddrs(b.Envelope.To)
					res.Cc = joinAddrs(b.Envelope.Cc)
				}
				if err := walkBody(raw, &res); err != nil {
					return err
				}

				out := cmd.OutOrStdout()
				if jsonOut {
					return output.WriteJSON(out, res)
				}
				return printReadText(out, res)
			})
		},
	}
	readCmd.Flags().Bool("raw", false, "print the original RFC822 message instead")
	return readCmd
}

func joinAddrs(addrs []imap.Address) string {
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		parts = append(parts, a.Addr())
	}
	return strings.Join(parts, ", ")
}

func walkBody(raw []byte, res *ReadResult) error {
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("parse message: %w", err)
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("walk mime: %w", err)
		}
		// PartHeader impls have pointer receivers: match on pointer types.
		switch h := part.Header.(type) {
		case *mail.AttachmentHeader:
			name, _ := h.Filename()
			size := 0
			if n, err := io.Copy(io.Discard, part.Body); err == nil {
				size = int(n)
			}
			res.Attachments = append(res.Attachments, attachment{Filename: name, Size: size})
		case *mail.InlineHeader:
			mediaType, _, err := h.ContentType()
			if err != nil {
				mediaType = "text/plain"
			}
			body, readErr := io.ReadAll(part.Body)
			if readErr != nil {
				return fmt.Errorf("read part: %w", readErr)
			}
			switch mediaType {
			case "text/plain":
				if res.TextBody == "" {
					res.TextBody = strings.TrimRight(string(body), "\r\n")
				}
			case "text/html":
				if res.HTMLBody == "" {
					res.HTMLBody = strings.TrimRight(string(body), "\r\n")
				}
			}
		}
	}
	if res.TextBody != "" {
		res.HTMLBody = "" // html is only a fallback when no text/plain part exists
	}
	return nil
}

func printReadText(w io.Writer, res ReadResult) error {
	fmt.Fprintf(w, "From:    %s\n", res.From)
	fmt.Fprintf(w, "To:      %s\n", res.To)
	if res.Cc != "" {
		fmt.Fprintf(w, "Cc:      %s\n", res.Cc)
	}
	fmt.Fprintf(w, "Subject: %s\n", res.Subject)
	fmt.Fprintf(w, "Date:    %s\n", res.Date)
	fmt.Fprintf(w, "\n%s\n", res.TextBody)
	if res.TextBody == "" && res.HTMLBody != "" {
		fmt.Fprintf(w, "[html]\n%s\n", res.HTMLBody)
	}
	if len(res.Attachments) > 0 {
		fmt.Fprintln(w, "\nAttachments:")
		for _, a := range res.Attachments {
			fmt.Fprintf(w, "  %s (%d bytes)\n", a.Filename, a.Size)
		}
	}
	return nil
}
