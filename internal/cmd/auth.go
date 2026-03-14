package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	atlassianAPITokenManageURL = "https://id.atlassian.com/manage-profile/security/api-tokens"
	atlassianAPITokenDocsURL   = "https://support.atlassian.com/bitbucket-cloud/docs/using-api-tokens/"
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
		Long:  "Store an Atlassian API token for Bitbucket Cloud. The username should be your Atlassian account email. Humans can run `bb auth login` interactively and paste the token securely. Agents can provide `BB_EMAIL` and `BB_TOKEN`, or pass `--username` and `--token` explicitly.",
		Example: "  bb auth login --username you@example.com --with-token\n" +
			"  bb auth login --username you@example.com --token $BITBUCKET_TOKEN\n" +
			"  BB_EMAIL=you@example.com BB_TOKEN=$BITBUCKET_TOKEN bb auth login\n" +
			"  printf '%s\\n' \"$BITBUCKET_TOKEN\" | bb auth login --username you@example.com --with-token",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedUsername, err := resolveUsernameValue(cmd, username)
			if err != nil {
				return err
			}
			resolvedToken, err := resolveTokenValue(cmd, token, tokenFromStdin)
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.SetHost(host, config.HostConfig{
				Username:  resolvedUsername,
				Token:     resolvedToken,
				AuthType:  config.AuthTypeAPIToken,
				UpdatedAt: time.Now().UTC(),
			}, setDefault)

			if err := config.Save(cfg); err != nil {
				return err
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Stored credentials for %s as %s\n", host, resolvedUsername); err != nil {
				return err
			}
			return writeNextStep(cmd.OutOrStdout(), fmt.Sprintf("bb auth status --check --host %s", host))
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
				return writeAuthStatusSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().BoolVar(&check, "check", false, "Validate stored credentials with the Bitbucket API")
	cmd.Flags().StringVar(&host, "host", "", "Only show status for a specific host")

	return cmd
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

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Removed credentials for %s\n", resolvedHost); err != nil {
				return err
			}
			if err := writeLabelValue(cmd.OutOrStdout(), "Create Token", atlassianAPITokenManageURL); err != nil {
				return err
			}
			return writeNextStep(cmd.OutOrStdout(), fmt.Sprintf("bb auth login --host %s", resolvedHost))
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to log out from")

	return cmd
}
