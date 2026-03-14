package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func authStatusAccountWidth(termWidth int) int {
	switch {
	case termWidth >= 160:
		return 42
	case termWidth >= 132:
		return 30
	default:
		return 22
	}
}

func writeAuthStatusSummary(w io.Writer, payload authStatusPayload) error {
	if len(payload.Hosts) == 0 {
		if _, err := io.WriteString(w, "No authenticated hosts.\n"); err != nil {
			return err
		}
		if err := writeLabelValue(w, "Create Token", atlassianAPITokenManageURL); err != nil {
			return err
		}
		return writeNextStep(w, "bb auth login")
	}

	accountWidth := authStatusAccountWidth(output.TerminalWidth(w))
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "host\tdefault\tuser\tauth\tok\taccount\tupdated"); err != nil {
		return err
	}
	for _, host := range payload.Hosts {
		defaultLabel := ""
		if host.Default {
			defaultLabel = "yes"
		}

		authLabel := ""
		if host.Authenticated != nil {
			if *host.Authenticated {
				authLabel = "yes"
			} else {
				authLabel = "no"
			}
		}

		accountLabel := host.DisplayName
		if accountLabel == "" && host.AccountID != "" {
			accountLabel = host.AccountID
		}
		if host.AuthenticationError != "" {
			accountLabel = host.AuthenticationError
		}

		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			output.Truncate(host.Host, 20),
			defaultLabel,
			output.Truncate(host.Username, 24),
			output.Truncate(host.AuthType, 12),
			authLabel,
			output.Truncate(accountLabel, accountWidth),
			host.UpdatedAt,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	for _, host := range payload.Hosts {
		if host.AuthenticationError == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "\n%s auth error: %s\n", host.Host, host.AuthenticationError); err != nil {
			return err
		}
	}

	return nil
}

func resolveUsernameValue(cmd *cobra.Command, username string) (string, error) {
	trimmed := strings.TrimSpace(username)
	if trimmed != "" {
		return trimmed, nil
	}

	if envEmail := strings.TrimSpace(os.Getenv("BB_EMAIL")); envEmail != "" {
		return envEmail, nil
	}

	if promptsEnabled(cmd) {
		return promptRequiredString(cmd, "Atlassian account email", "")
	}

	return "", fmt.Errorf("an Atlassian account email is required. Pass --username, set BB_EMAIL, or run in an interactive terminal. Create an API token at %s", atlassianAPITokenManageURL)
}

func resolveTokenValue(cmd *cobra.Command, token string, tokenFromStdin bool) (string, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed != "" {
		return trimmed, nil
	}

	if tokenFromStdin {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read token from stdin: %w", err)
		}

		trimmed = strings.TrimSpace(string(data))
		if trimmed == "" {
			return "", fmt.Errorf("stdin did not contain an API token")
		}
		return trimmed, nil
	}

	if envToken := strings.TrimSpace(os.Getenv("BB_TOKEN")); envToken != "" {
		return envToken, nil
	}

	if promptsEnabled(cmd) {
		return promptSecretString(cmd, "Atlassian API token")
	}

	return "", fmt.Errorf("an Atlassian API token is required. Pass --token, use --with-token, set BB_TOKEN, or run in an interactive terminal. Create one at %s", atlassianAPITokenManageURL)
}

func buildAuthStatusPayload(ctx context.Context, cfg config.Config, selectedHost string, check bool) (authStatusPayload, error) {
	payload := authStatusPayload{
		DefaultHost: cfg.DefaultHost,
		Hosts:       make([]authStatusHostRow, 0, len(cfg.Hosts)),
	}

	hostNames, err := statusHostNames(cfg, selectedHost)
	if err != nil {
		return authStatusPayload{}, err
	}

	for _, hostName := range hostNames {
		host := cfg.Hosts[hostName]
		row := authStatusHostRow{
			Host:            hostName,
			Default:         hostName == cfg.DefaultHost,
			Username:        host.Username,
			TokenConfigured: host.Token != "",
			AuthType:        host.AuthType,
		}
		if !host.UpdatedAt.IsZero() {
			row.UpdatedAt = host.UpdatedAt.Format(time.RFC3339)
		}

		if check {
			client, err := bitbucket.NewClient(hostName, host)
			if err != nil {
				authenticated := false
				row.Authenticated = &authenticated
				row.AuthenticationError = userFacingError(err).Error()
			} else {
				currentUser, err := client.CurrentUser(ctx)
				if err != nil {
					authenticated := false
					row.Authenticated = &authenticated
					row.AuthenticationError = userFacingError(err).Error()
				} else {
					authenticated := true
					row.Authenticated = &authenticated
					row.AccountID = currentUser.AccountID
					row.DisplayName = currentUser.DisplayName
					row.UUID = currentUser.UUID
				}
			}
		}

		payload.Hosts = append(payload.Hosts, row)
	}

	return payload, nil
}

func statusHostNames(cfg config.Config, selectedHost string) ([]string, error) {
	if selectedHost == "" {
		return cfg.HostNames(), nil
	}

	if _, ok := cfg.Hosts[selectedHost]; !ok {
		return nil, fmt.Errorf("no stored credentials found for %s; run `bb auth login --username <email> --with-token`", selectedHost)
	}

	return []string{selectedHost}, nil
}
