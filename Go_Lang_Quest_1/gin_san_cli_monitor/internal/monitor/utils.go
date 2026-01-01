package monitor

import (
	"fmt"
	"strings"
)

// ANSI color constants.
const (
	ANSIReset  = "\x1b[0m"
	ANSIDim    = "\x1b[2m"
	ANSIBold   = "\x1b[1m"
	ANSIRed    = "\x1b[31m"
	ANSIGreen  = "\x1b[32m"
	ANSIYellow = "\x1b[33m"
	ANSIBlue   = "\x1b[34m"
	ANSICyan   = "\x1b[36m"
	ANSIGray   = "\x1b[90m"
)

// ClearScreen clears terminal and moves cursor to top-left.
func ClearScreen() { fmt.Print("\x1b[2J\x1b[H") }

// PadRight pads s to width w (simple table alignment).
func PadRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// ParseURLs splits a comma-separated list into trimmed urls.
func ParseURLs(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		u := strings.TrimSpace(p)
		if u == "" {
			continue
		}
		out = append(out, u)
	}
	return out
}
