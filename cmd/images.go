package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newImagesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "images",
		Short: "Inspect localized Google Play store images",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newImagesListCommand(out, options, &packageName))
	return cmd
}

func newImagesListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		language  string
		imageType string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List store images for one language and image type",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedLanguage, err := play.NewListingLanguage(language)
			if err != nil {
				return err
			}
			typedImageType, err := play.NewImageType(imageType)
			if err != nil {
				return err
			}
			listOptions := play.ImageListOptions{
				PackageName: typedPackageName,
				Language:    typedLanguage,
				Type:        typedImageType,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListImages(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "BCP-47 listing language, for example en-US")
	cmd.Flags().StringVar(&imageType, "type", "", "Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots")
	return cmd
}
