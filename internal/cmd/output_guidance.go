package cmd

import (
	"fmt"
	"io"
)

func writeTargetHeader(w io.Writer, label, workspace, repo string) error {
	if label == "" || workspace == "" || repo == "" {
		return nil
	}
	_, err := fmt.Fprintf(w, "%s: %s/%s\n", label, workspace, repo)
	return err
}

func writeLabelValue(w io.Writer, label, value string) error {
	if label == "" || value == "" {
		return nil
	}
	_, err := fmt.Fprintf(w, "%s: %s\n", label, value)
	return err
}

func writeNextStep(w io.Writer, command string) error {
	if command == "" {
		return nil
	}
	_, err := fmt.Fprintf(w, "Next: %s\n", command)
	return err
}
