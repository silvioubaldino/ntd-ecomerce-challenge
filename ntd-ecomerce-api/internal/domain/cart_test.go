package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func TestAddItemInput_Validate(t *testing.T) {
	tests := map[string]struct {
		input    domain.AddItemInput
		expected map[string]string
	}{
		"should accept a positive quantity": {
			input:    domain.AddItemInput{Quantity: 2},
			expected: nil,
		},
		"should reject zero quantity": {
			input:    domain.AddItemInput{Quantity: 0},
			expected: map[string]string{"quantity": "must_be_positive_integer"},
		},
		"should reject negative quantity": {
			input:    domain.AddItemInput{Quantity: -1},
			expected: map[string]string{"quantity": "must_be_positive_integer"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			problems := tc.input.Validate()

			assert.Equal(t, tc.expected, problems)
		})
	}
}

func TestSetItemInput_Validate(t *testing.T) {
	tests := map[string]struct {
		input    domain.SetItemInput
		expected map[string]string
	}{
		"should accept a positive quantity": {
			input:    domain.SetItemInput{Quantity: 1},
			expected: nil,
		},
		"should reject zero quantity": {
			input:    domain.SetItemInput{Quantity: 0},
			expected: map[string]string{"quantity": "must_be_positive_integer"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			problems := tc.input.Validate()

			assert.Equal(t, tc.expected, problems)
		})
	}
}

func TestCart_Recalculate(t *testing.T) {
	type expected struct {
		subtotals []string
		total     string
	}

	tests := map[string]struct {
		items    []domain.CartItem
		expected expected
	}{
		"should sum item subtotals into the total": {
			items: []domain.CartItem{
				{UnitPrice: decimal.RequireFromString("10.00"), Quantity: 3},
				{UnitPrice: decimal.RequireFromString("5.50"), Quantity: 2},
			},
			expected: expected{subtotals: []string{"30", "11"}, total: "41"},
		},
		"should produce a zero total for an empty cart": {
			items:    []domain.CartItem{},
			expected: expected{subtotals: []string{}, total: "0"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cart := domain.Cart{Items: tc.items}

			cart.Recalculate()

			assert.Equal(t, tc.expected.total, cart.Total.String())
			for i, subtotal := range tc.expected.subtotals {
				assert.Equal(t, subtotal, cart.Items[i].Subtotal.String())
			}
		})
	}
}
