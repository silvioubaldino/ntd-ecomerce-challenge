package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func TestCustomer_Validate(t *testing.T) {
	tests := map[string]struct {
		customer domain.Customer
		expected map[string]string
	}{
		"should accept a valid name and email": {
			customer: domain.Customer{Name: "Ada Lovelace", Email: "ada@example.com"},
			expected: nil,
		},
		"should reject a blank name": {
			customer: domain.Customer{Name: "", Email: "ada@example.com"},
			expected: map[string]string{"name": "required"},
		},
		"should reject a missing email": {
			customer: domain.Customer{Name: "Ada", Email: ""},
			expected: map[string]string{"email": "required"},
		},
		"should reject a malformed email": {
			customer: domain.Customer{Name: "Ada", Email: "not-an-email"},
			expected: map[string]string{"email": "invalid"},
		},
		"should report every offending field": {
			customer: domain.Customer{Name: "", Email: "bad"},
			expected: map[string]string{"name": "required", "email": "invalid"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			problems := tc.customer.Validate()

			assert.Equal(t, tc.expected, problems)
		})
	}
}

func TestApprovedPayment(t *testing.T) {
	t.Run("should be a simulated, approved payment", func(t *testing.T) {
		payment := domain.ApprovedPayment()

		assert.Equal(t, "simulated", payment.Method)
		assert.Equal(t, "approved", payment.Status)
	})
}

func TestOrder_Recalculate(t *testing.T) {
	type expected struct {
		subtotals []string
		total     string
	}

	tests := map[string]struct {
		items    []domain.OrderItem
		expected expected
	}{
		"should sum item subtotals into the total": {
			items: []domain.OrderItem{
				{UnitPrice: decimal.RequireFromString("10.00"), Quantity: 3},
				{UnitPrice: decimal.RequireFromString("5.50"), Quantity: 2},
			},
			expected: expected{subtotals: []string{"30", "11"}, total: "41"},
		},
		"should produce a zero total for no items": {
			items:    []domain.OrderItem{},
			expected: expected{subtotals: []string{}, total: "0"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			order := domain.Order{Items: tc.items}

			order.Recalculate()

			assert.Equal(t, tc.expected.total, order.Total.String())
			for i, subtotal := range tc.expected.subtotals {
				assert.Equal(t, subtotal, order.Items[i].Subtotal.String())
			}
		})
	}
}
