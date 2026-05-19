package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/supply"
	"github.com/spf13/cobra"
)

func newMigrateCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Inspect and migrate local Google Play metadata",
	}
	cmd.AddCommand(newMigrateSupplyCommand(out, options))
	return cmd
}

func newMigrateSupplyCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply",
		Short: "Inspect fastlane supply metadata",
	}
	cmd.AddCommand(newMigrateSupplyInspectCommand(out, options))
	return cmd
}

func newMigrateSupplyInspectCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var directory string
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inventory a fastlane supply metadata directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inventory, err := supply.Inspect(cmd.Context(), supply.InspectOptions{Directory: directory})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, inventory)
		},
	}
	cmd.Flags().StringVar(&directory, "directory", supply.DefaultMetadataDirectory, "fastlane supply metadata directory")
	return cmd
}
