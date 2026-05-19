package reporting

import (
	"context"
	"fmt"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

type AnomalyListOptions struct {
	PackageName play.PackageName `json:"packageName"`
	Filter      string           `json:"filter,omitempty"`
	PageSize    int64            `json:"pageSize,omitempty"`
	PageToken   string           `json:"pageToken,omitempty"`
}

func (o AnomalyListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	return nil
}

type AnomalyListResult struct {
	PackageName   play.PackageName   `json:"packageName"`
	Anomalies     []Anomaly          `json:"anomalies"`
	NextPageToken string             `json:"nextPageToken,omitempty"`
	Options       AnomalyListOptions `json:"options"`
}

type Anomaly struct {
	Name         string           `json:"name"`
	MetricSet    string           `json:"metricSet,omitempty"`
	Metric       *MetricValue     `json:"metric,omitempty"`
	Dimensions   []DimensionValue `json:"dimensions"`
	TimelineSpec *TimelineSpec    `json:"timelineSpec,omitempty"`
}

type TimelineSpec struct {
	AggregationPeriod string    `json:"aggregationPeriod,omitempty"`
	StartTime         *DateTime `json:"startTime,omitempty"`
	EndTime           *DateTime `json:"endTime,omitempty"`
}

type AnomalyLister interface {
	ListAnomalies(ctx context.Context, options AnomalyListOptions) (AnomalyListResult, error)
}

func ListAnomalies(ctx context.Context, lister AnomalyLister, options AnomalyListOptions) (AnomalyListResult, error) {
	if err := options.Validate(); err != nil {
		return AnomalyListResult{}, err
	}
	if lister == nil {
		return AnomalyListResult{}, fmt.Errorf("anomaly lister is required")
	}
	return lister.ListAnomalies(ctx, options)
}

func (c Client) ListAnomalies(ctx context.Context, options AnomalyListOptions) (AnomalyListResult, error) {
	parent := fmt.Sprintf("apps/%s", options.PackageName)
	call := c.service.Anomalies.List(parent).Context(ctx)
	if options.Filter != "" {
		call.Filter(options.Filter)
	}
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return AnomalyListResult{}, fmt.Errorf("list anomalies for %s: %w", options.PackageName, err)
	}
	return AnomalyListResult{
		PackageName:   options.PackageName,
		Anomalies:     anomaliesFromAPI(response.Anomalies),
		NextPageToken: response.NextPageToken,
		Options:       options,
	}, nil
}

func anomaliesFromAPI(apiAnomalies []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly) []Anomaly {
	anomalies := make([]Anomaly, 0, len(apiAnomalies))
	for _, apiAnomaly := range apiAnomalies {
		if apiAnomaly == nil {
			continue
		}
		anomalies = append(anomalies, Anomaly{
			Name:         apiAnomaly.Name,
			MetricSet:    apiAnomaly.MetricSet,
			Metric:       metricValueFromAPI(apiAnomaly.Metric),
			Dimensions:   dimensionValuesFromAPI(apiAnomaly.Dimensions),
			TimelineSpec: timelineSpecFromAPI(apiAnomaly.TimelineSpec),
		})
	}
	return anomalies
}

func metricValueFromAPI(apiMetric *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue) *MetricValue {
	if apiMetric == nil {
		return nil
	}
	return &MetricValue{
		Metric:                         apiMetric.Metric,
		DecimalValue:                   decimalFromAPI(apiMetric.DecimalValue),
		DecimalValueConfidenceInterval: decimalConfidenceIntervalFromAPI(apiMetric.DecimalValueConfidenceInterval),
	}
}

func timelineSpecFromAPI(apiTimelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) *TimelineSpec {
	if apiTimelineSpec == nil {
		return nil
	}
	return &TimelineSpec{
		AggregationPeriod: apiTimelineSpec.AggregationPeriod,
		StartTime:         dateTimeFromAPI(apiTimelineSpec.StartTime),
		EndTime:           dateTimeFromAPI(apiTimelineSpec.EndTime),
	}
}
