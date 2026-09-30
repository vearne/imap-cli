// internal/output/output_test.go
package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintMessages(t *testing.T) {
	var buf bytes.Buffer
	msgs := []Message{{
		UID: 42, From: "a@x.com", Subject: "Hello 世界",
		Date: time.Date(2023, 5, 1, 10, 0, 0, 0, time.UTC), Seen: true, Size: 1234,
	}}
	if err := PrintMessages(&buf, msgs); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"42", "a@x.com", "Hello 世界", "2023-05-01"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output %q missing %q", out, want)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	err := WriteJSON(&buf, []Message{{UID: 1, From: "a@x.com", Subject: "s", Size: 10}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"uid": 1`) {
		t.Fatalf("json = %q", buf.String())
	}
}

func TestPrintFolders(t *testing.T) {
	var buf bytes.Buffer
	unseen := uint32(3)
	err := PrintFolders(&buf, []Folder{{Name: "INBOX", Delim: "/", Unseen: &unseen}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "INBOX") || !strings.Contains(buf.String(), "3") {
		t.Fatalf("folders output = %q", buf.String())
	}
}
