//go:build integration

package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
	"github.com/aurokin/bitbucket_cli/internal/config"
)

type integrationSession struct {
	Config     config.Config
	Client     *bitbucket.Client
	HostConfig config.HostConfig
	Workspace  string
	Binary     string
}

func newIntegrationSession(t *testing.T) integrationSession {
	t.Helper()

	requireManualIntegration(t)

	cfg, client, hostConfig := loadIntegrationClient(t)
	return integrationSession{
		Config:     cfg,
		Client:     client,
		HostConfig: hostConfig,
		Workspace:  resolveWorkspace(t, client),
		Binary:     buildBinary(t),
	}
}

func requireManualIntegration(t *testing.T) {
	t.Helper()

	if os.Getenv("BB_RUN_INTEGRATION") != "1" {
		t.Skip("set BB_RUN_INTEGRATION=1 to run Bitbucket Cloud integration tests")
	}
	if os.Getenv("CI") != "" {
		t.Skip("manual-only integration test")
	}
}

func (s integrationSession) Fixture(t *testing.T) integrationFixture {
	t.Helper()
	return ensureFixture(t, s.Client, s.HostConfig, s.Workspace)
}

func (s integrationSession) PipelineFixture(t *testing.T) pipelineFixture {
	t.Helper()
	return ensurePipelineFixture(t, s.Client, s.HostConfig, s.Workspace)
}

func (s integrationSession) Run(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	output, err := runExternalAllowFailure(t, dir, false, s.Binary, args...)
	if err != nil {
		t.Fatalf("bb %v failed: %v\n%s", args, err, output)
	}
	return output
}

func (s integrationSession) RunAllowFailure(t *testing.T, dir string, args ...string) ([]byte, error) {
	t.Helper()

	return runExternalAllowFailure(t, dir, false, s.Binary, args...)
}

func runExternal(t *testing.T, dir string, scrubOutput bool, name string, args ...string) []byte {
	t.Helper()

	output, err := runExternalAllowFailure(t, dir, scrubOutput, name, args...)
	if err != nil {
		if scrubOutput {
			t.Fatalf("%s %v failed: %v\n%s", name, args, err, scrub(output))
		}
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, output)
	}
	return output
}

func runExternalAllowFailure(t *testing.T, dir string, scrubOutput bool, name string, args ...string) ([]byte, error) {
	t.Helper()

	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if scrubOutput {
		output = scrub(output)
	}
	return output, err
}

func assertContainsOrdered(t *testing.T, got string, expected ...string) {
	t.Helper()

	searchFrom := 0
	for _, item := range expected {
		idx := strings.Index(got[searchFrom:], item)
		if idx < 0 {
			t.Fatalf("expected %q in output, got %q", item, got)
		}
		searchFrom += idx + len(item)
	}
}
