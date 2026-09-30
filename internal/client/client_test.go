// internal/client/client_test.go
package client

import (
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/vearne/imap-cli/internal/config"
	"github.com/vearne/imap-cli/internal/testenv"
)

func TestRunExecutesFnAndCleansUp(t *testing.T) {
	srv := testenv.Start(t)
	host, port := mustHostPort(srv.Addr)
	cfg := &config.Config{Host: host, Port: port, Username: "testuser", Password: "testpass", TLSMode: config.TLSNone}
	called := false
	err := Run(cfg, func(c *imapclient.Client) error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("Run() = %v, called=%v", err, called)
	}
}

func TestRunBadPassword(t *testing.T) {
	srv := testenv.Start(t)
	host, port := mustHostPort(srv.Addr)
	cfg := &config.Config{Host: host, Port: port, Username: "testuser", Password: "WRONG", TLSMode: config.TLSNone}
	err := Run(cfg, func(c *imapclient.Client) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "testuser") {
		t.Fatalf("Run() err = %v, want auth failure mentioning username testuser", err)
	}
}

func mustHostPort(addr string) (string, int) {
	h, p, _ := net.SplitHostPort(addr)
	n, _ := strconv.Atoi(p)
	return h, n
}
