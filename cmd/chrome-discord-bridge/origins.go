package main

import (
	_ "embed"
	"strings"
)

// originsDelimited is the newline-delimited set of URLs that are allowed to
// call the native messaging host.  It is initialized from the contents of
// "origins.txt".
//go:embed origins.txt
var originsDelimited string

// origins is a set of URLs that are allowed to call the native messaging host.
var origins = make(map[string]struct{})

// IsValidOrigin returns true if s matches a line in origins.txt.
func IsValidOrigin(s string) bool {
	_, ok := origins[s]
	return ok
}

func init() {
	for _, s := range strings.Split(originsDelimited, "\n") {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			continue
		}
		origins[trimmed] = struct{}{}
	}
}
