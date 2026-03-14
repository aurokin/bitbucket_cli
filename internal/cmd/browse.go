package cmd

import (
	"context"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/config"
	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

type browsePayload struct {
	Host      string   `json:"host"`
	Workspace string   `json:"workspace"`
	Repo      string   `json:"repo"`
	Type      string   `json:"type"`
	URL       string   `json:"url"`
	Warnings  []string `json:"warnings,omitempty"`
	Ref       string   `json:"ref,omitempty"`
	Path      string   `json:"path,omitempty"`
	Line      int      `json:"line,omitempty"`
	Commit    string   `json:"commit,omitempty"`
	PR        int      `json:"pr,omitempty"`
	Issue     int      `json:"issue,omitempty"`
	Opened    bool     `json:"opened"`
}

type browseOptions struct {
	PR        int
	Issue     int
	Settings  bool
	Pipelines bool
	NoBrowser bool
	Branch    string
	Commit    string
}

type browserOpener func(string) error

var (
	openBrowseURL    browserOpener = openURLInBrowser
	loadBrowseConfig               = config.Load
)

var commitSHAPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

func newBrowseCmd() *cobra.Command {
	var flags formatFlags
	var host string
	var workspace string
	var repo string
	var options browseOptions

	cmd := &cobra.Command{
		Use:   "browse [target]",
		Short: "Open or print Bitbucket web URLs",
		Long:  "Open Bitbucket repository, pull request, issue, commit, settings, pipelines, and source URLs. Default behavior opens the browser; use --no-browser to print the URL instead.",
		Example: "  bb browse --repo workspace-slug/repo-slug\n" +
			"  bb browse README.md:12 --repo workspace-slug/repo-slug --no-browser\n" +
			"  bb browse --pr 1 --repo workspace-slug/repo-slug\n" +
			"  bb browse --pipelines --repo workspace-slug/repo-slug --json '*'\n" +
			"  bb browse a1b2c3d --repo workspace-slug/repo-slug --no-browser",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			formatOptions, err := flags.options()
			if err != nil {
				return err
			}

			resolved, err := resolveRepoCommandTarget(context.Background(), host, workspace, repo, true)
			if err != nil {
				return err
			}
			client := resolved.Client
			target := resolved.Target

			payload, err := buildBrowsePayload(context.Background(), client, target, firstArg(args), options)
			if err != nil {
				return err
			}

			if !options.NoBrowser {
				if err := openBrowseURL(payload.URL); err != nil {
					return err
				}
				payload.Opened = true
			}

			return output.Render(cmd.OutOrStdout(), formatOptions, payload, func(w io.Writer) error {
				return writeBrowseSummary(w, payload)
			})
		},
	}

	addFormatFlags(cmd, &flags)
	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Optional workspace slug used only to disambiguate a bare repository target")
	cmd.Flags().StringVar(&repo, "repo", "", "Bitbucket repository target as <repo>, <workspace>/<repo>, or a repository URL")
	cmd.Flags().IntVar(&options.PR, "pr", 0, "Open one pull request by ID")
	cmd.Flags().IntVar(&options.Issue, "issue", 0, "Open one issue by ID")
	cmd.Flags().BoolVar(&options.Settings, "settings", false, "Open repository settings")
	cmd.Flags().BoolVar(&options.Pipelines, "pipelines", false, "Open repository pipelines")
	cmd.Flags().BoolVar(&options.NoBrowser, "no-browser", false, "Print the destination URL instead of opening the browser")
	cmd.Flags().StringVar(&options.Branch, "branch", "", "Branch name for source browsing")
	cmd.Flags().StringVar(&options.Commit, "commit", "", "Commit SHA for source browsing")

	return cmd
}

func configuredBrowserCommand() (string, error) {
	if browser := strings.TrimSpace(os.Getenv("BROWSER")); browser != "" {
		return browser, nil
	}

	cfg, err := loadBrowseConfig()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(cfg.Settings.Browser), nil
}
