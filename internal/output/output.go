// internal/output/output.go
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

type Message struct {
	UID     uint32    `json:"uid"`
	From    string    `json:"from"`
	Subject string    `json:"subject"`
	Date    time.Time `json:"date"`
	Seen    bool      `json:"seen"`
	Flagged bool      `json:"flagged"`
	Size    int64     `json:"size"`
}

type Folder struct {
	Name     string  `json:"name"`
	Delim    string  `json:"delim"`
	Messages *uint32 `json:"messages,omitempty"`
	Unseen   *uint32 `json:"unseen,omitempty"`
}

func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func PrintMessages(w io.Writer, msgs []Message) error {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "UID\tDATE\tFROM\tSUBJECT\tSIZE\tFLAGS"); err != nil {
		return err
	}
	for _, m := range msgs {
		flags := "-"
		if m.Seen && m.Flagged {
			flags = "seen,flagged"
		} else if m.Seen {
			flags = "seen"
		} else if m.Flagged {
			flags = "flagged"
		}
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%d\t%s\n",
			m.UID, m.Date.Format("2006-01-02"), m.From, m.Subject, m.Size, flags); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func PrintFolders(w io.Writer, folders []Folder) error {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tMESSAGES\tUNSEEN"); err != nil {
		return err
	}
	for _, f := range folders {
		msgs, unseen := "-", "-"
		if f.Messages != nil {
			msgs = fmt.Sprintf("%d", *f.Messages)
		}
		if f.Unseen != nil {
			unseen = fmt.Sprintf("%d", *f.Unseen)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", f.Name, msgs, unseen); err != nil {
			return err
		}
	}
	return tw.Flush()
}
