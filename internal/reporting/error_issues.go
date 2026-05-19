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

type ErrorIssueSearchOptions struct {
	PackageName            play.PackageName `json:"packageName"`
	Filter                 string           `json:"filter,omitempty"`
	OrderBy                string           `json:"orderBy,omitempty"`
	StartDate              QueryDate        `json:"startDate"`
	EndDate                QueryDate        `json:"endDate"`
	PageSize               int64            `json:"pageSize,omitempty"`
	PageToken              string           `json:"pageToken,omitempty"`
	SampleErrorReportLimit int64            `json:"sampleErrorReportLimit,omitempty"`
}

func (o ErrorIssueSearchOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
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
	if o.StartDate.TimeZone != o.EndDate.TimeZone {
		return fmt.Errorf("start and end time zones must match")
	}
	if err := validateErrorSearchIntervalTimeZone(o.StartDate.TimeZone); err != nil {
		return err
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	if o.SampleErrorReportLimit < 0 || o.SampleErrorReportLimit > 1 {
		return fmt.Errorf("sample error report limit must be 0 or 1")
	}
	if err := validateErrorIssueOrderBy(o.OrderBy); err != nil {
		return err
	}
	return nil
}

func validateErrorSearchIntervalTimeZone(value string) error {
	if value == "" {
		return nil
	}
	if value != "UTC" {
		return fmt.Errorf("error search intervals only support UTC time zone")
	}
	if _, err := time.LoadLocation(value); err != nil {
		return fmt.Errorf("invalid time zone %q: %w", value, err)
	}
	return nil
}

func validateErrorIssueOrderBy(value string) error {
	if value == "" {
		return nil
	}
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return fmt.Errorf("order by must use FIELD asc|desc")
	}
	if parts[0] != "errorReportCount" && parts[0] != "distinctUsers" {
		return fmt.Errorf("unsupported error issue order field %q; supported values: errorReportCount, distinctUsers", parts[0])
	}
	if parts[1] != "asc" && parts[1] != "desc" {
		return fmt.Errorf("unsupported error issue order direction %q; supported values: asc, desc", parts[1])
	}
	return nil
}

type ErrorIssueSearchResult struct {
	PackageName   play.PackageName        `json:"packageName"`
	Issues        []ErrorIssue            `json:"issues"`
	NextPageToken string                  `json:"nextPageToken,omitempty"`
	Options       ErrorIssueSearchOptions `json:"options"`
}

type ErrorIssue struct {
	Name                 string            `json:"name"`
	Type                 string            `json:"type,omitempty"`
	Cause                string            `json:"cause,omitempty"`
	Location             string            `json:"location,omitempty"`
	IssueURI             string            `json:"issueUri,omitempty"`
	DistinctUsers        string            `json:"distinctUsers,omitempty"`
	DistinctUsersPercent string            `json:"distinctUsersPercent,omitempty"`
	ErrorReportCount     string            `json:"errorReportCount,omitempty"`
	FirstAppVersion      string            `json:"firstAppVersion,omitempty"`
	LastAppVersion       string            `json:"lastAppVersion,omitempty"`
	FirstOSVersion       string            `json:"firstOsVersion,omitempty"`
	LastOSVersion        string            `json:"lastOsVersion,omitempty"`
	LastErrorReportTime  string            `json:"lastErrorReportTime,omitempty"`
	SampleErrorReports   []string          `json:"sampleErrorReports,omitempty"`
	Annotations          []IssueAnnotation `json:"annotations,omitempty"`
}

type IssueAnnotation struct {
	Category string `json:"category,omitempty"`
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
}

type ErrorIssueSearcher interface {
	SearchErrorIssues(ctx context.Context, options ErrorIssueSearchOptions) (ErrorIssueSearchResult, error)
}

func SearchErrorIssues(ctx context.Context, searcher ErrorIssueSearcher, options ErrorIssueSearchOptions) (ErrorIssueSearchResult, error) {
	if err := options.Validate(); err != nil {
		return ErrorIssueSearchResult{}, err
	}
	if searcher == nil {
		return ErrorIssueSearchResult{}, fmt.Errorf("error issue searcher is required")
	}
	return searcher.SearchErrorIssues(ctx, options)
}

func (c Client) SearchErrorIssues(ctx context.Context, options ErrorIssueSearchOptions) (ErrorIssueSearchResult, error) {
	parent := fmt.Sprintf("apps/%s", options.PackageName)
	call := c.service.Vitals.Errors.Issues.Search(parent).Context(ctx)
	if options.Filter != "" {
		call.Filter(options.Filter)
	}
	if options.OrderBy != "" {
		call.OrderBy(options.OrderBy)
	}
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	if options.SampleErrorReportLimit > 0 {
		call.SampleErrorReportLimit(options.SampleErrorReportLimit)
	}
	applyErrorIssueSearchInterval(call, options.StartDate, options.EndDate)
	response, err := call.Do()
	if err != nil {
		return ErrorIssueSearchResult{}, fmt.Errorf("search error issues for %s: %w", options.PackageName, err)
	}
	return ErrorIssueSearchResult{
		PackageName:   options.PackageName,
		Issues:        errorIssuesFromAPI(response.ErrorIssues),
		NextPageToken: response.NextPageToken,
		Options:       options,
	}, nil
}

func applyErrorIssueSearchInterval(call *playdeveloperreporting.VitalsErrorsIssuesSearchCall, startDate QueryDate, endDate QueryDate) {
	call.IntervalStartTimeYear(startDate.Year)
	call.IntervalStartTimeMonth(startDate.Month)
	call.IntervalStartTimeDay(startDate.Day)
	call.IntervalEndTimeYear(endDate.Year)
	call.IntervalEndTimeMonth(endDate.Month)
	call.IntervalEndTimeDay(endDate.Day)
	if startDate.TimeZone != "" {
		call.IntervalStartTimeTimeZoneId(startDate.TimeZone)
		call.IntervalEndTimeTimeZoneId(endDate.TimeZone)
	}
}

func errorIssuesFromAPI(apiIssues []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue) []ErrorIssue {
	issues := make([]ErrorIssue, 0, len(apiIssues))
	for _, apiIssue := range apiIssues {
		if apiIssue == nil {
			continue
		}
		issues = append(issues, ErrorIssue{
			Name:                 apiIssue.Name,
			Type:                 apiIssue.Type,
			Cause:                apiIssue.Cause,
			Location:             apiIssue.Location,
			IssueURI:             apiIssue.IssueUri,
			DistinctUsers:        countString(apiIssue.DistinctUsers),
			DistinctUsersPercent: decimalFromAPI(apiIssue.DistinctUsersPercent),
			ErrorReportCount:     countString(apiIssue.ErrorReportCount),
			FirstAppVersion:      appVersionFromAPI(apiIssue.FirstAppVersion),
			LastAppVersion:       appVersionFromAPI(apiIssue.LastAppVersion),
			FirstOSVersion:       osVersionFromAPI(apiIssue.FirstOsVersion),
			LastOSVersion:        osVersionFromAPI(apiIssue.LastOsVersion),
			LastErrorReportTime:  apiIssue.LastErrorReportTime,
			SampleErrorReports:   append([]string(nil), apiIssue.SampleErrorReports...),
			Annotations:          issueAnnotationsFromAPI(apiIssue.Annotations),
		})
	}
	return issues
}

func issueAnnotationsFromAPI(apiAnnotations []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1IssueAnnotation) []IssueAnnotation {
	annotations := make([]IssueAnnotation, 0, len(apiAnnotations))
	for _, apiAnnotation := range apiAnnotations {
		if apiAnnotation == nil {
			continue
		}
		annotations = append(annotations, IssueAnnotation{
			Category: apiAnnotation.Category,
			Title:    apiAnnotation.Title,
			Body:     apiAnnotation.Body,
		})
	}
	return annotations
}

func countString(value int64) string {
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}

func appVersionFromAPI(apiVersion *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1AppVersion) string {
	if apiVersion == nil {
		return ""
	}
	return strconv.FormatInt(apiVersion.VersionCode, 10)
}

func osVersionFromAPI(apiVersion *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1OsVersion) string {
	if apiVersion == nil {
		return ""
	}
	return strconv.FormatInt(apiVersion.ApiLevel, 10)
}
