package domain

import (
	"errors"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const MaxImportFileBytes int64 = 5 << 20 // 5 MB (AYD-002)

var (
	ErrInvalidCSVHeader = errors.New("invalid csv header")

	csvImportColumns = []string{"name", "sku", "description", "category", "price", "stock", "weight_kg"}
)

type ImportSummary struct {
	Total    int `json:"total"`
	Imported int `json:"imported"`
	Rejected int `json:"rejected"`
}

type RejectedRow struct {
	Row    int               `json:"row"`
	SKU    string            `json:"sku"`
	Errors map[string]string `json:"errors"`
}

type ImportReport struct {
	Summary  ImportSummary `json:"summary"`
	Rejected []RejectedRow `json:"rejected"`
}

func ValidateCSVHeader(header []string) error {
	if len(header) != len(csvImportColumns) {
		1
		return ErrInvalidCSVHeader
	}
	for i, column := range csvImportColumns {
		if strings.TrimSpace(header[i]) != column {
			return ErrInvalidCSVHeader
		}
	}
	return nil
}

// ParseProductCSVRecord turns one CSV data row into a ProductInput, returning a
// field->reason map of every problem found (per RN-02) instead of a single error,
// so a row with several bad fields is reported in one pass.
func ParseProductCSVRecord(record []string) (ProductInput, map[string]string) {
	record = padRecord(record, len(csvImportColumns))
	problems := map[string]string{}

	name := strings.TrimSpace(record[0])
	sku := strings.TrimSpace(record[1])
	description := strings.TrimSpace(record[2])
	category := strings.TrimSpace(record[3])

	price, err := decimal.NewFromString(strings.TrimSpace(record[4]))
	if err != nil {
		problems["price"] = "must_be_non_negative_decimal"
	}

	stock, err := strconv.Atoi(strings.TrimSpace(record[5]))
	if err != nil {
		problems["stock"] = "must_be_non_negative_integer"
	}

	weightKg, err := decimal.NewFromString(strings.TrimSpace(record[6]))
	if err != nil {
		problems["weight_kg"] = "must_be_non_negative_decimal"
	}

	if field := firstUnsafeField(name, description, category); field != "" {
		problems[field] = "unsafe_content"
	}

	input := ProductInput{
		Name:        name,
		SKU:         sku,
		Description: description,
		Category:    category,
		Price:       price,
		Stock:       stock,
		WeightKg:    weightKg,
	}

	for field, reason := range input.Validate() {
		if _, exists := problems[field]; !exists {
			problems[field] = reason
		}
	}

	return input, problems
}

func padRecord(record []string, size int) []string {
	if len(record) == size {
		return record
	}
	padded := make([]string, size)
	copy(padded, record)
	return padded
}

func firstUnsafeField(name, description, category string) string {
	fields := []struct {
		key   string
		value string
	}{
		{"name", name},
		{"description", description},
		{"category", category},
	}
	for _, f := range fields {
		if isUnsafeText(f.value) {
			return f.key
		}
	}
	return ""
}

// isUnsafeText flags an HTML/script tag or a leading CSV formula-injection
// character (RN-02 "unsafe content"). Plain SQL-injection-looking text is not
// flagged here — it's neutralized by parameterized queries and stored as-is.
func isUnsafeText(s string) bool {
	if strings.ContainsAny(s, "<>") {
		return true
	}
	if s == "" {
		return false
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t':
		return true
	default:
		return false
	}
}
