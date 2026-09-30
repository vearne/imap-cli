// cmd/move_test.go
package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/vearne/imap-cli/internal/testenv"
)

func countIn(t *testing.T, srv *testenv.Server, mbox string) int {
	t.Helper()
	c := srv.Client(t)
	defer func() { c.Logout().Wait() }() //nolint:errcheck
	data, err := c.Status(mbox, &imap.StatusOptions{NumMessages: true}).Wait()
	if err != nil {
		t.Fatalf("status %s: %v", mbox, err)
	}
	if data.NumMessages == nil {
		t.Fatalf("status %s returned no message count", mbox)
	}
	return int(*data.NumMessages)
}

func TestMoveWithNativeMOVE(t *testing.T) {
	srv := testenv.Start(t) // default caps include MOVE
	setEnvAddr(t, srv)
	if err := srv.User.Create("Archive", nil); err != nil {
		t.Fatal(err)
	}
	srv.Append(t, "INBOX", "Subject: m\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"move", "1", "Archive"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if n := countIn(t, srv, "INBOX"); n != 0 {
		t.Fatalf("INBOX has %d messages after move, want 0", n)
	}
	if n := countIn(t, srv, "Archive"); n != 1 {
		t.Fatalf("Archive has %d messages after move, want 1", n)
	}
}

func TestMoveFallbackWithoutMOVECap(t *testing.T) {
	srv := testenv.Start(t, imap.CapIMAP4rev1, imap.CapIMAP4rev2, imap.CapUIDPlus) // no CapMove
	setEnvAddr(t, srv)
	if err := srv.User.Create("Archive", nil); err != nil {
		t.Fatal(err)
	}
	srv.Append(t, "INBOX", "Subject: m\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"move", "1", "Archive"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if n := countIn(t, srv, "INBOX"); n != 0 {
		t.Fatalf("INBOX has %d messages after fallback move, want 0", n)
	}
	if n := countIn(t, srv, "Archive"); n != 1 {
		t.Fatalf("Archive has %d messages after fallback move, want 1", n)
	}
}
