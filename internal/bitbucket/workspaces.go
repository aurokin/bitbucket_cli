package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type WorkspaceMembership struct {
	Type         string                   `json:"type,omitempty"`
	Permission   string                   `json:"permission,omitempty"`
	AddedOn      string                   `json:"added_on,omitempty"`
	LastAccessed string                   `json:"last_accessed,omitempty"`
	User         RepositoryPermissionUser `json:"user,omitempty"`
	Workspace    Workspace                `json:"workspace,omitempty"`
	Links        PermissionLinks          `json:"links,omitempty"`
}

type WorkspaceRepositoryPermission struct {
	Type       string                        `json:"type,omitempty"`
	Permission string                        `json:"permission,omitempty"`
	User       RepositoryPermissionUser      `json:"user,omitempty"`
	Repository WorkspacePermissionRepository `json:"repository,omitempty"`
}

type WorkspacePermissionRepository struct {
	Name     string `json:"name,omitempty"`
	Slug     string `json:"slug,omitempty"`
	FullName string `json:"full_name,omitempty"`
	UUID     string `json:"uuid,omitempty"`
}

type workspaceMembershipListResponse struct {
	Values []WorkspaceMembership `json:"values"`
	Next   string                `json:"next,omitempty"`
}

type workspaceRepositoryPermissionListResponse struct {
	Values []WorkspaceRepositoryPermission `json:"values"`
	Next   string                          `json:"next,omitempty"`
}

func (c *Client) GetWorkspace(ctx context.Context, workspace string) (Workspace, error) {
	if workspace == "" {
		return Workspace{}, fmt.Errorf("workspace is required")
	}

	path := fmt.Sprintf("/workspaces/%s", url.PathEscape(workspace))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return Workspace{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Workspace{}, err
	}

	var item Workspace
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return Workspace{}, fmt.Errorf("decode workspace: %w", err)
	}

	return item, nil
}

func (c *Client) ListWorkspaceMembers(ctx context.Context, workspace string, limit int, query string) ([]WorkspaceMembership, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	if query != "" {
		values.Set("q", query)
	}

	nextPath := fmt.Sprintf("/workspaces/%s/members?%s", url.PathEscape(workspace), values.Encode())
	return listWorkspaceMemberships(ctx, c, nextPath, limit, "workspace members")
}

func (c *Client) GetWorkspaceMember(ctx context.Context, workspace, member string) (WorkspaceMembership, error) {
	if workspace == "" {
		return WorkspaceMembership{}, fmt.Errorf("workspace is required")
	}
	if member == "" {
		return WorkspaceMembership{}, fmt.Errorf("workspace member identifier is required")
	}

	path := fmt.Sprintf("/workspaces/%s/members/%s", url.PathEscape(workspace), url.PathEscape(member))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return WorkspaceMembership{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return WorkspaceMembership{}, err
	}

	var item WorkspaceMembership
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return WorkspaceMembership{}, fmt.Errorf("decode workspace member: %w", err)
	}

	return item, nil
}

func (c *Client) ListWorkspacePermissions(ctx context.Context, workspace string, limit int, query string) ([]WorkspaceMembership, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	if query != "" {
		values.Set("q", query)
	}

	nextPath := fmt.Sprintf("/workspaces/%s/permissions?%s", url.PathEscape(workspace), values.Encode())
	return listWorkspaceMemberships(ctx, c, nextPath, limit, "workspace permissions")
}

func (c *Client) ListWorkspaceRepositoryPermissions(ctx context.Context, workspace, repoSlug string, limit int, query, sort string) ([]WorkspaceRepositoryPermission, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	if query != "" {
		values.Set("q", query)
	}
	if sort != "" {
		values.Set("sort", sort)
	}

	path := fmt.Sprintf("/workspaces/%s/permissions/repositories", url.PathEscape(workspace))
	if repoSlug != "" {
		path = fmt.Sprintf("%s/%s", path, url.PathEscape(repoSlug))
	}
	nextPath := path + "?" + values.Encode()
	all := make([]WorkspaceRepositoryPermission, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page workspaceRepositoryPermissionListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode workspace repository permissions: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func listWorkspaceMemberships(ctx context.Context, client *Client, path string, limit int, label string) ([]WorkspaceMembership, error) {
	all := make([]WorkspaceMembership, 0)
	nextPath := path

	for nextPath != "" && len(all) < limit {
		resp, err := client.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page workspaceMembershipListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", label, err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
