package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	gitrepo "github.com/auro/bitbucket_cli/internal/git"
	"github.com/auro/bitbucket_cli/internal/output"
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

type browseRepositoryGetter interface {
	GetRepository(context.Context, string, string) (bitbucket.Repository, error)
}

var (
	openBrowseURL       browserOpener                                 = openURLInBrowser
	loadBrowseConfig                                                  = config.Load
	currentBrowseBranch func(context.Context, string) (string, error) = gitrepo.CurrentBranch
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
		Example: "  bb browse --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb browse README.md:12 --repo OhBizzle/bb-cli-integration-primary --no-browser\n" +
			"  bb browse --pr 1 --repo OhBizzle/bb-cli-integration-primary\n" +
			"  bb browse --pipelines --repo OhBizzle/bb-cli-integration-primary --json '*'\n" +
			"  bb browse a1b2c3d --repo OhBizzle/bb-cli-integration-primary --no-browser",
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
				if err := writeTargetHeader(w, "Repository", payload.Workspace, payload.Repo); err != nil {
					return err
				}
				if err := writeWarnings(w, payload.Warnings); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Type", payload.Type); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Ref", payload.Ref); err != nil {
					return err
				}
				if err := writeLabelValue(w, "Path", payload.Path); err != nil {
					return err
				}
				if payload.Line > 0 {
					if err := writeLabelValue(w, "Line", strconv.Itoa(payload.Line)); err != nil {
						return err
					}
				}
				if err := writeLabelValue(w, "Commit", payload.Commit); err != nil {
					return err
				}
				if payload.PR > 0 {
					if err := writeLabelValue(w, "Pull Request", strconv.Itoa(payload.PR)); err != nil {
						return err
					}
				}
				if payload.Issue > 0 {
					if err := writeLabelValue(w, "Issue", strconv.Itoa(payload.Issue)); err != nil {
						return err
					}
				}
				if err := writeLabelValue(w, "URL", payload.URL); err != nil {
					return err
				}
				if payload.Opened {
					return writeLabelValue(w, "Status", "opened")
				}
				return writeLabelValue(w, "Status", "printed")
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

func buildBrowsePayload(ctx context.Context, client browseRepositoryGetter, target resolvedRepoTarget, raw string, options browseOptions) (browsePayload, error) {
	if err := validateBrowseOptions(raw, options); err != nil {
		return browsePayload{}, err
	}

	base := browseBaseURL(target.Host, target.Workspace, target.Repo)
	payload := browsePayload{
		Host:      target.Host,
		Workspace: target.Workspace,
		Repo:      target.Repo,
	}

	switch {
	case options.Settings:
		payload.Type = "settings"
		payload.URL = base + "/admin"
		return payload, nil
	case options.Pipelines:
		payload.Type = "pipelines"
		payload.URL = base + "/pipelines"
		return payload, nil
	case options.PR > 0:
		payload.Type = "pull-request"
		payload.PR = options.PR
		payload.URL = fmt.Sprintf("%s/pull-requests/%d", base, options.PR)
		return payload, nil
	case options.Issue > 0:
		payload.Type = "issue"
		payload.Issue = options.Issue
		payload.URL = fmt.Sprintf("%s/issues/%d", base, options.Issue)
		return payload, nil
	}

	raw = strings.TrimSpace(raw)
	if raw == "" && strings.TrimSpace(options.Branch) == "" && strings.TrimSpace(options.Commit) == "" {
		payload.Type = "repository"
		payload.URL = base
		return payload, nil
	}

	if raw == "" {
		ref, warnings, err := resolveBrowseRef(ctx, client, target, options.Branch, options.Commit)
		if err != nil {
			return browsePayload{}, err
		}
		payload.Type = "path"
		payload.Warnings = append(payload.Warnings, warnings...)
		payload.Ref = ref
		payload.URL = fmt.Sprintf("%s/src/%s/", base, escapeURLPath(ref))
		return payload, nil
	}

	path, line, ok := parsePathLineReference(raw)
	if ok || shouldTreatAsPath(raw, target) {
		ref, warnings, err := resolveBrowseRef(ctx, client, target, options.Branch, options.Commit)
		if err != nil {
			return browsePayload{}, err
		}
		resolvedPath, err := resolveBrowsePath(target, coalesce(path, raw))
		if err != nil {
			return browsePayload{}, err
		}
		if target.LocalRepo == nil && resolvedPath != "" {
			warnings = append(warnings, fmt.Sprintf("resolved %q without local repository context; treating it as repository-relative", resolvedPath))
		}

		payload.Type = "path"
		payload.Warnings = append(payload.Warnings, warnings...)
		payload.Ref = ref
		payload.Path = resolvedPath
		payload.Line = line
		payload.URL = buildBrowsePathURL(base, ref, resolvedPath, line)
		return payload, nil
	}

	if isLikelyCommit(raw) {
		payload.Type = "commit"
		payload.Commit = raw
		payload.URL = fmt.Sprintf("%s/commits/%s", base, escapeURLPath(raw))
		return payload, nil
	}

	ref, warnings, err := resolveBrowseRef(ctx, client, target, options.Branch, options.Commit)
	if err != nil {
		return browsePayload{}, err
	}
	resolvedPath, err := resolveBrowsePath(target, raw)
	if err != nil {
		return browsePayload{}, err
	}
	if target.LocalRepo == nil && resolvedPath != "" {
		warnings = append(warnings, fmt.Sprintf("resolved %q without local repository context; treating it as repository-relative", resolvedPath))
	}

	payload.Type = "path"
	payload.Warnings = append(payload.Warnings, warnings...)
	payload.Ref = ref
	payload.Path = resolvedPath
	payload.URL = buildBrowsePathURL(base, ref, resolvedPath, 0)
	return payload, nil
}

func validateBrowseOptions(raw string, options browseOptions) error {
	targetKinds := 0
	if options.PR > 0 {
		targetKinds++
	}
	if options.Issue > 0 {
		targetKinds++
	}
	if options.Settings {
		targetKinds++
	}
	if options.Pipelines {
		targetKinds++
	}
	if targetKinds > 1 {
		return fmt.Errorf("choose only one of --pr, --issue, --settings, or --pipelines")
	}
	if options.PR < 0 || options.Issue < 0 {
		return fmt.Errorf("pull request and issue IDs must be greater than zero")
	}
	if strings.TrimSpace(options.Branch) != "" && strings.TrimSpace(options.Commit) != "" {
		return fmt.Errorf("--branch and --commit cannot be used together")
	}
	if targetKinds > 0 && strings.TrimSpace(raw) != "" {
		return fmt.Errorf("a positional browse target cannot be combined with --pr, --issue, --settings, or --pipelines")
	}
	return nil
}

func browseBaseURL(host, workspace, repo string) string {
	webHost := strings.TrimSpace(host)
	switch webHost {
	case "", "api.bitbucket.org":
		webHost = "bitbucket.org"
	}

	return fmt.Sprintf("https://%s/%s/%s", webHost, escapeURLPath(workspace), escapeURLPath(repo))
}

func resolveBrowseRef(ctx context.Context, client browseRepositoryGetter, target resolvedRepoTarget, branchFlag, commitFlag string) (string, []string, error) {
	if strings.TrimSpace(commitFlag) != "" {
		return strings.TrimSpace(commitFlag), nil, nil
	}
	if strings.TrimSpace(branchFlag) != "" {
		return strings.TrimSpace(branchFlag), nil, nil
	}
	warnings := make([]string, 0, 1)
	if target.LocalRepo != nil && target.LocalRepo.RootDir != "" {
		branch, err := currentBrowseBranch(ctx, target.LocalRepo.RootDir)
		if err == nil && strings.TrimSpace(branch) != "" {
			return strings.TrimSpace(branch), nil, nil
		}
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("could not determine the local branch; falling back to the repository main branch (%v)", err))
		} else {
			warnings = append(warnings, "could not determine the local branch; falling back to the repository main branch")
		}
	}

	repository, err := client.GetRepository(ctx, target.Workspace, target.Repo)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(repository.MainBranch.Name) == "" {
		return "", nil, fmt.Errorf("could not determine a branch for %s/%s", target.Workspace, target.Repo)
	}
	return strings.TrimSpace(repository.MainBranch.Name), warnings, nil
}

func parsePathLineReference(raw string) (string, int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, false
	}

	lastColon := strings.LastIndex(raw, ":")
	if lastColon <= 0 || lastColon == len(raw)-1 {
		return raw, 0, false
	}

	line, err := strconv.Atoi(raw[lastColon+1:])
	if err != nil || line <= 0 {
		return raw, 0, false
	}

	return raw[:lastColon], line, true
}

func shouldTreatAsPath(raw string, target resolvedRepoTarget) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	if strings.Contains(raw, "/") || strings.Contains(raw, ".") {
		return true
	}
	if target.LocalRepo == nil || target.LocalRepo.RootDir == "" {
		return false
	}

	candidate, err := browseAbsolutePath(target.LocalRepo.RootDir, raw)
	if err != nil {
		return false
	}
	_, err = os.Stat(candidate)
	return err == nil
}

func resolveBrowsePath(target resolvedRepoTarget, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if target.LocalRepo == nil || target.LocalRepo.RootDir == "" {
		return filepath.ToSlash(filepath.Clean(raw)), nil
	}

	absolutePath, err := browseAbsolutePath(target.LocalRepo.RootDir, raw)
	if err != nil {
		return "", err
	}
	relativePath, err := filepath.Rel(target.LocalRepo.RootDir, absolutePath)
	if err != nil {
		return "", fmt.Errorf("resolve repository-relative path: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside the repository root", raw)
	}
	return filepath.ToSlash(filepath.Clean(relativePath)), nil
}

func browseAbsolutePath(repoRoot, raw string) (string, error) {
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw), nil
	}

	cwd, err := getWorkingDirectory()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	if repoRoot == "" {
		return filepath.Join(cwd, raw), nil
	}

	if relToRoot, err := filepath.Rel(repoRoot, cwd); err == nil && relToRoot != ".." && !strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return filepath.Join(cwd, raw), nil
	}

	return filepath.Join(repoRoot, raw), nil
}

func buildBrowsePathURL(base, ref, path string, line int) string {
	var url string
	if strings.TrimSpace(path) == "" {
		url = fmt.Sprintf("%s/src/%s/", base, escapeURLPath(ref))
	} else {
		url = fmt.Sprintf("%s/src/%s/%s", base, escapeURLPath(ref), escapePathSegments(path))
	}
	if line > 0 {
		url += "#lines-" + strconv.Itoa(line)
	}
	return url
}

func escapePathSegments(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range parts {
		parts[i] = escapeURLPath(part)
	}
	return strings.Join(parts, "/")
}

func escapeURLPath(value string) string {
	return url.PathEscape(strings.TrimSpace(value))
}

func isLikelyCommit(raw string) bool {
	return commitSHAPattern.MatchString(strings.TrimSpace(raw))
}

func openURLInBrowser(rawURL string) error {
	commandLine, err := configuredBrowserCommand()
	if err != nil {
		return err
	}

	if commandLine != "" {
		args, err := splitCommandLine(commandLine)
		if err != nil {
			return fmt.Errorf("parse browser command: %w", err)
		}
		if len(args) == 0 {
			return fmt.Errorf("browser command is empty")
		}
		return startBrowserProcess(args[0], append(args[1:], rawURL)...)
	}

	name, args := defaultBrowserCommand(runtime.GOOS, rawURL)
	return startBrowserProcess(name, args...)
}

func defaultBrowserCommand(goos, rawURL string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{rawURL}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", rawURL}
	default:
		return "xdg-open", []string{rawURL}
	}
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

func startBrowserProcess(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open browser with %s: %w", name, err)
	}
	return nil
}
