package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

func resolvePipelineCommandTarget(ctx context.Context, host, workspace, repo, ref string) (resolvedRepoCommandTarget, bitbucket.Pipeline, error) {
	resolved, err := resolveRepoCommandTarget(ctx, host, workspace, repo, true)
	if err != nil {
		return resolvedRepoCommandTarget{}, bitbucket.Pipeline{}, err
	}

	pipeline, err := resolvePipelineReference(ctx, resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, ref)
	if err != nil {
		return resolvedRepoCommandTarget{}, bitbucket.Pipeline{}, err
	}

	return resolved, pipeline, nil
}

func resolvePipelineStepCommandTarget(ctx context.Context, host, workspace, repo, pipelineRef, stepRef string) (resolvedRepoCommandTarget, bitbucket.Pipeline, bitbucket.PipelineStep, error) {
	resolved, pipeline, err := resolvePipelineCommandTarget(ctx, host, workspace, repo, pipelineRef)
	if err != nil {
		return resolvedRepoCommandTarget{}, bitbucket.Pipeline{}, bitbucket.PipelineStep{}, err
	}

	step, err := resolvePipelineStepReference(ctx, resolved.Client, resolved.Target.Workspace, resolved.Target.Repo, pipeline.UUID, stepRef)
	if err != nil {
		return resolvedRepoCommandTarget{}, bitbucket.Pipeline{}, bitbucket.PipelineStep{}, err
	}

	return resolved, pipeline, step, nil
}

func resolvePipelineReference(ctx context.Context, client *bitbucket.Client, workspace, repo, raw string) (bitbucket.Pipeline, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return bitbucket.Pipeline{}, fmt.Errorf("pipeline reference is required")
	}

	if buildNumber, err := strconv.Atoi(raw); err == nil {
		return client.GetPipelineByBuildNumber(ctx, workspace, repo, buildNumber)
	}

	return client.GetPipeline(ctx, workspace, repo, raw)
}

func resolvePipelineStepReference(ctx context.Context, client *bitbucket.Client, workspace, repo, pipelineUUID, raw string) (bitbucket.PipelineStep, error) {
	steps, err := client.ListPipelineSteps(ctx, workspace, repo, pipelineUUID)
	if err != nil {
		return bitbucket.PipelineStep{}, err
	}
	return resolvePipelineStep(steps, raw)
}

func resolvePipelineStep(steps []bitbucket.PipelineStep, raw string) (bitbucket.PipelineStep, error) {
	raw = strings.TrimSpace(raw)
	if len(steps) == 0 {
		return bitbucket.PipelineStep{}, fmt.Errorf("no pipeline steps are available for this run")
	}
	if raw == "" {
		if len(steps) == 1 {
			return steps[0], nil
		}
		return bitbucket.PipelineStep{}, fmt.Errorf("multiple pipeline steps are available; pass --step with one of: %s", pipelineStepChoices(steps))
	}

	normalized := strings.Trim(raw, "{}")
	for _, step := range steps {
		if raw == step.UUID || normalized == strings.Trim(step.UUID, "{}") || raw == step.Name {
			return step, nil
		}
	}

	return bitbucket.PipelineStep{}, fmt.Errorf("pipeline step %q was not found; available steps: %s", raw, pipelineStepChoices(steps))
}

func waitForStoppedPipeline(ctx context.Context, client *bitbucket.Client, workspace, repo, pipelineUUID string) (bitbucket.Pipeline, error) {
	var last bitbucket.Pipeline
	for attempt := 0; attempt < 12; attempt++ {
		pipeline, err := client.GetPipeline(ctx, workspace, repo, pipelineUUID)
		if err != nil {
			return bitbucket.Pipeline{}, err
		}
		last = pipeline
		label := pipelineStateLabel(pipeline.State)
		if strings.EqualFold(label, "STOPPED") || strings.EqualFold(pipeline.State.Name, "STOPPED") {
			return pipeline, nil
		}
		if !pipelineStateActive(pipeline.State) {
			return pipeline, nil
		}
		time.Sleep(2 * time.Second)
	}
	return last, nil
}
