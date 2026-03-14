package cmd

import (
	"fmt"
	"io"

	"github.com/aurokin/bitbucket_cli/internal/output"
)

func writeProjectListSummary(w io.Writer, payload projectListPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if len(payload.Projects) == 0 {
		if _, err := fmt.Fprintf(w, "No projects found in %s.\n", payload.Workspace); err != nil {
			return err
		}
		return nil
	}
	tw := output.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "key\tname\tvis"); err != nil {
		return err
	}
	for _, project := range payload.Projects {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			output.Truncate(project.Key, 16),
			output.Truncate(project.Name, 28),
			repoVisibilityLabel(project.IsPrivate),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project view %s --workspace %s", payload.Projects[0].Key, payload.Workspace))
}

func writeProjectSummary(w io.Writer, payload projectPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.Project.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Project.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Host", payload.Host); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Project.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "UUID", payload.Project.UUID); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Project.Links.HTML.Href); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Description", payload.Project.Description); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project default-reviewer list %s --workspace %s", payload.Project.Key, payload.Workspace))
}

func writeProjectMutationSummary(w io.Writer, payload projectMutationPayload) error {
	if err := writeLabelValue(w, "Workspace", payload.Workspace); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Project", payload.Project.Key); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Name", payload.Project.Name); err != nil {
		return err
	}
	if err := writeLabelValue(w, "Action", payload.Action); err != nil {
		return err
	}
	if payload.Deleted {
		if err := writeLabelValue(w, "Status", "deleted"); err != nil {
			return err
		}
		return writeNextStep(w, fmt.Sprintf("bb project create %s --workspace %s --name %q", payload.Project.Key, payload.Workspace, payload.Project.Name))
	}
	if err := writeLabelValue(w, "Visibility", repoVisibilityLabel(payload.Project.IsPrivate)); err != nil {
		return err
	}
	if err := writeLabelValue(w, "URL", payload.Project.Links.HTML.Href); err != nil {
		return err
	}
	return writeNextStep(w, fmt.Sprintf("bb project view %s --workspace %s", payload.Project.Key, payload.Workspace))
}
