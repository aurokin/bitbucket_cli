package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type IssueMilestone struct {
	ID    int        `json:"id"`
	Name  string     `json:"name,omitempty"`
	Links IssueLinks `json:"links,omitempty"`
}

type IssueComponent struct {
	ID    int        `json:"id"`
	Name  string     `json:"name,omitempty"`
	Links IssueLinks `json:"links,omitempty"`
}

type issueMilestoneListResponse struct {
	Values []IssueMilestone `json:"values"`
	Next   string           `json:"next,omitempty"`
}

type issueComponentListResponse struct {
	Values []IssueComponent `json:"values"`
	Next   string           `json:"next,omitempty"`
}

func (c *Client) ListIssueMilestones(ctx context.Context, workspace, repoSlug string, limit int) ([]IssueMilestone, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 100
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/milestones?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]IssueMilestone, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page issueMilestoneListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode issue milestone list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetIssueMilestone(ctx context.Context, workspace, repoSlug string, milestoneID int) (IssueMilestone, error) {
	if workspace == "" || repoSlug == "" {
		return IssueMilestone{}, fmt.Errorf("workspace and repository are required")
	}
	if milestoneID <= 0 {
		return IssueMilestone{}, fmt.Errorf("issue milestone ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/milestones/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), milestoneID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return IssueMilestone{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return IssueMilestone{}, err
	}

	var milestone IssueMilestone
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return IssueMilestone{}, fmt.Errorf("decode issue milestone: %w", err)
	}
	return milestone, nil
}

func (c *Client) ListIssueComponents(ctx context.Context, workspace, repoSlug string, limit int) ([]IssueComponent, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 100
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/components?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]IssueComponent, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page issueComponentListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode issue component list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetIssueComponent(ctx context.Context, workspace, repoSlug string, componentID int) (IssueComponent, error) {
	if workspace == "" || repoSlug == "" {
		return IssueComponent{}, fmt.Errorf("workspace and repository are required")
	}
	if componentID <= 0 {
		return IssueComponent{}, fmt.Errorf("issue component ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/components/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), componentID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return IssueComponent{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return IssueComponent{}, err
	}

	var component IssueComponent
	if err := json.NewDecoder(resp.Body).Decode(&component); err != nil {
		return IssueComponent{}, fmt.Errorf("decode issue component: %w", err)
	}
	return component, nil
}
