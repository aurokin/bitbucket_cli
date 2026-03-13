package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/config"
)

func TestBranchAndTagClientFlows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/refs/branches":
			_, _ = w.Write([]byte(`{"values":[{"name":"main","type":"branch","target":{"hash":"abc1234"}}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/refs/branches/main":
			_, _ = w.Write([]byte(`{"name":"main","type":"branch","target":{"hash":"abc1234"},"default_merge_strategy":"squash"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/repositories/acme/widgets/refs/branches":
			_, _ = w.Write([]byte(`{"name":"feature/demo","type":"branch","target":{"hash":"abc1234"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/2.0/repositories/acme/widgets/refs/branches/feature/demo":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/refs/tags":
			_, _ = w.Write([]byte(`{"values":[{"name":"v1.0.0","type":"tag","target":{"hash":"abc1234"},"message":"release"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/repositories/acme/widgets/refs/tags/v1.0.0":
			_, _ = w.Write([]byte(`{"name":"v1.0.0","type":"tag","target":{"hash":"abc1234"},"message":"release"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/repositories/acme/widgets/refs/tags":
			_, _ = w.Write([]byte(`{"name":"v1.0.1","type":"tag","target":{"hash":"abc1234"},"message":"release"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/2.0/repositories/acme/widgets/refs/tags/v1.0.1":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client, err := NewClient("bitbucket.org", config.HostConfig{
		Username: "auro@example.com",
		Token:    "secret",
		AuthType: config.AuthTypeAPIToken,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	branches, err := client.ListBranches(context.Background(), "acme", "widgets", ListBranchesOptions{Limit: 20})
	if err != nil || len(branches) != 1 || branches[0].Name != "main" {
		t.Fatalf("ListBranches returned %v %+v", err, branches)
	}

	branch, err := client.GetBranch(context.Background(), "acme", "widgets", "main")
	if err != nil || branch.DefaultMergeStrategy != "squash" {
		t.Fatalf("GetBranch returned %v %+v", err, branch)
	}

	createdBranch, err := client.CreateBranch(context.Background(), "acme", "widgets", CreateBranchOptions{Name: "feature/demo", Target: "abc1234"})
	if err != nil || createdBranch.Name != "feature/demo" {
		t.Fatalf("CreateBranch returned %v %+v", err, createdBranch)
	}

	if err := client.DeleteBranch(context.Background(), "acme", "widgets", "feature/demo"); err != nil {
		t.Fatalf("DeleteBranch returned error: %v", err)
	}

	tags, err := client.ListTags(context.Background(), "acme", "widgets", ListTagsOptions{Limit: 20})
	if err != nil || len(tags) != 1 || tags[0].Name != "v1.0.0" {
		t.Fatalf("ListTags returned %v %+v", err, tags)
	}

	tag, err := client.GetTag(context.Background(), "acme", "widgets", "v1.0.0")
	if err != nil || tag.Message != "release" {
		t.Fatalf("GetTag returned %v %+v", err, tag)
	}

	createdTag, err := client.CreateTag(context.Background(), "acme", "widgets", CreateTagOptions{Name: "v1.0.1", Target: "abc1234", Message: "release"})
	if err != nil || createdTag.Name != "v1.0.1" {
		t.Fatalf("CreateTag returned %v %+v", err, createdTag)
	}

	if err := client.DeleteTag(context.Background(), "acme", "widgets", "v1.0.1"); err != nil {
		t.Fatalf("DeleteTag returned error: %v", err)
	}
}
