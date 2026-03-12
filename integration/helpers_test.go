//go:build integration

package integration

import (
	"os"
	"os/exec"
	"testing"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
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

	cmd := exec.Command(s.Binary, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bb %v failed: %v\n%s", args, err, output)
	}
	return output
}

func (s integrationSession) RunAllowFailure(t *testing.T, dir string, args ...string) ([]byte, error) {
	t.Helper()

	cmd := exec.Command(s.Binary, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	return output, err
}
