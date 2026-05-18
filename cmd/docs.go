package cmd

import (
	"fmt"
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/docs"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

type docsDocument struct {
	Name    string `json:"name"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

func newDocsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Print embedded gpc documentation",
	}
	cmd.AddCommand(newDocsParityCommand(out, options))
	return cmd
}

func newDocsParityCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parity",
		Short: "Print the asc-to-gpc parity matrix",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.output == output.Markdown {
				_, err := fmt.Fprint(out, docs.ParityMatrix)
				return err
			}
			return output.Write(out, options.output, options.pretty, docsDocument{
				Name:    "parity",
				Format:  "markdown",
				Content: docs.ParityMatrix,
			})
		},
	}
	return cmd
}
