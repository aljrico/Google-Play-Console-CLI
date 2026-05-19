package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newMetadataCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Apply app details and localized listings from a file",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newMetadataApplyCommand(out, options, &packageName))
	return cmd
}

func newMetadataApplyCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		filePath                string
		confirm                 bool
		dryRun                  bool
		changesNotSentForReview bool
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a JSON metadata file through one Google Play edit",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requestOptions := play.MetadataApplyOptions{
				PackageName:             typedPackageName,
				FilePath:                filePath,
				Confirm:                 confirm,
				DryRun:                  dryRun,
				ChangesNotSentForReview: changesNotSentForReview,
			}
			if err := requestOptions.ValidateRequest(); err != nil {
				return err
			}
			metadataFile, err := play.LoadMetadataFile(filePath)
			if err != nil {
				return err
			}
			applyOptions := play.MetadataApplyOptions{
				PackageName:             typedPackageName,
				FilePath:                filePath,
				Details:                 metadataFile.Details,
				Listings:                metadataFile.Listings,
				Confirm:                 confirm,
				DryRun:                  dryRun,
				ChangesNotSentForReview: changesNotSentForReview,
			}
			if dryRun {
				result, err := play.ApplyMetadata(cmd.Context(), nil, applyOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewMetadataApplyPlan(applyOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ApplyMetadata(cmd.Context(), publisher, applyOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to metadata JSON")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned metadata update without calling Google Play")
	cmd.Flags().BoolVar(&changesNotSentForReview, "changes-not-sent-for-review", false, "Commit without sending changes for review (required when the app is in an enforcement-required state)")
	return cmd
}
