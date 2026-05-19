package reporting

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
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

func TestClientQueryMetricsUsesReportingEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		metricSet  MetricSet
		pathSuffix string
		userCohort UserCohort
	}{
		{name: "ANR rate", metricSet: MetricSetANRRate, pathSuffix: "anrRateMetricSet", userCohort: UserCohortOSPublic},
		{name: "crash rate", metricSet: MetricSetCrashRate, pathSuffix: "crashRateMetricSet", userCohort: UserCohortOSPublic},
		{name: "error count", metricSet: MetricSetErrorCount, pathSuffix: "errorCountMetricSet"},
		{name: "excessive wakeup rate", metricSet: MetricSetExcessiveWakeupRate, pathSuffix: "excessiveWakeupRateMetricSet", userCohort: UserCohortOSPublic},
		{name: "LMK rate", metricSet: MetricSetLMKRate, pathSuffix: "lmkRateMetricSet", userCohort: UserCohortOSPublic},
		{name: "slow rendering rate", metricSet: MetricSetSlowRenderingRate, pathSuffix: "slowRenderingRateMetricSet", userCohort: UserCohortOSPublic},
		{name: "slow start rate", metricSet: MetricSetSlowStartRate, pathSuffix: "slowStartRateMetricSet", userCohort: UserCohortOSPublic},
		{name: "stuck background wakelock rate", metricSet: MetricSetStuckBackgroundWakeLockRate, pathSuffix: "stuckBackgroundWakelockRateMetricSet", userCohort: UserCohortOSPublic},
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
				if got := request.Metrics; len(got) != 1 || got[0] != "crashRate" {
					t.Fatalf("metrics = %#v, want crashRate", got)
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
									"metric": "crashRate",
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
				Metrics:           []string{"crashRate"},
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
			if row.Dimensions[0].Int64Value != 123 || row.Dimensions[0].ValueLabel != "1.2.3" {
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
