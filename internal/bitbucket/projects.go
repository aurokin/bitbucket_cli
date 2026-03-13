package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Project struct {
	Type        string       `json:"type,omitempty"`
	Key         string       `json:"key,omitempty"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	IsPrivate   bool         `json:"is_private,omitempty"`
	UUID        string       `json:"uuid,omitempty"`
	Links       ProjectLinks `json:"links,omitempty"`
}

type ProjectLinks struct {
	HTML   Link `json:"html,omitempty"`
	Self   Link `json:"self,omitempty"`
	Avatar Link `json:"avatar,omitempty"`
}

type CreateProjectOptions struct {
	Name        string
	Description string
	IsPrivate   *bool
}

type UpdateProjectOptions struct {
	Name        string
	Description string
	IsPrivate   *bool
}

type DefaultReviewer struct {
	Type         string                   `json:"type,omitempty"`
	ReviewerType string                   `json:"reviewer_type,omitempty"`
	User         RepositoryPermissionUser `json:"user,omitempty"`
}

type ProjectUserPermission struct {
	Type       string                   `json:"type,omitempty"`
	Permission string                   `json:"permission,omitempty"`
	User       RepositoryPermissionUser `json:"user,omitempty"`
	Links      PermissionLinks          `json:"links,omitempty"`
}

type ProjectGroupPermission struct {
	Type       string                    `json:"type,omitempty"`
	Permission string                    `json:"permission,omitempty"`
	Group      RepositoryPermissionGroup `json:"group,omitempty"`
	Links      PermissionLinks           `json:"links,omitempty"`
}

type projectListResponse struct {
	Values []Project `json:"values"`
	Next   string    `json:"next,omitempty"`
}

type defaultReviewerListResponse struct {
	Values []DefaultReviewer `json:"values"`
	Next   string            `json:"next,omitempty"`
}

type projectUserPermissionListResponse struct {
	Values []ProjectUserPermission `json:"values"`
	Next   string                  `json:"next,omitempty"`
}

type projectGroupPermissionListResponse struct {
	Values []ProjectGroupPermission `json:"values"`
	Next   string                   `json:"next,omitempty"`
}

func (c *Client) ListProjects(ctx context.Context, workspace string, limit int) ([]Project, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/workspaces/%s/projects?%s", url.PathEscape(workspace), values.Encode())
	all := make([]Project, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page projectListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode project list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetProject(ctx context.Context, workspace, projectKey string) (Project, error) {
	if workspace == "" || projectKey == "" {
		return Project{}, fmt.Errorf("workspace and project key are required")
	}

	path := fmt.Sprintf("/workspaces/%s/projects/%s", url.PathEscape(workspace), url.PathEscape(projectKey))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Project{}, err
	}

	var item Project
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return Project{}, fmt.Errorf("decode project: %w", err)
	}
	return item, nil
}

func (c *Client) CreateProject(ctx context.Context, workspace, projectKey string, options CreateProjectOptions) (Project, error) {
	if workspace == "" || projectKey == "" {
		return Project{}, fmt.Errorf("workspace and project key are required")
	}
	if options.Name == "" {
		return Project{}, fmt.Errorf("project name is required")
	}

	body := map[string]any{
		"key":  projectKey,
		"name": options.Name,
	}
	if options.Description != "" {
		body["description"] = options.Description
	}
	if options.IsPrivate != nil {
		body["is_private"] = *options.IsPrivate
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Project{}, fmt.Errorf("marshal create project request: %w", err)
	}

	path := fmt.Sprintf("/workspaces/%s/projects", url.PathEscape(workspace))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Project{}, err
	}

	var item Project
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return Project{}, fmt.Errorf("decode created project: %w", err)
	}
	return item, nil
}

func (c *Client) UpdateProject(ctx context.Context, workspace, projectKey string, options UpdateProjectOptions) (Project, error) {
	if workspace == "" || projectKey == "" {
		return Project{}, fmt.Errorf("workspace and project key are required")
	}
	if options.Name == "" && options.Description == "" && options.IsPrivate == nil {
		return Project{}, fmt.Errorf("at least one project field must be updated")
	}
	if options.Name == "" {
		current, err := c.GetProject(ctx, workspace, projectKey)
		if err != nil {
			return Project{}, err
		}
		options.Name = current.Name
	}

	body := map[string]any{}
	if options.Name != "" {
		body["name"] = options.Name
	}
	if options.Description != "" {
		body["description"] = options.Description
	}
	if options.IsPrivate != nil {
		body["is_private"] = *options.IsPrivate
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Project{}, fmt.Errorf("marshal update project request: %w", err)
	}

	path := fmt.Sprintf("/workspaces/%s/projects/%s", url.PathEscape(workspace), url.PathEscape(projectKey))
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return Project{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Project{}, err
	}

	var item Project
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return Project{}, fmt.Errorf("decode updated project: %w", err)
	}
	return item, nil
}

func (c *Client) DeleteProject(ctx context.Context, workspace, projectKey string) error {
	if workspace == "" || projectKey == "" {
		return fmt.Errorf("workspace and project key are required")
	}

	path := fmt.Sprintf("/workspaces/%s/projects/%s", url.PathEscape(workspace), url.PathEscape(projectKey))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}

func (c *Client) ListProjectDefaultReviewers(ctx context.Context, workspace, projectKey string, limit int) ([]DefaultReviewer, error) {
	if workspace == "" || projectKey == "" {
		return nil, fmt.Errorf("workspace and project key are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/workspaces/%s/projects/%s/default-reviewers?%s", url.PathEscape(workspace), url.PathEscape(projectKey), values.Encode())
	all := make([]DefaultReviewer, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page defaultReviewerListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode project default reviewers: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) ListProjectUserPermissions(ctx context.Context, workspace, projectKey string, limit int) ([]ProjectUserPermission, error) {
	if workspace == "" || projectKey == "" {
		return nil, fmt.Errorf("workspace and project key are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/workspaces/%s/projects/%s/permissions-config/users?%s", url.PathEscape(workspace), url.PathEscape(projectKey), values.Encode())
	all := make([]ProjectUserPermission, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page projectUserPermissionListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode project user permissions: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetProjectUserPermission(ctx context.Context, workspace, projectKey, accountID string) (ProjectUserPermission, error) {
	if workspace == "" || projectKey == "" {
		return ProjectUserPermission{}, fmt.Errorf("workspace and project key are required")
	}
	if accountID == "" {
		return ProjectUserPermission{}, fmt.Errorf("project user permission account ID is required")
	}

	path := fmt.Sprintf("/workspaces/%s/projects/%s/permissions-config/users/%s", url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(accountID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return ProjectUserPermission{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return ProjectUserPermission{}, err
	}

	var item ProjectUserPermission
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return ProjectUserPermission{}, fmt.Errorf("decode project user permission: %w", err)
	}
	return item, nil
}

func (c *Client) ListProjectGroupPermissions(ctx context.Context, workspace, projectKey string, limit int) ([]ProjectGroupPermission, error) {
	if workspace == "" || projectKey == "" {
		return nil, fmt.Errorf("workspace and project key are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/workspaces/%s/projects/%s/permissions-config/groups?%s", url.PathEscape(workspace), url.PathEscape(projectKey), values.Encode())
	all := make([]ProjectGroupPermission, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page projectGroupPermissionListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode project group permissions: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetProjectGroupPermission(ctx context.Context, workspace, projectKey, groupSlug string) (ProjectGroupPermission, error) {
	if workspace == "" || projectKey == "" {
		return ProjectGroupPermission{}, fmt.Errorf("workspace and project key are required")
	}
	if groupSlug == "" {
		return ProjectGroupPermission{}, fmt.Errorf("project group permission group slug is required")
	}

	path := fmt.Sprintf("/workspaces/%s/projects/%s/permissions-config/groups/%s", url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(groupSlug))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return ProjectGroupPermission{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return ProjectGroupPermission{}, err
	}

	var item ProjectGroupPermission
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return ProjectGroupPermission{}, fmt.Errorf("decode project group permission: %w", err)
	}
	return item, nil
}
