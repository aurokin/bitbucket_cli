package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
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
	TokenType           string `json:"token_type,omitempty"`
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
	var tokenType string
	var username string
	var setDefault bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store credentials for a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedToken, err := resolveTokenValue(cmd.InOrStdin(), token, tokenFromStdin)
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.SetHost(host, config.HostConfig{
				Username:  strings.TrimSpace(username),
				Token:     resolvedToken,
				TokenType: strings.TrimSpace(tokenType),
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
	cmd.Flags().StringVar(&token, "token", "", "Token or app password to store")
	cmd.Flags().BoolVar(&tokenFromStdin, "with-token", false, "Read the token from stdin")
	cmd.Flags().StringVar(&tokenType, "token-type", "bearer", "Credential type to store: bearer or app-password")
	cmd.Flags().StringVar(&username, "username", "", "Username associated with the credential")
	cmd.Flags().BoolVar(&setDefault, "default", true, "Set this host as the default")

	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	var flags formatFlags
	var check bool
	var host string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show stored authentication status",
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

				tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintln(tw, "HOST\tDEFAULT\tUSERNAME\tTOKEN TYPE\tAUTHENTICATED\tACCOUNT\tUPDATED"); err != nil {
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

					if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", host.Host, defaultLabel, host.Username, host.TokenType, authLabel, accountLabel, host.UpdatedAt); err != nil {
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

func newAuthLogoutCmd() *cobra.Command {
	var host string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials for a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			resolvedHost, err := cfg.ResolveHost(strings.TrimSpace(host))
			if err != nil {
				return err
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
		return "", fmt.Errorf("provide --token or --with-token")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read token from stdin: %w", err)
	}

	trimmed = strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", fmt.Errorf("stdin token was empty")
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
			TokenType:       host.TokenType,
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
		return nil, fmt.Errorf("no stored credentials found for %s", selectedHost)
	}

	return []string{selectedHost}, nil
}
