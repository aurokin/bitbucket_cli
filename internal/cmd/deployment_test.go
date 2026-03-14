package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
)

func TestResolveDeploymentEnvironmentBySlug(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/repositories/acme/widgets/environments/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{env-1}","name":"Test","slug":"test"},{"uuid":"{env-2}","name":"Staging","slug":"staging"}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client, err := bitbucket.NewClient("bitbucket.org", config.HostConfig{
		AuthType: "api-token",
		Username: "agent@example.com",
		Token:    "token",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	environment, err := resolveDeploymentEnvironment(context.Background(), client, "acme", "widgets", "staging")
	if err != nil {
		t.Fatalf("resolveDeploymentEnvironment returned error: %v", err)
	}
	if environment.UUID != "{env-2}" || environment.Slug != "staging" {
		t.Fatalf("unexpected environment %+v", environment)
	}
}

func TestResolveDeploymentVariableReferenceByKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/acme/widgets/deployments_config/environments/{env-1}/variables" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("pagelen") != "200" {
			t.Fatalf("unexpected query %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{var-1}","key":"APP_ENV","value":"production","secured":false}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client, err := bitbucket.NewClient("bitbucket.org", config.HostConfig{
		AuthType: "api-token",
		Username: "agent@example.com",
		Token:    "token",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	variable, err := resolveDeploymentVariableReference(context.Background(), client, "acme", "widgets", "{env-1}", "APP_ENV")
	if err != nil {
		t.Fatalf("resolveDeploymentVariableReference returned error: %v", err)
	}
	if variable.UUID != "{var-1}" || variable.Key != "APP_ENV" {
		t.Fatalf("unexpected deployment variable %+v", variable)
	}
}

func TestResolveDeploymentVariableReferenceAmbiguous(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{var-1}","key":"APP_ENV"},{"uuid":"{var-2}","key":"app_env"}]}`))
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client, err := bitbucket.NewClient("bitbucket.org", config.HostConfig{
		AuthType: "api-token",
		Username: "agent@example.com",
		Token:    "token",
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = resolveDeploymentVariableReference(context.Background(), client, "acme", "widgets", "{env-1}", "APP_ENV")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous deployment variable error, got %v", err)
	}
}

func TestWriteDeploymentSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := deploymentPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Warnings:  []string{"local repository context unavailable; continuing without local checkout metadata (not a repo)"},
		Deployment: bitbucket.Deployment{
			UUID: "{dep-1}",
			State: bitbucket.DeploymentState{
				Name: "SUCCESSFUL",
			},
			Environment: bitbucket.DeploymentEnvironment{
				UUID: "{env-1}",
				Name: "Test",
			},
			Release: bitbucket.DeploymentRelease{
				Name: "main",
			},
		},
	}

	if err := writeDeploymentSummary(&buf, payload); err != nil {
		t.Fatalf("writeDeploymentSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Warning: local repository context unavailable",
		"Deployment: {dep-1}",
		"State: SUCCESSFUL",
		"Environment: Test",
		"Release: main",
		"Next: bb deployment environment view {env-1} --repo acme/widgets",
	)
}

func TestWriteDeploymentVariableListSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := deploymentVariableListPayload{
		Workspace: "acme",
		Repo:      "widgets",
		Environment: bitbucket.DeploymentEnvironment{
			Name: "Production",
			Slug: "production",
		},
		Variables: []bitbucket.DeploymentVariable{
			{UUID: "{var-1}", Key: "APP_ENV", Secured: true},
		},
	}

	if err := writeDeploymentVariableListSummary(&buf, payload); err != nil {
		t.Fatalf("writeDeploymentVariableListSummary returned error: %v", err)
	}

	got := buf.String()
	for _, expected := range []string{"APP_ENV", "true", "Environment: Production"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output, got %q", expected, got)
		}
	}
	assertOrderedSubstrings(t, got,
		"Repository: acme/widgets",
		"Environment: Production",
		"APP_ENV",
		"Next: bb deployment environment variable view {var-1} --repo acme/widgets --environment production",
	)
}

func TestWriteDeploymentVariableSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	payload := deploymentVariablePayload{
		Workspace: "acme",
		Repo:      "widgets",
		Action:    "deleted",
		Deleted:   true,
		Environment: bitbucket.DeploymentEnvironment{
			Name: "Production",
			Slug: "production",
		},
		Variable: bitbucket.DeploymentVariable{
			UUID:    "{var-1}",
			Key:     "APP_ENV",
			Secured: true,
		},
	}

	if err := writeDeploymentVariableSummary(&buf, payload); err != nil {
		t.Fatalf("writeDeploymentVariableSummary returned error: %v", err)
	}

	assertOrderedSubstrings(t, buf.String(),
		"Repository: acme/widgets",
		"Environment: Production",
		"Action: deleted",
		"Variable: APP_ENV",
		"Status: deleted",
		"Next: bb deployment environment variable list --repo acme/widgets --environment production",
	)
}
