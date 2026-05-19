package analytics

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aljrico/Google-Play-Console-CLI/internal/decimal"
)

type StatsSummaryOptions struct {
	File string `json:"file"`
}

type StatsSummary struct {
	File        string          `json:"file"`
	Rows        int             `json:"rows"`
	PackageName string          `json:"packageName,omitempty"`
	StartDate   string          `json:"startDate,omitempty"`
	EndDate     string          `json:"endDate,omitempty"`
	Metrics     []MetricSummary `json:"metrics"`
}

type MetricSummary struct {
	Name        string `json:"name"`
	Aggregation string `json:"aggregation"`
	Value       string `json:"value"`
}

var dimensionHeaders = map[string]bool{
	"android os version": true,
	"app version":        true,
	"app version code":   true,
	"app version name":   true,
	"carrier":            true,
	"country":            true,
	"country/region":     true,
	"date":               true,
	"device":             true,
	"device model":       true,
	"form factor":        true,
	"installer":          true,
	"language":           true,
	"package name":       true,
	"search term":        true,
	"traffic source":     true,
	"utm campaign":       true,
	"utm source":         true,
}

func SummarizeStats(options StatsSummaryOptions) (StatsSummary, error) {
	if strings.TrimSpace(options.File) == "" {
		return StatsSummary{}, fmt.Errorf("file is required")
	}
	file, err := os.Open(options.File)
	if err != nil {
		return StatsSummary{}, fmt.Errorf("open analytics report %s: %w", options.File, err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return StatsSummary{}, fmt.Errorf("analytics report is empty")
		}
		return StatsSummary{}, fmt.Errorf("read analytics report %s: %w", options.File, err)
	}
	indexes := statsColumnIndexes(headers)
	if indexes.date < 0 {
		return StatsSummary{}, fmt.Errorf("analytics report requires a Date column")
	}
	if indexes.packageName < 0 {
		return StatsSummary{}, fmt.Errorf("analytics report requires a Package name column")
	}
	accumulators := map[string]metricAccumulator{}
	rowCount := 0
	packageName := ""
	startDate := ""
	endDate := ""
	for rowIndex := 2; ; rowIndex++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return StatsSummary{}, fmt.Errorf("read analytics report row %d: %w", rowIndex, err)
		}
		if isEmptyRecord(record) {
			continue
		}
		if len(record) <= indexes.date || len(record) <= indexes.packageName {
			return StatsSummary{}, fmt.Errorf("analytics report row %d is missing required columns", rowIndex)
		}
		date := strings.TrimSpace(record[indexes.date])
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return StatsSummary{}, fmt.Errorf("analytics report row %d date: %w", rowIndex, err)
		}
		if startDate == "" || date < startDate {
			startDate = date
		}
		if endDate == "" || date > endDate {
			endDate = date
		}
		rowPackageName := strings.TrimSpace(record[indexes.packageName])
		if rowPackageName == "" {
			return StatsSummary{}, fmt.Errorf("analytics report row %d package name is required", rowIndex)
		}
		if packageName == "" {
			packageName = rowPackageName
		} else if packageName != rowPackageName {
			packageName = "multiple"
		}
		for _, metric := range indexes.metrics {
			if len(record) <= metric.index {
				continue
			}
			value := strings.TrimSpace(record[metric.index])
			if value == "" {
				continue
			}
			amount, err := decimal.Parse(value)
			if err != nil {
				return StatsSummary{}, fmt.Errorf("analytics report row %d metric %s: %w", rowIndex, metric.name, err)
			}
			accumulator := accumulators[metric.name]
			accumulator.name = metric.name
			accumulator.aggregation = metric.aggregation
			accumulator.value = accumulator.value.Add(amount)
			accumulator.count++
			accumulators[metric.name] = accumulator
		}
		rowCount++
	}
	metrics := make([]MetricSummary, 0, len(accumulators))
	for _, accumulator := range accumulators {
		metrics = append(metrics, accumulator.Summary())
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return StatsSummary{
		File:        options.File,
		Rows:        rowCount,
		PackageName: packageName,
		StartDate:   startDate,
		EndDate:     endDate,
		Metrics:     metrics,
	}, nil
}

type statsIndexes struct {
	date        int
	packageName int
	metrics     []metricIndex
}

type metricIndex struct {
	name        string
	index       int
	aggregation metricAggregation
}

func statsColumnIndexes(headers []string) statsIndexes {
	indexes := statsIndexes{date: -1, packageName: -1}
	for index, header := range headers {
		name := cleanHeader(header)
		switch name {
		case "date":
			indexes.date = index
		case "package name":
			indexes.packageName = index
		default:
			if !dimensionHeaders[name] {
				indexes.metrics = append(indexes.metrics, metricIndex{
					name:        strings.TrimPrefix(strings.TrimSpace(header), "\ufeff"),
					index:       index,
					aggregation: aggregationForHeader(name),
				})
			}
		}
	}
	return indexes
}

type metricAggregation string

const (
	metricAggregationAverage metricAggregation = "average"
	metricAggregationSum     metricAggregation = "sum"
)

type metricAccumulator struct {
	name        string
	aggregation metricAggregation
	value       decimal.Amount
	count       int
}

func (a metricAccumulator) Summary() MetricSummary {
	value := a.value.String()
	if a.aggregation == metricAggregationAverage {
		value = a.value.Average(a.count)
	}
	return MetricSummary{
		Name:        a.name,
		Aggregation: string(a.aggregation),
		Value:       value,
	}
}

func aggregationForHeader(name string) metricAggregation {
	if strings.Contains(name, "average") ||
		strings.Contains(name, "avg") ||
		strings.Contains(name, "conversion") ||
		strings.Contains(name, "percentage") ||
		strings.Contains(name, "rating") ||
		strings.Contains(name, "rate") ||
		strings.Contains(name, "%") {
		return metricAggregationAverage
	}
	return metricAggregationSum
}

func cleanHeader(header string) string {
	header = strings.TrimPrefix(strings.TrimSpace(header), "\ufeff")
	return strings.ToLower(strings.Join(strings.Fields(header), " "))
}

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
