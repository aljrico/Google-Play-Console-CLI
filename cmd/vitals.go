package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/aljrico/Google-Play-Console-CLI/internal/reporting"
	"github.com/spf13/cobra"
)

func newVitalsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "vitals",
		Short: "Inspect Google Play Developer Reporting vitals",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newVitalsMetricSetCommand(out, options, &packageName))
	return cmd
}

func newVitalsMetricSetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metric-set",
		Short: "Inspect Android vitals metric set metadata",
	}
	cmd.AddCommand(newVitalsMetricSetGetCommand(out, options, packageName))
	return cmd
}

func newVitalsMetricSetGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var metricSet string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get Android vitals metric set freshness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedMetricSet, err := reporting.NewMetricSet(metricSet)
			if err != nil {
				return err
			}
			getOptions := reporting.MetricSetGetOptions{
				PackageName: typedPackageName,
				MetricSet:   typedMetricSet,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			client, err := reporting.NewClientFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := reporting.GetMetricSet(cmd.Context(), client, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&metricSet, "metric-set", "", "Vitals metric set: "+reporting.SupportedMetricSetValuesText())
	return cmd
}
