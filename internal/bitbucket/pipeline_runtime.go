package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type PipelineRunner struct {
	UUID        string                    `json:"uuid,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Labels      []string                  `json:"labels,omitempty"`
	State       PipelineRunnerState       `json:"state,omitempty"`
	CreatedOn   string                    `json:"created_on,omitempty"`
	UpdatedOn   string                    `json:"updated_on,omitempty"`
	OAuthClient PipelineRunnerOAuthClient `json:"oauth_client,omitempty"`
}

type PipelineRunnerState struct {
	Status    string                `json:"status,omitempty"`
	Version   PipelineRunnerVersion `json:"version,omitempty"`
	UpdatedOn string                `json:"updated_on,omitempty"`
	Cordoned  bool                  `json:"cordoned"`
}

type PipelineRunnerVersion struct {
	Version string `json:"version,omitempty"`
}

type PipelineRunnerOAuthClient struct {
	ID            string `json:"id,omitempty"`
	Secret        string `json:"secret,omitempty"`
	TokenEndpoint string `json:"token_endpoint,omitempty"`
	Audience      string `json:"audience,omitempty"`
}

type PipelineCache struct {
	UUID          string `json:"uuid,omitempty"`
	PipelineUUID  string `json:"pipeline_uuid,omitempty"`
	StepUUID      string `json:"step_uuid,omitempty"`
	Name          string `json:"name,omitempty"`
	KeyHash       string `json:"key_hash,omitempty"`
	Path          string `json:"path,omitempty"`
	FileSizeBytes int64  `json:"file_size_bytes,omitempty"`
	CreatedOn     string `json:"created_on,omitempty"`
}

type listPipelineRunnersResponse struct {
	Values []PipelineRunner `json:"values"`
	Next   string           `json:"next,omitempty"`
}

type listPipelineCachesResponse struct {
	Values []PipelineCache `json:"values"`
	Next   string          `json:"next,omitempty"`
}

func (c *Client) ListPipelineRunners(ctx context.Context, workspace, repoSlug string, limit int) ([]PipelineRunner, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 100
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), limit)
	nextPath := path
	all := make([]PipelineRunner, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page listPipelineRunnersResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pipeline runners: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetPipelineRunner(ctx context.Context, workspace, repoSlug, runnerUUID string) (PipelineRunner, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineRunner{}, fmt.Errorf("workspace and repository are required")
	}
	runnerUUID = normalizePipelineUUID(runnerUUID)
	if runnerUUID == "" {
		return PipelineRunner{}, fmt.Errorf("pipeline runner UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(runnerUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PipelineRunner{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineRunner{}, err
	}

	var runner PipelineRunner
	if err := json.NewDecoder(resp.Body).Decode(&runner); err != nil {
		return PipelineRunner{}, fmt.Errorf("decode pipeline runner: %w", err)
	}
	return runner, nil
}

func (c *Client) DeletePipelineRunner(ctx context.Context, workspace, repoSlug, runnerUUID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	runnerUUID = normalizePipelineUUID(runnerUUID)
	if runnerUUID == "" {
		return fmt.Errorf("pipeline runner UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(runnerUUID))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}

func (c *Client) ListPipelineCaches(ctx context.Context, workspace, repoSlug string, limit int) ([]PipelineCache, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 100
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines-config/caches?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), limit)
	nextPath := path
	all := make([]PipelineCache, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page listPipelineCachesResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pipeline caches: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) DeletePipelineCache(ctx context.Context, workspace, repoSlug, cacheUUID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	cacheUUID = normalizePipelineUUID(cacheUUID)
	if cacheUUID == "" {
		return fmt.Errorf("pipeline cache UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines-config/caches/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(cacheUUID))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}

func (c *Client) DeletePipelineCachesByName(ctx context.Context, workspace, repoSlug, name string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if name == "" {
		return fmt.Errorf("pipeline cache name is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines-config/caches?name=%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.QueryEscape(name))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}
