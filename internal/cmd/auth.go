package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"login-manager"},
		Short:   "Manage authentication",
		Long:    "Manage Bitbucket Cloud authentication using Atlassian API tokens.",
	}

	authCmd.AddCommand(
		newAuthLoginCmd(),
		newAuthStatusCmd(),
		newAuthLogoutCmd(),
	)

	return authCmd
}

func newStubCommand(use, short, feature string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s is not implemented yet", feature)
		},
	}
}

type authStatusPayload struct {
	DefaultHost string              `json:"default_host,omitempty"`
	Hosts       []authStatusHostRow `json:"hosts"`
}

type authStatusHostRow struct {
	Host                string `json:"host"`
	Default             bool   `json:"default"`
	Username            string `json:"username,omitempty"`
	TokenConfigured     bool   `json:"token_configured"`
	AuthType            string `json:"auth_type,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
	Authenticated       *bool  `json:"authenticated,omitempty"`
	AuthenticationError string `json:"authentication_error,omitempty"`
	AccountID           string `json:"account_id,omitempty"`
	DisplayName         string `json:"display_name,omitempty"`
	UUID                string `json:"uuid,omitempty"`
}

func newAuthLoginCmd() *cobra.Command {
	var host string
	var token string
	var tokenFromStdin bool
	var username string
	var setDefault bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store credentials for a Bitbucket host",
		Long:  "Store an Atlassian API token for Bitbucket Cloud. The username should be your Atlassian account email.",
		Example: "  bb auth login --username you@example.com --with-token\n" +
			"  bb auth login --username you@example.com --token $BITBUCKET_TOKEN",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedToken, err := resolveTokenValue(cmd.InOrStdin(), token, tokenFromStdin)
			if err != nil {
				return err
			}
			if strings.TrimSpace(username) == "" {
				return fmt.Errorf("--username is required and should be your Atlassian account email")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.SetHost(host, config.HostConfig{
				Username:  strings.TrimSpace(username),
				Token:     resolvedToken,
				AuthType:  config.AuthTypeAPIToken,
				UpdatedAt: time.Now().UTC(),
			}, setDefault)

			if err := config.Save(cfg); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Stored credentials for %s\n", host)
			return err
		},
	}

	cmd.Flags().StringVar(&host, "host", "bitbucket.org", "Bitbucket host to configure")
	cmd.Flags().StringVar(&token, "token", "", "Atlassian API token to store")
	cmd.Flags().BoolVar(&tokenFromStdin, "with-token", false, "Read the API token from stdin")
	cmd.Flags().StringVar(&username, "username", "", "Atlassian account email associated with the API token")
	cmd.Flags().BoolVar(&setDefault, "default", true, "Set this host as the default")
	cmd.MarkFlagsMutuallyExclusive("token", "with-token")

	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	var flags formatFlags
	var check bool
	var host string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show stored authentication status",
		Example: "  bb auth status\n" +
			"  bb auth status --check --json\n" +
			"  bb auth status --host bitbucket.org",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			payload, err := buildAuthStatusPayload(context.Background(), cfg, strings.TrimSpace(host), check)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				if len(payload.Hosts) == 0 {
					_, err := io.WriteString(w, "No authenticated hosts.\n")
					return err
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
				return tw.Flush()
			})
		},
	}

	cmd.Flags().StringVar(&flags.json, "json", "", "Output JSON with the specified comma-separated fields, or '*' for all fields")
	cmd.Flags().Lookup("json").NoOptDefVal = "*"
	cmd.Flags().StringVar(&flags.jq, "jq", "", "Filter JSON output using a jq expression")
	cmd.Flags().BoolVar(&check, "check", false, "Validate stored credentials with the Bitbucket API")
	cmd.Flags().StringVar(&host, "host", "", "Only show status for a specific host")

	return cmd
}

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

func newAuthLogoutCmd() *cobra.Command {
	var host string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials for a Bitbucket host",
		Example: "  bb auth logout\n" +
			"  bb auth logout --host bitbucket.org",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			resolvedHost, err := cfg.ResolveHost(strings.TrimSpace(host))
			if err != nil {
				return authGuidanceError(err)
			}

			if _, ok := cfg.Hosts[resolvedHost]; !ok {
				return fmt.Errorf("no stored credentials found for %s", resolvedHost)
			}

			cfg.RemoveHost(resolvedHost)
			if err := config.Save(cfg); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed credentials for %s\n", resolvedHost)
			return err
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to log out from")

	return cmd
}

func resolveTokenValue(r io.Reader, token string, tokenFromStdin bool) (string, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed != "" {
		return trimmed, nil
	}

	if !tokenFromStdin {
		return "", fmt.Errorf("provide an API token with --token or pass --with-token to read it from stdin")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read token from stdin: %w", err)
	}

	trimmed = strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", fmt.Errorf("stdin did not contain an API token")
	}

	return trimmed, nil
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
				row.AuthenticationError = err.Error()
			} else {
				currentUser, err := client.CurrentUser(ctx)
				if err != nil {
					authenticated := false
					row.Authenticated = &authenticated
					row.AuthenticationError = err.Error()
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
