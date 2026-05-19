package reporting

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

func TestErrorReportSearchOptionsValidate(t *testing.T) {
	packageName, err := play.NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	startDate, err := NewQueryDate("2026-05-01", "UTC")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}
	endDate, err := NewQueryDate("2026-05-19", "UTC")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}

	options := ErrorReportSearchOptions{
		PackageName: packageName,
		StartDate:   startDate,
		EndDate:     endDate,
		PageSize:    100,
	}
	if err := options.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	options.PageSize = -1
	if err := options.Validate(); err == nil {
		t.Fatal("expected page size validation error")
	}
	options.PageSize = 0
	options.StartDate = QueryDate{Year: 2026, Month: 5, Day: 1, TimeZone: "America/Los_Angeles"}
	options.EndDate = QueryDate{Year: 2026, Month: 5, Day: 19, TimeZone: "America/Los_Angeles"}
	if err := options.Validate(); err == nil {
		t.Fatal("expected unsupported timezone error")
	} else if !strings.Contains(err.Error(), "only support UTC") {
		t.Fatalf("error = %v, want UTC validation", err)
	}
}

func TestClientSearchErrorReportsUsesReportingEndpoint(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta1/apps/com.example.app/errorReports:search" {
			t.Fatalf("path = %q, want error report search endpoint", r.URL.Path)
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"filter":                         "errorIssueId = issue-1",
			"interval.startTime.year":        "2026",
			"interval.startTime.month":       "5",
			"interval.startTime.day":         "1",
			"interval.startTime.timeZone.id": "UTC",
			"interval.endTime.year":          "2026",
			"interval.endTime.month":         "5",
			"interval.endTime.day":           "19",
			"interval.endTime.timeZone.id":   "UTC",
			"pageSize":                       "25",
			"pageToken":                      "page-1",
		} {
			if got := query.Get(key); got != want {
				t.Fatalf("query[%s] = %q, want %q", key, got, want)
			}
		}
		_, _ = w.Write([]byte(`{
			"nextPageToken": "page-2",
			"errorReports": [
				{
					"name": "apps/com.example.app/errorReports/report-1",
					"issue": "apps/com.example.app/errorIssues/issue-1",
					"type": "CRASH",
					"eventTime": "2026-05-18T13:00:00Z",
					"appVersion": {"versionCode": "123"},
					"osVersion": {"apiLevel": "35"},
					"deviceModel": {
						"deviceId": {"buildBrand": "google", "buildDevice": "panther"},
						"marketingName": "Pixel 7",
						"deviceUri": "https://play.google.com/console/device"
					},
					"reportText": "java.lang.IllegalArgumentException: boom",
					"vcsInformation": "git abc123"
				}
			]
		}`))
	}))

	startDate, err := NewQueryDate("2026-05-01", "UTC")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}
	endDate, err := NewQueryDate("2026-05-19", "UTC")
	if err != nil {
		t.Fatalf("NewQueryDate() error = %v", err)
	}
	result, err := client.SearchErrorReports(context.Background(), ErrorReportSearchOptions{
		PackageName: "com.example.app",
		Filter:      "errorIssueId = issue-1",
		StartDate:   startDate,
		EndDate:     endDate,
		PageSize:    25,
		PageToken:   "page-1",
	})
	if err != nil {
		t.Fatalf("SearchErrorReports() error = %v", err)
	}
	if result.NextPageToken != "page-2" || len(result.Reports) != 1 {
		t.Fatalf("result = %#v, want one report and next token", result)
	}
	report := result.Reports[0]
	if report.Name != "apps/com.example.app/errorReports/report-1" || report.AppVersion != "123" || report.OSVersion != "35" {
		t.Fatalf("report = %#v, want mapped report", report)
	}
	if report.DeviceModel == nil || report.DeviceModel.BuildBrand != "google" || report.DeviceModel.MarketingName != "Pixel 7" {
		t.Fatalf("device model = %#v, want mapped device", report.DeviceModel)
	}
	if !strings.Contains(report.ReportText, "IllegalArgumentException") || report.VCSInformation != "git abc123" {
		t.Fatalf("report = %#v, want report text and VCS info", report)
	}
}
