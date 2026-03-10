package cmd

import (
	"fmt"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
)

func resolveAuthenticatedClient(host string) (string, *bitbucket.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", nil, err
	}

	resolvedHost, err := cfg.ResolveHost(strings.TrimSpace(host))
	if err != nil {
		return "", nil, authGuidanceError(err)
	}

	hostConfig, ok := cfg.Hosts[resolvedHost]
	if !ok {
		return "", nil, fmt.Errorf("no stored credentials found for %s; run `bb auth login --username <email> --with-token`", resolvedHost)
	}

	client, err := bitbucket.NewClient(resolvedHost, hostConfig)
	if err != nil {
		return "", nil, err
	}

	return resolvedHost, client, nil
}

func authGuidanceError(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "no authenticated hosts configured":
		return fmt.Errorf("no authenticated hosts configured; run `bb auth login --username <email> --with-token`")
	default:
		return err
	}
}
