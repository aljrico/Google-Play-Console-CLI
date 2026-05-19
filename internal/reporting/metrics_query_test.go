package reporting

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

func TestNewQueryDateParsesDateOnlyInput(t *testing.T) {
	date, err := NewQueryDate("2026-05-19", "America/Los_Angeles")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}
	if date.Year != 2026 || date.Month != 5 || date.Day != 19 || date.TimeZone != "America/Los_Angeles" {
		t.Fatalf("date = %#v, want parsed date with timezone", date)
	}

	if _, err := NewQueryDate("2026/05/19", ""); err == nil {
		t.Fatal("expected invalid date format error")
	}
}

func TestMetricQueryOptionsValidate(t *testing.T) {
	packageName, err := play.NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	startDate, err := NewQueryDate("2026-05-01", "")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}
	endDate, err := NewQueryDate("2026-05-19", "")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}

	options := MetricQueryOptions{
		PackageName:       packageName,
		MetricSet:         MetricSetCrashRate,
		Metrics:           []string{"crashRate"},
		AggregationPeriod: AggregationPeriodDaily,
		StartDate:         startDate,
		EndDate:           endDate,
	}
	if err := options.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	options.Metrics = nil
	if err := options.Validate(); err == nil {
		t.Fatal("expected missing metric validation error")
	}

	options.Metrics = []string{"errorReportCount"}
	options.MetricSet = MetricSetErrorCount
	options.UserCohort = UserCohortOSPublic
	if err := options.Validate(); err == nil {
		t.Fatal("expected unsupported user cohort validation error")
	}
}

func TestMetricQueryOptionsValidateAPISupportMatrix(t *testing.T) {
	packageName, err := play.NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	startDate, err := NewQueryDate("2026-05-01", "America/Los_Angeles")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}
	endDate, err := NewQueryDate("2026-05-19", "America/Los_Angeles")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}

	tests := []struct {
		name    string
		options MetricQueryOptions
		want    string
	}{
		{
			name: "unsupported aggregation",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetErrorCount,
				Metrics:           []string{"errorReportCount"},
				AggregationPeriod: AggregationPeriodHourly,
				StartDate:         startDate,
				EndDate:           endDate,
			},
			want: "aggregation period HOURLY is not supported",
		},
		{
			name: "unsupported cohort",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetSlowStartRate,
				Metrics:           []string{"slowStartRate"},
				AggregationPeriod: AggregationPeriodDaily,
				StartDate:         startDate,
				EndDate:           endDate,
				UserCohort:        UserCohortAppTesters,
			},
			want: "user cohort APP_TESTERS is not supported",
		},
		{
			name: "unsupported time zone",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetCrashRate,
				Metrics:           []string{"crashRate"},
				AggregationPeriod: AggregationPeriodDaily,
				StartDate:         QueryDate{Year: 2026, Month: 5, Day: 1, TimeZone: "UTC"},
				EndDate:           QueryDate{Year: 2026, Month: 5, Day: 19, TimeZone: "UTC"},
			},
			want: "DAILY aggregation only supports America/Los_Angeles",
		},
		{
			name: "reversed date range",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetCrashRate,
				Metrics:           []string{"crashRate"},
				AggregationPeriod: AggregationPeriodDaily,
				StartDate:         endDate,
				EndDate:           startDate,
			},
			want: "start date must be before end date",
		},
		{
			name: "unsupported metric",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetLMKRate,
				Metrics:           []string{"crashRate"},
				AggregationPeriod: AggregationPeriodDaily,
				StartDate:         startDate,
				EndDate:           endDate,
			},
			want: `metric "crashRate" is not supported`,
		},
		{
			name: "unsupported dimension",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetErrorCount,
				Metrics:           []string{"errorReportCount"},
				Dimensions:        []string{"countryCode"},
				AggregationPeriod: AggregationPeriodDaily,
				StartDate:         startDate,
				EndDate:           endDate,
			},
			want: `dimension "countryCode" is not supported`,
		},
		{
			name: "unsupported hourly rolling metric",
			options: MetricQueryOptions{
				PackageName:       packageName,
				MetricSet:         MetricSetCrashRate,
				Metrics:           []string{"crashRate28dUserWeighted"},
				AggregationPeriod: AggregationPeriodHourly,
				StartDate:         QueryDate{Year: 2026, Month: 5, Day: 1, TimeZone: "UTC"},
				EndDate:           QueryDate{Year: 2026, Month: 5, Day: 19, TimeZone: "UTC"},
			},
			want: `metric "crashRate28dUserWeighted" is not supported with HOURLY aggregation`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestDimensionValuesPreserveNumericZeroWithoutPollutingStringDimensions(t *testing.T) {
	rows := metricRowsFromAPI([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
		{
			Dimensions: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DimensionValue{
				{Dimension: "versionCode", Int64Value: 0},
				{Dimension: "countryCode", StringValue: "US"},
			},
		},
	})

	if len(rows) != 1 || len(rows[0].Dimensions) != 2 {
		t.Fatalf("rows = %#v, want two dimensions", rows)
	}
	if rows[0].Dimensions[0].Int64Value == nil || *rows[0].Dimensions[0].Int64Value != "0" {
		t.Fatalf("numeric dimension = %#v, want explicit zero", rows[0].Dimensions[0])
	}
	if rows[0].Dimensions[1].Int64Value != nil {
		t.Fatalf("string dimension = %#v, did not want int64Value", rows[0].Dimensions[1])
	}
}

func TestClientQueryMetricsUsesReportingEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		metricSet  MetricSet
		pathSuffix string
		metric     string
		userCohort UserCohort
	}{
		{name: "ANR rate", metricSet: MetricSetANRRate, pathSuffix: "anrRateMetricSet", metric: "anrRate", userCohort: UserCohortOSPublic},
		{name: "crash rate", metricSet: MetricSetCrashRate, pathSuffix: "crashRateMetricSet", metric: "crashRate", userCohort: UserCohortOSPublic},
		{name: "error count", metricSet: MetricSetErrorCount, pathSuffix: "errorCountMetricSet", metric: "errorReportCount"},
		{name: "excessive wakeup rate", metricSet: MetricSetExcessiveWakeupRate, pathSuffix: "excessiveWakeupRateMetricSet", metric: "excessiveWakeupRate", userCohort: UserCohortOSPublic},
		{name: "LMK rate", metricSet: MetricSetLMKRate, pathSuffix: "lmkRateMetricSet", metric: "userPerceivedLmkRate", userCohort: UserCohortOSPublic},
		{name: "slow rendering rate", metricSet: MetricSetSlowRenderingRate, pathSuffix: "slowRenderingRateMetricSet", metric: "slowRenderingRate20Fps", userCohort: UserCohortOSPublic},
		{name: "slow start rate", metricSet: MetricSetSlowStartRate, pathSuffix: "slowStartRateMetricSet", metric: "slowStartRate", userCohort: UserCohortOSPublic},
		{name: "stuck background wakelock rate", metricSet: MetricSetStuckBackgroundWakeLockRate, pathSuffix: "stuckBackgroundWakelockRateMetricSet", metric: "stuckBgWakelockRate", userCohort: UserCohortOSPublic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantPath := "/v1beta1/apps/com.example.app/" + tt.pathSuffix + ":query"
			client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != wantPath {
					t.Fatalf("path = %q, want %s", r.URL.Path, wantPath)
				}

				var request metricQueryRequestPayload
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := request.Metrics; len(got) != 1 || got[0] != tt.metric {
					t.Fatalf("metrics = %#v, want %s", got, tt.metric)
				}
				if got := request.Dimensions; len(got) != 1 || got[0] != "versionCode" {
					t.Fatalf("dimensions = %#v, want versionCode", got)
				}
				if request.Filter != `versionCode = 123` {
					t.Fatalf("filter = %q, want versionCode filter", request.Filter)
				}
				if request.TimelineSpec.AggregationPeriod != "DAILY" {
					t.Fatalf("aggregation = %q, want DAILY", request.TimelineSpec.AggregationPeriod)
				}
				if request.TimelineSpec.StartTime.Year != 2026 || request.TimelineSpec.StartTime.TimeZone.Id != "America/Los_Angeles" {
					t.Fatalf("startTime = %#v, want 2026 LA date", request.TimelineSpec.StartTime)
				}
				if request.PageSize != 50 || request.PageToken != "page-1" {
					t.Fatalf("pagination = (%d, %q), want (50, page-1)", request.PageSize, request.PageToken)
				}
				if request.UserCohort != tt.userCohort.String() {
					t.Fatalf("userCohort = %q, want %q", request.UserCohort, tt.userCohort)
				}

				_, _ = w.Write([]byte(`{
					"nextPageToken": "page-2",
					"rows": [
						{
							"aggregationPeriod": "DAILY",
							"startTime": {
								"year": 2026,
								"month": 5,
								"day": 1,
								"timeZone": {"id": "America/Los_Angeles"}
							},
							"dimensions": [
								{"dimension": "versionCode", "int64Value": "123", "valueLabel": "1.2.3"}
							],
							"metrics": [
								{
									"metric": "` + tt.metric + `",
									"decimalValue": {"value": "0.012"},
									"decimalValueConfidenceInterval": {
										"lowerBound": {"value": "0.010"},
										"upperBound": {"value": "0.014"}
									}
								}
							]
						}
					]
				}`))
			}))

			startDate, err := NewQueryDate("2026-05-01", "America/Los_Angeles")
			if err != nil {
				t.Fatalf("NewQueryDate() error = %v", err)
			}
			endDate, err := NewQueryDate("2026-05-19", "America/Los_Angeles")
			if err != nil {
				t.Fatalf("NewQueryDate() error = %v", err)
			}
			result, err := client.QueryMetrics(context.Background(), MetricQueryOptions{
				PackageName:       "com.example.app",
				MetricSet:         tt.metricSet,
				Metrics:           []string{tt.metric},
				Dimensions:        []string{"versionCode"},
				Filter:            `versionCode = 123`,
				AggregationPeriod: AggregationPeriodDaily,
				StartDate:         startDate,
				EndDate:           endDate,
				UserCohort:        tt.userCohort,
				PageSize:          50,
				PageToken:         "page-1",
			})
			if err != nil {
				t.Fatalf("QueryMetrics() error = %v", err)
			}
			if result.Name != "apps/com.example.app/"+tt.pathSuffix || result.NextPageToken != "page-2" {
				t.Fatalf("result = %#v, want name and next token", result)
			}
			if len(result.Rows) != 1 || len(result.Rows[0].Metrics) != 1 {
				t.Fatalf("rows = %#v, want one metric row", result.Rows)
			}
			row := result.Rows[0]
			if row.StartTime == nil || row.StartTime.TimeZone != "America/Los_Angeles" {
				t.Fatalf("startTime = %#v, want LA timezone", row.StartTime)
			}
			if row.Dimensions[0].Int64Value == nil || *row.Dimensions[0].Int64Value != "123" || row.Dimensions[0].ValueLabel != "1.2.3" {
				t.Fatalf("dimensions = %#v, want version code dimension", row.Dimensions)
			}
			if row.Metrics[0].DecimalValue != "0.012" || row.Metrics[0].DecimalValueConfidenceInterval.UpperBound != "0.014" {
				t.Fatalf("metrics = %#v, want decimal value and interval", row.Metrics)
			}
		})
	}
}

type metricQueryRequestPayload struct {
	Dimensions   []string                `json:"dimensions"`
	Filter       string                  `json:"filter"`
	Metrics      []string                `json:"metrics"`
	PageSize     int64                   `json:"pageSize"`
	PageToken    string                  `json:"pageToken"`
	TimelineSpec metricQueryTimelineSpec `json:"timelineSpec"`
	UserCohort   string                  `json:"userCohort"`
}

type metricQueryTimelineSpec struct {
	AggregationPeriod string              `json:"aggregationPeriod"`
	StartTime         metricQueryDateTime `json:"startTime"`
	EndTime           metricQueryDateTime `json:"endTime"`
}

type metricQueryDateTime struct {
	Year     int64               `json:"year"`
	Month    int64               `json:"month"`
	Day      int64               `json:"day"`
	TimeZone metricQueryTimeZone `json:"timeZone"`
}

type metricQueryTimeZone struct {
	Id string `json:"id"`
}
