package uidset

import (
	"testing"

	"github.com/emersion/go-imap/v2"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in      string
		want    imap.UIDSet
		wantStr string
	}{
		{"7", imap.UIDSetNum(7), "7"},
		{"1:5", func() imap.UIDSet { s := &imap.UIDSet{}; s.AddRange(1, 5); return *s }(), "1:5"},
		{"1:500,800,1000:1200", func() imap.UIDSet {
			s := &imap.UIDSet{}
			s.AddRange(1, 500)
			s.AddNum(800)
			s.AddRange(1000, 1200)
			return *s
		}(), "1:500,800,1000:1200"},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Fatalf("Parse(%q) = %v", tc.in, err)
		}
		if got.String() != tc.wantStr {
			t.Fatalf("Parse(%q).String() = %q, want %q", tc.in, got.String(), tc.wantStr)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "abc", "0", "5:1", "1:2,", "1,,2", "-3", "1:"} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) should fail", in)
		}
	}
}
