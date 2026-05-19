package reporting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"google.golang.org/api/option"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

func TestNewMetricSetValidatesSupportedValues(t *testing.T) {
	metricSet, err := NewMetricSet("crash-rate")
	if err != nil {
		t.Fatalf("NewMetricSet() error = %v", err)
	}
	if metricSet != MetricSetCrashRate {
		t.Fatalf("MetricSet = %q, want crash-rate", metricSet)
	}
	if _, err := NewMetricSet("crashes"); err == nil {
		t.Fatal("expected unsupported metric set error")
	} else if !strings.Contains(err.Error(), "anr-rate, crash-rate") {
		t.Fatalf("error = %v, want supported values text", err)
	}
	if _, err := NewMetricSet(""); err == nil {
		t.Fatal("expected missing metric set error")
	} else if strings.Contains(err.Error(), "%!") {
		t.Fatalf("error = %v, want clean formatting", err)
	}
}

func TestGetMetricSetPassesOptionsToGetter(t *testing.T) {
	packageName, err := play.NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeMetricSetGetter{
		result: MetricSetMetadata{Name: "apps/com.example.app/crashRateMetricSet"},
	}

	result, err := GetMetricSet(context.Background(), getter, MetricSetGetOptions{
		PackageName: packageName,
		MetricSet:   MetricSetCrashRate,
	})
	if err != nil {
		t.Fatalf("GetMetricSet() error = %v", err)
	}
	if result.Name != "apps/com.example.app/crashRateMetricSet" {
		t.Fatalf("Name = %q, want crash metric set", result.Name)
	}
	if !reflect.DeepEqual(getter.options, MetricSetGetOptions{PackageName: packageName, MetricSet: MetricSetCrashRate}) {
		t.Fatalf("options = %#v", getter.options)
	}
}

func TestClientGetMetricSetUsesReportingEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		metricSet  MetricSet
		pathSuffix string
	}{
		{name: "ANR rate", metricSet: MetricSetANRRate, pathSuffix: "anrRateMetricSet"},
		{name: "crash rate", metricSet: MetricSetCrashRate, pathSuffix: "crashRateMetricSet"},
		{name: "error count", metricSet: MetricSetErrorCount, pathSuffix: "errorCountMetricSet"},
		{name: "excessive wakeup rate", metricSet: MetricSetExcessiveWakeupRate, pathSuffix: "excessiveWakeupRateMetricSet"},
		{name: "LMK rate", metricSet: MetricSetLMKRate, pathSuffix: "lmkRateMetricSet"},
		{name: "slow rendering rate", metricSet: MetricSetSlowRenderingRate, pathSuffix: "slowRenderingRateMetricSet"},
		{name: "slow start rate", metricSet: MetricSetSlowStartRate, pathSuffix: "slowStartRateMetricSet"},
		{name: "stuck background wakelock rate", metricSet: MetricSetStuckBackgroundWakeLockRate, pathSuffix: "stuckBackgroundWakelockRateMetricSet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantPath := "/v1beta1/apps/com.example.app/" + tt.pathSuffix
			client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != wantPath {
					t.Fatalf("path = %q, want %s", r.URL.Path, wantPath)
				}
				_, _ = w.Write([]byte(`{
					"name": "apps/com.example.app/` + tt.pathSuffix + `",
					"freshnessInfo": {
						"freshnesses": [
							{
								"aggregationPeriod": "HOURLY",
								"latestEndTime": {
									"year": 2026,
									"month": 5,
									"day": 19,
									"hours": 1,
									"timeZone": {"id": "America/Los_Angeles"}
								}
							},
							{
								"aggregationPeriod": "DAILY",
								"latestEndTime": {
									"year": 2026,
									"month": 5,
									"day": 19,
									"hours": 0,
									"timeZone": {"id": "America/Los_Angeles"}
								}
							}
						]
					}
				}`))
			}))

			result, err := client.GetMetricSet(context.Background(), MetricSetGetOptions{
				PackageName: "com.example.app",
				MetricSet:   tt.metricSet,
			})
			if err != nil {
				t.Fatalf("GetMetricSet() error = %v", err)
			}
			if result.Name != "apps/com.example.app/"+tt.pathSuffix {
				t.Fatalf("Name = %q, want metric set name", result.Name)
			}
			if len(result.Freshnesses) != 2 {
				t.Fatalf("len(Freshnesses) = %d, want 2", len(result.Freshnesses))
			}
			if result.Freshnesses[0].AggregationPeriod != "DAILY" || result.Freshnesses[1].AggregationPeriod != "HOURLY" {
				t.Fatalf("Freshnesses = %#v, want sorted by aggregation period", result.Freshnesses)
			}
			if result.Freshnesses[0].LatestEndTime == nil || result.Freshnesses[0].LatestEndTime.TimeZone != "America/Los_Angeles" {
				t.Fatalf("LatestEndTime = %#v, want LA timezone", result.Freshnesses[0].LatestEndTime)
			}
		})
	}
}

type fakeMetricSetGetter struct {
	options MetricSetGetOptions
	result  MetricSetMetadata
}

func (g *fakeMetricSetGetter) GetMetricSet(ctx context.Context, options MetricSetGetOptions) (MetricSetMetadata, error) {
	g.options = options
	return g.result, nil
}

func newTestClient(t *testing.T, handler http.Handler) Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	service, err := playdeveloperreporting.NewService(
		context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return Client{service: service}
}
