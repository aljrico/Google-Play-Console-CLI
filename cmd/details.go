package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newDetailsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "details",
		Short: "Manage app-level Google Play details",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newDetailsGetCommand(out, options, &packageName), newDetailsUpdateCommand(out, options, &packageName))
	return cmd
}

func newDetailsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get app-level details",
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
			details, err := play.GetAppDetails(cmd.Context(), publisher, typedPackageName)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, details)
		},
	}
}

func newDetailsUpdateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		defaultLanguage string
		contactWebsite  string
		contactEmail    string
		contactPhone    string
		confirm         bool
		dryRun          bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Patch app-level details",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			updateOptions := play.UpdateDetailsOptions{
				PackageName: typedPackageName,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if cmd.Flags().Changed("default-language") {
				updateOptions.Details.DefaultLanguage = &defaultLanguage
			}
			if cmd.Flags().Changed("contact-website") {
				updateOptions.Details.ContactWebsite = &contactWebsite
			}
			if cmd.Flags().Changed("contact-email") {
				updateOptions.Details.ContactEmail = &contactEmail
			}
			if cmd.Flags().Changed("contact-phone") {
				updateOptions.Details.ContactPhone = &contactPhone
			}
			if dryRun {
				result, err := play.UpdateAppDetails(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateAppDetails(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&defaultLanguage, "default-language", "", "Default BCP-47 language, for example en-US")
	cmd.Flags().StringVar(&contactWebsite, "contact-website", "", "User-visible support website")
	cmd.Flags().StringVar(&contactEmail, "contact-email", "", "User-visible support email")
	cmd.Flags().StringVar(&contactPhone, "contact-phone", "", "User-visible support phone")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Commit the edit after validation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned details update without calling Google Play")
	return cmd
}
