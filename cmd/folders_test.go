// cmd/folders_test.go
package cmd

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/vearne/imap-cli/internal/testenv"
)

// setEnvAddr points the CLI at the in-memory test server via env vars.
func setEnvAddr(t *testing.T, srv *testenv.Server) {
	t.Helper()
	host, port, err := net.SplitHostPort(srv.Addr)
	if err != nil {
		t.Fatalf("split %q: %v", srv.Addr, err)
	}
	t.Setenv("IMAP_HOST", host)
	t.Setenv("IMAP_PORT", port)
	t.Setenv("IMAP_USERNAME", "testuser")
	t.Setenv("IMAP_PASSWORD", "testpass")
	t.Setenv("IMAP_TLS", "none")
}

func TestFoldersJSON(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"folders", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("folders: %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "INBOX"`) {
		t.Fatalf("json output = %q", buf.String())
	}
}

func TestFoldersCounts(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", "Subject: x\r\n\r\nbody", nil, time.Now())

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"folders", "--counts"})
	if err := root.Execute(); err != nil {
		t.Fatalf("folders --counts: %v", err)
	}
	if !strings.Contains(buf.String(), "INBOX") || !strings.Contains(buf.String(), "1") {
		t.Fatalf("counts output = %q", buf.String())
	}
}
