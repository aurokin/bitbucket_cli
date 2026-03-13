package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type PipelineTestReportSummary map[string]any
type PipelineTestCase map[string]any

type PipelineVariable struct {
	UUID    string `json:"uuid,omitempty"`
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
	Secured bool   `json:"secured"`
}

type ListPipelineVariablesOptions struct {
	Limit int
}

type paginatedPipelineVariablesResponse struct {
	Values []PipelineVariable `json:"values"`
	Next   string             `json:"next,omitempty"`
}

type paginatedPipelineTestCasesResponse struct {
	Values []PipelineTestCase `json:"values"`
	Next   string             `json:"next,omitempty"`
}

func (c *Client) GetPipelineTestReports(ctx context.Context, workspace, repoSlug, pipelineUUID, stepUUID string) (PipelineTestReportSummary, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	pipelineUUID = normalizePipelineUUID(pipelineUUID)
	stepUUID = normalizePipelineUUID(stepUUID)
	if pipelineUUID == "" || stepUUID == "" {
		return nil, fmt.Errorf("pipeline UUID and step UUID are required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/%s/test_reports", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(pipelineUUID), url.PathEscape(stepUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return nil, err
	}

	var payload PipelineTestReportSummary
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode pipeline test report summary: %w", err)
	}
	return payload, nil
}

func (c *Client) ListPipelineTestCases(ctx context.Context, workspace, repoSlug, pipelineUUID, stepUUID string, limit int) ([]PipelineTestCase, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	pipelineUUID = normalizePipelineUUID(pipelineUUID)
	stepUUID = normalizePipelineUUID(stepUUID)
	if pipelineUUID == "" || stepUUID == "" {
		return nil, fmt.Errorf("pipeline UUID and step UUID are required")
	}
	if limit <= 0 {
		limit = 100
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/%s/test_reports/test_cases?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(pipelineUUID), url.PathEscape(stepUUID), limit)
	nextPath := path
	all := make([]PipelineTestCase, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page paginatedPipelineTestCasesResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pipeline test cases: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) ListPipelineVariables(ctx context.Context, workspace, repoSlug string, options ListPipelineVariablesOptions) ([]PipelineVariable, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if options.Limit <= 0 {
		options.Limit = 100
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), options.Limit)
	nextPath := path
	all := make([]PipelineVariable, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page paginatedPipelineVariablesResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pipeline variables: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) GetPipelineVariable(ctx context.Context, workspace, repoSlug, variableUUID string) (PipelineVariable, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineVariable{}, fmt.Errorf("workspace and repository are required")
	}
	variableUUID = normalizePipelineUUID(variableUUID)
	if variableUUID == "" {
		return PipelineVariable{}, fmt.Errorf("pipeline variable UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(variableUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PipelineVariable{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineVariable{}, err
	}

	var variable PipelineVariable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return PipelineVariable{}, fmt.Errorf("decode pipeline variable: %w", err)
	}
	return variable, nil
}

func (c *Client) CreatePipelineVariable(ctx context.Context, workspace, repoSlug string, variable PipelineVariable) (PipelineVariable, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineVariable{}, fmt.Errorf("workspace and repository are required")
	}
	if strings.TrimSpace(variable.Key) == "" {
		return PipelineVariable{}, fmt.Errorf("pipeline variable key is required")
	}

	payload, err := json.Marshal(variable)
	if err != nil {
		return PipelineVariable{}, fmt.Errorf("marshal pipeline variable request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PipelineVariable{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineVariable{}, err
	}

	var created PipelineVariable
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return PipelineVariable{}, fmt.Errorf("decode created pipeline variable: %w", err)
	}
	return created, nil
}

func (c *Client) UpdatePipelineVariable(ctx context.Context, workspace, repoSlug, variableUUID string, variable PipelineVariable) (PipelineVariable, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineVariable{}, fmt.Errorf("workspace and repository are required")
	}
	variableUUID = normalizePipelineUUID(variableUUID)
	if variableUUID == "" {
		return PipelineVariable{}, fmt.Errorf("pipeline variable UUID is required")
	}
	if strings.TrimSpace(variable.Key) == "" {
		return PipelineVariable{}, fmt.Errorf("pipeline variable key is required")
	}

	payload, err := json.Marshal(variable)
	if err != nil {
		return PipelineVariable{}, fmt.Errorf("marshal pipeline variable request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(variableUUID))
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return PipelineVariable{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineVariable{}, err
	}

	var updated PipelineVariable
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return PipelineVariable{}, fmt.Errorf("decode updated pipeline variable: %w", err)
	}
	return updated, nil
}

func (c *Client) DeletePipelineVariable(ctx context.Context, workspace, repoSlug, variableUUID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	variableUUID = normalizePipelineUUID(variableUUID)
	if variableUUID == "" {
		return fmt.Errorf("pipeline variable UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(variableUUID))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}

func parsePipelineBuildNumber(raw string) (int, bool) {
	id, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
