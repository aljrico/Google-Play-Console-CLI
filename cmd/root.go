package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type globalOptions struct {
	output output.Format
	pretty bool
}

func newRootCommand(out io.Writer) *cobra.Command {
	options := &globalOptions{}
	cmd := &cobra.Command{
		Use:           "gpc",
		Short:         "Fast, scriptable CLI for the Google Play Developer API",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().VarP(&options.output, "output", "o", "Output format: json, table, markdown")
	cmd.PersistentFlags().BoolVar(&options.pretty, "pretty", false, "Pretty-print JSON output")
	cmd.SetOut(out)
	cmd.SetErr(out)

	cmd.AddCommand(
		newVersionCommand(out, options),
		newAuthCommand(out, options),
		newAppsCommand(out),
		newTracksCommand(out, options),
		newReleasesCommand(out, options),
		newPublishCommand(out, options),
		newListingsCommand(out, options),
		newDetailsCommand(out, options),
	)

	return cmd
}

func Execute() error {
	cmd := newRootCommand(os.Stdout)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
