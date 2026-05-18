package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newDeviceTierConfigsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "device-tier-configs",
		Short: "Inspect Google Play device tier configs",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newDeviceTierConfigsListCommand(out, options, &packageName),
		newDeviceTierConfigsGetCommand(out, options, &packageName),
	)
	return cmd
}

func newDeviceTierConfigsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		pageSize  int64
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List device tier configs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.DeviceTierConfigListOptions{
				PackageName: typedPackageName,
				PageSize:    pageSize,
				PageToken:   pageToken,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListDeviceTierConfigs(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum configs to return, 0 uses the Google default")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}

func newDeviceTierConfigsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var configID int64

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one device tier config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			getOptions := play.DeviceTierConfigGetOptions{
				PackageName:        typedPackageName,
				DeviceTierConfigID: configID,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.GetDeviceTierConfig(cmd.Context(), publisher, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&configID, "id", 0, "Device tier config ID")
	return cmd
}
