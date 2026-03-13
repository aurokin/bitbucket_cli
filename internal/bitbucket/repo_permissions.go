package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type RepositoryPermissionUser struct {
	DisplayName string `json:"display_name,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	UUID        string `json:"uuid,omitempty"`
}

type RepositoryPermissionGroup struct {
	Name     string `json:"name,omitempty"`
	Slug     string `json:"slug,omitempty"`
	FullSlug string `json:"full_slug,omitempty"`
}

type RepositoryUserPermission struct {
	Type       string                   `json:"type,omitempty"`
	Permission string                   `json:"permission,omitempty"`
	User       RepositoryPermissionUser `json:"user,omitempty"`
	Links      PermissionLinks          `json:"links,omitempty"`
}

type RepositoryGroupPermission struct {
	Type       string                    `json:"type,omitempty"`
	Permission string                    `json:"permission,omitempty"`
	Group      RepositoryPermissionGroup `json:"group,omitempty"`
	Links      PermissionLinks           `json:"links,omitempty"`
}

type PermissionLinks struct {
	Self Link `json:"self,omitempty"`
}

type repositoryUserPermissionListResponse struct {
	Values []RepositoryUserPermission `json:"values"`
	Next   string                     `json:"next,omitempty"`
}

type repositoryGroupPermissionListResponse struct {
	Values []RepositoryGroupPermission `json:"values"`
	Next   string                      `json:"next,omitempty"`
}

func (c *Client) ListRepositoryUserPermissions(ctx context.Context, workspace, repoSlug string, limit int) ([]RepositoryUserPermission, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository slug are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/permissions-config/users?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]RepositoryUserPermission, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page repositoryUserPermissionListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode repository user permission list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetRepositoryUserPermission(ctx context.Context, workspace, repoSlug, accountID string) (RepositoryUserPermission, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryUserPermission{}, fmt.Errorf("workspace and repository slug are required")
	}
	if accountID == "" {
		return RepositoryUserPermission{}, fmt.Errorf("repository user permission account ID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/permissions-config/users/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(accountID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryUserPermission{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return RepositoryUserPermission{}, err
	}

	var permission RepositoryUserPermission
	if err := json.NewDecoder(resp.Body).Decode(&permission); err != nil {
		return RepositoryUserPermission{}, fmt.Errorf("decode repository user permission: %w", err)
	}
	return permission, nil
}

func (c *Client) ListRepositoryGroupPermissions(ctx context.Context, workspace, repoSlug string, limit int) ([]RepositoryGroupPermission, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository slug are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/permissions-config/groups?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]RepositoryGroupPermission, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page repositoryGroupPermissionListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode repository group permission list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetRepositoryGroupPermission(ctx context.Context, workspace, repoSlug, groupSlug string) (RepositoryGroupPermission, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryGroupPermission{}, fmt.Errorf("workspace and repository slug are required")
	}
	if groupSlug == "" {
		return RepositoryGroupPermission{}, fmt.Errorf("repository group permission group slug is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/permissions-config/groups/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(groupSlug))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryGroupPermission{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return RepositoryGroupPermission{}, err
	}

	var permission RepositoryGroupPermission
	if err := json.NewDecoder(resp.Body).Decode(&permission); err != nil {
		return RepositoryGroupPermission{}, fmt.Errorf("decode repository group permission: %w", err)
	}
	return permission, nil
}
