package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Repository struct {
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	FullName    string            `json:"full_name,omitempty"`
	Description string            `json:"description,omitempty"`
	IsPrivate   bool              `json:"is_private"`
	UpdatedOn   string            `json:"updated_on,omitempty"`
	Project     RepositoryProject `json:"project,omitempty"`
	MainBranch  RepositoryBranch  `json:"mainbranch,omitempty"`
	Links       RepositoryLinks   `json:"links,omitempty"`
}

type RepositoryProject struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

type RepositoryBranch struct {
	Name string `json:"name,omitempty"`
}

type RepositoryLinks struct {
	HTML  Link               `json:"html"`
	Clone []NamedCloneTarget `json:"clone,omitempty"`
}

type NamedCloneTarget struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

type CreateRepositoryOptions struct {
	Name          string
	Description   string
	ProjectKey    string
	IsPrivate     bool
	ReuseExisting bool
}

type Workspace struct {
	Slug string `json:"slug"`
	Name string `json:"name,omitempty"`
}

type workspaceListResponse struct {
	Values []Workspace `json:"values"`
}

type ListRepositoriesOptions struct {
	Query string
	Sort  string
	Limit int
}

type repositoryListResponse struct {
	Values []Repository `json:"values"`
	Next   string       `json:"next,omitempty"`
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/workspaces?role=member", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return nil, err
	}

	var payload workspaceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode workspaces: %w", err)
	}

	return payload.Values, nil
}

func (c *Client) GetRepository(ctx context.Context, workspace, repoSlug string) (Repository, error) {
	path := fmt.Sprintf("/repositories/%s/%s", url.PathEscape(workspace), url.PathEscape(repoSlug))

	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return Repository{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Repository{}, err
	}

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return Repository{}, fmt.Errorf("decode repository: %w", err)
	}

	return repo, nil
}

func (c *Client) ListRepositories(ctx context.Context, workspace string, options ListRepositoriesOptions) ([]Repository, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace is required")
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

	nextPath := fmt.Sprintf("/repositories/%s?%s", url.PathEscape(workspace), values.Encode())
	var all []Repository

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page repositoryListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode repository list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) CreateRepository(ctx context.Context, workspace, repoSlug string, options CreateRepositoryOptions) (Repository, error) {
	if workspace == "" || repoSlug == "" {
		return Repository{}, fmt.Errorf("workspace and repository slug are required")
	}

	if options.ReuseExisting {
		repo, err := c.GetRepository(ctx, workspace, repoSlug)
		if err == nil {
			return repo, nil
		}
	}

	body := map[string]any{
		"scm":        "git",
		"is_private": options.IsPrivate,
	}
	if options.Name != "" {
		body["name"] = options.Name
	}
	if options.Description != "" {
		body["description"] = options.Description
	}
	if options.ProjectKey != "" {
		body["project"] = map[string]string{
			"key": options.ProjectKey,
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return Repository{}, fmt.Errorf("marshal create repository request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return Repository{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Repository{}, err
	}

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return Repository{}, fmt.Errorf("decode created repository: %w", err)
	}

	return repo, nil
}

func (c *Client) DeleteRepository(ctx context.Context, workspace, repoSlug string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository slug are required")
	}

	path := fmt.Sprintf("/repositories/%s/%s", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return err
	}

	return nil
}
