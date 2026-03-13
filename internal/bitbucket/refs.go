package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ListBranchesOptions struct {
	Limit int
	Query string
	Sort  string
}

type ListTagsOptions struct {
	Limit int
	Query string
	Sort  string
}

type CreateBranchOptions struct {
	Name   string
	Target string
}

type CreateTagOptions struct {
	Name    string
	Target  string
	Message string
}

type RefLinks struct {
	Self    Link `json:"self,omitempty"`
	HTML    Link `json:"html,omitempty"`
	Commits Link `json:"commits,omitempty"`
}

type RepositoryTag struct {
	Name    string                 `json:"name,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Message string                 `json:"message,omitempty"`
	Date    string                 `json:"date,omitempty"`
	Tagger  RepositoryCommitAuthor `json:"tagger,omitempty"`
	Target  RepositoryCommit       `json:"target,omitempty"`
	Links   RefLinks               `json:"links,omitempty"`
}

type branchListResponse struct {
	Values []RepositoryBranch `json:"values"`
	Next   string             `json:"next,omitempty"`
}

type tagListResponse struct {
	Values []RepositoryTag `json:"values"`
	Next   string          `json:"next,omitempty"`
}

func (c *Client) ListBranches(ctx context.Context, workspace, repoSlug string, options ListBranchesOptions) ([]RepositoryBranch, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	query := url.Values{}
	query.Set("pagelen", strconv.Itoa(options.Limit))
	if strings.TrimSpace(options.Query) != "" {
		query.Set("q", strings.TrimSpace(options.Query))
	}
	if strings.TrimSpace(options.Sort) != "" {
		query.Set("sort", strings.TrimSpace(options.Sort))
	}
	nextPath := fmt.Sprintf("/repositories/%s/%s/refs/branches?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), query.Encode())
	all := make([]RepositoryBranch, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page branchListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode branch list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) GetBranch(ctx context.Context, workspace, repoSlug, name string) (RepositoryBranch, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryBranch{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(name) == "" {
		return RepositoryBranch{}, fmt.Errorf("branch name is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(name))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryBranch{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return RepositoryBranch{}, err
	}

	var branch RepositoryBranch
	if err := json.NewDecoder(resp.Body).Decode(&branch); err != nil {
		return RepositoryBranch{}, fmt.Errorf("decode branch: %w", err)
	}
	return branch, nil
}

func (c *Client) CreateBranch(ctx context.Context, workspace, repoSlug string, options CreateBranchOptions) (RepositoryBranch, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryBranch{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(options.Name) == "" {
		return RepositoryBranch{}, fmt.Errorf("branch name is required")
	}
	if strings.TrimSpace(options.Target) == "" {
		return RepositoryBranch{}, fmt.Errorf("branch target is required")
	}

	payload, err := json.Marshal(map[string]any{
		"name": options.Name,
		"target": map[string]string{
			"hash": options.Target,
		},
	})
	if err != nil {
		return RepositoryBranch{}, fmt.Errorf("marshal create branch request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/refs/branches", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return RepositoryBranch{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return RepositoryBranch{}, err
	}

	var branch RepositoryBranch
	if err := json.NewDecoder(resp.Body).Decode(&branch); err != nil {
		return RepositoryBranch{}, fmt.Errorf("decode created branch: %w", err)
	}
	return branch, nil
}

func (c *Client) DeleteBranch(ctx context.Context, workspace, repoSlug, name string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("branch name is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(name))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}

func (c *Client) ListTags(ctx context.Context, workspace, repoSlug string, options ListTagsOptions) ([]RepositoryTag, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	query := url.Values{}
	query.Set("pagelen", strconv.Itoa(options.Limit))
	if strings.TrimSpace(options.Query) != "" {
		query.Set("q", strings.TrimSpace(options.Query))
	}
	if strings.TrimSpace(options.Sort) != "" {
		query.Set("sort", strings.TrimSpace(options.Sort))
	}
	nextPath := fmt.Sprintf("/repositories/%s/%s/refs/tags?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), query.Encode())
	all := make([]RepositoryTag, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page tagListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode tag list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) GetTag(ctx context.Context, workspace, repoSlug, name string) (RepositoryTag, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryTag{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(name) == "" {
		return RepositoryTag{}, fmt.Errorf("tag name is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/refs/tags/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(name))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryTag{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return RepositoryTag{}, err
	}

	var tag RepositoryTag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return RepositoryTag{}, fmt.Errorf("decode tag: %w", err)
	}
	return tag, nil
}

func (c *Client) CreateTag(ctx context.Context, workspace, repoSlug string, options CreateTagOptions) (RepositoryTag, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryTag{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(options.Name) == "" {
		return RepositoryTag{}, fmt.Errorf("tag name is required")
	}
	if strings.TrimSpace(options.Target) == "" {
		return RepositoryTag{}, fmt.Errorf("tag target is required")
	}

	body := map[string]any{
		"name": options.Name,
		"target": map[string]string{
			"hash": options.Target,
		},
	}
	if strings.TrimSpace(options.Message) != "" {
		body["message"] = strings.TrimSpace(options.Message)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return RepositoryTag{}, fmt.Errorf("marshal create tag request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/refs/tags", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return RepositoryTag{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return RepositoryTag{}, err
	}

	var tag RepositoryTag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return RepositoryTag{}, fmt.Errorf("decode created tag: %w", err)
	}
	return tag, nil
}

func (c *Client) DeleteTag(ctx context.Context, workspace, repoSlug, name string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("tag name is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/refs/tags/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(name))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}
