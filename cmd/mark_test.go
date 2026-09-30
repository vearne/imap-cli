// cmd/mark_test.go
package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/vearne/imap-cli/internal/testenv"
)

func flagsOf(t *testing.T, srv *testenv.Server, uid imap.UID) []string {
	// Re-fetch the message flags through a fresh client SELECT + FETCH via library call.
	c := srv.Client(t)
	defer func() { c.Logout().Wait() }() //nolint:errcheck
	if _, err := c.Select("INBOX", nil).Wait(); err != nil {
		t.Fatal(err)
	}
	var set imap.UIDSet
	set.AddNum(uid)
	bufs, err := c.Fetch(set, &imap.FetchOptions{UID: true, Flags: true}).Collect()
	if err != nil || len(bufs) == 0 {
		t.Fatalf("fetch flags: %v", err)
	}
	out := make([]string, 0, len(bufs[0].Flags))
	for _, f := range bufs[0].Flags {
		out = append(out, string(f))
	}
	return out
}

func has(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestMarkReadAndUnread(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	uid := srv.Append(t, "INBOX", "Subject: m\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"mark", "1", "read"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if fl := flagsOf(t, srv, uid); !has(fl, "\\Seen") {
		t.Fatalf("flags after mark read: %v", fl)
	}

	root = NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"mark", "1", "unread"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if fl := flagsOf(t, srv, uid); has(fl, "\\Seen") {
		t.Fatalf("flags after mark unread: %v", fl)
	}
}

func TestMarkFlaggedAndInvalidState(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", "Subject: m\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"mark", "1", "flagged"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	root = NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"mark", "1", "archive"}) // invalid
	if err := root.Execute(); err == nil {
		t.Fatal("invalid state should fail")
	}
}
