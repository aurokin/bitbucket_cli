package cmd

import (
	"fmt"
	"io"
	"runtime"

	"github.com/aurokin/bitbucket_cli/internal/output"
	"github.com/aurokin/bitbucket_cli/internal/version"
	"github.com/spf13/cobra"
)

type versionPayload struct {
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
	GoVersion string `json:"go_version"`
}

func newVersionCmd() *cobra.Command {
	var flags formatFlags

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show bb version information",
		Example: "  bb version\n" +
			"  bb version --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := flags.options()
			if err != nil {
				return err
			}

			payload := versionPayload{
				Version:   version.Version,
				Commit:    version.Commit,
				BuildDate: version.BuildDate,
				GoVersion: runtime.Version(),
			}

			return output.Render(cmd.OutOrStdout(), opts, payload, func(w io.Writer) error {
				_, err := fmt.Fprintf(w, "bb %s\n", version.Version)
				return err
			})
		},
	}

	addFormatFlags(cmd, &flags)

	return cmd
}
