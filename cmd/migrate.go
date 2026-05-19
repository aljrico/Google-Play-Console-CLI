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
		Short: "Inspect local metadata for migration",
	}
	cmd.AddCommand(newMigrateSupplyCommand(out, options))
	return cmd
}

func newMigrateSupplyCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supply",
		Short: "Inspect fastlane supply metadata",
	}
	cmd.AddCommand(
		newMigrateSupplyInspectCommand(out, options),
		newMigrateSupplyConvertCommand(out, options),
		newMigrateSupplyChangelogsCommand(out, options),
		newMigrateSupplyImagesCommand(out, options),
	)
	return cmd
}

func newMigrateSupplyConvertCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var directory string
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert fastlane supply listings to gpc metadata JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			metadata, err := supply.Convert(cmd.Context(), supply.ConvertOptions{Directory: directory})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, metadata)
		},
	}
	cmd.Flags().StringVar(&directory, "directory", supply.DefaultMetadataDirectory, "fastlane supply metadata directory")
	return cmd
}

func newMigrateSupplyChangelogsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		directory   string
		versionCode int64
	)
	cmd := &cobra.Command{
		Use:   "changelogs",
		Short: "Convert fastlane supply changelogs to release-note payloads",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			changelogs, err := supply.ConvertChangelogs(cmd.Context(), supply.ConvertChangelogsOptions{
				Directory:   directory,
				VersionCode: versionCode,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, changelogs)
		},
	}
	cmd.Flags().StringVar(&directory, "directory", supply.DefaultMetadataDirectory, "fastlane supply metadata directory")
	cmd.Flags().Int64Var(&versionCode, "version-code", 0, "Only include changelogs for this version code")
	return cmd
}

func newMigrateSupplyImagesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		directory string
		language  string
		imageType string
	)
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Convert fastlane supply images to image upload payloads",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			images, err := supply.ConvertImages(cmd.Context(), supply.ConvertImagesOptions{
				Directory: directory,
				Language:  language,
				Type:      imageType,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, images)
		},
	}
	cmd.Flags().StringVar(&directory, "directory", supply.DefaultMetadataDirectory, "fastlane supply metadata directory")
	cmd.Flags().StringVar(&language, "language", "", "Only include images for this BCP-47 listing language")
	cmd.Flags().StringVar(&imageType, "type", "", "Only include this image type")
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
