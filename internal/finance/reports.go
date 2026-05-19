package finance

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"os"
	"regexp"
	"sort"
	"strings"
)

type ReportSummaryOptions struct {
	File string `json:"file"`
}

type ReportSummary struct {
	File             string             `json:"file"`
	ReportType       string             `json:"reportType"`
	Rows             int                `json:"rows"`
	GroupColumn      string             `json:"groupColumn"`
	AmountColumn     string             `json:"amountColumn"`
	CurrencyColumn   string             `json:"currencyColumn,omitempty"`
	TransactionTypes []TransactionTotal `json:"transactionTypes"`
}

type TransactionTotal struct {
	TransactionType string `json:"transactionType"`
	Count           int    `json:"count"`
	Total           string `json:"total"`
	Currency        string `json:"currency,omitempty"`
}

type columnIndexes struct {
	transactionType int
	amount          int
	currency        int
	reportType      string
	groupName       string
	amountName      string
	currencyName    string
}

type reportSchema struct {
	reportType        string
	transactionHeader string
	amountHeader      string
	currencyHeader    string
}

var reportSchemas = []reportSchema{
	{reportType: "earnings", transactionHeader: "transaction type", amountHeader: "amount (merchant currency)", currencyHeader: "merchant currency"},
	{reportType: "estimated-sales", transactionHeader: "financial status", amountHeader: "charged amount", currencyHeader: "currency of sale"},
}

var decimalPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)

func SummarizeReport(options ReportSummaryOptions) (ReportSummary, error) {
	if strings.TrimSpace(options.File) == "" {
		return ReportSummary{}, fmt.Errorf("file is required")
	}
	file, err := os.Open(options.File)
	if err != nil {
		return ReportSummary{}, fmt.Errorf("open finance report %s: %w", options.File, err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return ReportSummary{}, fmt.Errorf("finance report is empty")
		}
		return ReportSummary{}, fmt.Errorf("read finance report %s: %w", options.File, err)
	}
	indexes, err := findColumns(headers)
	if err != nil {
		return ReportSummary{}, err
	}
	totals := map[string]decimalAmount{}
	counts := map[string]int{}
	currencies := map[string]string{}
	rowCount := 0
	for rowIndex := 2; ; rowIndex++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ReportSummary{}, fmt.Errorf("read finance report row %d: %w", rowIndex, err)
		}
		if isEmptyRecord(record) {
			continue
		}
		if len(record) <= indexes.transactionType || len(record) <= indexes.amount {
			return ReportSummary{}, fmt.Errorf("finance report row %d is missing required columns", rowIndex)
		}
		transactionType := strings.TrimSpace(record[indexes.transactionType])
		if transactionType == "" {
			return ReportSummary{}, fmt.Errorf("finance report row %d transaction type is required", rowIndex)
		}
		amount, err := parseAmount(record[indexes.amount])
		if err != nil {
			return ReportSummary{}, fmt.Errorf("finance report row %d amount: %w", rowIndex, err)
		}
		key := transactionType
		currency := ""
		if indexes.currency >= 0 && len(record) > indexes.currency {
			currency = strings.TrimSpace(record[indexes.currency])
			if currency != "" {
				key = transactionType + "\x00" + currency
			}
		}
		totals[key] = totals[key].Add(amount)
		counts[key]++
		currencies[key] = currency
		rowCount++
	}
	transactionTypes := make([]TransactionTotal, 0, len(totals))
	for key, total := range totals {
		transactionType := strings.Split(key, "\x00")[0]
		transactionTypes = append(transactionTypes, TransactionTotal{
			TransactionType: transactionType,
			Count:           counts[key],
			Total:           total.String(),
			Currency:        currencies[key],
		})
	}
	sort.Slice(transactionTypes, func(i, j int) bool {
		if transactionTypes[i].TransactionType == transactionTypes[j].TransactionType {
			return transactionTypes[i].Currency < transactionTypes[j].Currency
		}
		return transactionTypes[i].TransactionType < transactionTypes[j].TransactionType
	})
	return ReportSummary{
		File:             options.File,
		ReportType:       indexes.reportType,
		Rows:             rowCount,
		GroupColumn:      indexes.groupName,
		AmountColumn:     indexes.amountName,
		CurrencyColumn:   indexes.currencyName,
		TransactionTypes: transactionTypes,
	}, nil
}

func findColumns(headers []string) (columnIndexes, error) {
	normalized := map[string]int{}
	original := map[int]string{}
	for index, header := range headers {
		cleanHeader := normalizeHeader(header)
		normalized[cleanHeader] = index
		original[index] = strings.TrimPrefix(strings.TrimSpace(header), "\ufeff")
	}
	for _, schema := range reportSchemas {
		transactionType, ok := normalized[schema.transactionHeader]
		if !ok {
			continue
		}
		amount, ok := normalized[schema.amountHeader]
		if !ok {
			continue
		}
		currency, ok := normalized[schema.currencyHeader]
		if !ok {
			continue
		}
		if amount == currency {
			return columnIndexes{}, fmt.Errorf("finance report amount and currency columns must be distinct")
		}
		return columnIndexes{
			transactionType: transactionType,
			amount:          amount,
			currency:        currency,
			reportType:      schema.reportType,
			groupName:       original[transactionType],
			amountName:      original[amount],
			currencyName:    original[currency],
		}, nil
	}
	return columnIndexes{}, fmt.Errorf("finance report must match earnings or estimated sales CSV headers")
}

func normalizeHeader(header string) string {
	header = strings.TrimPrefix(strings.TrimSpace(header), "\ufeff")
	return strings.ToLower(strings.Join(strings.Fields(header), " "))
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
	if !decimalPattern.MatchString(cleanValue) {
		return decimalAmount{}, fmt.Errorf("invalid decimal %q", value)
	}
	sign := 1
	if strings.HasPrefix(cleanValue, "-") {
		sign = -1
		cleanValue = strings.TrimPrefix(cleanValue, "-")
	}
	scale := 0
	parts := strings.Split(cleanValue, ".")
	digits := parts[0]
	if len(parts) == 2 {
		scale = len(parts[1])
		digits += parts[1]
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
	if text == "" {
		return "0"
	}
	if text == "0" {
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

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
