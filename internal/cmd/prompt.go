package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func isInteractiveIO(in io.Reader, out io.Writer) bool {
	stdin, ok := in.(interface{ Fd() uintptr })
	if !ok || !term.IsTerminal(int(stdin.Fd())) {
		return false
	}

	stdout, ok := out.(interface{ Fd() uintptr })
	if !ok || !term.IsTerminal(int(stdout.Fd())) {
		return false
	}

	return true
}

func promptRequiredString(cmd *cobra.Command, label, defaultValue string) (string, error) {
	reader := bufio.NewReader(cmd.InOrStdin())

	for {
		prompt := label
		if defaultValue != "" {
			prompt += " [" + defaultValue + "]"
		}
		prompt += ": "

		if _, err := io.WriteString(cmd.OutOrStdout(), prompt); err != nil {
			return "", err
		}

		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("read prompt input: %w", err)
		}

		value := strings.TrimSpace(line)
		if value == "" {
			value = defaultValue
		}
		if value != "" {
			return value, nil
		}

		if _, writeErr := io.WriteString(cmd.OutOrStdout(), "A value is required.\n"); writeErr != nil {
			return "", writeErr
		}

		if err == io.EOF {
			return "", fmt.Errorf("prompt input ended before a value was provided")
		}
	}
}
