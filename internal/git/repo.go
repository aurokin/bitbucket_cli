package git

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

type RepoContext struct {
	Host       string `json:"host"`
	Workspace  string `json:"workspace"`
	RepoSlug   string `json:"repo"`
	RemoteName string `json:"remote"`
	CloneURL   string `json:"clone_url"`
	RootDir    string `json:"root"`
}

func ResolveRepoContext(ctx context.Context, dir string) (RepoContext, error) {
	rootDir, err := gitOutput(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return RepoContext{}, fmt.Errorf("resolve git root: %w", err)
	}

	remoteName, err := detectRemoteName(ctx, rootDir)
	if err != nil {
		return RepoContext{}, err
	}

	cloneURL, err := gitOutput(ctx, rootDir, "remote", "get-url", remoteName)
	if err != nil {
		return RepoContext{}, fmt.Errorf("resolve remote URL for %q: %w", remoteName, err)
	}

	parsed, err := ParseRemoteURL(cloneURL)
	if err != nil {
		return RepoContext{}, err
	}

	return RepoContext{
		Host:       parsed.Host,
		Workspace:  parsed.Workspace,
		RepoSlug:   parsed.RepoSlug,
		RemoteName: remoteName,
		CloneURL:   cloneURL,
		RootDir:    rootDir,
	}, nil
}

func CurrentBranch(ctx context.Context, dir string) (string, error) {
	branch, err := gitOutput(ctx, dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}

	return branch, nil
}

func CheckoutRemoteBranch(ctx context.Context, dir, remoteName, branch string) error {
	if remoteName == "" {
		return fmt.Errorf("remote name is required")
	}
	if branch == "" {
		return fmt.Errorf("branch name is required")
	}

	if _, err := gitOutput(ctx, dir, "fetch", remoteName, branch); err != nil {
		return err
	}

	if branchExists(ctx, dir, branch) {
		if _, err := gitOutput(ctx, dir, "switch", branch); err != nil {
			return err
		}
		if _, err := gitOutput(ctx, dir, "branch", "--set-upstream-to", remoteName+"/"+branch, branch); err != nil {
			return err
		}
		return nil
	}

	if _, err := gitOutput(ctx, dir, "switch", "-c", branch, "--track", remoteName+"/"+branch); err != nil {
		return err
	}

	return nil
}

type ParsedRemote struct {
	Host      string
	Workspace string
	RepoSlug  string
}

const bitbucketAPITokenUsername = "x-bitbucket-api-token-auth"

func ParseRemoteURL(raw string) (ParsedRemote, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParsedRemote{}, fmt.Errorf("remote URL is empty")
	}

	if strings.Contains(raw, "://") {
		return parseStandardURL(raw)
	}

	if strings.Contains(raw, "@") && strings.Contains(raw, ":") {
		return parseSCPStyleURL(raw)
	}

	return ParsedRemote{}, fmt.Errorf("unsupported remote URL format %q", raw)
}

func CloneRepository(ctx context.Context, cloneURL, token, dir string) error {
	if strings.TrimSpace(cloneURL) == "" {
		return fmt.Errorf("clone URL is required")
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("API token is required to clone private repositories")
	}

	authenticatedURL, err := authenticatedHTTPSURL(cloneURL, bitbucketAPITokenUsername, token)
	if err != nil {
		return fmt.Errorf("prepare clone URL: %w", err)
	}

	args := []string{"clone", authenticatedURL}
	if strings.TrimSpace(dir) != "" {
		args = append(args, dir)
	}

	if _, err := gitOutput(ctx, ".", args...); err != nil {
		return err
	}

	targetDir := strings.TrimSpace(dir)
	if targetDir == "" {
		parsed, err := ParseRemoteURL(cloneURL)
		if err != nil {
			return fmt.Errorf("resolve clone directory from repository slug: %w", err)
		}
		targetDir = parsed.RepoSlug
	}

	absoluteDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolve clone directory: %w", err)
	}

	sanitizedURL, err := sanitizedHTTPSURL(cloneURL, bitbucketAPITokenUsername)
	if err != nil {
		return fmt.Errorf("prepare sanitized remote URL: %w", err)
	}

	if _, err := gitOutput(ctx, absoluteDir, "remote", "set-url", "origin", sanitizedURL); err != nil {
		return err
	}

	return nil
}

func parseStandardURL(raw string) (ParsedRemote, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ParsedRemote{}, fmt.Errorf("parse remote URL %q: %w", raw, err)
	}

	return remoteFromParts(parsed.Hostname(), parsed.Path)
}

func authenticatedHTTPSURL(rawURL, username, password string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("parse HTTPS URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported clone URL scheme %q", parsed.Scheme)
	}

	parsed.User = url.UserPassword(username, password)
	return parsed.String(), nil
}

func sanitizedHTTPSURL(rawURL, username string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("parse HTTPS URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported clone URL scheme %q", parsed.Scheme)
	}

	parsed.User = url.User(username)
	return parsed.String(), nil
}

func parseSCPStyleURL(raw string) (ParsedRemote, error) {
	at := strings.LastIndex(raw, "@")
	colon := strings.LastIndex(raw, ":")
	if at == -1 || colon == -1 || colon < at {
		return ParsedRemote{}, fmt.Errorf("unsupported SCP-style remote URL %q", raw)
	}

	host := raw[at+1 : colon]
	path := raw[colon+1:]

	return remoteFromParts(host, path)
}

func remoteFromParts(host, path string) (ParsedRemote, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return ParsedRemote{}, fmt.Errorf("remote host is empty")
	}

	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, ".git")

	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ParsedRemote{}, fmt.Errorf("remote path %q does not look like a Bitbucket Cloud repository", path)
	}

	return ParsedRemote{
		Host:      host,
		Workspace: parts[0],
		RepoSlug:  parts[1],
	}, nil
}

func detectRemoteName(ctx context.Context, dir string) (string, error) {
	branch, err := gitOutput(ctx, dir, "branch", "--show-current")
	if err == nil && branch != "" {
		remoteName, remoteErr := gitOutput(ctx, dir, "config", "--get", "branch."+branch+".remote")
		if remoteErr == nil && remoteName != "" {
			return remoteName, nil
		}
	}

	remotes, err := gitOutput(ctx, dir, "remote")
	if err != nil {
		return "", fmt.Errorf("list remotes: %w", err)
	}

	names := strings.Fields(remotes)
	if len(names) == 0 {
		return "", fmt.Errorf("no git remotes configured")
	}

	for _, name := range names {
		if name == "origin" {
			return name, nil
		}
	}

	return names[0], nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

func branchExists(ctx context.Context, dir, branch string) bool {
	_, err := gitOutput(ctx, dir, "rev-parse", "--verify", "refs/heads/"+branch)
	return err == nil
}
