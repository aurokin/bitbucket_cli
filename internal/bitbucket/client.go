package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/config"
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
	auth = config.NormalizeHostConfig(auth)

	token := strings.TrimSpace(auth.Token)
	if token == "" {
		return fmt.Errorf("no token configured for host")
	}

	switch auth.AuthType {
	case config.AuthTypeAPIToken:
		if strings.TrimSpace(auth.Username) == "" {
			return fmt.Errorf("username is required for api-token auth")
		}
		req.SetBasicAuth(auth.Username, token)
		return nil
	default:
		return fmt.Errorf("unsupported auth type %q", auth.AuthType)
	}
}

func requireSuccess(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		body = nil
	}

	return newAPIError(resp.StatusCode, resp.Status, body)
}

type APIError struct {
	StatusCode int
	Status     string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Message != "":
		return fmt.Sprintf("bitbucket API returned %s: %s", e.Status, e.Message)
	case e.Status != "":
		return fmt.Sprintf("bitbucket API returned %s", e.Status)
	default:
		return "bitbucket API request failed"
	}
}

func (e *APIError) Is(target error) bool {
	other, ok := target.(*APIError)
	if !ok {
		return false
	}
	if other.StatusCode != 0 && e.StatusCode != other.StatusCode {
		return false
	}
	return true
}

func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return nil, false
	}
	return apiErr, true
}

func NewAPIError(statusCode int, status string, body []byte) error {
	return newAPIError(statusCode, status, body)
}

func newAPIError(statusCode int, status string, body []byte) error {
	return &APIError{
		StatusCode: statusCode,
		Status:     strings.TrimSpace(status),
		Message:    parseAPIErrorMessage(body),
		Body:       strings.TrimSpace(string(body)),
	}
}

func parseAPIErrorMessage(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return trimmed
	}

	if message := nestedString(payload, "error", "message"); message != "" {
		return message
	}
	if message := nestedString(payload, "error", "detail"); message != "" {
		return message
	}
	if message := stringValue(payload["error"]); message != "" {
		return message
	}
	if message := stringValue(payload["message"]); message != "" {
		return message
	}

	return trimmed
}

func nestedString(payload map[string]any, keys ...string) string {
	current := any(payload)
	for _, key := range keys {
		typed, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = typed[key]
		if !ok {
			return ""
		}
	}
	return stringValue(current)
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if part := stringValue(item); part != "" {
				parts = append(parts, part)
			}
		}
		return strings.Join(parts, "; ")
	case map[string]any:
		if message := nestedString(typed, "message"); message != "" {
			return message
		}
	}
	return ""
}

type CurrentUser struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username,omitempty"`
	UUID        string `json:"uuid,omitempty"`
}
