package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type RepositoryWebhook struct {
	UUID        string   `json:"uuid,omitempty"`
	URL         string   `json:"url,omitempty"`
	Description string   `json:"description,omitempty"`
	SubjectType string   `json:"subject_type,omitempty"`
	Active      bool     `json:"active"`
	CreatedAt   string   `json:"created_at,omitempty"`
	Events      []string `json:"events,omitempty"`
	SecretSet   bool     `json:"secret_set,omitempty"`
}

type RepositoryWebhookMutationOptions struct {
	URL         string
	Description string
	Active      *bool
	Events      []string
	Secret      *string
	ClearSecret bool
}

type RepositoryDeployKey struct {
	ID         int                  `json:"id,omitempty"`
	Label      string               `json:"label,omitempty"`
	Key        string               `json:"key,omitempty"`
	CreatedOn  string               `json:"created_on,omitempty"`
	LastUsed   string               `json:"last_used,omitempty"`
	Comment    string               `json:"comment,omitempty"`
	Type       string               `json:"type,omitempty"`
	Repository RepositoryKeySubject `json:"repository,omitempty"`
	Links      RepositoryKeyLinks   `json:"links,omitempty"`
}

type RepositoryKeySubject struct {
	Name     string `json:"name,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

type RepositoryKeyLinks struct {
	Self Link `json:"self,omitempty"`
}

type CreateRepositoryDeployKeyOptions struct {
	Label   string
	Key     string
	Comment string
}

type repositoryWebhookListResponse struct {
	Values []RepositoryWebhook `json:"values"`
	Next   string              `json:"next,omitempty"`
}

type repositoryDeployKeyListResponse struct {
	Values []RepositoryDeployKey `json:"values"`
	Next   string                `json:"next,omitempty"`
}

func (c *Client) ListRepositoryWebhooks(ctx context.Context, workspace, repoSlug string, limit int) ([]RepositoryWebhook, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository slug are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/hooks?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]RepositoryWebhook, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page repositoryWebhookListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode repository webhook list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetRepositoryWebhook(ctx context.Context, workspace, repoSlug, webhookID string) (RepositoryWebhook, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryWebhook{}, fmt.Errorf("workspace and repository slug are required")
	}
	if webhookID == "" {
		return RepositoryWebhook{}, fmt.Errorf("repository webhook ID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/hooks/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(webhookID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryWebhook{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return RepositoryWebhook{}, err
	}

	var hook RepositoryWebhook
	if err := json.NewDecoder(resp.Body).Decode(&hook); err != nil {
		return RepositoryWebhook{}, fmt.Errorf("decode repository webhook: %w", err)
	}
	return hook, nil
}

func (c *Client) CreateRepositoryWebhook(ctx context.Context, workspace, repoSlug string, options RepositoryWebhookMutationOptions) (RepositoryWebhook, error) {
	return c.mutateRepositoryWebhook(ctx, http.MethodPost, workspace, repoSlug, "", options)
}

func (c *Client) UpdateRepositoryWebhook(ctx context.Context, workspace, repoSlug, webhookID string, options RepositoryWebhookMutationOptions) (RepositoryWebhook, error) {
	existing, err := c.GetRepositoryWebhook(ctx, workspace, repoSlug, webhookID)
	if err != nil {
		return RepositoryWebhook{}, err
	}
	if options.URL == "" {
		options.URL = existing.URL
	}
	if options.Description == "" {
		options.Description = existing.Description
	}
	if options.Active == nil {
		active := existing.Active
		options.Active = &active
	}
	if len(options.Events) == 0 {
		options.Events = append([]string(nil), existing.Events...)
	}
	return c.mutateRepositoryWebhook(ctx, http.MethodPut, workspace, repoSlug, webhookID, options)
}

func (c *Client) mutateRepositoryWebhook(ctx context.Context, method, workspace, repoSlug, webhookID string, options RepositoryWebhookMutationOptions) (RepositoryWebhook, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryWebhook{}, fmt.Errorf("workspace and repository slug are required")
	}
	if method == http.MethodPut && webhookID == "" {
		return RepositoryWebhook{}, fmt.Errorf("repository webhook ID is required")
	}
	if method == http.MethodPost {
		if options.URL == "" {
			return RepositoryWebhook{}, fmt.Errorf("repository webhook URL is required")
		}
		if len(options.Events) == 0 {
			return RepositoryWebhook{}, fmt.Errorf("at least one repository webhook event is required")
		}
	}

	body := map[string]any{}
	if options.URL != "" {
		body["url"] = options.URL
	}
	if options.Description != "" {
		body["description"] = options.Description
	}
	if options.Active != nil {
		body["active"] = *options.Active
	}
	if len(options.Events) > 0 {
		body["events"] = options.Events
	}
	if options.Secret != nil {
		body["secret"] = *options.Secret
	}
	if options.ClearSecret {
		body["secret"] = nil
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return RepositoryWebhook{}, fmt.Errorf("marshal repository webhook request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/hooks", url.PathEscape(workspace), url.PathEscape(repoSlug))
	if method == http.MethodPut {
		path = path + "/" + url.PathEscape(webhookID)
	}
	resp, err := c.Do(ctx, method, path, payload, nil)
	if err != nil {
		return RepositoryWebhook{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return RepositoryWebhook{}, err
	}

	var hook RepositoryWebhook
	if err := json.NewDecoder(resp.Body).Decode(&hook); err != nil {
		return RepositoryWebhook{}, fmt.Errorf("decode repository webhook: %w", err)
	}
	return hook, nil
}

func (c *Client) DeleteRepositoryWebhook(ctx context.Context, workspace, repoSlug, webhookID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository slug are required")
	}
	if webhookID == "" {
		return fmt.Errorf("repository webhook ID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/hooks/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(webhookID))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}

func (c *Client) ListRepositoryDeployKeys(ctx context.Context, workspace, repoSlug string, limit int) ([]RepositoryDeployKey, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository slug are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/deploy-keys?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]RepositoryDeployKey, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page repositoryDeployKeyListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode repository deploy key list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetRepositoryDeployKey(ctx context.Context, workspace, repoSlug string, keyID int) (RepositoryDeployKey, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryDeployKey{}, fmt.Errorf("workspace and repository slug are required")
	}
	if keyID <= 0 {
		return RepositoryDeployKey{}, fmt.Errorf("repository deploy key ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), keyID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return RepositoryDeployKey{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return RepositoryDeployKey{}, err
	}

	var key RepositoryDeployKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return RepositoryDeployKey{}, fmt.Errorf("decode repository deploy key: %w", err)
	}
	return key, nil
}

func (c *Client) CreateRepositoryDeployKey(ctx context.Context, workspace, repoSlug string, options CreateRepositoryDeployKeyOptions) (RepositoryDeployKey, error) {
	if workspace == "" || repoSlug == "" {
		return RepositoryDeployKey{}, fmt.Errorf("workspace and repository slug are required")
	}
	if options.Label == "" {
		return RepositoryDeployKey{}, fmt.Errorf("repository deploy key label is required")
	}
	if options.Key == "" {
		return RepositoryDeployKey{}, fmt.Errorf("repository deploy key is required")
	}

	body := map[string]any{
		"label": options.Label,
		"key":   options.Key,
	}
	if options.Comment != "" {
		body["comment"] = options.Comment
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return RepositoryDeployKey{}, fmt.Errorf("marshal repository deploy key request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return RepositoryDeployKey{}, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return RepositoryDeployKey{}, err
	}

	var key RepositoryDeployKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return RepositoryDeployKey{}, fmt.Errorf("decode created repository deploy key: %w", err)
	}
	return key, nil
}

func (c *Client) DeleteRepositoryDeployKey(ctx context.Context, workspace, repoSlug string, keyID int) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository slug are required")
	}
	if keyID <= 0 {
		return fmt.Errorf("repository deploy key ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), keyID)
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}
