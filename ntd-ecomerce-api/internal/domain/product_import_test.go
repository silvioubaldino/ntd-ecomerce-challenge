package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func validRecord() []string {
	return []string{"Running Shoes", "RS-001", "Lightweight running shoes", "Footwear", "89.99", "150", "0.35"}
}

func TestValidateCSVHeader(t *testing.T) {
	type (
		input struct {
			header []string
		}
		expected struct {
			err error
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should accept the exact expected header": {
			input:    input{header: []string{"name", "sku", "description", "category", "price", "stock", "weight_kg"}},
			expected: expected{err: nil},
		},
		"should reject a renamed column": {
			input:    input{header: []string{"title", "sku", "description", "category", "price", "stock", "weight_kg"}},
			expected: expected{err: domain.ErrInvalidCSVHeader},
		},
		"should reject a reordered header": {
			input:    input{header: []string{"sku", "name", "description", "category", "price", "stock", "weight_kg"}},
			expected: expected{err: domain.ErrInvalidCSVHeader},
		},
		"should reject a short header": {
			input:    input{header: []string{"name", "sku"}},
			expected: expected{err: domain.ErrInvalidCSVHeader},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange (input already set in the table)

			// Act
			err := domain.ValidateCSVHeader(tc.input.header)

			// Assert
			assert.ErrorIs(t, err, tc.expected.err)
		})
	}
}

func TestParseProductCSVRecord_Problems(t *testing.T) {
	type (
		input struct {
			record []string
		}
		expected struct {
			problems map[string]string
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should return no problems for a fully valid row": {
			input:    input{record: validRecord()},
			expected: expected{problems: map[string]string{}},
		},
		"should flag price as must_be_non_negative_decimal when prefixed with a currency sign": {
			input:    input{record: []string{"Wireless Mouse", "WM-042", "desc", "Electronics", "$29.99", "75", "0.12"}},
			expected: expected{problems: map[string]string{"price": "must_be_non_negative_decimal"}},
		},
		"should flag price as must_be_non_negative_decimal when non-numeric": {
			input:    input{record: []string{"Yoga Mat", "YM-015", "desc", "Sports", "free", "200", "1.2"}},
			expected: expected{problems: map[string]string{"price": "must_be_non_negative_decimal"}},
		},
		"should flag stock as must_be_non_negative_integer when negative": {
			input:    input{record: []string{"Desk Lamp", "DL-007", "desc", "Home & Office", "45.50", "-5", "2.1"}},
			expected: expected{problems: map[string]string{"stock": "must_be_non_negative_integer"}},
		},
		"should flag weight_kg as must_be_non_negative_decimal when blank": {
			input:    input{record: []string{"Gaming Keyboard", "GK-088", "desc", "Electronics", "129.99", "45", ""}},
			expected: expected{problems: map[string]string{"weight_kg": "must_be_non_negative_decimal"}},
		},
		"should flag name as required when blank": {
			input:    input{record: []string{"", "HD-099", "desc", "Electronics", "149.99", "30", "0.25"}},
			expected: expected{problems: map[string]string{"name": "required"}},
		},
		"should flag name as required when whitespace-only": {
			input:    input{record: []string{"     ", "WS-001", "desc", "Misc", "5.00", "10", "0.1"}},
			expected: expected{problems: map[string]string{"name": "required"}},
		},
		"should flag category as required when blank": {
			input:    input{record: []string{"Gift Card", "GC-025", "desc", "", "25.00", "99999", "0"}},
			expected: expected{problems: map[string]string{"category": "required"}},
		},
		"should return no problems for zero-valued price stock and weight": {
			input:    input{record: []string{"Mystery Box", "MB-001", "desc", "Gifts", "0.00", "0", "0"}},
			expected: expected{problems: map[string]string{}},
		},
		"should flag name as unsafe_content when it contains a script tag": {
			input:    input{record: []string{"<script>alert('xss')</script>", "XS-001", "desc", "Electronics", "19.99", "100", "0.1"}},
			expected: expected{problems: map[string]string{"name": "unsafe_content"}},
		},
		"should not flag sql-injection-looking text as unsafe": {
			input:    input{record: []string{"Robert'); DROP TABLE products;--", "SQL-001", "desc", "Games", "9.99", "50", "0.5"}},
			expected: expected{problems: map[string]string{}},
		},
		"should return no problems for a quoted comma in the name": {
			input:    input{record: []string{"Comma, In Product Name", "CI-001", "desc", "Accessories", "14.99", "300", "0.1"}},
			expected: expected{problems: map[string]string{}},
		},
		"should flag multiple problems on the same row": {
			input: input{record: []string{"", "BAD-001", "desc", "Misc", "free", "-5", ""}},
			expected: expected{problems: map[string]string{
				"name":      "required",
				"price":     "must_be_non_negative_decimal",
				"stock":     "must_be_non_negative_integer",
				"weight_kg": "must_be_non_negative_decimal",
			}},
		},
		"should pad a short row and flag every missing required field": {
			input: input{record: []string{"", "", "", "", "", "", ""}},
			expected: expected{problems: map[string]string{
				"name":      "required",
				"sku":       "required",
				"category":  "required",
				"price":     "must_be_non_negative_decimal",
				"stock":     "must_be_non_negative_integer",
				"weight_kg": "must_be_non_negative_decimal",
			}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange (input already set in the table)

			// Act
			_, problems := domain.ParseProductCSVRecord(tc.input.record)

			// Assert
			assert.Equal(t, tc.expected.problems, problems)
		})
	}
}

func TestParseProductCSVRecord_Input(t *testing.T) {
	type (
		input struct {
			record []string
		}
		expected struct {
			output domain.ProductInput
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should trim fields and parse decimals and int for a valid row": {
			input: input{record: validRecord()},
			expected: expected{output: domain.ProductInput{
				Name:        "Running Shoes",
				SKU:         "RS-001",
				Description: "Lightweight running shoes",
				Category:    "Footwear",
				Price:       decimal.RequireFromString("89.99"),
				Stock:       150,
				WeightKg:    decimal.RequireFromString("0.35"),
			}},
		},
		"should preserve a quoted comma in the name": {
			input: input{record: []string{"Comma, In Product Name", "CI-001", "desc", "Accessories", "14.99", "300", "0.1"}},
			expected: expected{output: domain.ProductInput{
				Name:        "Comma, In Product Name",
				SKU:         "CI-001",
				Description: "desc",
				Category:    "Accessories",
				Price:       decimal.RequireFromString("14.99"),
				Stock:       300,
				WeightKg:    decimal.RequireFromString("0.1"),
			}},
		},
		"should keep zero-value price and stock when parsing fails": {
			input: input{record: []string{"Yoga Mat", "YM-015", "desc", "Sports", "free", "-5", "1.2"}},
			expected: expected{output: domain.ProductInput{
				Name:        "Yoga Mat",
				SKU:         "YM-015",
				Description: "desc",
				Category:    "Sports",
				Price:       decimal.Decimal{},
				Stock:       -5,
				WeightKg:    decimal.RequireFromString("1.2"),
			}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange (input already set in the table)

			// Act
			output, _ := domain.ParseProductCSVRecord(tc.input.record)

			// Assert
			assert.Equal(t, tc.expected.output, output)
		})
	}
}
