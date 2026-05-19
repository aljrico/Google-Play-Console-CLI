package reporting

import (
	"context"
	"net/http"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

func TestAnomalyListOptionsValidate(t *testing.T) {
	packageName, err := play.NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	options := AnomalyListOptions{
		PackageName: packageName,
		PageSize:    10,
	}
	if err := options.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	options.PageSize = -1
	if err := options.Validate(); err == nil {
		t.Fatal("expected page size validation error")
	}
}

func TestClientListAnomaliesUsesReportingEndpoint(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta1/apps/com.example.app/anomalies" {
			t.Fatalf("path = %q, want anomalies endpoint", r.URL.Path)
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"filter":    `activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")`,
			"pageSize":  "25",
			"pageToken": "page-1",
		} {
			if got := query.Get(key); got != want {
				t.Fatalf("query[%s] = %q, want %q", key, got, want)
			}
		}
		_, _ = w.Write([]byte(`{
			"nextPageToken": "page-2",
			"anomalies": [
				{
					"name": "apps/com.example.app/anomalies/anomaly-1",
					"metricSet": "apps/com.example.app/crashRateMetricSet",
					"metric": {
						"metric": "crashRate",
						"decimalValue": {"value": "0.05"},
						"decimalValueConfidenceInterval": {
							"lowerBound": {"value": "0.04"},
							"upperBound": {"value": "0.06"}
						}
					},
					"dimensions": [
						{"dimension": "versionCode", "int64Value": "123", "valueLabel": "1.2.3"},
						{"dimension": "apiLevel", "stringValue": "35"}
					],
					"timelineSpec": {
						"aggregationPeriod": "DAILY",
						"startTime": {
							"year": 2026,
							"month": 5,
							"day": 1,
							"timeZone": {"id": "America/Los_Angeles"}
						},
						"endTime": {
							"year": 2026,
							"month": 5,
							"day": 3,
							"timeZone": {"id": "America/Los_Angeles"}
						}
					}
				}
			]
		}`))
	}))

	result, err := client.ListAnomalies(context.Background(), AnomalyListOptions{
		PackageName: "com.example.app",
		Filter:      `activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")`,
		PageSize:    25,
		PageToken:   "page-1",
	})
	if err != nil {
		t.Fatalf("ListAnomalies() error = %v", err)
	}
	if result.NextPageToken != "page-2" || len(result.Anomalies) != 1 {
		t.Fatalf("result = %#v, want one anomaly and next token", result)
	}
	anomaly := result.Anomalies[0]
	if anomaly.Name != "apps/com.example.app/anomalies/anomaly-1" || anomaly.MetricSet != "apps/com.example.app/crashRateMetricSet" {
		t.Fatalf("anomaly = %#v, want name and metric set", anomaly)
	}
	if anomaly.Metric == nil || anomaly.Metric.DecimalValue != "0.05" || anomaly.Metric.DecimalValueConfidenceInterval.LowerBound != "0.04" {
		t.Fatalf("metric = %#v, want decimal metric with interval", anomaly.Metric)
	}
	if len(anomaly.Dimensions) != 2 || anomaly.Dimensions[0].Int64Value == nil || *anomaly.Dimensions[0].Int64Value != "123" {
		t.Fatalf("dimensions = %#v, want version code dimension", anomaly.Dimensions)
	}
	if anomaly.Dimensions[1].StringValue != "35" || anomaly.Dimensions[1].Int64Value != nil {
		t.Fatalf("api level dimension = %#v, want string-only apiLevel", anomaly.Dimensions[1])
	}
	if anomaly.TimelineSpec == nil || anomaly.TimelineSpec.StartTime == nil || anomaly.TimelineSpec.StartTime.TimeZone != "America/Los_Angeles" {
		t.Fatalf("timeline = %#v, want mapped LA timeline", anomaly.TimelineSpec)
	}
}
