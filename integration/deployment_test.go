//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func TestBitbucketCloudDeploymentRead(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	repoTarget := session.Workspace + "/" + pipelines.Repo.Slug

	listOutput := session.Run(t, "", "deployment", "list", "--repo", repoTarget, "--json", "*")
	var listed struct {
		Workspace   string                 `json:"workspace"`
		Repo        string                 `json:"repo"`
		Deployments []bitbucket.Deployment `json:"deployments"`
	}
	if err := json.Unmarshal(listOutput, &listed); err != nil {
		t.Fatalf("parse deployment list JSON: %v\n%s", err, listOutput)
	}
	if listed.Workspace != session.Workspace || listed.Repo != pipelines.Repo.Slug {
		t.Fatalf("unexpected deployment list payload %+v", listed)
	}

	environmentListOutput := session.Run(t, "", "deployment", "environment", "list", "--repo", repoTarget, "--json", "*")
	var environments struct {
		Environments []bitbucket.DeploymentEnvironment `json:"environments"`
	}
	if err := json.Unmarshal(environmentListOutput, &environments); err != nil {
		t.Fatalf("parse deployment environment list JSON: %v\n%s", err, environmentListOutput)
	}
	if len(environments.Environments) == 0 {
		t.Fatalf("expected at least one deployment environment")
	}

	viewOutput := session.Run(t, "", "deployment", "environment", "view", environments.Environments[0].Slug, "--repo", repoTarget, "--json", "*")
	var viewed struct {
		Environment bitbucket.DeploymentEnvironment `json:"environment"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse deployment environment view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Environment.UUID == "" {
		t.Fatalf("unexpected deployment environment payload %+v", viewed)
	}

	variableListOutput := session.Run(t, "", "deployment", "environment", "variable", "list", "--repo", repoTarget, "--environment", environments.Environments[0].Slug, "--json", "*")
	var variables struct {
		Environment bitbucket.DeploymentEnvironment `json:"environment"`
		Variables   []bitbucket.DeploymentVariable  `json:"variables"`
	}
	if err := json.Unmarshal(variableListOutput, &variables); err != nil {
		t.Fatalf("parse deployment variable list JSON: %v\n%s", err, variableListOutput)
	}
	if variables.Environment.UUID != viewed.Environment.UUID {
		t.Fatalf("unexpected deployment variable payload %+v", variables)
	}

	humanOutput := session.Run(t, "", "deployment", "environment", "view", environments.Environments[0].Slug, "--repo", repoTarget)
	assertContainsOrdered(t, string(humanOutput),
		"Repository: "+repoTarget,
		"Environment: "+viewed.Environment.Name,
		"Next: bb deployment environment variable list --repo "+repoTarget+" --environment "+viewed.Environment.Slug,
	)
}

func TestBitbucketCloudDeploymentVariableFlow(t *testing.T) {
	session := newIntegrationSession(t)
	pipelines := session.PipelineFixture(t)
	repoTarget := session.Workspace + "/" + pipelines.Repo.Slug

	environmentsOutput := session.Run(t, "", "deployment", "environment", "list", "--repo", repoTarget, "--json", "*")
	var environments struct {
		Environments []bitbucket.DeploymentEnvironment `json:"environments"`
	}
	if err := json.Unmarshal(environmentsOutput, &environments); err != nil {
		t.Fatalf("parse deployment environment list JSON: %v\n%s", err, environmentsOutput)
	}
	if len(environments.Environments) == 0 {
		t.Fatalf("expected at least one deployment environment")
	}
	environment := environments.Environments[0]
	variableKey := fmt.Sprintf("BB_DEPLOY_IT_%d", time.Now().UTC().UnixNano())

	createOutput, err := session.RunAllowFailure(t, "", "deployment", "environment", "variable", "create", "--repo", repoTarget, "--environment", environment.Slug, "--key", variableKey, "--value", "created-value", "--json", "*")
	if err != nil {
		if bytes.Contains(createOutput, []byte("Missing Token Scopes Or Insufficient Access")) {
			t.Skipf("deployment variable create requires broader repository administration scopes:\n%s", createOutput)
		}
		t.Fatalf("bb deployment environment variable create failed: %v\n%s", err, createOutput)
	}

	var created struct {
		Environment bitbucket.DeploymentEnvironment `json:"environment"`
		Variable    bitbucket.DeploymentVariable    `json:"variable"`
	}
	if err := json.Unmarshal(createOutput, &created); err != nil {
		t.Fatalf("parse deployment variable create JSON: %v\n%s", err, createOutput)
	}
	if created.Environment.UUID != environment.UUID || created.Variable.Key != variableKey {
		t.Fatalf("unexpected created deployment variable %+v", created)
	}

	var found bool
	for attempt := 0; attempt < 12; attempt++ {
		listOutput := session.Run(t, "", "deployment", "environment", "variable", "list", "--repo", repoTarget, "--environment", environment.Slug, "--json", "*")
		var listed struct {
			Variables []bitbucket.DeploymentVariable `json:"variables"`
		}
		if err := json.Unmarshal(listOutput, &listed); err != nil {
			t.Fatalf("parse deployment variable list JSON: %v\n%s", err, listOutput)
		}
		for _, variable := range listed.Variables {
			if variable.UUID == created.Variable.UUID {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if !found {
		t.Fatalf("expected created deployment variable in list after create")
	}

	viewOutput := session.Run(t, "", "deployment", "environment", "variable", "view", variableKey, "--repo", repoTarget, "--environment", environment.Slug, "--json", "*")
	var viewed struct {
		Variable bitbucket.DeploymentVariable `json:"variable"`
	}
	if err := json.Unmarshal(viewOutput, &viewed); err != nil {
		t.Fatalf("parse deployment variable view JSON: %v\n%s", err, viewOutput)
	}
	if viewed.Variable.UUID != created.Variable.UUID {
		t.Fatalf("unexpected viewed deployment variable %+v", viewed)
	}

	editOutput := session.Run(t, "", "deployment", "environment", "variable", "edit", variableKey, "--repo", repoTarget, "--environment", environment.Slug, "--value", "updated-value", "--secured", "false", "--json", "*")
	var edited struct {
		Variable bitbucket.DeploymentVariable `json:"variable"`
	}
	if err := json.Unmarshal(editOutput, &edited); err != nil {
		t.Fatalf("parse deployment variable edit JSON: %v\n%s", err, editOutput)
	}
	if edited.Variable.UUID != created.Variable.UUID || edited.Variable.Value != "updated-value" {
		t.Fatalf("unexpected edited deployment variable %+v", edited)
	}

	deleteOutput := session.Run(t, "", "deployment", "environment", "variable", "delete", variableKey, "--repo", repoTarget, "--environment", environment.Slug, "--yes", "--json", "*")
	var deleted struct {
		Deleted  bool                         `json:"deleted"`
		Variable bitbucket.DeploymentVariable `json:"variable"`
	}
	if err := json.Unmarshal(deleteOutput, &deleted); err != nil {
		t.Fatalf("parse deployment variable delete JSON: %v\n%s", err, deleteOutput)
	}
	if !deleted.Deleted || deleted.Variable.UUID != created.Variable.UUID {
		t.Fatalf("unexpected deleted deployment variable %+v", deleted)
	}

	humanOutput := session.Run(t, "", "deployment", "environment", "variable", "create", "--repo", repoTarget, "--environment", environment.Slug, "--key", variableKey, "--value", "final-value")
	assertContainsOrdered(t, string(humanOutput),
		"Repository: "+repoTarget,
		"Environment: "+environment.Name,
		"Action: created",
		"Variable: "+variableKey,
		"Next: bb deployment environment variable list --repo "+repoTarget+" --environment "+environment.Slug,
	)
	var cleanupOutput []byte
	var cleanupErr error
	for attempt := 0; attempt < 12; attempt++ {
		cleanupOutput, cleanupErr = session.RunAllowFailure(t, "", "deployment", "environment", "variable", "delete", variableKey, "--repo", repoTarget, "--environment", environment.Slug, "--yes", "--json", "*")
		if cleanupErr == nil {
			break
		}
		if !bytes.Contains(cleanupOutput, []byte("was not found")) {
			t.Fatalf("bb deployment environment variable delete failed: %v\n%s", cleanupErr, cleanupOutput)
		}
		time.Sleep(2 * time.Second)
	}
	if cleanupErr != nil {
		t.Fatalf("bb deployment environment variable cleanup failed: %v\n%s", cleanupErr, cleanupOutput)
	}
	var cleanup struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal(cleanupOutput, &cleanup); err != nil {
		t.Fatalf("parse deployment variable cleanup JSON: %v\n%s", err, cleanupOutput)
	}
	if !cleanup.Deleted {
		t.Fatalf("expected cleanup delete payload %+v", cleanup)
	}
}
