package cmd

import (
	"io"
	"os"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newDataSafetyCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "data-safety",
		Short: "Update Google Play data safety declarations",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newDataSafetyUpdateCommand(out, options, &packageName))
	return cmd
}

func newDataSafetyUpdateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		csvPath string
		confirm bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Upload a data safety CSV declaration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			if err := play.ValidateReadableFile(csvPath); err != nil {
				return err
			}
			csvContent, err := os.ReadFile(csvPath)
			if err != nil {
				return err
			}
			updateOptions := play.DataSafetyUpdateOptions{
				PackageName:  typedPackageName,
				CSVPath:      csvPath,
				SafetyLabels: string(csvContent),
				Confirm:      confirm,
				DryRun:       dryRun,
			}
			if dryRun {
				result, err := play.UpdateDataSafety(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewDataSafetyUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateDataSafety(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&csvPath, "csv", "", "Path to the data safety CSV export")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the data safety update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned data safety update without calling Google Play")
	return cmd
}
