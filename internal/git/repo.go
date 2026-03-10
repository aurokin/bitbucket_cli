package git

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
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

type ParsedRemote struct {
	Host      string
	Workspace string
	RepoSlug  string
}

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

func parseStandardURL(raw string) (ParsedRemote, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ParsedRemote{}, fmt.Errorf("parse remote URL %q: %w", raw, err)
	}

	return remoteFromParts(parsed.Hostname(), parsed.Path)
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
