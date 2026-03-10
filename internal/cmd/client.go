package cmd

import (
	"fmt"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
)

func resolveAuthenticatedClient(host string) (string, *bitbucket.Client, error) {
	resolvedHost, hostConfig, err := resolveAuthenticatedHostConfig(host)
	if err != nil {
		return "", nil, err
	}

	client, err := bitbucket.NewClient(resolvedHost, hostConfig)
	if err != nil {
		return "", nil, err
	}

	return resolvedHost, client, nil
}

func resolveAuthenticatedHostConfig(host string) (string, config.HostConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", config.HostConfig{}, err
	}

	resolvedHost, err := cfg.ResolveHost(strings.TrimSpace(host))
	if err != nil {
		return "", config.HostConfig{}, authGuidanceError(err)
	}

	hostConfig, ok := cfg.Hosts[resolvedHost]
	if !ok {
		return "", config.HostConfig{}, fmt.Errorf("no stored credentials found for %s; run `bb auth login --username <email> --with-token`", resolvedHost)
	}

	return resolvedHost, hostConfig, nil
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
