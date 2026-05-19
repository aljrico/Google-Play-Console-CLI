package reporting

import (
	"context"
	"fmt"

	"github.com/aljrico/Google-Play-Console-CLI/internal/googleclient"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"google.golang.org/api/option"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

type MetricSet string

const (
	MetricSetANRRate                     MetricSet = "anr-rate"
	MetricSetCrashRate                   MetricSet = "crash-rate"
	MetricSetErrorCount                  MetricSet = "error-count"
	MetricSetExcessiveWakeupRate         MetricSet = "excessive-wakeup-rate"
	MetricSetLMKRate                     MetricSet = "lmk-rate"
	MetricSetSlowRenderingRate           MetricSet = "slow-rendering-rate"
	MetricSetSlowStartRate               MetricSet = "slow-start-rate"
	MetricSetStuckBackgroundWakeLockRate MetricSet = "stuck-background-wakelock-rate"
)

func NewMetricSet(value string) (MetricSet, error) {
	switch MetricSet(value) {
	case MetricSetANRRate,
		MetricSetCrashRate,
		MetricSetErrorCount,
		MetricSetExcessiveWakeupRate,
		MetricSetLMKRate,
		MetricSetSlowRenderingRate,
		MetricSetSlowStartRate,
		MetricSetStuckBackgroundWakeLockRate:
		return MetricSet(value), nil
	default:
		return "", fmt.Errorf("unsupported vitals metric set %q", value)
	}
}

func (m MetricSet) String() string {
	return string(m)
}

func (m MetricSet) ResourceName(packageName play.PackageName) string {
	return fmt.Sprintf("apps/%s/%s", packageName, m.resourceSuffix())
}

func (m MetricSet) resourceSuffix() string {
	switch m {
	case MetricSetANRRate:
		return "anrRateMetricSet"
	case MetricSetCrashRate:
		return "crashRateMetricSet"
	case MetricSetErrorCount:
		return "errorCountMetricSet"
	case MetricSetExcessiveWakeupRate:
		return "excessiveWakeupRateMetricSet"
	case MetricSetLMKRate:
		return "lmkRateMetricSet"
	case MetricSetSlowRenderingRate:
		return "slowRenderingRateMetricSet"
	case MetricSetSlowStartRate:
		return "slowStartRateMetricSet"
	case MetricSetStuckBackgroundWakeLockRate:
		return "stuckBackgroundWakelockRateMetricSet"
	default:
		return ""
	}
}

type Client struct {
	service *playdeveloperreporting.Service
}

func NewClientFromActiveProfile(ctx context.Context) (*Client, error) {
	httpClient, err := googleclient.ActiveProfileHTTPClient(ctx, playdeveloperreporting.PlaydeveloperreportingScope)
	if err != nil {
		return nil, err
	}
	service, err := playdeveloperreporting.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create Play Developer Reporting API service: %w", err)
	}
	return &Client{service: service}, nil
}

type MetricSetGetOptions struct {
	PackageName play.PackageName `json:"packageName"`
	MetricSet   MetricSet        `json:"metricSet"`
}

func (o MetricSetGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewMetricSet(o.MetricSet.String()); err != nil {
		return err
	}
	return nil
}

type MetricSetFreshness struct {
	AggregationPeriod string    `json:"aggregationPeriod,omitempty"`
	LatestEndTime     *DateTime `json:"latestEndTime,omitempty"`
}

type DateTime struct {
	Year      int64  `json:"year,omitempty"`
	Month     int64  `json:"month,omitempty"`
	Day       int64  `json:"day,omitempty"`
	Hours     int64  `json:"hours,omitempty"`
	Minutes   int64  `json:"minutes,omitempty"`
	Seconds   int64  `json:"seconds,omitempty"`
	Nanos     int64  `json:"nanos,omitempty"`
	UtcOffset string `json:"utcOffset,omitempty"`
	TimeZone  string `json:"timeZone,omitempty"`
}

type MetricSetMetadata struct {
	PackageName play.PackageName     `json:"packageName"`
	MetricSet   MetricSet            `json:"metricSet"`
	Name        string               `json:"name"`
	Freshnesses []MetricSetFreshness `json:"freshnesses"`
	Options     MetricSetGetOptions  `json:"options"`
}

type MetricSetGetter interface {
	GetMetricSet(ctx context.Context, options MetricSetGetOptions) (MetricSetMetadata, error)
}

func GetMetricSet(ctx context.Context, getter MetricSetGetter, options MetricSetGetOptions) (MetricSetMetadata, error) {
	if err := options.Validate(); err != nil {
		return MetricSetMetadata{}, err
	}
	if getter == nil {
		return MetricSetMetadata{}, fmt.Errorf("vitals metric set getter is required")
	}
	return getter.GetMetricSet(ctx, options)
}

func (c Client) GetMetricSet(ctx context.Context, options MetricSetGetOptions) (MetricSetMetadata, error) {
	name := options.MetricSet.ResourceName(options.PackageName)
	var (
		apiName      string
		apiFreshness *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1FreshnessInfo
		err          error
	)
	switch options.MetricSet {
	case MetricSetANRRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1AnrRateMetricSet
		response, err = c.service.Vitals.Anrrate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetCrashRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1CrashRateMetricSet
		response, err = c.service.Vitals.Crashrate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetErrorCount:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorCountMetricSet
		response, err = c.service.Vitals.Errors.Counts.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetExcessiveWakeupRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ExcessiveWakeupRateMetricSet
		response, err = c.service.Vitals.Excessivewakeuprate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetLMKRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1LmkRateMetricSet
		response, err = c.service.Vitals.Lmkrate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetSlowRenderingRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SlowRenderingRateMetricSet
		response, err = c.service.Vitals.Slowrenderingrate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetSlowStartRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SlowStartRateMetricSet
		response, err = c.service.Vitals.Slowstartrate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	case MetricSetStuckBackgroundWakeLockRate:
		var response *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1StuckBackgroundWakelockRateMetricSet
		response, err = c.service.Vitals.Stuckbackgroundwakelockrate.Get(name).Context(ctx).Do()
		if response != nil {
			apiName = response.Name
			apiFreshness = response.FreshnessInfo
		}
	default:
		return MetricSetMetadata{}, fmt.Errorf("unsupported vitals metric set %q", options.MetricSet)
	}
	if err != nil {
		return MetricSetMetadata{}, fmt.Errorf("get vitals metric set %s for %s: %w", options.MetricSet, options.PackageName, err)
	}
	return MetricSetMetadata{
		PackageName: options.PackageName,
		MetricSet:   options.MetricSet,
		Name:        apiName,
		Freshnesses: freshnessesFromAPI(apiFreshness),
		Options:     options,
	}, nil
}

func freshnessesFromAPI(apiFreshness *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1FreshnessInfo) []MetricSetFreshness {
	if apiFreshness == nil {
		return []MetricSetFreshness{}
	}
	freshnesses := make([]MetricSetFreshness, 0, len(apiFreshness.Freshnesses))
	for _, apiFreshness := range apiFreshness.Freshnesses {
		if apiFreshness == nil {
			continue
		}
		freshnesses = append(freshnesses, MetricSetFreshness{
			AggregationPeriod: apiFreshness.AggregationPeriod,
			LatestEndTime:     dateTimeFromAPI(apiFreshness.LatestEndTime),
		})
	}
	return freshnesses
}

func dateTimeFromAPI(apiDateTime *playdeveloperreporting.GoogleTypeDateTime) *DateTime {
	if apiDateTime == nil {
		return nil
	}
	dateTime := &DateTime{
		Year:      apiDateTime.Year,
		Month:     apiDateTime.Month,
		Day:       apiDateTime.Day,
		Hours:     apiDateTime.Hours,
		Minutes:   apiDateTime.Minutes,
		Seconds:   apiDateTime.Seconds,
		Nanos:     apiDateTime.Nanos,
		UtcOffset: apiDateTime.UtcOffset,
	}
	if apiDateTime.TimeZone != nil {
		dateTime.TimeZone = apiDateTime.TimeZone.Id
	}
	return dateTime
}
