package reporting

import (
	"context"
	"fmt"

	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

type ErrorReportSearchOptions struct {
	PackageName play.PackageName `json:"packageName"`
	Filter      string           `json:"filter,omitempty"`
	StartDate   QueryDate        `json:"startDate"`
	EndDate     QueryDate        `json:"endDate"`
	PageSize    int64            `json:"pageSize,omitempty"`
	PageToken   string           `json:"pageToken,omitempty"`
}

func (o ErrorReportSearchOptions) Validate() error {
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
	return nil
}

type ErrorReportSearchResult struct {
	PackageName   play.PackageName         `json:"packageName"`
	Reports       []ErrorReport            `json:"reports"`
	NextPageToken string                   `json:"nextPageToken,omitempty"`
	Options       ErrorReportSearchOptions `json:"options"`
}

type ErrorReport struct {
	Name           string              `json:"name"`
	Issue          string              `json:"issue,omitempty"`
	Type           string              `json:"type,omitempty"`
	EventTime      string              `json:"eventTime,omitempty"`
	AppVersion     string              `json:"appVersion,omitempty"`
	OSVersion      string              `json:"osVersion,omitempty"`
	DeviceModel    *DeviceModelSummary `json:"deviceModel,omitempty"`
	ReportText     string              `json:"reportText,omitempty"`
	VCSInformation string              `json:"vcsInformation,omitempty"`
}

type DeviceModelSummary struct {
	BuildBrand    string `json:"buildBrand,omitempty"`
	BuildDevice   string `json:"buildDevice,omitempty"`
	MarketingName string `json:"marketingName,omitempty"`
	DeviceURI     string `json:"deviceUri,omitempty"`
}

type ErrorReportSearcher interface {
	SearchErrorReports(ctx context.Context, options ErrorReportSearchOptions) (ErrorReportSearchResult, error)
}

func SearchErrorReports(ctx context.Context, searcher ErrorReportSearcher, options ErrorReportSearchOptions) (ErrorReportSearchResult, error) {
	if err := options.Validate(); err != nil {
		return ErrorReportSearchResult{}, err
	}
	if searcher == nil {
		return ErrorReportSearchResult{}, fmt.Errorf("error report searcher is required")
	}
	return searcher.SearchErrorReports(ctx, options)
}

func (c Client) SearchErrorReports(ctx context.Context, options ErrorReportSearchOptions) (ErrorReportSearchResult, error) {
	parent := fmt.Sprintf("apps/%s", options.PackageName)
	call := c.service.Vitals.Errors.Reports.Search(parent).Context(ctx)
	if options.Filter != "" {
		call.Filter(options.Filter)
	}
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	applyErrorReportSearchInterval(call, options.StartDate, options.EndDate)
	response, err := call.Do()
	if err != nil {
		return ErrorReportSearchResult{}, fmt.Errorf("search error reports for %s: %w", options.PackageName, err)
	}
	return ErrorReportSearchResult{
		PackageName:   options.PackageName,
		Reports:       errorReportsFromAPI(response.ErrorReports),
		NextPageToken: response.NextPageToken,
		Options:       options,
	}, nil
}

func applyErrorReportSearchInterval(call *playdeveloperreporting.VitalsErrorsReportsSearchCall, startDate QueryDate, endDate QueryDate) {
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

func errorReportsFromAPI(apiReports []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport) []ErrorReport {
	reports := make([]ErrorReport, 0, len(apiReports))
	for _, apiReport := range apiReports {
		if apiReport == nil {
			continue
		}
		reports = append(reports, ErrorReport{
			Name:           apiReport.Name,
			Issue:          apiReport.Issue,
			Type:           apiReport.Type,
			EventTime:      apiReport.EventTime,
			AppVersion:     appVersionFromAPI(apiReport.AppVersion),
			OSVersion:      osVersionFromAPI(apiReport.OsVersion),
			DeviceModel:    deviceModelSummaryFromAPI(apiReport.DeviceModel),
			ReportText:     apiReport.ReportText,
			VCSInformation: apiReport.VcsInformation,
		})
	}
	return reports
}

func deviceModelSummaryFromAPI(apiDeviceModel *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceModelSummary) *DeviceModelSummary {
	if apiDeviceModel == nil {
		return nil
	}
	summary := &DeviceModelSummary{
		MarketingName: apiDeviceModel.MarketingName,
		DeviceURI:     apiDeviceModel.DeviceUri,
	}
	if apiDeviceModel.DeviceId != nil {
		summary.BuildBrand = apiDeviceModel.DeviceId.BuildBrand
		summary.BuildDevice = apiDeviceModel.DeviceId.BuildDevice
	}
	return summary
}
