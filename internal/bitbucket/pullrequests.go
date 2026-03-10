package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ListPullRequestsOptions struct {
	State string
	Limit int
}

type PullRequest struct {
	ID          int              `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	State       string           `json:"state"`
	Author      PullRequestActor `json:"author"`
	Source      PullRequestRef   `json:"source"`
	Destination PullRequestRef   `json:"destination"`
	UpdatedOn   string           `json:"updated_on,omitempty"`
	CreatedOn   string           `json:"created_on,omitempty"`
	Links       PullRequestLinks `json:"links,omitempty"`
}

type PullRequestActor struct {
	DisplayName string `json:"display_name,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
}

type PullRequestRef struct {
	Branch PullRequestBranch `json:"branch"`
	Commit PullRequestCommit `json:"commit"`
	Repo   PullRequestRepo   `json:"repository"`
}

type PullRequestBranch struct {
	Name string `json:"name"`
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
