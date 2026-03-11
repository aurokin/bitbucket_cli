package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Issue struct {
	ID        int          `json:"id"`
	Title     string       `json:"title"`
	State     string       `json:"state,omitempty"`
	Kind      string       `json:"kind,omitempty"`
	Priority  string       `json:"priority,omitempty"`
	Content   IssueContent `json:"content,omitempty"`
	Reporter  IssueActor   `json:"reporter,omitempty"`
	Assignee  IssueActor   `json:"assignee,omitempty"`
	CreatedOn string       `json:"created_on,omitempty"`
	UpdatedOn string       `json:"updated_on,omitempty"`
	Links     IssueLinks   `json:"links,omitempty"`
}

type IssueContent struct {
	Raw string `json:"raw,omitempty"`
}

type IssueActor struct {
	DisplayName string `json:"display_name,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
}

type IssueLinks struct {
	HTML Link `json:"html"`
}

type ListIssuesOptions struct {
	Query string
	Sort  string
	State string
	Limit int
}

type CreateIssueOptions struct {
	Title    string
	Body     string
	Kind     string
	Priority string
}

type UpdateIssueOptions struct {
	Title    string
	Body     string
	State    string
	Kind     string
	Priority string
}

type issueListResponse struct {
	Values []Issue `json:"values"`
	Next   string  `json:"next,omitempty"`
}

type IssueChangeOptions struct {
	State   string
	Message string
}

func (c *Client) ListIssues(ctx context.Context, workspace, repoSlug string, options ListIssuesOptions) ([]Issue, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(options.Limit))
	if options.Query != "" {
		values.Set("q", options.Query)
	}
	if options.Sort != "" {
		values.Set("sort", options.Sort)
	}
	if options.State != "" && options.State != "ALL" {
		values.Set("state", options.State)
	}

	nextPath := fmt.Sprintf("/repositories/%s/%s/issues?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	var all []Issue

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page issueListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode issue list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) GetIssue(ctx context.Context, workspace, repoSlug string, id int) (Issue, error) {
	if workspace == "" || repoSlug == "" {
		return Issue{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return Issue{}, fmt.Errorf("issue ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return Issue{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Issue{}, err
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return Issue{}, fmt.Errorf("decode issue: %w", err)
	}

	return issue, nil
}

func (c *Client) CreateIssue(ctx context.Context, workspace, repoSlug string, options CreateIssueOptions) (Issue, error) {
	if workspace == "" || repoSlug == "" {
		return Issue{}, fmt.Errorf("workspace and repository are required")
	}
	if options.Title == "" {
		return Issue{}, fmt.Errorf("issue title is required")
	}

	body := map[string]any{
		"title": options.Title,
	}
	if options.Body != "" {
		body["content"] = map[string]string{"raw": options.Body}
	}
	if options.Kind != "" {
		body["kind"] = options.Kind
	}
	if options.Priority != "" {
		body["priority"] = options.Priority
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return Issue{}, fmt.Errorf("marshal create issue request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return Issue{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Issue{}, err
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return Issue{}, fmt.Errorf("decode created issue: %w", err)
	}

	return issue, nil
}

func (c *Client) UpdateIssue(ctx context.Context, workspace, repoSlug string, id int, options UpdateIssueOptions) (Issue, error) {
	if workspace == "" || repoSlug == "" {
		return Issue{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return Issue{}, fmt.Errorf("issue ID must be greater than zero")
	}

	body := map[string]any{}
	if options.Title != "" {
		body["title"] = options.Title
	}
	if options.Body != "" {
		body["content"] = map[string]string{"raw": options.Body}
	}
	if options.State != "" {
		body["state"] = options.State
	}
	if options.Kind != "" {
		body["kind"] = options.Kind
	}
	if options.Priority != "" {
		body["priority"] = options.Priority
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return Issue{}, fmt.Errorf("marshal update issue request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return Issue{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Issue{}, err
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return Issue{}, fmt.Errorf("decode updated issue: %w", err)
	}

	return issue, nil
}

func (c *Client) ChangeIssueState(ctx context.Context, workspace, repoSlug string, id int, options IssueChangeOptions) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return fmt.Errorf("issue ID must be greater than zero")
	}
	if options.State == "" {
		return fmt.Errorf("issue state is required")
	}

	body := map[string]any{
		"changes": map[string]any{
			"state": map[string]string{
				"new": options.State,
			},
		},
	}
	if options.Message != "" {
		body["message"] = map[string]string{"raw": options.Message}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal issue change request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/changes", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return err
	}

	return nil
}
