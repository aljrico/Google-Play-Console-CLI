package reporting

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
)

func TestErrorIssueSearchOptionsValidate(t *testing.T) {
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

	options := ErrorIssueSearchOptions{
		PackageName:            packageName,
		StartDate:              startDate,
		EndDate:                endDate,
		OrderBy:                "errorReportCount desc",
		SampleErrorReportLimit: 1,
	}
	if err := options.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	options.OrderBy = "lastSeen desc"
	if err := options.Validate(); err == nil {
		t.Fatal("expected unsupported order field error")
	}
	options.OrderBy = ""
	options.SampleErrorReportLimit = 2
	if err := options.Validate(); err == nil {
		t.Fatal("expected sample error report limit error")
	}
	options.SampleErrorReportLimit = 0
	options.StartDate = endDate
	options.EndDate = startDate
	if err := options.Validate(); err == nil {
		t.Fatal("expected date range error")
	}
	options.StartDate = QueryDate{Year: 2026, Month: 5, Day: 1, TimeZone: "America/Los_Angeles"}
	options.EndDate = QueryDate{Year: 2026, Month: 5, Day: 19, TimeZone: "America/Los_Angeles"}
	if err := options.Validate(); err == nil {
		t.Fatal("expected unsupported timezone error")
	} else if !strings.Contains(err.Error(), "only support UTC") {
		t.Fatalf("error = %v, want UTC validation", err)
	}
}

func TestClientSearchErrorIssuesUsesReportingEndpoint(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta1/apps/com.example.app/errorIssues:search" {
			t.Fatalf("path = %q, want error issue search endpoint", r.URL.Path)
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"filter":                         "versionCode = 123 AND errorIssueType = CRASH",
			"orderBy":                        "errorReportCount desc",
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
			"sampleErrorReportLimit":         "1",
		} {
			if got := query.Get(key); got != want {
				t.Fatalf("query[%s] = %q, want %q", key, got, want)
			}
		}
		_, _ = w.Write([]byte(`{
			"nextPageToken": "page-2",
			"errorIssues": [
				{
					"name": "apps/com.example.app/errorIssues/issue-1",
					"type": "CRASH",
					"cause": "IllegalArgumentException",
					"location": "MainActivity.onCreate",
					"issueUri": "https://play.google.com/console/issue",
					"distinctUsers": "12",
					"distinctUsersPercent": {"value": "0.34"},
					"errorReportCount": "42",
					"firstAppVersion": {"versionCode": "100"},
					"lastAppVersion": {"versionCode": "123"},
					"firstOsVersion": {"apiLevel": "28"},
					"lastOsVersion": {"apiLevel": "35"},
					"lastErrorReportTime": "2026-05-18T13:00:00Z",
					"sampleErrorReports": ["apps/com.example.app/errorReports/report-1"],
					"annotations": [
						{"category": "Insight", "title": "Spike", "body": "Recent increase"}
					]
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
	result, err := client.SearchErrorIssues(context.Background(), ErrorIssueSearchOptions{
		PackageName:            "com.example.app",
		Filter:                 "versionCode = 123 AND errorIssueType = CRASH",
		OrderBy:                "errorReportCount desc",
		StartDate:              startDate,
		EndDate:                endDate,
		PageSize:               25,
		PageToken:              "page-1",
		SampleErrorReportLimit: 1,
	})
	if err != nil {
		t.Fatalf("SearchErrorIssues() error = %v", err)
	}
	if result.NextPageToken != "page-2" || len(result.Issues) != 1 {
		t.Fatalf("result = %#v, want one issue and next token", result)
	}
	issue := result.Issues[0]
	if issue.Name != "apps/com.example.app/errorIssues/issue-1" || issue.ErrorReportCount != "42" {
		t.Fatalf("issue = %#v, want mapped issue", issue)
	}
	if issue.DistinctUsersPercent != "0.34" || issue.FirstAppVersion != "100" || issue.LastOSVersion != "35" {
		t.Fatalf("issue = %#v, want decimal and version fields", issue)
	}
	if len(issue.Annotations) != 1 || issue.Annotations[0].Title != "Spike" {
		t.Fatalf("annotations = %#v, want mapped annotation", issue.Annotations)
	}
}
