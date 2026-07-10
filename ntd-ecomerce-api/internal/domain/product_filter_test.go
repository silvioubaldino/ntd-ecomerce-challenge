package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func decimalPtr(s string) *decimal.Decimal {
	v := decimal.RequireFromString(s)
	return &v
}

func TestParseProductSort(t *testing.T) {
	type (
		input struct {
			raw string
		}
		expected struct {
			sort domain.ProductSort
			ok   bool
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should accept price_asc": {
			input:    input{raw: "price_asc"},
			expected: expected{sort: domain.ProductSortPriceAsc, ok: true},
		},
		"should accept price_desc": {
			input:    input{raw: "price_desc"},
			expected: expected{sort: domain.ProductSortPriceDesc, ok: true},
		},
		"should accept name_asc": {
			input:    input{raw: "name_asc"},
			expected: expected{sort: domain.ProductSortNameAsc, ok: true},
		},
		"should accept name_desc": {
			input:    input{raw: "name_desc"},
			expected: expected{sort: domain.ProductSortNameDesc, ok: true},
		},
		"should accept newest": {
			input:    input{raw: "newest"},
			expected: expected{sort: domain.ProductSortNewest, ok: true},
		},
		"should reject an unknown value": {
			input:    input{raw: "cheapest"},
			expected: expected{sort: "", ok: false},
		},
		"should reject a blank value": {
			input:    input{raw: ""},
			expected: expected{sort: "", ok: false},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sort, ok := domain.ParseProductSort(tc.input.raw)

			assert.Equal(t, tc.expected.sort, sort)
			assert.Equal(t, tc.expected.ok, ok)
		})
	}
}

func TestProductFilter_Validate(t *testing.T) {
	type (
		input struct {
			filter domain.ProductFilter
		}
		expected struct {
			problems map[string]string
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should accept an empty filter": {
			input:    input{filter: domain.ProductFilter{}},
			expected: expected{problems: map[string]string{}},
		},
		"should accept a fully populated valid filter": {
			input: input{filter: domain.ProductFilter{
				Query:    "shirt",
				Category: "Apparel",
				PriceMin: decimalPtr("10.00"),
				PriceMax: decimalPtr("25.50"),
				Sort:     domain.ProductSortPriceAsc,
			}},
			expected: expected{problems: map[string]string{}},
		},
		"should reject a negative price_min": {
			input:    input{filter: domain.ProductFilter{PriceMin: decimalPtr("-1")}},
			expected: expected{problems: map[string]string{"price_min": "must_be_non_negative_decimal"}},
		},
		"should reject a negative price_max": {
			input:    input{filter: domain.ProductFilter{PriceMax: decimalPtr("-1")}},
			expected: expected{problems: map[string]string{"price_max": "must_be_non_negative_decimal"}},
		},
		"should reject price_min greater than price_max": {
			input: input{filter: domain.ProductFilter{
				PriceMin: decimalPtr("30"),
				PriceMax: decimalPtr("10"),
			}},
			expected: expected{problems: map[string]string{"price_min": "must_not_exceed_price_max"}},
		},
		"should accept price_min equal to price_max": {
			input: input{filter: domain.ProductFilter{
				PriceMin: decimalPtr("10"),
				PriceMax: decimalPtr("10"),
			}},
			expected: expected{problems: map[string]string{}},
		},
		"should reject an unknown sort value": {
			input:    input{filter: domain.ProductFilter{Sort: "cheapest"}},
			expected: expected{problems: map[string]string{"sort": "invalid_sort"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			problems := tc.input.filter.Validate()

			assert.Equal(t, tc.expected.problems, problems)
		})
	}
}
