package bitbucket

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auro/bitbucket_cli/internal/config"
)

func TestCurrentUserAPITokenAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("auro@example.com:api-token-secret"))
		if got := r.Header.Get("Authorization"); got != expected {
			t.Fatalf("unexpected authorization header %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"account_id":"123","display_name":"Auro"}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL)

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "api-token-secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("CurrentUser returned error: %v", err)
	}
	if user.AccountID != "123" || user.DisplayName != "Auro" {
		t.Fatalf("unexpected current user payload %+v", user)
	}
}

func TestResolveBaseURLRejectsUnsupportedHosts(t *testing.T) {
	t.Setenv("BB_API_BASE_URL", "")

	if _, err := resolveBaseURL("example.com"); err == nil {
		t.Fatalf("expected unsupported host error")
	}
}

func TestApplyAuthorizationRejectsUnsupportedAuthTypes(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	err = applyAuthorization(req, config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: "oauth-web",
	})
	if err == nil {
		t.Fatalf("expected unsupported auth type error")
	}
}

func TestResolveURLPreservesQuery(t *testing.T) {
	t.Setenv("BB_API_BASE_URL", "https://api.bitbucket.org/2.0")

	client, err := NewClient("bitbucket.org", config.HostConfig{
		Token: "secret-token",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	rawURL, err := client.resolveURL("/repositories/acme/widgets/pullrequests?q=state=%22OPEN%22")
	if err != nil {
		t.Fatalf("resolveURL returned error: %v", err)
	}

	expected := "https://api.bitbucket.org/2.0/repositories/acme/widgets/pullrequests?q=state=%22OPEN%22"
	if rawURL != expected {
		t.Fatalf("expected URL %q, got %q", expected, rawURL)
	}
}

func TestNewAPIErrorParsesNestedMessage(t *testing.T) {
	t.Parallel()

	err := NewAPIError(401, "401 Unauthorized", []byte(`{"type":"error","error":{"message":"Token is invalid or expired"}}`))
	apiErr, ok := AsAPIError(err)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Message != "Token is invalid or expired" {
		t.Fatalf("unexpected message %q", apiErr.Message)
	}
}

func TestNewAPIErrorFallsBackToRawBody(t *testing.T) {
	t.Parallel()

	err := NewAPIError(403, "403 Forbidden", []byte(`plain text error`))
	apiErr, ok := AsAPIError(err)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Message != "plain text error" {
		t.Fatalf("unexpected message %q", apiErr.Message)
	}
}
