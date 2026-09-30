// internal/testenv/testenv.go
package testenv

import (
	"net"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"
)

// Server is an in-memory IMAP server for integration tests.
type Server struct {
	Addr string
	User *imapmemserver.User
}

func Start(t *testing.T, caps ...imap.Cap) *Server {
	t.Helper()
	if len(caps) == 0 {
		caps = []imap.Cap{imap.CapIMAP4rev1, imap.CapIMAP4rev2, imap.CapUIDPlus, imap.CapMove}
	}
	capSet := imap.CapSet{}
	for _, c := range caps {
		capSet[c] = struct{}{}
	}
	memSrv := imapmemserver.New()
	user := imapmemserver.NewUser("testuser", "testpass")
	if err := user.Create("INBOX", nil); err != nil {
		t.Fatalf("create INBOX: %v", err)
	}
	memSrv.AddUser(user)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := imapserver.New(&imapserver.Options{
		NewSession: func(conn *imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			return memSrv.NewSession(), nil, nil
		},
		Caps:         capSet,
		InsecureAuth: true,
	})
	go srv.Serve(ln)                 //nolint:errcheck — test server
	t.Cleanup(func() { ln.Close() }) //nolint:errcheck
	return &Server{Addr: ln.Addr().String(), User: user}
}

func (s *Server) Client(t *testing.T) *imapclient.Client {
	t.Helper()
	c, err := imapclient.DialInsecure(s.Addr, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if err := c.Login("testuser", "testpass").Wait(); err != nil {
		t.Fatalf("login: %v", err)
	}
	t.Cleanup(func() { c.Logout().Wait() }) //nolint:errcheck
	return c
}

// Append seeds one message and returns its UID.
func (s *Server) Append(t *testing.T, mbox string, raw string, flags []imap.Flag, when time.Time) imap.UID {
	t.Helper()
	c := s.Client(t)
	defer func() { c.Logout().Wait() }() //nolint:errcheck
	cmd := c.Append(mbox, int64(len(raw)), &imap.AppendOptions{Flags: flags, Time: when})
	if _, err := cmd.Write([]byte(raw)); err != nil {
		t.Fatalf("append write: %v", err)
	}
	if err := cmd.Close(); err != nil {
		t.Fatalf("append close: %v", err)
	}
	data, err := cmd.Wait()
	if err != nil {
		t.Fatalf("append wait: %v", err)
	}
	if data == nil || data.UID == 0 {
		t.Fatal("append returned no UID (server must advertise UIDPLUS)")
	}
	return data.UID
}
