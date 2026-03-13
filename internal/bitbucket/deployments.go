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

type Deployment struct {
	UUID        string                `json:"uuid,omitempty"`
	State       DeploymentState       `json:"state,omitempty"`
	Environment DeploymentEnvironment `json:"environment,omitempty"`
	Release     DeploymentRelease     `json:"release,omitempty"`
}

type DeploymentState struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type DeploymentRelease struct {
	Name string `json:"name,omitempty"`
}

type DeploymentEnvironment struct {
	UUID                   string                        `json:"uuid,omitempty"`
	Name                   string                        `json:"name,omitempty"`
	Slug                   string                        `json:"slug,omitempty"`
	Type                   string                        `json:"type,omitempty"`
	Hidden                 bool                          `json:"hidden,omitempty"`
	Rank                   int                           `json:"rank,omitempty"`
	EnvironmentLockEnabled bool                          `json:"environment_lock_enabled,omitempty"`
	DeploymentGateEnabled  bool                          `json:"deployment_gate_enabled,omitempty"`
	Category               DeploymentEnvironmentCategory `json:"category,omitempty"`
	EnvironmentType        DeploymentEnvironmentType     `json:"environment_type,omitempty"`
	Lock                   DeploymentEnvironmentLock     `json:"lock,omitempty"`
}

type DeploymentEnvironmentCategory struct {
	Name string `json:"name,omitempty"`
}

type DeploymentEnvironmentType struct {
	Name string `json:"name,omitempty"`
	Rank int    `json:"rank,omitempty"`
	Type string `json:"type,omitempty"`
}

type DeploymentEnvironmentLock struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type DeploymentVariable = PipelineVariable

type ListDeploymentVariablesOptions struct {
	Limit int
}

type deploymentListResponse struct {
	Values []Deployment `json:"values"`
	Next   string       `json:"next,omitempty"`
}

type deploymentEnvironmentListResponse struct {
	Values []DeploymentEnvironment `json:"values"`
	Next   string                  `json:"next,omitempty"`
}

func (c *Client) ListDeployments(ctx context.Context, workspace, repoSlug string, limit int) ([]Deployment, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/deployments/?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]Deployment, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page deploymentListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode deployment list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetDeployment(ctx context.Context, workspace, repoSlug, deploymentUUID string) (Deployment, error) {
	if workspace == "" || repoSlug == "" {
		return Deployment{}, fmt.Errorf("workspace and repository are required")
	}
	deploymentUUID = normalizePipelineUUID(deploymentUUID)
	if deploymentUUID == "" {
		return Deployment{}, fmt.Errorf("deployment UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/deployments/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(deploymentUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return Deployment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return Deployment{}, err
	}

	var item Deployment
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return Deployment{}, fmt.Errorf("decode deployment: %w", err)
	}
	return item, nil
}

func (c *Client) ListDeploymentEnvironments(ctx context.Context, workspace, repoSlug string, limit int) ([]DeploymentEnvironment, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/environments/?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode())
	all := make([]DeploymentEnvironment, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page deploymentEnvironmentListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode deployment environments: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetDeploymentEnvironment(ctx context.Context, workspace, repoSlug, environmentUUID string) (DeploymentEnvironment, error) {
	if workspace == "" || repoSlug == "" {
		return DeploymentEnvironment{}, fmt.Errorf("workspace and repository are required")
	}
	environmentUUID = normalizePipelineUUID(environmentUUID)
	if environmentUUID == "" {
		return DeploymentEnvironment{}, fmt.Errorf("environment UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/environments/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(environmentUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return DeploymentEnvironment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return DeploymentEnvironment{}, err
	}

	var item DeploymentEnvironment
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return DeploymentEnvironment{}, fmt.Errorf("decode deployment environment: %w", err)
	}
	return item, nil
}

func (c *Client) ListDeploymentVariables(ctx context.Context, workspace, repoSlug, environmentUUID string, options ListDeploymentVariablesOptions) ([]DeploymentVariable, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	environmentUUID = normalizePipelineUUID(environmentUUID)
	if environmentUUID == "" {
		return nil, fmt.Errorf("environment UUID is required")
	}
	if options.Limit <= 0 {
		options.Limit = 100
	}

	nextPath := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables?pagelen=%d", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(environmentUUID), options.Limit)
	all := make([]DeploymentVariable, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page paginatedPipelineVariablesResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode deployment variables: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}
	return all, nil
}

func (c *Client) GetDeploymentVariable(ctx context.Context, workspace, repoSlug, environmentUUID, variableUUID string) (DeploymentVariable, error) {
	if workspace == "" || repoSlug == "" {
		return DeploymentVariable{}, fmt.Errorf("workspace and repository are required")
	}
	environmentUUID = normalizePipelineUUID(environmentUUID)
	if environmentUUID == "" {
		return DeploymentVariable{}, fmt.Errorf("environment UUID is required")
	}
	variableUUID = normalizePipelineUUID(variableUUID)
	if variableUUID == "" {
		return DeploymentVariable{}, fmt.Errorf("deployment variable UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(environmentUUID), url.PathEscape(variableUUID))
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return DeploymentVariable{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return DeploymentVariable{}, err
	}

	var item DeploymentVariable
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return DeploymentVariable{}, fmt.Errorf("decode deployment variable: %w", err)
	}
	return item, nil
}

func (c *Client) CreateDeploymentVariable(ctx context.Context, workspace, repoSlug, environmentUUID string, variable DeploymentVariable) (DeploymentVariable, error) {
	if workspace == "" || repoSlug == "" {
		return DeploymentVariable{}, fmt.Errorf("workspace and repository are required")
	}
	environmentUUID = normalizePipelineUUID(environmentUUID)
	if environmentUUID == "" {
		return DeploymentVariable{}, fmt.Errorf("environment UUID is required")
	}
	if strings.TrimSpace(variable.Key) == "" {
		return DeploymentVariable{}, fmt.Errorf("deployment variable key is required")
	}

	payload, err := json.Marshal(variable)
	if err != nil {
		return DeploymentVariable{}, fmt.Errorf("marshal deployment variable request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(environmentUUID))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return DeploymentVariable{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return DeploymentVariable{}, err
	}

	var created DeploymentVariable
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return DeploymentVariable{}, fmt.Errorf("decode created deployment variable: %w", err)
	}
	return created, nil
}

func (c *Client) UpdateDeploymentVariable(ctx context.Context, workspace, repoSlug, environmentUUID, variableUUID string, variable DeploymentVariable) (DeploymentVariable, error) {
	if workspace == "" || repoSlug == "" {
		return DeploymentVariable{}, fmt.Errorf("workspace and repository are required")
	}
	environmentUUID = normalizePipelineUUID(environmentUUID)
	if environmentUUID == "" {
		return DeploymentVariable{}, fmt.Errorf("environment UUID is required")
	}
	variableUUID = normalizePipelineUUID(variableUUID)
	if variableUUID == "" {
		return DeploymentVariable{}, fmt.Errorf("deployment variable UUID is required")
	}
	if strings.TrimSpace(variable.Key) == "" {
		return DeploymentVariable{}, fmt.Errorf("deployment variable key is required")
	}

	payload, err := json.Marshal(variable)
	if err != nil {
		return DeploymentVariable{}, fmt.Errorf("marshal deployment variable request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(environmentUUID), url.PathEscape(variableUUID))
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return DeploymentVariable{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return DeploymentVariable{}, err
	}

	var updated DeploymentVariable
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return DeploymentVariable{}, fmt.Errorf("decode updated deployment variable: %w", err)
	}
	return updated, nil
}

func (c *Client) DeleteDeploymentVariable(ctx context.Context, workspace, repoSlug, environmentUUID, variableUUID string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	environmentUUID = normalizePipelineUUID(environmentUUID)
	if environmentUUID == "" {
		return fmt.Errorf("environment UUID is required")
	}
	variableUUID = normalizePipelineUUID(variableUUID)
	if variableUUID == "" {
		return fmt.Errorf("deployment variable UUID is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/%s", url.PathEscape(workspace), url.PathEscape(repoSlug), url.PathEscape(environmentUUID), url.PathEscape(variableUUID))
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}
