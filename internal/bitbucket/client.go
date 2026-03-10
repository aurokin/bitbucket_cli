package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/auro/bitbucket_cli/internal/config"
)

const defaultCloudAPIBaseURL = "https://api.bitbucket.org/2.0"

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       config.HostConfig
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

func NewClient(host string, auth config.HostConfig, options ...Option) (*Client, error) {
	baseURL, err := resolveBaseURL(strings.TrimSpace(host))
	if err != nil {
		return nil, err
	}

	client := &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		auth:       auth,
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

func (c *Client) CurrentUser(ctx context.Context) (CurrentUser, error) {
	resp, err := c.Do(ctx, http.MethodGet, "/user", nil, nil)
	if err != nil {
		return CurrentUser{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return CurrentUser{}, err
	}

	var payload CurrentUser
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return CurrentUser{}, fmt.Errorf("decode current user response: %w", err)
	}

	return payload, nil
}

func (c *Client) Do(ctx context.Context, method, requestPath string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := c.NewRequest(ctx, method, requestPath, body, headers)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", req.Method, req.URL.String(), err)
	}

	return resp, nil
}

func (c *Client) NewRequest(ctx context.Context, method, requestPath string, body []byte, headers map[string]string) (*http.Request, error) {
	targetURL, err := c.resolveURL(requestPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if err := applyAuthorization(req, c.auth); err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) resolveURL(requestPath string) (string, error) {
	trimmed := strings.TrimSpace(requestPath)
	if trimmed == "" {
		return "", fmt.Errorf("request path is required")
	}

	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("parse request URL: %w", err)
		}
		return parsed.String(), nil
	}

	relative, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse request path: %w", err)
	}

	base := *c.baseURL
	base.Path = path.Join(c.baseURL.Path, strings.TrimPrefix(relative.Path, "/"))
	base.RawQuery = relative.RawQuery
	base.Fragment = relative.Fragment

	return base.String(), nil
}

func resolveBaseURL(host string) (*url.URL, error) {
	if override := strings.TrimSpace(os.Getenv("BB_API_BASE_URL")); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return nil, fmt.Errorf("parse BB_API_BASE_URL: %w", err)
		}
		return parsed, nil
	}

	switch host {
	case "", "bitbucket.org", "api.bitbucket.org":
		parsed, err := url.Parse(defaultCloudAPIBaseURL)
		if err != nil {
			return nil, fmt.Errorf("parse default API base URL: %w", err)
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("host %q is not supported yet; only Bitbucket Cloud is implemented", host)
	}
}

func applyAuthorization(req *http.Request, auth config.HostConfig) error {
	token := strings.TrimSpace(auth.Token)
	if token == "" {
		return fmt.Errorf("no token configured for host")
	}

	switch normalizeTokenType(auth.TokenType) {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	case "basic":
		if strings.TrimSpace(auth.Username) == "" {
			return fmt.Errorf("username is required for basic auth")
		}
		req.SetBasicAuth(auth.Username, token)
		return nil
	default:
		return fmt.Errorf("unsupported token type %q", auth.TokenType)
	}
}

func normalizeTokenType(tokenType string) string {
	switch strings.ToLower(strings.TrimSpace(tokenType)) {
	case "", "bearer", "token", "api-token", "oauth":
		return "bearer"
	case "basic", "app-password":
		return "basic"
	default:
		return strings.ToLower(strings.TrimSpace(tokenType))
	}
}

func requireSuccess(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return fmt.Errorf("bitbucket API returned %s", resp.Status)
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("bitbucket API returned %s", resp.Status)
	}

	return fmt.Errorf("bitbucket API returned %s: %s", resp.Status, trimmed)
}

type CurrentUser struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username,omitempty"`
	UUID        string `json:"uuid,omitempty"`
}
