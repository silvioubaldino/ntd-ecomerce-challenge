package domain_test

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func validInput() domain.ProductInput {
	return domain.ProductInput{
		Name:        "Running Shoes",
		SKU:         "RS-001",
		Description: "Lightweight running shoes",
		Category:    "Footwear",
		Price:       decimal.NewFromFloat(89.99),
		Stock:       150,
		WeightKg:    decimal.NewFromFloat(0.35),
	}
}

func TestProductInput_Validate(t *testing.T) {
	type (
		input struct {
			mutate func(domain.ProductInput) domain.ProductInput
		}
		expected struct {
			problems map[string]string
		}
	)

	tests := map[string]struct {
		// input
		input input
		// expected
		expected expected
	}{
		"should return no problems when input is valid": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput { return in }},
			expected: expected{
				problems: map[string]string{},
			},
		},
		"should flag name as required when empty": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput { in.Name = ""; return in }},
			expected: expected{
				problems: map[string]string{"name": "required"},
			},
		},
		"should flag name as too_long when over 255 chars": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput {
				in.Name = strings.Repeat("a", 256)
				return in
			}},
			expected: expected{
				problems: map[string]string{"name": "too_long"},
			},
		},
		"should flag sku as required when empty": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput { in.SKU = ""; return in }},
			expected: expected{
				problems: map[string]string{"sku": "required"},
			},
		},
		"should flag sku as too_long when over 64 chars": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput {
				in.SKU = strings.Repeat("a", 65)
				return in
			}},
			expected: expected{
				problems: map[string]string{"sku": "too_long"},
			},
		},
		"should flag category as required when empty": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput { in.Category = ""; return in }},
			expected: expected{
				problems: map[string]string{"category": "required"},
			},
		},
		"should flag price as must_be_non_negative_decimal when negative": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput {
				in.Price = decimal.NewFromFloat(-1)
				return in
			}},
			expected: expected{
				problems: map[string]string{"price": "must_be_non_negative_decimal"},
			},
		},
		"should flag stock as must_be_non_negative_integer when negative": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput { in.Stock = -1; return in }},
			expected: expected{
				problems: map[string]string{"stock": "must_be_non_negative_integer"},
			},
		},
		"should flag weight_kg as must_be_non_negative_decimal when negative": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput {
				in.WeightKg = decimal.NewFromFloat(-1)
				return in
			}},
			expected: expected{
				problems: map[string]string{"weight_kg": "must_be_non_negative_decimal"},
			},
		},
		"should flag multiple fields at once": {
			input: input{mutate: func(in domain.ProductInput) domain.ProductInput {
				in.Name = ""
				in.Stock = -5
				return in
			}},
			expected: expected{
				problems: map[string]string{"name": "required", "stock": "must_be_non_negative_integer"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			in := tc.input.mutate(validInput())

			// Act
			problems := in.Validate()

			// Assert
			assert.Equal(t, tc.expected.problems, problems)
		})
	}
}

func TestProductInput_ToProduct(t *testing.T) {
	type expected struct {
		output domain.Product
	}

	tests := map[string]struct {
		// expected
		expected expected
	}{
		"should copy every field except id and timestamps": {
			expected: expected{
				output: domain.Product{
					Name:        "Running Shoes",
					SKU:         "RS-001",
					Description: "Lightweight running shoes",
					Category:    "Footwear",
					Price:       decimal.NewFromFloat(89.99),
					Stock:       150,
					WeightKg:    decimal.NewFromFloat(0.35),
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			in := validInput()

			// Act
			product := in.ToProduct()

			// Assert
			assert.Equal(t, tc.expected.output, product)
		})
	}
}
