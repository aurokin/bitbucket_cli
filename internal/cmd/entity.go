package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type resolvedEntity struct {
	Host         string `json:"host"`
	Workspace    string `json:"workspace"`
	Repo         string `json:"repo"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	CanonicalURL string `json:"canonical_url"`
	Ref          string `json:"ref,omitempty"`
	Path         string `json:"path,omitempty"`
	Line         int    `json:"line,omitempty"`
	Commit       string `json:"commit,omitempty"`
	PR           int    `json:"pr,omitempty"`
	Comment      int    `json:"comment,omitempty"`
	Issue        int    `json:"issue,omitempty"`
}

func parseBitbucketEntityURL(raw string) (resolvedEntity, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL is required")
	}

	parsedURL, err := url.Parse(raw)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" || parsedURL.Hostname() == "" {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q is invalid", raw)
	}

	path := strings.Trim(parsedURL.EscapedPath(), "/")
	rawParts := strings.Split(path, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		decoded, err := url.PathUnescape(part)
		if err != nil {
			return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q is invalid", raw)
		}
		parts = append(parts, decoded)
	}
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q must point to a repository-scoped entity", raw)
	}

	entity := resolvedEntity{
		Host:      parsedURL.Hostname(),
		Workspace: parts[0],
		Repo:      strings.TrimSuffix(parts[1], ".git"),
		URL:       raw,
	}
	base := browseBaseURL(entity.Host, entity.Workspace, entity.Repo)

	switch {
	case len(parts) == 2:
		return resolveRepositoryEntity(entity, base), nil
	case len(parts) >= 4 && parts[2] == "pull-requests":
		return resolvePullRequestEntity(raw, entity, base, parts[3], parsedURL.Fragment)
	case len(parts) >= 4 && parts[2] == "issues":
		return resolveIssueEntity(raw, entity, base, parts[3])
	case len(parts) >= 4 && parts[2] == "commits":
		return resolveCommitEntity(raw, entity, base, parts[3])
	case len(parts) >= 4 && parts[2] == "src":
		return resolvePathEntity(raw, entity, base, parts[3:], parsedURL.Fragment)
	default:
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q is not a supported repository, pull request, comment, issue, commit, or source URL", raw)
	}
}

func resolveRepositoryEntity(entity resolvedEntity, base string) resolvedEntity {
	entity.Type = "repository"
	entity.CanonicalURL = base
	return entity
}

func resolvePullRequestEntity(raw string, entity resolvedEntity, base, idRaw, fragment string) (resolvedEntity, error) {
	id, err := parsePositiveID(idRaw, "pull request")
	if err != nil {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q does not contain a valid pull request ID", raw)
	}
	entity.PR = id
	entity.CanonicalURL = fmt.Sprintf("%s/pull-requests/%d", base, id)
	if commentID := parseCommentFragment(fragment); commentID > 0 {
		entity.Type = "pull-request-comment"
		entity.Comment = commentID
		entity.CanonicalURL += fmt.Sprintf("#comment-%d", commentID)
		return entity, nil
	}
	entity.Type = "pull-request"
	return entity, nil
}

func resolveIssueEntity(raw string, entity resolvedEntity, base, idRaw string) (resolvedEntity, error) {
	id, err := parsePositiveID(idRaw, "issue")
	if err != nil {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q does not contain a valid issue ID", raw)
	}
	entity.Type = "issue"
	entity.Issue = id
	entity.CanonicalURL = fmt.Sprintf("%s/issues/%d", base, id)
	return entity, nil
}

func resolveCommitEntity(raw string, entity resolvedEntity, base, commitRaw string) (resolvedEntity, error) {
	commit := strings.TrimSpace(commitRaw)
	if commit == "" {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q does not contain a valid commit SHA", raw)
	}
	entity.Type = "commit"
	entity.Commit = commit
	entity.CanonicalURL = fmt.Sprintf("%s/commits/%s", base, escapeURLPath(commit))
	return entity, nil
}

func resolvePathEntity(raw string, entity resolvedEntity, base string, parts []string, fragment string) (resolvedEntity, error) {
	ref := strings.TrimSpace(parts[0])
	if ref == "" {
		return resolvedEntity{}, fmt.Errorf("Bitbucket URL %q does not contain a valid source ref", raw)
	}
	entity.Type = "path"
	entity.Ref = ref
	if len(parts) > 1 {
		entity.Path = filepath.ToSlash(strings.Join(parts[1:], "/"))
	}
	entity.Line = parseSourceLineFragment(fragment)
	entity.CanonicalURL = buildBrowsePathURL(base, entity.Ref, entity.Path, entity.Line)
	return entity, nil
}

func parsePositiveID(raw, kind string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s ID %q", kind, raw)
	}
	return id, nil
}

func parseCommentFragment(fragment string) int {
	fragment = strings.TrimSpace(fragment)
	if !strings.HasPrefix(fragment, "comment-") {
		return 0
	}
	id, err := strconv.Atoi(strings.TrimPrefix(fragment, "comment-"))
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func parseSourceLineFragment(fragment string) int {
	fragment = strings.TrimSpace(fragment)
	if !strings.HasPrefix(fragment, "lines-") {
		return 0
	}
	lineRange := strings.TrimPrefix(fragment, "lines-")
	end := len(lineRange)
	for i, r := range lineRange {
		if r < '0' || r > '9' {
			end = i
			break
		}
	}
	if end == 0 {
		return 0
	}
	line, err := strconv.Atoi(lineRange[:end])
	if err != nil || line <= 0 {
		return 0
	}
	return line
}
