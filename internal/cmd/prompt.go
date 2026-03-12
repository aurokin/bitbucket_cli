package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/auro/bitbucket_cli/internal/config"
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

func promptsDisabled(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	flag := cmd.Flag("no-prompt")
	if flag == nil {
		return configDisablesPrompts()
	}

	if flag.Value.String() == "true" {
		return true
	}

	return configDisablesPrompts()
}

func promptsEnabled(cmd *cobra.Command) bool {
	if promptsDisabled(cmd) {
		return false
	}

	return isInteractiveIO(cmd.InOrStdin(), cmd.OutOrStdout())
}

func configDisablesPrompts() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}

	return !cfg.PromptEnabled()
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

func confirmExactMatch(cmd *cobra.Command, expected string) error {
	value, err := promptRequiredString(cmd, "Type "+expected+" to confirm", "")
	if err != nil {
		return err
	}
	if value != expected {
		return fmt.Errorf("confirmation did not match %s", expected)
	}
	return nil
}

func promptSecretString(cmd *cobra.Command, label string) (string, error) {
	if cmd == nil {
		return "", fmt.Errorf("prompt command is required")
	}

	stdin, ok := cmd.InOrStdin().(interface{ Fd() uintptr })
	if !ok {
		return "", fmt.Errorf("stdin does not support secure input")
	}

	for {
		if _, err := io.WriteString(cmd.OutOrStdout(), label+": "); err != nil {
			return "", err
		}

		value, err := term.ReadPassword(int(stdin.Fd()))
		if _, writeErr := io.WriteString(cmd.OutOrStdout(), "\n"); writeErr != nil {
			return "", writeErr
		}
		if err != nil {
			return "", fmt.Errorf("read prompt input: %w", err)
		}

		trimmed := strings.TrimSpace(string(value))
		if trimmed != "" {
			return trimmed, nil
		}

		if _, err := io.WriteString(cmd.OutOrStdout(), "A value is required.\n"); err != nil {
			return "", err
		}
	}
}
