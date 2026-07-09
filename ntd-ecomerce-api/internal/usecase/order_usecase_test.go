package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
	"ntd-ecomerce-api/internal/usecase"
)

var orderID = uuid.New()

func validCustomer() domain.Customer {
	return domain.Customer{Name: "Ada Lovelace", Email: "ada@example.com"}
}

func confirmedOrder() domain.Order {
	return domain.Order{
		ID:       orderID,
		Status:   domain.OrderStatusConfirmed,
		Customer: validCustomer(),
		Items: []domain.OrderItem{{
			ProductID: productID,
			SKU:       "RS-001",
			Name:      "Running Shoes",
			UnitPrice: decimal.RequireFromString("10.00"),
			Quantity:  3,
		}},
		Payment: domain.ApprovedPayment(),
	}
}

func TestOrder_Checkout(t *testing.T) {
	type expected struct {
		err     error
		code    string
		details map[string]string
	}

	tests := map[string]struct {
		input     domain.CheckoutInput
		mockSetup func(orders *MockOrderRepository)
		expected  expected
		wantTotal string
	}{
		"should create a confirmed order from a cart": {
			input: domain.CheckoutInput{CartID: cartID, Customer: validCustomer()},
			mockSetup: func(orders *MockOrderRepository) {
				orders.On("Checkout", cartID, validCustomer()).Return(confirmedOrder(), nil)
			},
			wantTotal: "30",
		},
		"should reject an invalid customer with a validation error": {
			input:     domain.CheckoutInput{CartID: cartID, Customer: domain.Customer{Name: "", Email: "bad"}},
			mockSetup: func(_ *MockOrderRepository) {},
			expected: expected{
				err:     usecase.ErrInvalidCustomerInput,
				code:    "validation_error",
				details: map[string]string{"name": "required", "email": "invalid"},
			},
		},
		"should return cart_not_found when the cart is missing": {
			input: domain.CheckoutInput{CartID: cartID, Customer: validCustomer()},
			mockSetup: func(orders *MockOrderRepository) {
				orders.On("Checkout", cartID, validCustomer()).
					Return(domain.Order{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			expected: expected{err: domain.ErrNotFound, code: "cart_not_found"},
		},
		"should return cart_empty when the cart has no items": {
			input: domain.CheckoutInput{CartID: cartID, Customer: validCustomer()},
			mockSetup: func(orders *MockOrderRepository) {
				orders.On("Checkout", cartID, validCustomer()).
					Return(domain.Order{}, domain.WrapValidationCode(assert.AnError, "cart_empty", "cart is empty"))
			},
			expected: expected{err: domain.ErrValidation, code: "cart_empty"},
		},
		"should return insufficient_stock when a line exceeds current stock": {
			input: domain.CheckoutInput{CartID: cartID, Customer: validCustomer()},
			mockSetup: func(orders *MockOrderRepository) {
				orders.On("Checkout", cartID, validCustomer()).
					Return(domain.Order{}, domain.WrapConflictDetails(assert.AnError, "insufficient_stock", "insufficient stock",
						map[string]string{productID.String(): "requested=5, available=3"}))
			},
			expected: expected{
				err:     domain.ErrConflict,
				code:    "insufficient_stock",
				details: map[string]string{productID.String(): "requested=5, available=3"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockOrders := &MockOrderRepository{}
			defer mockOrders.AssertExpectations(t)
			tc.mockSetup(mockOrders)
			svc := usecase.NewOrder(mockOrders)

			order, err := svc.Checkout(context.Background(), tc.input)

			assertOrderError(t, err, tc.expected.err, tc.expected.code, tc.expected.details)
			if tc.expected.err == nil {
				assert.Equal(t, tc.wantTotal, order.Total.String())
				assert.Equal(t, domain.OrderStatusConfirmed, order.Status)
			}
		})
	}
}

func TestOrder_Get(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(orders *MockOrderRepository)
		expectErr error
		code      string
	}{
		"should return the order with recomputed totals": {
			mockSetup: func(orders *MockOrderRepository) {
				orders.On("FindByID", orderID).Return(confirmedOrder(), nil)
			},
		},
		"should return order_not_found when the order does not exist": {
			mockSetup: func(orders *MockOrderRepository) {
				orders.On("FindByID", orderID).
					Return(domain.Order{}, domain.WrapNotFound(assert.AnError, "order_not_found", "order not found"))
			},
			expectErr: domain.ErrNotFound,
			code:      "order_not_found",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockOrders := &MockOrderRepository{}
			defer mockOrders.AssertExpectations(t)
			tc.mockSetup(mockOrders)
			svc := usecase.NewOrder(mockOrders)

			order, err := svc.Get(context.Background(), orderID)

			if tc.expectErr != nil {
				assertOrderError(t, err, tc.expectErr, tc.code, nil)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "30", order.Total.String())
			assert.Equal(t, "30", order.Items[0].Subtotal.String())
		})
	}
}

func assertOrderError(t *testing.T, err, wantErr error, wantCode string, wantDetails map[string]string) {
	t.Helper()

	if wantErr == nil {
		assert.NoError(t, err)
		return
	}

	assert.ErrorIs(t, err, wantErr)

	var domainErr *domain.Error
	if assert.ErrorAs(t, err, &domainErr) {
		assert.Equal(t, wantCode, domainErr.Code)
		if wantDetails != nil {
			assert.Equal(t, wantDetails, domainErr.Details)
		}
	}
}
