package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ListPipelinesOptions struct {
	State string
	Sort  string
	Query string
	Limit int
}

type Pipeline struct {
	UUID        string         `json:"uuid"`
	BuildNumber int            `json:"build_number,omitempty"`
	Creator     PipelineActor  `json:"creator,omitempty"`
	Target      PipelineTarget `json:"target,omitempty"`
	State       PipelineState  `json:"state,omitempty"`
	CreatedOn   string         `json:"created_on,omitempty"`
	CompletedOn string         `json:"completed_on,omitempty"`
	Links       PipelineLinks  `json:"links,omitempty"`
}

type PipelineConfig struct {
	Type       string                   `json:"type,omitempty"`
	Enabled    bool                     `json:"enabled"`
	Repository PipelineConfigRepository `json:"repository,omitempty"`
}

type PipelineConfigRepository struct {
	Type string `json:"type,omitempty"`
}

type TriggerPipelineOptions struct {
	RefType string
	RefName string
}

type PipelineActor struct {
	DisplayName string `json:"display_name,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
}

type PipelineTarget struct {
	Type     string             `json:"type,omitempty"`
	RefType  string             `json:"ref_type,omitempty"`
	RefName  string             `json:"ref_name,omitempty"`
	Commit   PipelineCommit     `json:"commit,omitempty"`
	Selector PipelineTargetExpr `json:"selector,omitempty"`
}

type PipelineCommit struct {
	Hash string `json:"hash,omitempty"`
}

type PipelineTargetExpr struct {
	Type    string `json:"type,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

type PipelineState struct {
	Name   string         `json:"name,omitempty"`
	Type   string         `json:"type,omitempty"`
	Result PipelineResult `json:"result,omitempty"`
	Stage  PipelineStage  `json:"stage,omitempty"`
}

type PipelineResult struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type PipelineStage struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type PipelineLinks struct {
	HTML Link `json:"html"`
}

type PipelineStep struct {
	UUID        string        `json:"uuid"`
	Name        string        `json:"name,omitempty"`
	State       PipelineState `json:"state,omitempty"`
	StartedOn   string        `json:"started_on,omitempty"`
	CompletedOn string        `json:"completed_on,omitempty"`
}

type pipelineListResponse struct {
	Values []Pipeline `json:"values"`
	Next   string     `json:"next,omitempty"`
}

type pipelineStepListResponse struct {
	Values []PipelineStep `json:"values"`
	Next   string         `json:"next,omitempty"`
}

func (c *Client) ListPipelines(ctx context.Context, workspace, repoSlug string, options ListPipelinesOptions) ([]Pipeline, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(options.Limit))
	if options.Sort != "" {
		values.Set("sort", options.Sort)
	}
	if options.State != "" {
		values.Set("status", options.State)
	}
	if options.Query != "" {
		values.Set("q", options.Query)
	}

	nextPath := fmt.Sprintf("/repositories/%s/%s/pipelines/?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	var all []Pipeline

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pipelineListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pipeline list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) GetPipeline(ctx context.Context, workspace, repoSlug, pipelineUUID string) (Pipeline, error) {
	if workspace == "" || repoSlug == "" {
		return Pipeline{}, fmt.Errorf("workspace and repository are required")
	}
	pipelineUUID = normalizePipelineUUID(pipelineUUID)
	if pipelineUUID == "" {
		return Pipeline{}, fmt.Errorf("pipeline UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(pipelineUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return Pipeline{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Pipeline{}, err
	}

	var pipeline Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return Pipeline{}, fmt.Errorf("decode pipeline: %w", err)
	}

	return pipeline, nil
}

func (c *Client) GetPipelineConfig(ctx context.Context, workspace, repoSlug string) (PipelineConfig, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineConfig{}, fmt.Errorf("workspace and repository are required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PipelineConfig{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineConfig{}, err
	}

	var config PipelineConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return PipelineConfig{}, fmt.Errorf("decode pipeline config: %w", err)
	}

	return config, nil
}

func (c *Client) UpdatePipelineConfig(ctx context.Context, workspace, repoSlug string, config PipelineConfig) (PipelineConfig, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineConfig{}, fmt.Errorf("workspace and repository are required")
	}

	if strings.TrimSpace(config.Type) == "" {
		config.Type = "repository_pipelines_configuration"
	}
	if strings.TrimSpace(config.Repository.Type) == "" {
		config.Repository.Type = "repository"
	}

	payload, err := json.Marshal(config)
	if err != nil {
		return PipelineConfig{}, fmt.Errorf("marshal pipeline config request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return PipelineConfig{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineConfig{}, err
	}

	var updated PipelineConfig
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return PipelineConfig{}, fmt.Errorf("decode updated pipeline config: %w", err)
	}

	return updated, nil
}

func (c *Client) GetPipelineByBuildNumber(ctx context.Context, workspace, repoSlug string, buildNumber int) (Pipeline, error) {
	if buildNumber <= 0 {
		return Pipeline{}, fmt.Errorf("pipeline build number must be greater than zero")
	}

	values := url.Values{}
	values.Set("pagelen", "50")
	values.Set("sort", "-created_on")

	nextPath := fmt.Sprintf("/repositories/%s/%s/pipelines/?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	for nextPath != "" {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return Pipeline{}, err
		}

		var page pipelineListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return Pipeline{}, fmt.Errorf("decode pipeline list: %w", err)
		}

		for _, pipeline := range page.Values {
			if pipeline.BuildNumber == buildNumber {
				return pipeline, nil
			}
		}

		nextPath = page.Next
	}

	return Pipeline{}, fmt.Errorf("pipeline #%d was not found", buildNumber)
}

func (c *Client) TriggerPipeline(ctx context.Context, workspace, repoSlug string, options TriggerPipelineOptions) (Pipeline, error) {
	if workspace == "" || repoSlug == "" {
		return Pipeline{}, fmt.Errorf("workspace and repository are required")
	}
	options.RefType = strings.TrimSpace(options.RefType)
	if options.RefType == "" {
		options.RefType = "branch"
	}
	options.RefName = strings.TrimSpace(options.RefName)
	if options.RefName == "" {
		return Pipeline{}, fmt.Errorf("pipeline ref name is required")
	}

	body := map[string]any{
		"target": map[string]any{
			"type":     "pipeline_ref_target",
			"ref_type": options.RefType,
			"ref_name": options.RefName,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return Pipeline{}, fmt.Errorf("marshal pipeline trigger request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return Pipeline{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Pipeline{}, err
	}

	var pipeline Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return Pipeline{}, fmt.Errorf("decode triggered pipeline: %w", err)
	}

	return pipeline, nil
}

func (c *Client) ListPipelineSteps(ctx context.Context, workspace, repoSlug, pipelineUUID string) ([]PipelineStep, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	pipelineUUID = normalizePipelineUUID(pipelineUUID)
	if pipelineUUID == "" {
		return nil, fmt.Errorf("pipeline UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(pipelineUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return nil, err
	}

	var page pipelineStepListResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("decode pipeline steps: %w", err)
	}

	return page.Values, nil
}

func (c *Client) GetPipelineStepLog(ctx context.Context, workspace, repoSlug, pipelineUUID, stepUUID string) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}
	pipelineUUID = normalizePipelineUUID(pipelineUUID)
	if pipelineUUID == "" {
		return "", fmt.Errorf("pipeline UUID is required")
	}
	stepUUID = normalizePipelineUUID(stepUUID)
	if stepUUID == "" {
		return "", fmt.Errorf("pipeline step UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/%s/log", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(pipelineUUID), url.PathEscape(stepUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, map[string]string{"Accept": "*/*"})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return "", err
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read pipeline step log: %w", err)
	}

	return string(payload), nil
}

func (c *Client) StopPipeline(ctx context.Context, workspace, repoSlug, pipelineUUID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	pipelineUUID = normalizePipelineUUID(pipelineUUID)
	if pipelineUUID == "" {
		return fmt.Errorf("pipeline UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/stopPipeline", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(pipelineUUID))
	resp, err := c.Do(ctx, http.MethodPost, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return err
	}

	return nil
}

func normalizePipelineUUID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		return value
	}
	return "{" + strings.Trim(value, "{}") + "}"
}
