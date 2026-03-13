package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ListPullRequestActivityOptions struct {
	Limit int
}

type ListPullRequestCommitsOptions struct {
	Limit int
}

type ListPullRequestStatusesOptions struct {
	Limit int
	Query string
	Sort  string
}

type PullRequestReviewAction string

const (
	PullRequestReviewApprove             PullRequestReviewAction = "approve"
	PullRequestReviewUnapprove           PullRequestReviewAction = "unapprove"
	PullRequestReviewRequestChanges      PullRequestReviewAction = "request-changes"
	PullRequestReviewClearRequestChanges PullRequestReviewAction = "clear-request-changes"
)

type PullRequestActivity struct {
	Comment        *PullRequestComment       `json:"comment,omitempty"`
	Update         *PullRequestActivityUpdate `json:"update,omitempty"`
	Approval       *PullRequestActivityEvent `json:"approval,omitempty"`
	ChangesRequest *PullRequestActivityEvent `json:"changes_request,omitempty"`
	RequestChanges *PullRequestActivityEvent `json:"request_changes,omitempty"`
}

type PullRequestActivityUpdate struct {
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Reason      string           `json:"reason,omitempty"`
	State       string           `json:"state,omitempty"`
	Date        string           `json:"date,omitempty"`
	Author      PullRequestActor `json:"author"`
	Source      PullRequestRef   `json:"source"`
	Destination PullRequestRef   `json:"destination"`
}

type PullRequestActivityEvent struct {
	Date string           `json:"date,omitempty"`
	User PullRequestActor `json:"user"`
}

type RepositoryCommit struct {
	Hash    string                  `json:"hash,omitempty"`
	Date    string                  `json:"date,omitempty"`
	Message string                  `json:"message,omitempty"`
	Summary RepositoryCommitSummary `json:"summary,omitempty"`
	Author  RepositoryCommitAuthor  `json:"author,omitempty"`
	Links   RepositoryCommitLinks   `json:"links,omitempty"`
}

type RepositoryCommitSummary struct {
	Raw string `json:"raw,omitempty"`
}

type RepositoryCommitAuthor struct {
	Raw  string           `json:"raw,omitempty"`
	User PullRequestActor `json:"user,omitempty"`
}

type RepositoryCommitLinks struct {
	HTML Link `json:"html,omitempty"`
}

type pullRequestActivityListResponse struct {
	Values []PullRequestActivity `json:"values"`
	Next   string                `json:"next,omitempty"`
}

type repositoryCommitListResponse struct {
	Values []RepositoryCommit `json:"values"`
	Next   string             `json:"next,omitempty"`
}

type CommitStatus struct {
	Key         string `json:"key,omitempty"`
	State       string `json:"state,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	RefName     string `json:"refname,omitempty"`
	CreatedOn   string `json:"created_on,omitempty"`
	UpdatedOn   string `json:"updated_on,omitempty"`
	Links       struct {
		Self   Link `json:"self,omitempty"`
		Commit Link `json:"commit,omitempty"`
	} `json:"links,omitempty"`
}

type commitStatusListResponse struct {
	Values []CommitStatus `json:"values"`
	Next   string         `json:"next,omitempty"`
}

func (c *Client) ReviewPullRequest(ctx context.Context, workspace, repoSlug string, id int, action PullRequestReviewAction) (PullRequestParticipant, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestParticipant{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequestParticipant{}, fmt.Errorf("pull request ID must be greater than zero")
	}

	var method string
	var suffix string
	switch action {
	case PullRequestReviewApprove:
		method = http.MethodPost
		suffix = "approve"
	case PullRequestReviewUnapprove:
		method = http.MethodDelete
		suffix = "approve"
	case PullRequestReviewRequestChanges:
		method = http.MethodPost
		suffix = "request-changes"
	case PullRequestReviewClearRequestChanges:
		method = http.MethodDelete
		suffix = "request-changes"
	default:
		return PullRequestParticipant{}, fmt.Errorf("unsupported pull request review action %q", action)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), id, suffix)
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
		return PullRequestParticipant{}, fmt.Errorf("decode pull request review participant: %w", err)
	}

	return participant, nil
}

func (c *Client) ListPullRequestActivity(ctx context.Context, workspace, repoSlug string, id int, options ListPullRequestActivityOptions) ([]PullRequestActivity, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return nil, fmt.Errorf("pull request ID must be greater than zero")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/activity?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), id, options.Limit)
	nextPath := path
	all := make([]PullRequestActivity, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pullRequestActivityListResponse
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
			return nil, fmt.Errorf("decode pull request activity list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) ListPullRequestCommits(ctx context.Context, workspace, repoSlug string, id int, options ListPullRequestCommitsOptions) ([]RepositoryCommit, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return nil, fmt.Errorf("pull request ID must be greater than zero")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/commits?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), id, options.Limit)
	nextPath := path
	all := make([]RepositoryCommit, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page repositoryCommitListResponse
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
			return nil, fmt.Errorf("decode pull request commits: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) ListPullRequestStatuses(ctx context.Context, workspace, repoSlug string, id int, options ListPullRequestStatusesOptions) ([]CommitStatus, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return nil, fmt.Errorf("pull request ID must be greater than zero")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", fmt.Sprintf("%d", options.Limit))
	if q := strings.TrimSpace(options.Query); q != "" {
		values.Set("q", q)
	}
	if sort := strings.TrimSpace(options.Sort); sort != "" {
		values.Set("sort", sort)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/statuses?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), id, values.Encode())
	nextPath := path
	all := make([]CommitStatus, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page commitStatusListResponse
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
			return nil, fmt.Errorf("decode pull request statuses: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (a PullRequestActivity) RequestChangesEvent() *PullRequestActivityEvent {
	if a.ChangesRequest != nil {
		return a.ChangesRequest
	}
	return a.RequestChanges
}
