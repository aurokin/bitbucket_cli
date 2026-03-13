package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CommitDiffOptions struct {
	Context          int
	Path             []string
	IgnoreWhitespace bool
	Binary           *bool
	Renames          *bool
}

type ListCommitCommentsOptions struct {
	Limit int
	Sort  string
	Query string
}

type ListCommitStatusesOptions struct {
	Limit   int
	Sort    string
	Query   string
	RefName string
}

type ListCommitReportsOptions struct {
	Limit int
}

type CommitComment struct {
	ID        int                  `json:"id"`
	Content   CommitCommentBody    `json:"content,omitempty"`
	User      PullRequestActor     `json:"user,omitempty"`
	Inline    *CommitCommentInline `json:"inline,omitempty"`
	Deleted   bool                 `json:"deleted,omitempty"`
	Pending   bool                 `json:"pending,omitempty"`
	CreatedOn string               `json:"created_on,omitempty"`
	UpdatedOn string               `json:"updated_on,omitempty"`
	Links     CommitCommentLinks   `json:"links,omitempty"`
}

type CommitCommentBody struct {
	Raw string `json:"raw,omitempty"`
}

type CommitCommentInline struct {
	Path string `json:"path,omitempty"`
	From int    `json:"from,omitempty"`
	To   int    `json:"to,omitempty"`
}

type CommitCommentLinks struct {
	HTML Link `json:"html,omitempty"`
}

type CommitReport struct {
	UUID              string             `json:"uuid,omitempty"`
	ExternalID        string             `json:"external_id,omitempty"`
	Title             string             `json:"title,omitempty"`
	Details           string             `json:"details,omitempty"`
	Reporter          string             `json:"reporter,omitempty"`
	Link              string             `json:"link,omitempty"`
	RemoteLinkEnabled bool               `json:"remote_link_enabled,omitempty"`
	LogoURL           string             `json:"logo_url,omitempty"`
	ReportType        string             `json:"report_type,omitempty"`
	Result            string             `json:"result,omitempty"`
	Data              []CommitReportData `json:"data,omitempty"`
	CreatedOn         string             `json:"created_on,omitempty"`
	UpdatedOn         string             `json:"updated_on,omitempty"`
}

type CommitReportData struct {
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
	Value any    `json:"value,omitempty"`
}

type commitCommentListResponse struct {
	Values []CommitComment `json:"values"`
	Next   string          `json:"next,omitempty"`
}

type commitReportListResponse struct {
	Values []CommitReport `json:"values"`
	Next   string         `json:"next,omitempty"`
}

func (c *Client) GetCommit(ctx context.Context, workspace, repoSlug, commit string) (RepositoryCommit, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryCommit{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return RepositoryCommit{}, fmt.Errorf("commit reference is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/commit/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryCommit{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return RepositoryCommit{}, err
	}

	var value RepositoryCommit
	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		return RepositoryCommit{}, fmt.Errorf("decode commit: %w", err)
	}
	return value, nil
}

func (c *Client) GetCommitPatch(ctx context.Context, workspace, repoSlug, commit string) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return "", fmt.Errorf("commit reference is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/patch/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, map[string]string{"Accept": "text/plain"})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read commit patch: %w", err)
	}
	return string(body), nil
}

func (c *Client) GetCommitDiff(ctx context.Context, workspace, repoSlug, commit string, options CommitDiffOptions) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return "", fmt.Errorf("commit reference is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/diff/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit))
	query := url.Values{}
	if options.Context > 0 {
		query.Set("context", strconv.Itoa(options.Context))
	}
	for _, pathValue := range options.Path {
		pathValue = strings.TrimSpace(pathValue)
		if pathValue != "" {
			query.Add("path", pathValue)
		}
	}
	if options.IgnoreWhitespace {
		query.Set("ignore_whitespace", "true")
	}
	if options.Binary != nil {
		query.Set("binary", strconv.FormatBool(*options.Binary))
	}
	if options.Renames != nil {
		query.Set("renames", strconv.FormatBool(*options.Renames))
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	resp, err := c.Do(ctx, http.MethodGet, path, nil, map[string]string{"Accept": "text/plain"})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read commit diff: %w", err)
	}
	return string(body), nil
}

func (c *Client) ListCommitDiffStats(ctx context.Context, workspace, repoSlug, commit string) ([]PullRequestDiffStat, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return nil, fmt.Errorf("commit reference is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/diffstat/%s?pagelen=100", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit))
	nextPath := path
	var all []PullRequestDiffStat

	for nextPath != "" {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pullRequestDiffStatResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode commit diffstat: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	return all, nil
}

func (c *Client) ListCommitComments(ctx context.Context, workspace, repoSlug, commit string, options ListCommitCommentsOptions) ([]CommitComment, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return nil, fmt.Errorf("commit reference is required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	query := url.Values{}
	query.Set("pagelen", strconv.Itoa(options.Limit))
	if strings.TrimSpace(options.Sort) != "" {
		query.Set("sort", strings.TrimSpace(options.Sort))
	}
	if strings.TrimSpace(options.Query) != "" {
		query.Set("q", strings.TrimSpace(options.Query))
	}
	nextPath := fmt.Sprintf("/repositories/%s/%s/commit/%s/comments?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit), query.Encode())
	all := make([]CommitComment, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page commitCommentListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode commit comment list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) GetCommitComment(ctx context.Context, workspace, repoSlug, commit string, commentID int) (CommitComment, error) {
	if workspace == "" || repoSlug == "" {
		return CommitComment{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return CommitComment{}, fmt.Errorf("commit reference is required")
	}
	if commentID <= 0 {
		return CommitComment{}, fmt.Errorf("commit comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit), commentID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return CommitComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return CommitComment{}, err
	}

	var comment CommitComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return CommitComment{}, fmt.Errorf("decode commit comment: %w", err)
	}
	return comment, nil
}

func (c *Client) ReviewCommit(ctx context.Context, workspace, repoSlug, commit string, approve bool) (PullRequestParticipant, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestParticipant{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return PullRequestParticipant{}, fmt.Errorf("commit reference is required")
	}

	method := http.MethodPost
	if !approve {
		method = http.MethodDelete
	}
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/approve", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit))
	resp, err := c.Do(ctx, method, path, nil, nil)
	if err != nil {
		return PullRequestParticipant{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestParticipant{}, err
	}

	if method == http.MethodDelete || resp.StatusCode == http.StatusNoContent {
		return PullRequestParticipant{}, nil
	}

	var participant PullRequestParticipant
	if err := json.NewDecoder(resp.Body).Decode(&participant); err != nil {
		return PullRequestParticipant{}, fmt.Errorf("decode commit approval participant: %w", err)
	}
	return participant, nil
}

func (c *Client) ListCommitStatuses(ctx context.Context, workspace, repoSlug, commit string, options ListCommitStatusesOptions) ([]CommitStatus, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return nil, fmt.Errorf("commit reference is required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	query := url.Values{}
	query.Set("pagelen", strconv.Itoa(options.Limit))
	if strings.TrimSpace(options.Sort) != "" {
		query.Set("sort", strings.TrimSpace(options.Sort))
	}
	if strings.TrimSpace(options.Query) != "" {
		query.Set("q", strings.TrimSpace(options.Query))
	}
	if strings.TrimSpace(options.RefName) != "" {
		query.Set("refname", strings.TrimSpace(options.RefName))
	}

	nextPath := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit), query.Encode())
	all := make([]CommitStatus, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page commitStatusListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode commit status list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) ListCommitReports(ctx context.Context, workspace, repoSlug, commit string, options ListCommitReportsOptions) ([]CommitReport, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return nil, fmt.Errorf("commit reference is required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	nextPath := fmt.Sprintf("/repositories/%s/%s/commit/%s/reports?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit), options.Limit)
	all := make([]CommitReport, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page commitReportListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode commit report list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) GetCommitReport(ctx context.Context, workspace, repoSlug, commit, reportID string) (CommitReport, error) {
	if workspace == "" || repoSlug == "" {
		return CommitReport{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(commit) == "" {
		return CommitReport{}, fmt.Errorf("commit reference is required")
	}
	if strings.TrimSpace(reportID) == "" {
		return CommitReport{}, fmt.Errorf("commit report ID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/reports/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(commit), url.PathEscape(reportID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return CommitReport{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return CommitReport{}, err
	}

	var report CommitReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return CommitReport{}, fmt.Errorf("decode commit report: %w", err)
	}
	return report, nil
}
