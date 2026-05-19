package analytics

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"
)

type StatsSummaryOptions struct {
	File string `json:"file"`
}

type StatsSummary struct {
	File        string        `json:"file"`
	Rows        int           `json:"rows"`
	PackageName string        `json:"packageName,omitempty"`
	StartDate   string        `json:"startDate,omitempty"`
	EndDate     string        `json:"endDate,omitempty"`
	Metrics     []MetricTotal `json:"metrics"`
}

type MetricTotal struct {
	Name  string `json:"name"`
	Total string `json:"total"`
}

var dimensionHeaders = map[string]bool{
	"date":           true,
	"package name":   true,
	"country/region": true,
	"traffic source": true,
	"search term":    true,
	"utm campaign":   true,
	"utm source":     true,
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
	totals := map[string]decimalAmount{}
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
			amount, err := parseAmount(value)
			if err != nil {
				return StatsSummary{}, fmt.Errorf("analytics report row %d metric %s: %w", rowIndex, metric.name, err)
			}
			totals[metric.name] = totals[metric.name].Add(amount)
		}
		rowCount++
	}
	metrics := make([]MetricTotal, 0, len(totals))
	for name, total := range totals {
		metrics = append(metrics, MetricTotal{Name: name, Total: total.String()})
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
	name  string
	index int
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
				indexes.metrics = append(indexes.metrics, metricIndex{name: strings.TrimPrefix(strings.TrimSpace(header), "\ufeff"), index: index})
			}
		}
	}
	return indexes
}

type decimalAmount struct {
	value *big.Int
	scale int
}

func parseAmount(value string) (decimalAmount, error) {
	cleanValue := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if cleanValue == "" {
		return decimalAmount{}, fmt.Errorf("value is required")
	}
	if strings.Contains(cleanValue, "%") {
		cleanValue = strings.TrimSuffix(cleanValue, "%")
	}
	sign := 1
	if strings.HasPrefix(cleanValue, "-") {
		sign = -1
		cleanValue = strings.TrimPrefix(cleanValue, "-")
	}
	if cleanValue == "" {
		return decimalAmount{}, fmt.Errorf("invalid decimal %q", value)
	}
	parts := strings.Split(cleanValue, ".")
	if len(parts) > 2 {
		return decimalAmount{}, fmt.Errorf("invalid decimal %q", value)
	}
	digits := parts[0]
	scale := 0
	if len(parts) == 2 {
		scale = len(parts[1])
		digits += parts[1]
	}
	if strings.Trim(digits, "0123456789") != "" {
		return decimalAmount{}, fmt.Errorf("invalid decimal %q", value)
	}
	amount := new(big.Int)
	amount.SetString(digits, 10)
	if sign < 0 {
		amount.Neg(amount)
	}
	return decimalAmount{value: amount, scale: scale}, nil
}

func (a decimalAmount) Add(b decimalAmount) decimalAmount {
	if a.value == nil {
		return b
	}
	if b.value == nil {
		return a
	}
	scale := max(a.scale, b.scale)
	left := scaleDecimal(a.value, scale-a.scale)
	right := scaleDecimal(b.value, scale-b.scale)
	return decimalAmount{value: left.Add(left, right), scale: scale}
}

func (a decimalAmount) String() string {
	if a.value == nil {
		return "0"
	}
	value := new(big.Int).Set(a.value)
	sign := ""
	if value.Sign() < 0 {
		sign = "-"
		value.Abs(value)
	}
	text := value.String()
	if a.scale > 0 {
		for len(text) <= a.scale {
			text = "0" + text
		}
		split := len(text) - a.scale
		text = text[:split] + "." + text[split:]
	}
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" || text == "0" {
		return "0"
	}
	return sign + text
}

func scaleDecimal(value *big.Int, places int) *big.Int {
	scaled := new(big.Int).Set(value)
	if places <= 0 {
		return scaled
	}
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return scaled.Mul(scaled, multiplier)
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
