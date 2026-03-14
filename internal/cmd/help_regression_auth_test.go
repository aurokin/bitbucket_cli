package cmd

import (
	"strings"
	"testing"
)

func TestAuthLoginHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "auth", "login", "--help")
	for _, fragment := range []string{
		"BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login",
		"--username string",
		"--with-token",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("auth login help missing %q\n%s", fragment, output)
		}
	}
}

func TestAuthStatusHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "auth", "status", "--help")
	for _, fragment := range []string{
		"bb auth status --check --json",
		"--check",
		"--host string",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("auth status help missing %q\n%s", fragment, output)
		}
	}
}

func TestConfigSetHelpRegression(t *testing.T) {
	t.Parallel()

	output := renderHelp(t, "config", "set", "--help")
	for _, fragment := range []string{
		"bb config set browser 'firefox --new-window'",
		"bb config set output.format json",
		"bb config get output.format",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("config set help missing %q\n%s", fragment, output)
		}
	}
}
