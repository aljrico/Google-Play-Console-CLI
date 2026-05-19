package finance

import (
	"encoding/csv"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
)

type ReportSummaryOptions struct {
	File string `json:"file"`
}

type ReportSummary struct {
	File             string             `json:"file"`
	Rows             int                `json:"rows"`
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
	amountName      string
	currencyName    string
}

var (
	transactionTypeHeaders = []string{"transaction type", "transaction_type"}
	amountHeaders          = []string{"merchant currency", "merchant amount", "merchant currency amount", "amount (merchant currency)"}
	currencyHeaders        = []string{"merchant currency code", "currency of merchant", "merchant currency currency", "currency"}
)

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
	records, err := reader.ReadAll()
	if err != nil {
		return ReportSummary{}, fmt.Errorf("read finance report %s: %w", options.File, err)
	}
	if len(records) == 0 {
		return ReportSummary{}, fmt.Errorf("finance report is empty")
	}
	indexes, err := findColumns(records[0])
	if err != nil {
		return ReportSummary{}, err
	}
	totals := map[string]*big.Rat{}
	counts := map[string]int{}
	currencies := map[string]string{}
	rowCount := 0
	for rowIndex, record := range records[1:] {
		if isEmptyRecord(record) {
			continue
		}
		if len(record) <= indexes.transactionType || len(record) <= indexes.amount {
			return ReportSummary{}, fmt.Errorf("finance report row %d is missing required columns", rowIndex+2)
		}
		transactionType := strings.TrimSpace(record[indexes.transactionType])
		if transactionType == "" {
			return ReportSummary{}, fmt.Errorf("finance report row %d transaction type is required", rowIndex+2)
		}
		amount, err := parseAmount(record[indexes.amount])
		if err != nil {
			return ReportSummary{}, fmt.Errorf("finance report row %d amount: %w", rowIndex+2, err)
		}
		key := transactionType
		currency := ""
		if indexes.currency >= 0 && len(record) > indexes.currency {
			currency = strings.TrimSpace(record[indexes.currency])
			if currency != "" {
				key = transactionType + "\x00" + currency
			}
		}
		if totals[key] == nil {
			totals[key] = new(big.Rat)
		}
		totals[key].Add(totals[key], amount)
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
			Total:           decimalString(total),
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
		Rows:             rowCount,
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
	transactionType, ok := firstHeaderIndex(normalized, transactionTypeHeaders)
	if !ok {
		return columnIndexes{}, fmt.Errorf("finance report requires a Transaction Type column")
	}
	amount, ok := firstHeaderIndex(normalized, amountHeaders)
	if !ok {
		return columnIndexes{}, fmt.Errorf("finance report requires a Merchant Currency amount column")
	}
	currency, hasCurrency := firstHeaderIndex(normalized, currencyHeaders)
	if !hasCurrency {
		currency = -1
	}
	indexes := columnIndexes{
		transactionType: transactionType,
		amount:          amount,
		currency:        currency,
		amountName:      original[amount],
	}
	if hasCurrency {
		indexes.currencyName = original[currency]
	}
	return indexes, nil
}

func firstHeaderIndex(headers map[string]int, candidates []string) (int, bool) {
	for _, candidate := range candidates {
		index, ok := headers[candidate]
		if ok {
			return index, true
		}
	}
	return 0, false
}

func normalizeHeader(header string) string {
	header = strings.TrimPrefix(strings.TrimSpace(header), "\ufeff")
	return strings.ToLower(strings.Join(strings.Fields(header), " "))
}

func parseAmount(value string) (*big.Rat, error) {
	cleanValue := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if cleanValue == "" {
		return nil, fmt.Errorf("value is required")
	}
	amount, ok := new(big.Rat).SetString(cleanValue)
	if !ok {
		return nil, fmt.Errorf("invalid decimal %q", value)
	}
	return amount, nil
}

func decimalString(value *big.Rat) string {
	if value == nil {
		return "0"
	}
	text := value.FloatString(6)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" || text == "-0" {
		return "0"
	}
	return text
}

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
