package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/auro/bitbucket_cli/internal/bitbucket"
	"github.com/auro/bitbucket_cli/internal/config"
	"github.com/auro/bitbucket_cli/internal/output"
	"github.com/spf13/cobra"
)

func newAPICmd() *cobra.Command {
	var host string
	var method string
	var inputFile string
	var jq string

	cmd := &cobra.Command{
		Use:   "api <path-or-url>",
		Short: "Make an authenticated Bitbucket API request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			resolvedHost, err := cfg.ResolveHost(strings.TrimSpace(host))
			if err != nil {
				return err
			}

			hostConfig, ok := cfg.Hosts[resolvedHost]
			if !ok {
				return fmt.Errorf("no stored credentials found for %s", resolvedHost)
			}

			client, err := bitbucket.NewClient(resolvedHost, hostConfig)
			if err != nil {
				return err
			}

			body, err := readRequestBody(cmd.InOrStdin(), inputFile)
			if err != nil {
				return err
			}

			resp, err := client.Do(context.Background(), strings.ToUpper(strings.TrimSpace(method)), args[0], body, nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read API response: %w", err)
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("bitbucket API returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
			}

			return writeAPIResponse(cmd.OutOrStdout(), responseBody, jq)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Bitbucket host to use")
	cmd.Flags().StringVarP(&method, "method", "X", http.MethodGet, "HTTP method")
	cmd.Flags().StringVar(&inputFile, "input", "", "Read request body from a file, or '-' for stdin")
	cmd.Flags().StringVar(&jq, "jq", "", "Filter JSON output using a jq expression")

	return cmd
}

func readRequestBody(stdin io.Reader, inputFile string) ([]byte, error) {
	switch strings.TrimSpace(inputFile) {
	case "":
		return nil, nil
	case "-":
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read request body from stdin: %w", err)
		}
		return data, nil
	default:
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("read request body from %q: %w", inputFile, err)
		}
		return data, nil
	}
}

func writeAPIResponse(w io.Writer, body []byte, jq string) error {
	if strings.TrimSpace(jq) == "" {
		_, err := w.Write(body)
		if err == nil && len(body) > 0 && body[len(body)-1] != '\n' {
			_, err = io.WriteString(w, "\n")
		}
		return err
	}

	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return fmt.Errorf("cannot apply --jq to non-JSON response: %w", err)
	}

	filtered, err := output.ApplyJQ(value, jq)
	if err != nil {
		return err
	}

	return output.WriteJSON(w, filtered)
}
