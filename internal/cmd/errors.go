package cmd

import (
	"fmt"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
)

func userFacingError(err error) error {
	if err == nil {
		return nil
	}

	if guided := authGuidanceError(err); guided != nil && guided.Error() != err.Error() {
		return guided
	}

	if apiErr, ok := bitbucket.AsAPIError(err); ok {
		return apiGuidanceError(apiErr)
	}

	message := err.Error()
	switch {
	case message == "multiple workspaces available; specify --workspace":
		return fmt.Errorf("multiple workspaces are available; pass --repo <workspace>/<repo> or add --workspace to disambiguate")
	case strings.HasPrefix(message, "repository target ") && strings.HasSuffix(message, " requires --workspace"):
		return fmt.Errorf("%s; pass --repo <workspace>/<repo> or add --workspace", message)
	case strings.HasPrefix(message, "could not determine the repository from the current directory;"):
		return fmt.Errorf("%s Try `bb repo view --repo <workspace>/<repo>` or run inside a Bitbucket git checkout.", strings.TrimSuffix(message, ";"))
	case strings.HasPrefix(message, "--workspace requires --repo"):
		return fmt.Errorf("--workspace only disambiguates a repository target; pass --repo <repo> or --repo <workspace>/<repo>")
	default:
		return err
	}
}

func apiGuidanceError(apiErr *bitbucket.APIError) error {
	if apiErr == nil {
		return nil
	}

	detail := apiErrorDetail(apiErr)
	if strings.Contains(strings.ToLower(detail), "no issue tracker") {
		return fmt.Errorf("this repository does not have Bitbucket issue tracking enabled. Enable the issue tracker in the repository settings or use a repository with issues enabled. Bitbucket said: %s", detail)
	}

	switch apiErr.StatusCode {
	case 401:
		return fmt.Errorf("authentication failed: the stored API token may be invalid, expired, or revoked. Run `bb auth login --username <email> --with-token` to replace it, then verify with `bb auth status --check`. Bitbucket said: %s", detail)
	case 403:
		return fmt.Errorf("request denied by Bitbucket Cloud: the API token may be missing required Bitbucket scopes, or the account may not have access to this workspace or repository. Create a token with the needed Bitbucket scopes, then verify with `bb auth status --check`. Bitbucket said: %s", detail)
	case 404:
		return fmt.Errorf("Bitbucket could not find that resource, or your token cannot see it. Check the repository target, workspace slug, and pull request ID. Bitbucket said: %s", detail)
	default:
		return apiErr
	}
}

func apiErrorDetail(apiErr *bitbucket.APIError) string {
	if apiErr == nil {
		return ""
	}
	if apiErr.Message != "" {
		return apiErr.Message
	}
	if apiErr.Status != "" {
		return apiErr.Status
	}
	return "request failed"
}
