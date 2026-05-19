package cmd

import (
	"fmt"
	"io"

	gpcdiff "github.com/aljrico/Google-Play-Console-CLI/internal/diff"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newDiffCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare local Google Play payloads",
	}
	cmd.AddCommand(newDiffJSONCommand(out, options))
	return cmd
}

func newDiffJSONCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var failOnChange bool

	cmd := &cobra.Command{
		Use:   "json FROM TO",
		Short: "Compare two JSON files with stable JSON Pointer paths",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := gpcdiff.CompareJSONFiles(gpcdiff.JSONOptions{
				FromPath: args[0],
				ToPath:   args[1],
			})
			if err != nil {
				return err
			}
			if err := output.Write(out, options.output, options.pretty, result); err != nil {
				return err
			}
			if failOnChange && !result.Equal {
				return fmt.Errorf("JSON files differ: %d change(s)", len(result.Changes))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&failOnChange, "fail-on-change", false, "Exit nonzero when the JSON files differ")
	return cmd
}
