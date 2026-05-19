package insights

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/reporting"
)

type AnomalySummaryOptions struct {
	File string `json:"file"`
}

type AnomalySummary struct {
	File        string   `json:"file"`
	Total       int      `json:"total"`
	MetricSets  []Count  `json:"metricSets,omitempty"`
	Metrics     []Count  `json:"metrics,omitempty"`
	PackageName string   `json:"packageName,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
}

type Count struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func SummarizeAnomalies(options AnomalySummaryOptions) (AnomalySummary, error) {
	if strings.TrimSpace(options.File) == "" {
		return AnomalySummary{}, fmt.Errorf("file is required")
	}
	content, err := os.ReadFile(options.File)
	if err != nil {
		return AnomalySummary{}, fmt.Errorf("read anomaly file %s: %w", options.File, err)
	}
	var result reporting.AnomalyListResult
	if err := json.Unmarshal(content, &result); err != nil {
		return AnomalySummary{}, fmt.Errorf("parse anomaly file %s: %w", options.File, err)
	}
	return summarizeAnomalyList(options.File, result), nil
}

func summarizeAnomalyList(file string, result reporting.AnomalyListResult) AnomalySummary {
	metricSetCounts := map[string]int{}
	metricCounts := map[string]int{}
	for _, anomaly := range result.Anomalies {
		if anomaly.MetricSet != "" {
			metricSetCounts[anomaly.MetricSet]++
		}
		if anomaly.Metric != nil && anomaly.Metric.Metric != "" {
			metricCounts[anomaly.Metric.Metric]++
		}
	}
	summary := AnomalySummary{
		File:        file,
		Total:       len(result.Anomalies),
		MetricSets:  countsFromMap(metricSetCounts),
		Metrics:     countsFromMap(metricCounts),
		PackageName: string(result.PackageName),
	}
	summary.Highlights = highlights(summary)
	return summary
}

func countsFromMap(values map[string]int) []Count {
	counts := make([]Count, 0, len(values))
	for name, count := range values {
		counts = append(counts, Count{Name: name, Count: count})
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count == counts[j].Count {
			return counts[i].Name < counts[j].Name
		}
		return counts[i].Count > counts[j].Count
	})
	return counts
}

func highlights(summary AnomalySummary) []string {
	values := []string{fmt.Sprintf("%d anomalies", summary.Total)}
	if len(summary.MetricSets) > 0 {
		values = append(values, fmt.Sprintf("top metric set: %s (%d)", summary.MetricSets[0].Name, summary.MetricSets[0].Count))
	}
	if len(summary.Metrics) > 0 {
		values = append(values, fmt.Sprintf("top metric: %s (%d)", summary.Metrics[0].Name, summary.Metrics[0].Count))
	}
	return values
}
