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
		Short: "Manage localized Google Play store images",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newImagesListCommand(out, options, &packageName),
		newImagesUploadCommand(out, options, &packageName),
		newImagesDeleteCommand(out, options, &packageName),
		newImagesDeleteAllCommand(out, options, &packageName),
	)
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

func newImagesUploadCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		language  string
		imageType string
		path      string
		confirm   bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload one store image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			uploadOptions, err := imageUploadOptions(*packageName, language, imageType, path, confirm, dryRun)
			if err != nil {
				return err
			}
			if err := uploadOptions.Validate(); err != nil {
				return err
			}
			if err := play.ValidateReadableImageFile(path); err != nil {
				return err
			}
			if dryRun {
				result, err := play.UploadImage(cmd.Context(), nil, uploadOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UploadImage(cmd.Context(), publisher, uploadOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "BCP-47 listing language, for example en-US")
	cmd.Flags().StringVar(&imageType, "type", "", "Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots")
	cmd.Flags().StringVar(&path, "file", "", "Path to a .jpg, .jpeg, or .png image")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned image upload without calling Google Play")
	return cmd
}

func newImagesDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		language  string
		imageType string
		imageID   string
		confirm   bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete one store image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteOptions, err := imageDeleteOptions(*packageName, language, imageType, imageID, false, confirm, dryRun)
			if err != nil {
				return err
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteImages(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteImages(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "BCP-47 listing language, for example en-US")
	cmd.Flags().StringVar(&imageType, "type", "", "Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots")
	cmd.Flags().StringVar(&imageID, "image-id", "", "Google Play image ID to delete")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned image deletion without calling Google Play")
	return cmd
}

func newImagesDeleteAllCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		language  string
		imageType string
		confirm   bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "delete-all",
		Short: "Delete all store images for one language and image type",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteOptions, err := imageDeleteOptions(*packageName, language, imageType, "", true, confirm, dryRun)
			if err != nil {
				return err
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteImages(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteImages(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "BCP-47 listing language, for example en-US")
	cmd.Flags().StringVar(&imageType, "type", "", "Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned image deletion without calling Google Play")
	return cmd
}

func imageDeleteOptions(packageName string, language string, imageType string, imageID string, all bool, confirm bool, dryRun bool) (play.ImageDeleteOptions, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return play.ImageDeleteOptions{}, err
	}
	typedLanguage, err := play.NewListingLanguage(language)
	if err != nil {
		return play.ImageDeleteOptions{}, err
	}
	typedImageType, err := play.NewImageType(imageType)
	if err != nil {
		return play.ImageDeleteOptions{}, err
	}
	return play.ImageDeleteOptions{
		PackageName: typedPackageName,
		Language:    typedLanguage,
		Type:        typedImageType,
		ImageID:     imageID,
		All:         all,
		Confirm:     confirm,
		DryRun:      dryRun,
	}, nil
}

func imageUploadOptions(packageName string, language string, imageType string, path string, confirm bool, dryRun bool) (play.ImageUploadOptions, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return play.ImageUploadOptions{}, err
	}
	typedLanguage, err := play.NewListingLanguage(language)
	if err != nil {
		return play.ImageUploadOptions{}, err
	}
	typedImageType, err := play.NewImageType(imageType)
	if err != nil {
		return play.ImageUploadOptions{}, err
	}
	return play.ImageUploadOptions{
		PackageName: typedPackageName,
		Language:    typedLanguage,
		Type:        typedImageType,
		Path:        path,
		Confirm:     confirm,
		DryRun:      dryRun,
	}, nil
}
