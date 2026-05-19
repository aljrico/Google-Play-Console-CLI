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
	cmd.AddCommand(
		newVitalsMetricSetCommand(out, options, &packageName),
		newVitalsErrorsCommand(out, options, &packageName),
	)
	return cmd
}

func newVitalsMetricSetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metric-set",
		Short: "Inspect Android vitals metric set metadata",
	}
	cmd.AddCommand(
		newVitalsMetricSetGetCommand(out, options, packageName),
		newVitalsMetricSetQueryCommand(out, options, packageName),
	)
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

func newVitalsMetricSetQueryCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		metricSet         string
		metrics           []string
		dimensions        []string
		filter            string
		aggregationPeriod string
		startDate         string
		endDate           string
		timeZone          string
		userCohort        string
		pageSize          int64
		pageToken         string
	)

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query Android vitals metric rows",
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
			typedAggregationPeriod, err := reporting.NewAggregationPeriod(aggregationPeriod)
			if err != nil {
				return err
			}
			typedStartDate, err := reporting.NewQueryDate(startDate, timeZone)
			if err != nil {
				return err
			}
			typedEndDate, err := reporting.NewQueryDate(endDate, timeZone)
			if err != nil {
				return err
			}
			typedUserCohort, err := reporting.NewUserCohort(userCohort)
			if err != nil {
				return err
			}
			queryOptions := reporting.MetricQueryOptions{
				PackageName:       typedPackageName,
				MetricSet:         typedMetricSet,
				Metrics:           metrics,
				Dimensions:        dimensions,
				Filter:            filter,
				AggregationPeriod: typedAggregationPeriod,
				StartDate:         typedStartDate,
				EndDate:           typedEndDate,
				UserCohort:        typedUserCohort,
				PageSize:          pageSize,
				PageToken:         pageToken,
			}
			if err := queryOptions.Validate(); err != nil {
				return err
			}
			client, err := reporting.NewClientFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := reporting.QueryMetrics(cmd.Context(), client, queryOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&metricSet, "metric-set", "", "Vitals metric set: "+reporting.SupportedMetricSetValuesText())
	cmd.Flags().StringArrayVar(&metrics, "metric", nil, "Metric to request; repeat for multiple metrics")
	cmd.Flags().StringArrayVar(&dimensions, "dimension", nil, "Dimension to break down by; repeat for multiple dimensions")
	cmd.Flags().StringVar(&filter, "filter", "", "AIP-160 filter expression over supported dimensions")
	cmd.Flags().StringVar(&aggregationPeriod, "aggregation", "", "Aggregation period: DAILY, or HOURLY where the metric set supports it")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date, inclusive, in YYYY-MM-DD format")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date, exclusive, in YYYY-MM-DD format")
	cmd.Flags().StringVar(&timeZone, "time-zone", "", "IANA time zone for daily aggregation, for example America/Los_Angeles")
	cmd.Flags().StringVar(&userCohort, "user-cohort", "", "User cohort where supported: OS_PUBLIC, OS_BETA, or APP_TESTERS")
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum rows to return, capped by Google at 100000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}

func newVitalsErrorsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "errors",
		Short: "Search Android vitals error issues",
	}
	cmd.AddCommand(newVitalsErrorsIssuesCommand(out, options, packageName))
	return cmd
}

func newVitalsErrorsIssuesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issues",
		Short: "Search grouped Android vitals error issues",
	}
	cmd.AddCommand(newVitalsErrorsIssuesSearchCommand(out, options, packageName))
	return cmd
}

func newVitalsErrorsIssuesSearchCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		filter                 string
		orderBy                string
		startDate              string
		endDate                string
		timeZone               string
		pageSize               int64
		pageToken              string
		sampleErrorReportLimit int64
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search grouped Android vitals error issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedStartDate, err := reporting.NewQueryDate(startDate, timeZone)
			if err != nil {
				return err
			}
			typedEndDate, err := reporting.NewQueryDate(endDate, timeZone)
			if err != nil {
				return err
			}
			searchOptions := reporting.ErrorIssueSearchOptions{
				PackageName:            typedPackageName,
				Filter:                 filter,
				OrderBy:                orderBy,
				StartDate:              typedStartDate,
				EndDate:                typedEndDate,
				PageSize:               pageSize,
				PageToken:              pageToken,
				SampleErrorReportLimit: sampleErrorReportLimit,
			}
			if err := searchOptions.Validate(); err != nil {
				return err
			}
			client, err := reporting.NewClientFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := reporting.SearchErrorIssues(cmd.Context(), client, searchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "AIP-160 filter expression for issue fields")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Order issues by errorReportCount|distinctUsers asc|desc")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date, inclusive, in YYYY-MM-DD format")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date, exclusive, in YYYY-MM-DD format")
	cmd.Flags().StringVar(&timeZone, "time-zone", "", "IANA time zone for the interval, for example America/Los_Angeles")
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum issues to return, capped by Google at 1000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	cmd.Flags().Int64Var(&sampleErrorReportLimit, "sample-error-report-limit", 0, "Sample reports per issue; Google currently supports 0 or 1")
	return cmd
}
