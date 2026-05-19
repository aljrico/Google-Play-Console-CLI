package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/websurface"
	"github.com/spf13/cobra"
)

func newWebCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Inspect Play Console browser automation support",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Explain the Play Console browser automation boundary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.Write(out, options.output, options.pretty, websurface.CurrentStatus())
		},
	})
	return cmd
}
