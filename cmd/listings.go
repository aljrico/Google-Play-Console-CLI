package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newListingsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "listings",
		Short: "Manage localized Google Play store listings",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newListingsListCommand(out, options, &packageName),
		newListingsGetCommand(out, options, &packageName),
		newListingsUpdateCommand(out, options, &packageName),
	)
	return cmd
}

func newListingsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List localized store listings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			listings, err := play.ListListings(cmd.Context(), publisher, typedPackageName)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, listings)
		},
	}
}

func newListingsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var language string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one localized store listing",
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
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			listing, err := play.GetListing(cmd.Context(), publisher, typedPackageName, typedLanguage)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, listing)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "BCP-47 listing language, for example en-US")
	return cmd
}

func newListingsUpdateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		language         string
		title            string
		shortDescription string
		fullDescription  string
		video            string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Create or update one localized store listing",
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
			updateOptions := play.UpdateListingOptions{
				PackageName: typedPackageName,
				Listing: play.Listing{
					Language: typedLanguage,
				},
				Confirm: confirm,
				DryRun:  dryRun,
			}
			if cmd.Flags().Changed("title") {
				updateOptions.Listing.Title = &title
			}
			if cmd.Flags().Changed("short-description") {
				updateOptions.Listing.ShortDescription = &shortDescription
			}
			if cmd.Flags().Changed("full-description") {
				updateOptions.Listing.FullDescription = &fullDescription
			}
			if cmd.Flags().Changed("video") {
				updateOptions.Listing.Video = &video
			}
			if dryRun {
				result, err := play.UpdateListing(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}

			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateListing(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "BCP-47 listing language, for example en-US")
	cmd.Flags().StringVar(&title, "title", "", "Localized app title")
	cmd.Flags().StringVar(&shortDescription, "short-description", "", "Localized short description")
	cmd.Flags().StringVar(&fullDescription, "full-description", "", "Localized full description")
	cmd.Flags().StringVar(&video, "video", "", "Promotional YouTube video URL")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned listing update without calling Google Play")
	return cmd
}
