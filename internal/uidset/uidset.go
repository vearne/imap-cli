package uidset

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/emersion/go-imap/v2"
)

// Parse parses a user-supplied UID set like "1:500,800,1000:1200".
func Parse(s string) (imap.UIDSet, error) {
	var set imap.UIDSet
	for _, part := range strings.Split(s, ",") {
		if part == "" {
			return nil, fmt.Errorf("invalid UID set %q: empty element", s)
		}
		lo, hi, ok := strings.Cut(part, ":")
		start, err := parseUID(lo)
		if err != nil {
			return nil, fmt.Errorf("invalid UID set %q: %w", s, err)
		}
		if !ok {
			set.AddNum(start)
			continue
		}
		stop, err := parseUID(hi)
		if err != nil {
			return nil, fmt.Errorf("invalid UID set %q: %w", s, err)
		}
		if stop < start {
			return nil, fmt.Errorf("invalid UID set %q: range %d:%d is reversed", s, start, stop)
		}
		set.AddRange(start, stop)
	}
	return set, nil
}

func parseUID(s string) (imap.UID, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("%q is not a positive UID", s)
	}
	return imap.UID(n), nil
}
