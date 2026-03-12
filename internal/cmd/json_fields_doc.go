package cmd

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
)

type jsonFieldDocEntry struct {
	Command string
	Type    any
}

func GenerateJSONFieldsDoc() (string, error) {
	entries := []jsonFieldDocEntry{
		{Command: "bb repo view", Type: repoViewPayload{}},
		{Command: "bb repo clone", Type: repoClonePayload{}},
		{Command: "bb repo delete", Type: repoDeletePayload{}},
		{Command: "bb browse", Type: browsePayload{}},
		{Command: "bb pipeline list", Type: bitbucket.Pipeline{}},
		{Command: "bb pipeline log", Type: pipelineLogPayload{}},
		{Command: "bb pipeline stop", Type: pipelineStopPayload{}},
		{Command: "bb pipeline view", Type: pipelineViewPayload{}},
		{Command: "bb pr list", Type: bitbucket.PullRequest{}},
		{Command: "bb pr view", Type: bitbucket.PullRequest{}},
		{Command: "bb pr status", Type: prStatusPayload{}},
		{Command: "bb pr diff", Type: prDiffPayload{}},
		{Command: "bb issue list", Type: bitbucket.Issue{}},
		{Command: "bb issue view", Type: bitbucket.Issue{}},
		{Command: "bb search repos", Type: bitbucket.Repository{}},
		{Command: "bb search prs", Type: bitbucket.PullRequest{}},
		{Command: "bb search issues", Type: bitbucket.Issue{}},
		{Command: "bb status", Type: crossRepoStatusPayload{}},
		{Command: "bb auth status", Type: authStatusPayload{}},
		{Command: "bb config list", Type: configValue{}},
		{Command: "bb alias list", Type: aliasEntry{}},
		{Command: "bb extension list", Type: extensionEntry{}},
	}

	var b strings.Builder
	b.WriteString("# JSON Field Index\n\n")
	b.WriteString("Generated from the current payload structs and Bitbucket response models.\n\n")
	b.WriteString("Use this file to discover top-level field names for `--json` selection.\n\n")
	b.WriteString("| Command | Top-level fields | Example |\n")
	b.WriteString("|---|---|---|\n")

	for _, entry := range entries {
		fields := topLevelJSONFields(reflect.TypeOf(entry.Type))
		example := exampleJSONSelection(fields)
		fmt.Fprintf(&b, "| `%s` | %s | `%s --json %s` |\n", entry.Command, strings.Join(wrapCode(fields), ", "), entry.Command, example)
	}

	return b.String(), nil
}

func topLevelJSONFields(t reflect.Type) []string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	fields := make([]string, 0)
	if t.Kind() != reflect.Struct {
		return fields
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, ok := jsonFieldName(field)
		if !ok {
			continue
		}
		fields = append(fields, name)
	}
	slices.Sort(fields)
	return fields
}

func exampleJSONSelection(fields []string) string {
	if len(fields) == 0 {
		return "*"
	}
	limit := 3
	if len(fields) < limit {
		limit = len(fields)
	}
	return strings.Join(fields[:limit], ",")
}

func wrapCode(values []string) []string {
	wrapped := make([]string, 0, len(values))
	for _, value := range values {
		wrapped = append(wrapped, fmt.Sprintf("`%s`", value))
	}
	return wrapped
}
