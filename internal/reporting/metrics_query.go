package reporting

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

type AggregationPeriod string

const (
	AggregationPeriodHourly AggregationPeriod = "HOURLY"
	AggregationPeriodDaily  AggregationPeriod = "DAILY"
)

func NewAggregationPeriod(value string) (AggregationPeriod, error) {
	switch AggregationPeriod(value) {
	case AggregationPeriodHourly, AggregationPeriodDaily:
		return AggregationPeriod(value), nil
	case "":
		return "", fmt.Errorf("aggregation period is required")
	default:
		return "", fmt.Errorf("unsupported aggregation period %q; supported values: HOURLY, DAILY", value)
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

func (d QueryDate) Before(other QueryDate) bool {
	return d.dateTime().Before(other.dateTime())
}

func (d QueryDate) dateTime() time.Time {
	return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), 0, 0, 0, 0, time.UTC)
}

func validateQueryTimeZone(aggregationPeriod AggregationPeriod, timeZone string) error {
	if timeZone == "" {
		return nil
	}
	if _, err := time.LoadLocation(timeZone); err != nil {
		return fmt.Errorf("invalid time zone %q: %w", timeZone, err)
	}
	switch aggregationPeriod {
	case AggregationPeriodHourly:
		if timeZone != "UTC" {
			return fmt.Errorf("HOURLY aggregation only supports UTC time zone")
		}
	case AggregationPeriodDaily:
		if timeZone != "America/Los_Angeles" {
			return fmt.Errorf("DAILY aggregation only supports America/Los_Angeles time zone")
		}
	}
	return nil
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
	support, ok := metricQuerySupportByMetricSet[o.MetricSet]
	if !ok {
		return fmt.Errorf("unsupported vitals metric set %q", o.MetricSet)
	}
	if len(o.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	for _, metric := range o.Metrics {
		if metric == "" {
			return fmt.Errorf("metric cannot be empty")
		}
		if !support.Metrics.Contains(metric) {
			return fmt.Errorf("metric %q is not supported for vitals metric set %s; supported values: %s", metric, o.MetricSet, support.Metrics.Text())
		}
		if o.AggregationPeriod == AggregationPeriodHourly && support.UnsupportedHourlyMetrics.Contains(metric) {
			return fmt.Errorf("metric %q is not supported with HOURLY aggregation for vitals metric set %s", metric, o.MetricSet)
		}
	}
	aggregationPeriod, err := NewAggregationPeriod(o.AggregationPeriod.String())
	if err != nil {
		return err
	}
	if !support.AggregationPeriods.Contains(aggregationPeriod.String()) {
		return fmt.Errorf("aggregation period %s is not supported for vitals metric set %s; supported values: %s", aggregationPeriod, o.MetricSet, support.AggregationPeriods.Text())
	}
	userCohort, err := NewUserCohort(o.UserCohort.String())
	if err != nil {
		return err
	}
	if userCohort != "" && !support.UserCohorts.Contains(userCohort.String()) {
		if len(support.UserCohorts) == 0 {
			return fmt.Errorf("user cohort is not supported for vitals metric set %s", o.MetricSet)
		}
		return fmt.Errorf("user cohort %s is not supported for vitals metric set %s; supported values: %s", userCohort, o.MetricSet, support.UserCohorts.Text())
	}
	dimensions := support.Dimensions
	if userCohort == UserCohortOSBeta {
		dimensions = support.OSBetaDimensions
	}
	for _, dimension := range o.Dimensions {
		if dimension == "" {
			return fmt.Errorf("dimension cannot be empty")
		}
		if !dimensions.Contains(dimension) {
			return fmt.Errorf("dimension %q is not supported for vitals metric set %s; supported values: %s", dimension, o.MetricSet, dimensions.Text())
		}
	}
	if o.StartDate.Year == 0 || o.StartDate.Month == 0 || o.StartDate.Day == 0 {
		return fmt.Errorf("start date is required")
	}
	if o.EndDate.Year == 0 || o.EndDate.Month == 0 || o.EndDate.Day == 0 {
		return fmt.Errorf("end date is required")
	}
	if !o.StartDate.Before(o.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}
	if err := validateQueryTimeZone(o.AggregationPeriod, o.StartDate.TimeZone); err != nil {
		return err
	}
	if err := validateQueryTimeZone(o.AggregationPeriod, o.EndDate.TimeZone); err != nil {
		return err
	}
	if o.StartDate.TimeZone != o.EndDate.TimeZone {
		return fmt.Errorf("start and end time zones must match")
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
	Dimension   string  `json:"dimension"`
	StringValue string  `json:"stringValue,omitempty"`
	Int64Value  *string `json:"int64Value,omitempty"`
	ValueLabel  string  `json:"valueLabel,omitempty"`
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
			Int64Value:  int64ValueFromAPI(apiDimension.Dimension, apiDimension.Int64Value),
			ValueLabel:  apiDimension.ValueLabel,
		})
	}
	return dimensions
}

func int64ValueFromAPI(dimension string, value int64) *string {
	if !numericDimensions.Contains(dimension) {
		return nil
	}
	formatted := strconv.FormatInt(value, 10)
	return &formatted
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

type stringSet []string

func (s stringSet) Contains(value string) bool {
	for _, item := range s {
		if item == value {
			return true
		}
	}
	return false
}

func (s stringSet) Text() string {
	return strings.Join(s, ", ")
}

type metricQuerySupport struct {
	Metrics                  stringSet
	UnsupportedHourlyMetrics stringSet
	Dimensions               stringSet
	OSBetaDimensions         stringSet
	AggregationPeriods       stringSet
	UserCohorts              stringSet
}

var metricQuerySupportByMetricSet = map[MetricSet]metricQuerySupport{
	MetricSetANRRate: {
		Metrics: stringSet{
			"anrRate",
			"anrRate7dUserWeighted",
			"anrRate28dUserWeighted",
			"userPerceivedAnrRate",
			"userPerceivedAnrRate7dUserWeighted",
			"userPerceivedAnrRate28dUserWeighted",
			"distinctUsers",
		},
		UnsupportedHourlyMetrics: stringSet{
			"anrRate7dUserWeighted",
			"anrRate28dUserWeighted",
			"userPerceivedAnrRate7dUserWeighted",
			"userPerceivedAnrRate28dUserWeighted",
		},
		Dimensions:         deviceMetricDimensions,
		OSBetaDimensions:   betaMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String(), AggregationPeriodHourly.String()},
		UserCohorts:        allUserCohorts,
	},
	MetricSetCrashRate: {
		Metrics: stringSet{
			"crashRate",
			"crashRate7dUserWeighted",
			"crashRate28dUserWeighted",
			"userPerceivedCrashRate",
			"userPerceivedCrashRate7dUserWeighted",
			"userPerceivedCrashRate28dUserWeighted",
			"distinctUsers",
		},
		UnsupportedHourlyMetrics: stringSet{
			"crashRate28dUserWeighted",
			"userPerceivedCrashRate7dUserWeighted",
			"userPerceivedCrashRate28dUserWeighted",
		},
		Dimensions:         deviceMetricDimensions,
		OSBetaDimensions:   betaMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String(), AggregationPeriodHourly.String()},
		UserCohorts:        allUserCohorts,
	},
	MetricSetErrorCount: {
		Metrics: stringSet{
			"errorReportCount",
			"distinctUsers",
		},
		Dimensions: stringSet{
			"apiLevel",
			"versionCode",
			"deviceModel",
			"deviceType",
			"reportType",
			"issueId",
			"deviceRamBucket",
			"deviceSocMake",
			"deviceSocModel",
			"deviceCpuMake",
			"deviceCpuModel",
			"deviceGpuMake",
			"deviceGpuModel",
			"deviceGpuVersion",
			"deviceVulkanVersion",
			"deviceGlEsVersion",
			"deviceScreenSize",
			"deviceScreenDpi",
		},
		AggregationPeriods: stringSet{AggregationPeriodDaily.String()},
	},
	MetricSetExcessiveWakeupRate: {
		Metrics: stringSet{
			"excessiveWakeupRate",
			"excessiveWakeupRate7dUserWeighted",
			"excessiveWakeupRate28dUserWeighted",
			"distinctUsers",
		},
		Dimensions:         deviceMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String()},
		UserCohorts:        osPublicOnlyUserCohort,
	},
	MetricSetLMKRate: {
		Metrics: stringSet{
			"userPerceivedLmkRate",
			"userPerceivedLmkRate7dUserWeighted",
			"userPerceivedLmkRate28dUserWeighted",
			"distinctUsers",
		},
		Dimensions:         deviceMetricDimensions,
		OSBetaDimensions:   betaMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String()},
		UserCohorts:        allUserCohorts,
	},
	MetricSetSlowRenderingRate: {
		Metrics: stringSet{
			"slowRenderingRate20Fps",
			"slowRenderingRate20Fps7dUserWeighted",
			"slowRenderingRate20Fps28dUserWeighted",
			"slowRenderingRate30Fps",
			"slowRenderingRate30Fps7dUserWeighted",
			"slowRenderingRate30Fps28dUserWeighted",
			"distinctUsers",
		},
		Dimensions:         deviceMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String()},
		UserCohorts:        osPublicOnlyUserCohort,
	},
	MetricSetSlowStartRate: {
		Metrics: stringSet{
			"slowStartRate",
			"slowStartRate7dUserWeighted",
			"slowStartRate28dUserWeighted",
			"distinctUsers",
		},
		Dimensions:         deviceMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String()},
		UserCohorts:        osPublicOnlyUserCohort,
	},
	MetricSetStuckBackgroundWakeLockRate: {
		Metrics: stringSet{
			"stuckBgWakelockRate",
			"stuckBgWakelockRate7dUserWeighted",
			"stuckBgWakelockRate28dUserWeighted",
			"distinctUsers",
		},
		Dimensions:         deviceMetricDimensions,
		AggregationPeriods: stringSet{AggregationPeriodDaily.String()},
		UserCohorts:        osPublicOnlyUserCohort,
	},
}

var deviceMetricDimensions = stringSet{
	"apiLevel",
	"versionCode",
	"deviceModel",
	"deviceBrand",
	"deviceType",
	"countryCode",
	"deviceRamBucket",
	"deviceSocMake",
	"deviceSocModel",
	"deviceCpuMake",
	"deviceCpuModel",
	"deviceGpuMake",
	"deviceGpuModel",
	"deviceGpuVersion",
	"deviceVulkanVersion",
	"deviceGlEsVersion",
	"deviceScreenSize",
	"deviceScreenDpi",
}

var betaMetricDimensions = stringSet{
	"versionCode",
	"osBuild",
}

var numericDimensions = stringSet{
	"apiLevel",
	"versionCode",
	"deviceRamBucket",
}

var allUserCohorts = stringSet{
	UserCohortOSPublic.String(),
	UserCohortOSBeta.String(),
	UserCohortAppTesters.String(),
}

var osPublicOnlyUserCohort = stringSet{
	UserCohortOSPublic.String(),
}
