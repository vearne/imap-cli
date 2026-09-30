// cmd/list_test.go
package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/vearne/imap-cli/internal/output"
	"github.com/vearne/imap-cli/internal/testenv"
)

func seedMailbox(t *testing.T) *testenv.Server {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	old := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	// Date: headers are required: IMAP SENTSINCE/SENTBEFORE filter on the
	// message Date header, not InternalDate.
	srv.Append(t, "INBOX", "From: news@list.example\r\nSubject: old promo\r\nDate: Sat, 01 Jan 2022 00:00:00 +0000\r\n\r\nhi", nil, old)
	srv.Append(t, "INBOX", "From: boss@example.com\r\nSubject: fresh report\r\nDate: Sat, 01 Jun 2024 00:00:00 +0000\r\n\r\nhi", nil, newer)
	srv.Append(t, "INBOX", "From: boss@example.com\r\nSubject: unseen one\r\nDate: Sat, 01 Jun 2024 00:00:00 +0000\r\n\r\nhi", []imap.Flag{imap.FlagSeen}, newer)
	return srv
}

// resetListFlags clears package-global flag state; cobra only overwrites
// globals for flags present on the command line, so values leak across tests.
func resetListFlags() {
	listFrom, listSubject, listSince, listBefore = "", "", "", ""
	listUnread, listLimit = false, 50
}

func runList(t *testing.T, args ...string) []output.Message {
	resetListFlags()
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"list", "--json"}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("list %v: %v", args, err)
	}
	var msgs []output.Message
	if err := json.Unmarshal(buf.Bytes(), &msgs); err != nil {
		t.Fatalf("bad json %q: %v", buf.String(), err)
	}
	return msgs
}

func TestListNewestFirst(t *testing.T) {
	seedMailbox(t)
	msgs := runList(t)
	if len(msgs) != 3 || msgs[0].Subject != "fresh report" || msgs[2].Subject != "old promo" {
		t.Fatalf("order wrong: %+v", msgs)
	}
}

func TestListBeforeFilter(t *testing.T) {
	seedMailbox(t)
	msgs := runList(t, "--before", "2023-01-01")
	if len(msgs) != 1 || msgs[0].Subject != "old promo" {
		t.Fatalf("before filter wrong: %+v", msgs)
	}
}

func TestListFromAndUnread(t *testing.T) {
	seedMailbox(t)
	msgs := runList(t, "--from", "boss@example.com", "--unread")
	if len(msgs) != 1 || msgs[0].Subject != "fresh report" {
		t.Fatalf("from+unread wrong: %+v", msgs)
	}
}

func TestListLimit(t *testing.T) {
	seedMailbox(t)
	if msgs := runList(t, "--limit", "1"); len(msgs) != 1 {
		t.Fatalf("limit wrong: %+v", msgs)
	}
}

func TestListFlagParsing(t *testing.T) {
	seedMailbox(t)
	resetListFlags()
	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"list", "--since", "not-a-date"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "YYYY-MM-DD") {
		t.Fatalf("bad date err = %v", err)
	}
}
