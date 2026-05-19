package reporting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestClientGetMetricSetUsesReportingEndpoint(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta1/apps/com.example.app/crashRateMetricSet" {
			t.Fatalf("path = %q, want crash metric set endpoint", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"name": "apps/com.example.app/crashRateMetricSet",
			"freshnessInfo": {
				"freshnesses": [
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
		MetricSet:   MetricSetCrashRate,
	})
	if err != nil {
		t.Fatalf("GetMetricSet() error = %v", err)
	}
	if result.Name != "apps/com.example.app/crashRateMetricSet" {
		t.Fatalf("Name = %q, want crash metric set", result.Name)
	}
	if len(result.Freshnesses) != 1 {
		t.Fatalf("len(Freshnesses) = %d, want 1", len(result.Freshnesses))
	}
	if result.Freshnesses[0].LatestEndTime == nil || result.Freshnesses[0].LatestEndTime.TimeZone != "America/Los_Angeles" {
		t.Fatalf("LatestEndTime = %#v, want LA timezone", result.Freshnesses[0].LatestEndTime)
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
