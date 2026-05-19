package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/capabilities"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newCapabilitiesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		status  string
		section string
	)

	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List gpc command parity and capability status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := capabilities.List(capabilities.ListOptions{
				Status:  capabilities.Status(status),
				Section: section,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, items)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: planned, implemented, tested, documented, blocked, not applicable")
	cmd.Flags().StringVar(&section, "section", "", "Filter by parity matrix section")
	return cmd
}
