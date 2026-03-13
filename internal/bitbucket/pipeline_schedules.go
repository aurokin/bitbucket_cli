package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type PipelineSchedule struct {
	Type        string         `json:"type,omitempty"`
	UUID        string         `json:"uuid,omitempty"`
	Enabled     bool           `json:"enabled"`
	Target      PipelineTarget `json:"target,omitempty"`
	CronPattern string         `json:"cron_pattern,omitempty"`
	CreatedOn   string         `json:"created_on,omitempty"`
	UpdatedOn   string         `json:"updated_on,omitempty"`
}

type PipelineSelector struct {
	Type    string `json:"type,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

type CreatePipelineScheduleOptions struct {
	RefName         string
	CronPattern     string
	Enabled         bool
	SelectorType    string
	SelectorPattern string
}

type listPipelineSchedulesResponse struct {
	Values []PipelineSchedule `json:"values"`
	Next   string             `json:"next,omitempty"`
}

func (c *Client) ListPipelineSchedules(ctx context.Context, workspace, repoSlug string, limit int) ([]PipelineSchedule, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 100
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/schedules?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), limit)
	nextPath := path
	all := make([]PipelineSchedule, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page listPipelineSchedulesResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pipeline schedules: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetPipelineSchedule(ctx context.Context, workspace, repoSlug, scheduleUUID string) (PipelineSchedule, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineSchedule{}, fmt.Errorf("workspace and repository are required")
	}
	scheduleUUID = normalizePipelineUUID(scheduleUUID)
	if scheduleUUID == "" {
		return PipelineSchedule{}, fmt.Errorf("pipeline schedule UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/schedules/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(scheduleUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PipelineSchedule{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineSchedule{}, err
	}

	var schedule PipelineSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return PipelineSchedule{}, fmt.Errorf("decode pipeline schedule: %w", err)
	}
	return schedule, nil
}

func (c *Client) CreatePipelineSchedule(ctx context.Context, workspace, repoSlug string, options CreatePipelineScheduleOptions) (PipelineSchedule, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineSchedule{}, fmt.Errorf("workspace and repository are required")
	}
	if options.RefName == "" {
		return PipelineSchedule{}, fmt.Errorf("pipeline schedule ref name is required")
	}
	if options.CronPattern == "" {
		return PipelineSchedule{}, fmt.Errorf("pipeline schedule cron pattern is required")
	}
	if options.SelectorType == "" {
		options.SelectorType = "branches"
	}
	if options.SelectorPattern == "" {
		options.SelectorPattern = options.RefName
	}

	body := map[string]any{
		"type": "pipeline_schedule",
		"target": map[string]any{
			"type":     "pipeline_ref_target",
			"ref_type": "branch",
			"ref_name": options.RefName,
			"selector": map[string]any{
				"type":    options.SelectorType,
				"pattern": options.SelectorPattern,
			},
		},
		"cron_pattern": options.CronPattern,
		"enabled":      options.Enabled,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PipelineSchedule{}, fmt.Errorf("marshal pipeline schedule request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/schedules", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PipelineSchedule{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineSchedule{}, err
	}

	var schedule PipelineSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return PipelineSchedule{}, fmt.Errorf("decode created pipeline schedule: %w", err)
	}
	return schedule, nil
}

func (c *Client) UpdatePipelineScheduleEnabled(ctx context.Context, workspace, repoSlug, scheduleUUID string, enabled bool) (PipelineSchedule, error) {
	if workspace == "" || repoSlug == "" {
		return PipelineSchedule{}, fmt.Errorf("workspace and repository are required")
	}
	scheduleUUID = normalizePipelineUUID(scheduleUUID)
	if scheduleUUID == "" {
		return PipelineSchedule{}, fmt.Errorf("pipeline schedule UUID is required")
	}

	payload, err := json.Marshal(map[string]any{
		"type":    "pipeline_schedule",
		"enabled": enabled,
	})
	if err != nil {
		return PipelineSchedule{}, fmt.Errorf("marshal pipeline schedule update request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/schedules/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(scheduleUUID))
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return PipelineSchedule{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PipelineSchedule{}, err
	}

	var schedule PipelineSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return PipelineSchedule{}, fmt.Errorf("decode updated pipeline schedule: %w", err)
	}
	return schedule, nil
}

func (c *Client) DeletePipelineSchedule(ctx context.Context, workspace, repoSlug, scheduleUUID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	scheduleUUID = normalizePipelineUUID(scheduleUUID)
	if scheduleUUID == "" {
		return fmt.Errorf("pipeline schedule UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/schedules/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(scheduleUUID))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}
