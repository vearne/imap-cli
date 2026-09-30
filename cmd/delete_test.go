// cmd/delete_test.go
package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/vearne/imap-cli/internal/testenv"
)

func TestDeleteDryRunTouchesNothing(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", "Subject: a\r\n\r\nb", nil, time.Now())
	srv.Append(t, "INBOX", "Subject: b\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"delete", "1:2", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "2") || !strings.Contains(buf.String(), "1:2") {
		t.Fatalf("dry-run output = %q", buf.String())
	}
	if n := countIn(t, srv, "INBOX"); n != 2 {
		t.Fatalf("dry-run deleted mail! INBOX has %d", n)
	}
}

func TestDeleteWithUIDExpunge(t *testing.T) {
	srv := testenv.Start(t) // caps include UIDPLUS
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", "Subject: a\r\n\r\nb", nil, time.Now())
	srv.Append(t, "INBOX", "Subject: b\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"delete", "1"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if n := countIn(t, srv, "INBOX"); n != 1 {
		t.Fatalf("INBOX has %d after delete, want 1", n)
	}
}

func TestDeleteWithoutUIDPLUS(t *testing.T) {
	srv := testenv.Start(t, imap.CapIMAP4rev1, imap.CapIMAP4rev2, imap.CapMove) // no UIDPLUS
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", "Subject: a\r\n\r\nb", nil, time.Now())
	srv.Append(t, "INBOX", "Subject: b\r\n\r\nb", nil, time.Now())

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"delete", "1"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if n := countIn(t, srv, "INBOX"); n != 1 {
		t.Fatalf("INBOX has %d after delete, want 1", n)
	}
}
