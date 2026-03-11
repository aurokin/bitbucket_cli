package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ListPullRequestsOptions struct {
	State string
	Limit int
}

type CreatePullRequestOptions struct {
	Title             string
	Description       string
	SourceBranch      string
	DestinationBranch string
	CloseSourceBranch bool
	Draft             bool
	ReuseExisting     bool
}

type MergePullRequestOptions struct {
	Message           string
	CloseSourceBranch bool
	MergeStrategy     string
	PollInterval      time.Duration
	PollTimeout       time.Duration
}

type PullRequest struct {
	ID                int                      `json:"id"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	State             string                   `json:"state"`
	Author            PullRequestActor         `json:"author"`
	Reviewers         []PullRequestActor       `json:"reviewers,omitempty"`
	Participants      []PullRequestParticipant `json:"participants,omitempty"`
	Source            PullRequestRef           `json:"source"`
	Destination       PullRequestRef           `json:"destination"`
	CloseSourceBranch bool                     `json:"close_source_branch,omitempty"`
	Queued            bool                     `json:"queued,omitempty"`
	MergeCommit       PullRequestCommit        `json:"merge_commit,omitempty"`
	Draft             bool                     `json:"draft,omitempty"`
	CommentCount      int                      `json:"comment_count,omitempty"`
	TaskCount         int                      `json:"task_count,omitempty"`
	UpdatedOn         string                   `json:"updated_on,omitempty"`
	CreatedOn         string                   `json:"created_on,omitempty"`
	Links             PullRequestLinks         `json:"links,omitempty"`
}

type PullRequestActor struct {
	DisplayName string `json:"display_name,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
}

type PullRequestParticipant struct {
	User     PullRequestActor `json:"user"`
	Role     string           `json:"role,omitempty"`
	Approved bool             `json:"approved,omitempty"`
	State    string           `json:"state,omitempty"`
}

type PullRequestRef struct {
	Branch PullRequestBranch `json:"branch"`
	Commit PullRequestCommit `json:"commit"`
	Repo   PullRequestRepo   `json:"repository"`
}

type PullRequestBranch struct {
	Name                 string   `json:"name"`
	MergeStrategies      []string `json:"merge_strategies,omitempty"`
	DefaultMergeStrategy string   `json:"default_merge_strategy,omitempty"`
}

type PullRequestCommit struct {
	Hash string `json:"hash,omitempty"`
}

type PullRequestRepo struct {
	Name     string `json:"name,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

type PullRequestLinks struct {
	HTML Link `json:"html"`
}

type Link struct {
	Href string `json:"href"`
}

type pullRequestListResponse struct {
	Values []PullRequest `json:"values"`
	Next   string        `json:"next,omitempty"`
}

type mergeTaskStatusResponse struct {
	TaskStatus  string      `json:"task_status"`
	MergeResult PullRequest `json:"merge_result,omitempty"`
}

const (
	mergeTaskStatusPending = "PENDING"
	mergeTaskStatusSuccess = "SUCCESS"

	defaultMergePollInterval = time.Second
	defaultMergePollTimeout  = 2 * time.Minute
)

func (c *Client) ListPullRequests(ctx context.Context, workspace, repoSlug string, options ListPullRequestsOptions) ([]PullRequest, error) {
	if options.Limit <= 0 {
		options.Limit = 20
	}

	path, err := listPullRequestsPath(workspace, repoSlug, options)
	if err != nil {
		return nil, err
	}

	var all []PullRequest
	nextPath := path

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pullRequestListResponse
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
			return nil, fmt.Errorf("decode pull request list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) GetPullRequest(ctx context.Context, workspace, repoSlug string, id int) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequest{}, fmt.Errorf("pull request ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, workspace, repoSlug string, options CreatePullRequestOptions) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if options.Title == "" {
		return PullRequest{}, fmt.Errorf("pull request title is required")
	}
	if options.SourceBranch == "" {
		return PullRequest{}, fmt.Errorf("source branch is required")
	}
	if options.DestinationBranch == "" {
		return PullRequest{}, fmt.Errorf("destination branch is required")
	}

	if options.ReuseExisting {
		existing, err := c.ListPullRequests(ctx, workspace, repoSlug, ListPullRequestsOptions{
			State: "OPEN",
			Limit: 50,
		})
		if err != nil {
			return PullRequest{}, err
		}
		for _, pr := range existing {
			if pr.Title == options.Title && pr.Source.Branch.Name == options.SourceBranch && pr.Destination.Branch.Name == options.DestinationBranch {
				return pr, nil
			}
		}
	}

	body := map[string]any{
		"title": options.Title,
		"source": map[string]any{
			"branch": map[string]string{
				"name": options.SourceBranch,
			},
		},
		"destination": map[string]any{
			"branch": map[string]string{
				"name": options.DestinationBranch,
			},
		},
	}
	if options.Description != "" {
		body["description"] = options.Description
	}
	if options.CloseSourceBranch {
		body["close_source_branch"] = true
	}
	if options.Draft {
		body["draft"] = true
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PullRequest{}, fmt.Errorf("marshal create pull request request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode created pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) MergePullRequest(ctx context.Context, workspace, repoSlug string, id int, options MergePullRequestOptions) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequest{}, fmt.Errorf("pull request ID must be greater than zero")
	}

	body := map[string]any{
		"type": "pullrequest",
	}
	if strings.TrimSpace(options.Message) != "" {
		body["message"] = strings.TrimSpace(options.Message)
	}
	if options.CloseSourceBranch {
		body["close_source_branch"] = true
	}
	if strings.TrimSpace(options.MergeStrategy) != "" {
		body["merge_strategy"] = strings.TrimSpace(options.MergeStrategy)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PullRequest{}, fmt.Errorf("marshal merge pull request request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge?async=true", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		location := strings.TrimSpace(resp.Header.Get("Location"))
		if location == "" {
			return PullRequest{}, fmt.Errorf("bitbucket API returned 202 Accepted without a merge task location")
		}
		return c.waitForMergeTask(ctx, location, options.PollInterval, options.PollTimeout)
	default:
		if err := requireSuccess(resp); err != nil {
			return PullRequest{}, err
		}
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode merged pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) waitForMergeTask(ctx context.Context, taskURL string, interval, timeout time.Duration) (PullRequest, error) {
	if interval <= 0 {
		interval = defaultMergePollInterval
	}
	if timeout <= 0 {
		timeout = defaultMergePollTimeout
	}

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		resp, err := c.Do(pollCtx, http.MethodGet, taskURL, nil, nil)
		if err != nil {
			return PullRequest{}, err
		}

		var status mergeTaskStatusResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&status)
		}()
		if err != nil {
			return PullRequest{}, fmt.Errorf("check merge task status: %w", err)
		}

		switch status.TaskStatus {
		case mergeTaskStatusSuccess:
			if status.MergeResult.ID == 0 {
				return PullRequest{}, fmt.Errorf("merge task completed without a merged pull request payload")
			}
			return status.MergeResult, nil
		case "", mergeTaskStatusPending:
			if status.MergeResult.ID != 0 {
				return status.MergeResult, nil
			}
		default:
			return PullRequest{}, fmt.Errorf("merge task returned unexpected status %q", status.TaskStatus)
		}

		select {
		case <-pollCtx.Done():
			return PullRequest{}, fmt.Errorf("timed out waiting for merge task: %w", pollCtx.Err())
		case <-time.After(interval):
		}
	}
}

func listPullRequestsPath(workspace, repoSlug string, options ListPullRequestsOptions) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(options.Limit))
	if options.State != "" && options.State != "ALL" {
		values.Set("state", options.State)
	}

	return fmt.Sprintf("/repositories/%s/%s/pullrequests?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode()), nil
}
