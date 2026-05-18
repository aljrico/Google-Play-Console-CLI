package cmd

import (
	"io"
	"net/http"
	"time"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/schema"
	"github.com/spf13/cobra"
)

var schemaDiscoveryURL = schema.DefaultDiscoveryURL
var schemaHTTPClient = &http.Client{Timeout: 30 * time.Second}

func newSchemaCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		resource string
		method   string
	)

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the Google Play discovery schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			document, err := schema.Fetch(cmd.Context(), schemaHTTPClient, schema.FetchOptions{
				DiscoveryURL: schemaDiscoveryURL,
				Resource:     resource,
				Method:       method,
			})
			if err != nil {
				return err
			}
			if options.output == output.Table || options.output == output.Markdown {
				return output.Write(out, options.output, options.pretty, schema.MethodSummaries(document))
			}
			return output.Write(out, options.output, options.pretty, document)
		},
	}
	cmd.Flags().StringVar(&resource, "resource", "", "Filter by dotted discovery resource path, for example edits.tracks")
	cmd.Flags().StringVar(&method, "method", "", "Filter by discovery method name or ID, for example list or androidpublisher.edits.tracks.list")
	return cmd
}
