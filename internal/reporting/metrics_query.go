package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

type AggregationPeriod string

const (
	AggregationPeriodHourly    AggregationPeriod = "HOURLY"
	AggregationPeriodDaily     AggregationPeriod = "DAILY"
	AggregationPeriodFullRange AggregationPeriod = "FULL_RANGE"
)

func NewAggregationPeriod(value string) (AggregationPeriod, error) {
	switch AggregationPeriod(value) {
	case AggregationPeriodHourly, AggregationPeriodDaily, AggregationPeriodFullRange:
		return AggregationPeriod(value), nil
	case "":
		return "", fmt.Errorf("aggregation period is required")
	default:
		return "", fmt.Errorf("unsupported aggregation period %q; supported values: HOURLY, DAILY, FULL_RANGE", value)
	}
}

func (a AggregationPeriod) String() string {
	return string(a)
}

type UserCohort string

const (
	UserCohortOSPublic   UserCohort = "OS_PUBLIC"
	UserCohortOSBeta     UserCohort = "OS_BETA"
	UserCohortAppTesters UserCohort = "APP_TESTERS"
)

func NewUserCohort(value string) (UserCohort, error) {
	switch UserCohort(value) {
	case UserCohortOSPublic, UserCohortOSBeta, UserCohortAppTesters:
		return UserCohort(value), nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("unsupported user cohort %q; supported values: OS_PUBLIC, OS_BETA, APP_TESTERS", value)
	}
}

func (u UserCohort) String() string {
	return string(u)
}

type QueryDate struct {
	Year     int64  `json:"year"`
	Month    int64  `json:"month"`
	Day      int64  `json:"day"`
	TimeZone string `json:"timeZone,omitempty"`
}

func NewQueryDate(value string, timeZone string) (QueryDate, error) {
	if value == "" {
		return QueryDate{}, fmt.Errorf("date is required")
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return QueryDate{}, fmt.Errorf("date %q must use YYYY-MM-DD", value)
	}
	return QueryDate{
		Year:     int64(parsed.Year()),
		Month:    int64(parsed.Month()),
		Day:      int64(parsed.Day()),
		TimeZone: timeZone,
	}, nil
}

func (d QueryDate) toAPI() *playdeveloperreporting.GoogleTypeDateTime {
	apiDateTime := &playdeveloperreporting.GoogleTypeDateTime{
		Year:  d.Year,
		Month: d.Month,
		Day:   d.Day,
	}
	if d.TimeZone != "" {
		apiDateTime.TimeZone = &playdeveloperreporting.GoogleTypeTimeZone{Id: d.TimeZone}
	}
	return apiDateTime
}

type MetricQueryOptions struct {
	PackageName       play.PackageName  `json:"packageName"`
	MetricSet         MetricSet         `json:"metricSet"`
	Metrics           []string          `json:"metrics"`
	Dimensions        []string          `json:"dimensions,omitempty"`
	Filter            string            `json:"filter,omitempty"`
	AggregationPeriod AggregationPeriod `json:"aggregationPeriod"`
	StartDate         QueryDate         `json:"startDate"`
	EndDate           QueryDate         `json:"endDate"`
	UserCohort        UserCohort        `json:"userCohort,omitempty"`
	PageSize          int64             `json:"pageSize,omitempty"`
	PageToken         string            `json:"pageToken,omitempty"`
}

func (o MetricQueryOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewMetricSet(o.MetricSet.String()); err != nil {
		return err
	}
	if len(o.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	for _, metric := range o.Metrics {
		if metric == "" {
			return fmt.Errorf("metric cannot be empty")
		}
	}
	for _, dimension := range o.Dimensions {
		if dimension == "" {
			return fmt.Errorf("dimension cannot be empty")
		}
	}
	if _, err := NewAggregationPeriod(o.AggregationPeriod.String()); err != nil {
		return err
	}
	if _, err := NewUserCohort(o.UserCohort.String()); err != nil {
		return err
	}
	if o.MetricSet == MetricSetErrorCount && o.UserCohort != "" {
		return fmt.Errorf("user cohort is not supported for vitals metric set %s", o.MetricSet)
	}
	if o.StartDate.Year == 0 || o.StartDate.Month == 0 || o.StartDate.Day == 0 {
		return fmt.Errorf("start date is required")
	}
	if o.EndDate.Year == 0 || o.EndDate.Month == 0 || o.EndDate.Day == 0 {
		return fmt.Errorf("end date is required")
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	return nil
}

type MetricQueryResult struct {
	PackageName   play.PackageName   `json:"packageName"`
	MetricSet     MetricSet          `json:"metricSet"`
	Name          string             `json:"name"`
	Rows          []MetricRow        `json:"rows"`
	NextPageToken string             `json:"nextPageToken,omitempty"`
	Options       MetricQueryOptions `json:"options"`
}

type MetricRow struct {
	AggregationPeriod string           `json:"aggregationPeriod,omitempty"`
	StartTime         *DateTime        `json:"startTime,omitempty"`
	Dimensions        []DimensionValue `json:"dimensions"`
	Metrics           []MetricValue    `json:"metrics"`
}

type DimensionValue struct {
	Dimension   string `json:"dimension"`
	StringValue string `json:"stringValue,omitempty"`
	Int64Value  int64  `json:"int64Value,omitempty"`
	ValueLabel  string `json:"valueLabel,omitempty"`
}

type MetricValue struct {
	Metric                         string                     `json:"metric"`
	DecimalValue                   string                     `json:"decimalValue,omitempty"`
	DecimalValueConfidenceInterval *DecimalConfidenceInterval `json:"decimalValueConfidenceInterval,omitempty"`
}

type DecimalConfidenceInterval struct {
	LowerBound string `json:"lowerBound,omitempty"`
	UpperBound string `json:"upperBound,omitempty"`
}

type MetricQuerier interface {
	QueryMetrics(ctx context.Context, options MetricQueryOptions) (MetricQueryResult, error)
}

func QueryMetrics(ctx context.Context, querier MetricQuerier, options MetricQueryOptions) (MetricQueryResult, error) {
	if err := options.Validate(); err != nil {
		return MetricQueryResult{}, err
	}
	if querier == nil {
		return MetricQueryResult{}, fmt.Errorf("vitals metric querier is required")
	}
	return querier.QueryMetrics(ctx, options)
}

func (c Client) QueryMetrics(ctx context.Context, options MetricQueryOptions) (MetricQueryResult, error) {
	name := options.MetricSet.ResourceName(options.PackageName)
	var (
		rows          []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
		nextPageToken string
		err           error
	)
	switch options.MetricSet {
	case MetricSetANRRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetResponse
		response, err = c.service.Vitals.Anrrate.Query(name, queryAnrRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetCrashRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetResponse
		response, err = c.service.Vitals.Crashrate.Query(name, queryCrashRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetErrorCount:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetResponse
		response, err = c.service.Vitals.Errors.Counts.Query(name, queryErrorCountMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetExcessiveWakeupRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetResponse
		response, err = c.service.Vitals.Excessivewakeuprate.Query(name, queryExcessiveWakeupRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetLMKRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryLmkRateMetricSetResponse
		response, err = c.service.Vitals.Lmkrate.Query(name, queryLmkRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetSlowRenderingRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetResponse
		response, err = c.service.Vitals.Slowrenderingrate.Query(name, querySlowRenderingRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetSlowStartRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetResponse
		response, err = c.service.Vitals.Slowstartrate.Query(name, querySlowStartRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	case MetricSetStuckBackgroundWakeLockRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetResponse
		response, err = c.service.Vitals.Stuckbackgroundwakelockrate.Query(name, queryStuckBackgroundWakelockRateMetricSetRequest(options)).Context(ctx).Do()
		if response != nil {
			rows = response.Rows
			nextPageToken = response.NextPageToken
		}
	default:
		return MetricQueryResult{}, fmt.Errorf("unsupported vitals metric set %q", options.MetricSet)
	}
	if err != nil {
		return MetricQueryResult{}, fmt.Errorf("query vitals metric set %s for %s: %w", options.MetricSet, options.PackageName, err)
	}
	return MetricQueryResult{
		PackageName:   options.PackageName,
		MetricSet:     options.MetricSet,
		Name:          name,
		Rows:          metricRowsFromAPI(rows),
		NextPageToken: nextPageToken,
		Options:       options,
	}, nil
}

func queryAnrRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func queryCrashRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func queryErrorCountMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
	}
}

func queryExcessiveWakeupRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func queryLmkRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryLmkRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryLmkRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func querySlowRenderingRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func querySlowStartRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func queryStuckBackgroundWakelockRateMetricSetRequest(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
		Dimensions:   options.Dimensions,
		Filter:       options.Filter,
		Metrics:      options.Metrics,
		PageSize:     options.PageSize,
		PageToken:    options.PageToken,
		TimelineSpec: timelineSpecFromOptions(options),
		UserCohort:   options.UserCohort.String(),
	}
}

func timelineSpecFromOptions(options MetricQueryOptions) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: options.AggregationPeriod.String(),
		StartTime:         options.StartDate.toAPI(),
		EndTime:           options.EndDate.toAPI(),
	}
}

func metricRowsFromAPI(apiRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow) []MetricRow {
	rows := make([]MetricRow, 0, len(apiRows))
	for _, apiRow := range apiRows {
		if apiRow == nil {
			continue
		}
		rows = append(rows, MetricRow{
			AggregationPeriod: apiRow.AggregationPeriod,
			StartTime:         dateTimeFromAPI(apiRow.StartTime),
			Dimensions:        dimensionValuesFromAPI(apiRow.Dimensions),
			Metrics:           metricValuesFromAPI(apiRow.Metrics),
		})
	}
	return rows
}

func dimensionValuesFromAPI(apiDimensions []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DimensionValue) []DimensionValue {
	dimensions := make([]DimensionValue, 0, len(apiDimensions))
	for _, apiDimension := range apiDimensions {
		if apiDimension == nil {
			continue
		}
		dimensions = append(dimensions, DimensionValue{
			Dimension:   apiDimension.Dimension,
			StringValue: apiDimension.StringValue,
			Int64Value:  apiDimension.Int64Value,
			ValueLabel:  apiDimension.ValueLabel,
		})
	}
	return dimensions
}

func metricValuesFromAPI(apiMetrics []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue) []MetricValue {
	metrics := make([]MetricValue, 0, len(apiMetrics))
	for _, apiMetric := range apiMetrics {
		if apiMetric == nil {
			continue
		}
		metrics = append(metrics, MetricValue{
			Metric:                         apiMetric.Metric,
			DecimalValue:                   decimalFromAPI(apiMetric.DecimalValue),
			DecimalValueConfidenceInterval: decimalConfidenceIntervalFromAPI(apiMetric.DecimalValueConfidenceInterval),
		})
	}
	return metrics
}

func decimalFromAPI(apiDecimal *playdeveloperreporting.GoogleTypeDecimal) string {
	if apiDecimal == nil {
		return ""
	}
	return apiDecimal.Value
}

func decimalConfidenceIntervalFromAPI(apiInterval *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DecimalConfidenceInterval) *DecimalConfidenceInterval {
	if apiInterval == nil {
		return nil
	}
	return &DecimalConfidenceInterval{
		LowerBound: decimalFromAPI(apiInterval.LowerBound),
		UpperBound: decimalFromAPI(apiInterval.UpperBound),
	}
}
