// cmd/read_test.go
package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vearne/imap-cli/internal/testenv"
)

const multipartRaw = "From: a@x.com\r\nTo: b@y.com\r\nSubject: report\r\n" +
	"Content-Type: multipart/mixed; boundary=XYZ\r\n\r\n" +
	"--XYZ\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nhello body\r\n" +
	"--XYZ\r\nContent-Type: application/pdf; name=\"r.pdf\"\r\n" +
	"Content-Disposition: attachment; filename=\"r.pdf\"\r\n\r\n%PDF-data\r\n" +
	"--XYZ--\r\n"

func readJSON(t *testing.T, args ...string) ReadResult {
	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"read", "--json"}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("read %v: %v", args, err)
	}
	var res ReadResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("bad json %q: %v", buf.String(), err)
	}
	return res
}

func TestReadTextBodyAndAttachment(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", multipartRaw, nil, time.Now())

	res := readJSON(t, "1") // UID 1: fresh mailbox
	if res.TextBody != "hello body" {
		t.Fatalf("text body = %q", res.TextBody)
	}
	if len(res.Attachments) != 1 || res.Attachments[0].Filename != "r.pdf" || res.Attachments[0].Size == 0 {
		t.Fatalf("attachments = %+v", res.Attachments)
	}
}

func TestReadHTMLFallback(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	html := "From: a@x.com\r\nSubject: html only\r\nContent-Type: text/html; charset=utf-8\r\n\r\n<p>hi</p>\r\n"
	srv.Append(t, "INBOX", html, nil, time.Now())

	res := readJSON(t, "1")
	if res.TextBody != "" || !strings.Contains(res.HTMLBody, "<p>hi</p>") {
		t.Fatalf("html fallback wrong: %+v", res)
	}
}

func TestReadRaw(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	srv.Append(t, "INBOX", multipartRaw, nil, time.Now())

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"read", "1", "--raw"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "boundary=XYZ") {
		t.Fatalf("raw output = %q", buf.String())
	}
}

func TestReadBadUID(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"read", "zz"})
	if err := root.Execute(); err == nil {
		t.Fatal("read zz should fail")
	}
}

func TestReadRejectsUIDRange(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	raw := "From: a@x.com\r\nSubject: m\r\n\r\nbody\r\n"
	for i := 0; i < 5; i++ {
		srv.Append(t, "INBOX", raw, nil, time.Now())
	}

	root := NewRootCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"read", "1:5"})
	if err := root.Execute(); err == nil {
		t.Fatal("read 1:5 should fail instead of truncating to the first message")
	}
}

func TestReadGBKCharset(t *testing.T) {
	srv := testenv.Start(t)
	setEnvAddr(t, srv)
	// GBK bytes for 你好
	raw := "From: a@x.com\r\nSubject: gbk\r\nContent-Type: text/plain; charset=gbk\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n\r\n\xc4\xe3\xba\xc3\r\n"
	srv.Append(t, "INBOX", raw, nil, time.Now())

	res := readJSON(t, "1")
	if res.TextBody != "你好" {
		t.Fatalf("gbk body = %q, want 你好", res.TextBody)
	}
}
