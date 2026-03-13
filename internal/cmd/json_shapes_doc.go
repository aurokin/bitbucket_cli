package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/aurokin/bitbucket_cli/internal/bitbucket"
)

type jsonShapeSection struct {
	Title       string
	Description string
	Commands    []string
	Type        any
}

// GenerateJSONShapesDoc renders representative JSON shapes from the current
// payload structs and Bitbucket response models.
func GenerateJSONShapesDoc() (string, error) {
	sections := []jsonShapeSection{
		{
			Title:       "Workspace Inspection",
			Description: "Representative shapes for workspace list, view, membership, and workspace-scoped repository permission payloads.",
			Commands: []string{
				"bb workspace list --json '*'",
				"bb workspace view workspace-slug --json '*'",
				"bb workspace member view 557058:example --workspace workspace-slug --json '*'",
				"bb workspace repo-permission list workspace-slug --repo workspace-slug/repo-slug --json '*'",
			},
			Type: []any{workspaceListPayload{}, workspacePayload{}, workspaceMembershipPayload{}, workspaceRepoPermissionListPayload{}},
		},
		{
			Title:       "Project Inspection And Mutation",
			Description: "Representative shapes for project list, view, create or edit, default reviewer, and permission inspection payloads.",
			Commands: []string{
				"bb project list workspace-slug --json '*'",
				"bb project view BBCLI --workspace workspace-slug --json '*'",
				"bb project create TMP --workspace workspace-slug --name 'Temp project' --json '*'",
				"bb project default-reviewer list BBCLI --workspace workspace-slug --json '*'",
				"bb project permissions user view BBCLI 557058:example --workspace workspace-slug --json '*'",
			},
			Type: []any{projectListPayload{}, projectPayload{}, projectMutationPayload{}, projectDefaultReviewerListPayload{}, projectUserPermissionPayload{}, projectGroupPermissionPayload{}},
		},
		{
			Title:       "Deployments And Environments",
			Description: "Representative shapes for deployment history, deployment environment inspection, and deployment environment variable inspection.",
			Commands: []string{
				"bb deployment list --repo workspace-slug/pipelines-repo-slug --json '*'",
				"bb deployment environment list --repo workspace-slug/pipelines-repo-slug --json '*'",
				"bb deployment environment variable list --repo workspace-slug/pipelines-repo-slug --environment test --json '*'",
			},
			Type: []any{deploymentListPayload{}, deploymentPayload{}, deploymentEnvironmentListPayload{}, deploymentEnvironmentPayload{}, deploymentVariableListPayload{}, deploymentVariablePayload{}},
		},
		{
			Title:       "Repository View",
			Description: "Representative shape for the repository view payload.",
			Commands: []string{
				"bb repo view --repo workspace-slug/repo-slug --json '*'",
			},
			Type: repoViewPayload{},
		},
		{
			Title:       "Repository Clone And Delete",
			Description: "Representative shapes for repository clone and delete payloads.",
			Commands: []string{
				"bb repo clone workspace-slug/repo-slug /tmp/repo-slug --json '*'",
				"bb --no-prompt repo delete workspace-slug/delete-repo-slug --yes --json '*'",
			},
			Type: []any{repoClonePayload{}, repoDeletePayload{}},
		},
		{
			Title:       "Branches And Tags",
			Description: "Representative shapes for branch and tag list, view, create, and delete payloads.",
			Commands: []string{
				"bb branch list --repo workspace-slug/repo-slug --json '*'",
				"bb branch create feature/demo --repo workspace-slug/repo-slug --target main --json '*'",
				"bb tag list --repo workspace-slug/repo-slug --json '*'",
				"bb tag create v1.0.0 --repo workspace-slug/repo-slug --target main --json '*'",
			},
			Type: []any{branchListPayload{}, branchPayload{}, branchDeletePayload{}, tagListPayload{}, tagPayload{}, tagDeletePayload{}},
		},
		{
			Title:       "Browse",
			Description: "Representative shape for browse payloads when printing or emitting JSON instead of opening the browser.",
			Commands: []string{
				"bb browse --repo workspace-slug/repo-slug --no-browser --json '*'",
				"bb browse README.md:12 --repo workspace-slug/repo-slug --no-browser --json '*'",
				"bb browse --pr 1 --repo workspace-slug/repo-slug --no-browser --json '*'",
			},
			Type: browsePayload{},
		},
		{
			Title:       "Resolve",
			Description: "Representative shape for URL-to-entity resolution payloads used by agents and humans to normalize Bitbucket URLs.",
			Commands: []string{
				"bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'",
			},
			Type: resolvedEntity{},
		},
		{
			Title:       "Commit View, Diff, And Statuses",
			Description: "Representative shapes for commit view, diff, and commit status payloads.",
			Commands: []string{
				"bb commit view https://bitbucket.org/workspace-slug/repo-slug/commits/abc1234 --json '*'",
				"bb commit diff abc1234 --repo workspace-slug/repo-slug --json '*'",
				"bb commit statuses abc1234 --repo workspace-slug/repo-slug --json '*'",
			},
			Type: []any{commitViewPayload{}, commitDiffPayload{}, commitStatusesPayload{}},
		},
		{
			Title:       "Commit Comments And Reports",
			Description: "Representative shapes for commit comment inspection, commit approvals, and commit report inspection payloads.",
			Commands: []string{
				"bb commit comment view 15 --commit abc1234 --repo workspace-slug/repo-slug --json '*'",
				"bb commit approve abc1234 --repo workspace-slug/repo-slug --json '*'",
				"bb commit report view bb-cli-report --commit abc1234 --repo workspace-slug/repo-slug --json '*'",
			},
			Type: []any{commitCommentPayload{}, commitReviewPayload{}, commitReportPayload{}},
		},
		{
			Title:       "Pipeline List And View",
			Description: "Representative shapes for pipeline list items plus pipeline log, stop, and view payloads.",
			Commands: []string{
				"bb pipeline list --repo workspace-slug/pipelines-repo-slug --json '*'",
				"bb pipeline log 1 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}' --json '*'",
				"bb --no-prompt pipeline stop 1 --repo workspace-slug/pipelines-repo-slug --yes --json '*'",
				"bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug --json '*'",
			},
			Type: []any{bitbucket.Pipeline{}, pipelineLogPayload{}, pipelineStopPayload{}, pipelineViewPayload{}},
		},
		{
			Title:       "Pull Request List And View",
			Description: "Representative shape for pull request list items and the pull request view payload.",
			Commands: []string{
				"bb pr list --repo workspace-slug/repo-slug --json '*'",
				"bb pr view 1 --repo workspace-slug/repo-slug --json '*'",
			},
			Type: bitbucket.PullRequest{},
		},
		{
			Title:       "Pull Request Comment View And Resolution",
			Description: "Representative shape for pull request comment detail, edit, delete, resolve, and reopen payloads.",
			Commands: []string{
				"bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --json '*'",
				"bb pr comment resolve 15 --pr 1 --repo workspace-slug/repo-slug --json '*'",
			},
			Type: prCommentPayload{},
		},
		{
			Title:       "Pull Request Tasks",
			Description: "Representative shapes for pull request task lists and task detail, create, edit, resolve, reopen, and delete payloads.",
			Commands: []string{
				"bb pr task list 1 --repo workspace-slug/repo-slug --json '*'",
				"bb pr task create 1 --repo workspace-slug/repo-slug --comment 15 --body 'Handle this thread' --json '*'",
				"bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug --json '*'",
			},
			Type: []any{prTaskListPayload{}, prTaskPayload{}},
		},
		{
			Title:       "Pull Request Status",
			Description: "Representative shape for the pull request status payload.",
			Commands: []string{
				"bb pr status --repo workspace-slug/repo-slug --json '*'",
			},
			Type: prStatusPayload{},
		},
		{
			Title:       "Pull Request Diff",
			Description: "Representative shape for the pull request diff payload.",
			Commands: []string{
				"bb pr diff 1 --repo workspace-slug/repo-slug --json '*'",
			},
			Type: prDiffPayload{},
		},
		{
			Title:       "Issue List And View",
			Description: "Representative shape for issue list items and the issue view payload.",
			Commands: []string{
				"bb issue list --repo workspace-slug/issues-repo-slug --json '*'",
				"bb issue view 1 --repo workspace-slug/issues-repo-slug --json '*'",
			},
			Type: bitbucket.Issue{},
		},
		{
			Title:       "Search Results",
			Description: "Representative shapes for repository, pull request, and issue search results.",
			Commands: []string{
				"bb search repos bb-cli --workspace workspace-slug --json '*'",
				"bb search prs fixture --repo workspace-slug/repo-slug --json '*'",
				"bb search issues broken --repo workspace-slug/issues-repo-slug --json '*'",
			},
			Type: []any{bitbucket.Repository{}, bitbucket.PullRequest{}, bitbucket.Issue{}},
		},
		{
			Title:       "Cross-Repository Status",
			Description: "Representative shape for the cross-repository status payload.",
			Commands: []string{
				"bb status --workspace workspace-slug --json '*'",
			},
			Type: crossRepoStatusPayload{},
		},
	}

	var buf bytes.Buffer
	buf.WriteString("# JSON Shapes\n\n")
	buf.WriteString("Representative JSON payload shapes for common commands.\n\n")
	buf.WriteString("Generated from the current payload structs and Bitbucket response models. Field order follows the Go structs. Omitted fields in live output still depend on the selected command, flags, and Bitbucket data.\n\n")
	buf.WriteString("Use [automation.md](./automation.md) for deterministic command patterns and [cli-reference.md](./cli-reference.md) for the full command surface.\n")

	for _, section := range sections {
		buf.WriteString("\n## ")
		buf.WriteString(section.Title)
		buf.WriteString("\n\n")
		if section.Description != "" {
			buf.WriteString(section.Description)
			buf.WriteString("\n\n")
		}

		if len(section.Commands) == 1 {
			buf.WriteString("Command:\n\n```bash\n")
			buf.WriteString(section.Commands[0])
			buf.WriteString("\n```\n")
		} else {
			buf.WriteString("Commands:\n\n```bash\n")
			for i, command := range section.Commands {
				if i > 0 {
					buf.WriteByte('\n')
				}
				buf.WriteString(command)
			}
			buf.WriteString("\n```\n")
		}

		switch typed := section.Type.(type) {
		case []any:
			buf.WriteString("\nRepresentative shapes:\n")
			for _, item := range typed {
				rendered, err := representativeJSON(item)
				if err != nil {
					return "", err
				}
				buf.WriteString("\n```json\n")
				buf.WriteString(rendered)
				buf.WriteString("\n```\n")
			}
		default:
			rendered, err := representativeJSON(section.Type)
			if err != nil {
				return "", err
			}
			buf.WriteString("\nRepresentative shape:\n\n```json\n")
			buf.WriteString(rendered)
			buf.WriteString("\n```\n")
		}
	}

	return buf.String(), nil
}

func representativeJSON(example any) (string, error) {
	value := buildRepresentativeValue(reflect.TypeOf(example), nil, 0)
	raw, err := json.Marshal(value.Interface())
	if err != nil {
		return "", err
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", err
	}
	trimmed := trimRepresentativeValue(decoded, nil, 0)
	data, err := json.MarshalIndent(trimmed, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildRepresentativeValue(t reflect.Type, path []string, depth int) reflect.Value {
	if t == nil {
		return reflect.Value{}
	}
	if depth > 8 {
		return reflect.Zero(t)
	}

	switch t.Kind() {
	case reflect.Pointer:
		value := reflect.New(t.Elem())
		value.Elem().Set(buildRepresentativeValue(t.Elem(), path, depth+1))
		return value
	case reflect.Struct:
		value := reflect.New(t).Elem()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			name, ok := jsonFieldName(field)
			if !ok {
				continue
			}
			fieldValue := buildRepresentativeValue(field.Type, append(path, name), depth+1)
			if fieldValue.IsValid() && value.Field(i).CanSet() {
				value.Field(i).Set(fieldValue)
			}
		}
		return value
	case reflect.Slice:
		slice := reflect.MakeSlice(t, 1, 1)
		slice.Index(0).Set(buildRepresentativeValue(t.Elem(), append(path, "item"), depth+1))
		return slice
	case reflect.String:
		return reflect.ValueOf(sampleString(path)).Convert(t)
	case reflect.Bool:
		return reflect.ValueOf(sampleBool(path)).Convert(t)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(sampleInt(path)).Convert(t)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(uint64(sampleInt(path))).Convert(t)
	default:
		return reflect.Zero(t)
	}
}

func trimRepresentativeValue(value any, path []string, depth int) any {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		slices.Sort(keys)

		result := make(map[string]any, len(keys))
		for _, key := range keys {
			if shouldTrimJSONField(key, depth, path) {
				continue
			}
			result[key] = trimRepresentativeValue(typed[key], append(path, key), depth+1)
		}
		return result
	case []any:
		if len(typed) == 0 {
			return typed
		}
		return []any{trimRepresentativeValue(typed[0], append(path, "item"), depth+1)}
	default:
		return typed
	}
}

func shouldTrimJSONField(name string, depth int, path []string) bool {
	if isSummaryContext(path) {
		if _, ok := summaryEssentialFields[name]; ok {
			return false
		}
		return true
	}

	if depth < 3 {
		return false
	}

	if _, ok := summaryEssentialFields[name]; ok {
		return false
	}
	return true
}

var summaryEssentialFields = map[string]struct{}{
	"account_id":       {},
	"author":           {},
	"branch":           {},
	"created":          {},
	"current_branch":   {},
	"current_user":     {},
	"description":      {},
	"display_name":     {},
	"host":             {},
	"html":             {},
	"href":             {},
	"id":               {},
	"issue":            {},
	"links":            {},
	"name":             {},
	"nickname":         {},
	"path":             {},
	"pull_request":     {},
	"raw":              {},
	"repo":             {},
	"repository":       {},
	"review_requested": {},
	"source":           {},
	"state":            {},
	"title":            {},
	"user":             {},
	"username":         {},
	"workspace":        {},
}

func isSummaryContext(path []string) bool {
	for _, part := range path {
		switch part {
		case "current_branch", "created", "review_requested", "authored_prs", "review_requested_prs", "your_issues", "pull_request", "issue":
			return true
		}
	}
	return false
}

func jsonFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}
	if tag == "" {
		return field.Name, true
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return field.Name, true
	}
	return name, true
}

func sampleString(path []string) string {
	key := lastPathElement(path)
	parent := ""
	if len(path) > 1 {
		parent = path[len(path)-2]
	}

	switch key {
	case "host":
		return "bitbucket.org"
	case "workspace":
		return "workspace-slug"
	case "repo":
		return "repo-slug"
	case "full_name":
		return "workspace-slug/repo-slug"
	case "name":
		switch parent {
		case "branch":
			return "main"
		case "project":
			return "BBCLI"
		default:
			return "Example Name"
		}
	case "project_key", "key":
		return "BBCLI"
	case "project_name":
		return "Bitbucket CLI"
	case "main_branch":
		return "main"
	case "item":
		if parent == "merge_strategies" {
			return "merge_commit"
		}
		return "<item>"
	case "html_url", "href":
		return "https://bitbucket.org/workspace-slug/repo-slug"
	case "https_clone", "clone_url":
		return "https://bitbucket.org/workspace-slug/repo-slug.git"
	case "ssh_clone", "local_clone_url":
		return "git@bitbucket.org:workspace-slug/repo-slug.git"
	case "remote":
		return "origin"
	case "root", "directory":
		return "/path/to/repo"
	case "title":
		return "Example title"
	case "description", "raw", "message":
		return "Example text"
	case "state":
		if parent == "stats" || parent == "item" && containsPath(path, "stats") {
			return "modified"
		}
		return "OPEN"
	case "status":
		return "modified"
	case "display_name":
		return "Example User"
	case "account_id":
		return "account-id"
	case "nickname":
		return "example-user"
	case "username":
		return "user@example.com"
	case "uuid":
		return "{uuid}"
	case "hash":
		return "abc123def456"
	case "path", "escaped_path":
		return "file.txt"
	case "type":
		return "commit_file"
	case "role":
		return "REVIEWER"
	case "default_merge_strategy":
		return "merge_commit"
	case "patch":
		return "diff --git a/file.txt b/file.txt\n..."
	case "created_on", "updated_on", "updated_at":
		return "2026-03-11T00:00:00Z"
	case "user":
		return "Example User"
	default:
		return fmt.Sprintf("<%s>", strings.ReplaceAll(key, "_", "-"))
	}
}

func sampleBool(path []string) bool {
	key := lastPathElement(path)
	switch key {
	case "deleted", "approved", "private", "is_private", "close_source_branch":
		return true
	default:
		return true
	}
}

func sampleInt(path []string) int64 {
	key := lastPathElement(path)
	switch {
	case strings.Contains(key, "count"):
		return 2
	case strings.Contains(key, "limit"):
		return 20
	case strings.Contains(key, "total"):
		return 3
	case strings.Contains(key, "repositories_scanned"):
		return 4
	default:
		return 1
	}
}

func lastPathElement(path []string) string {
	if len(path) == 0 {
		return "value"
	}
	return path[len(path)-1]
}

func containsPath(path []string, target string) bool {
	for _, part := range path {
		if part == target {
			return true
		}
	}
	return false
}
