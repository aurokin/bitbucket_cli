package output

import (
	"io"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

func NewTableWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
}

func TerminalWidth(w io.Writer) int {
	fd, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return 0
	}
	if !term.IsTerminal(int(fd.Fd())) {
		return 0
	}

	width, _, err := term.GetSize(int(fd.Fd()))
	if err != nil || width <= 0 {
		return 0
	}

	return width
}

func Truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	if max == 1 {
		return "…"
	}

	return string(runes[:max-1]) + "…"
}

func TruncateMiddle(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	if max == 1 {
		return "…"
	}

	left := (max - 1) / 2
	right := max - 1 - left
	return string(runes[:left]) + "…" + string(runes[len(runes)-right:])
}
