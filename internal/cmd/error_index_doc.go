package cmd

import (
	"fmt"
	"strings"
)

func GenerateErrorIndexDoc() (string, error) {
	var b strings.Builder
	b.WriteString("# Error Index\n\n")
	b.WriteString("Generated from the current recovery guidance catalog.\n\n")
	b.WriteString("Use this file when you have an error fragment and want the fastest likely next command.\n\n")
	b.WriteString("| Error fragment | Recovery focus | First next command |\n")
	b.WriteString("|---|---|---|\n")

	for _, entry := range recoveryDocEntries() {
		first := ""
		if len(entry.Recovery) > 0 {
			first = fmt.Sprintf("`%s`", entry.Recovery[0])
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", singleLine(entry.TypicalFailure), entry.Title, first)
	}

	return b.String(), nil
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
