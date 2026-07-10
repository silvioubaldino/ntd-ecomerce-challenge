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

var (
	cartID    = uuid.New()
	productID = uuid.New()
)

func productWithStock(stock int, price string) domain.Product {
	return domain.Product{
		ID:    productID,
		Name:  "Running Shoes",
		SKU:   "RS-001",
		Price: decimal.RequireFromString(price),
		Stock: stock,
	}
}

func cartWithItem(quantity int) domain.Cart {
	return domain.Cart{
		ID:    cartID,
		Items: []domain.CartItem{{ProductID: productID, Quantity: quantity}},
	}
}

func emptyCart() domain.Cart {
	return domain.Cart{ID: cartID, Items: []domain.CartItem{}}
}

func TestCart_Create(t *testing.T) {
	t.Run("should create an empty cart with zero total", func(t *testing.T) {
		mockCarts := &MockCartRepository{}
		mockProducts := &MockProductReader{}
		defer mockCarts.AssertExpectations(t)
		mockCarts.On("Create").Return(emptyCart(), nil)
		svc := usecase.NewCart(mockCarts, mockProducts)

		cart, err := svc.Create(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, "0", cart.Total.String())
		assert.Empty(t, cart.Items)
	})
}

func TestCart_Get(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(carts *MockCartRepository, products *MockProductReader)
		expectErr error
		expectSub string
	}{
		"should return the enriched cart with subtotals and total": {
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(cartWithItem(3), nil)
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil)
			},
			expectSub: "30",
		},
		"should return not found when the cart does not exist": {
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			expectErr: domain.ErrNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockCarts := &MockCartRepository{}
			mockProducts := &MockProductReader{}
			defer mockCarts.AssertExpectations(t)
			defer mockProducts.AssertExpectations(t)
			tc.mockSetup(mockCarts, mockProducts)
			svc := usecase.NewCart(mockCarts, mockProducts)

			cart, err := svc.Get(context.Background(), cartID)

			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectSub, cart.Items[0].Subtotal.String())
			assert.Equal(t, tc.expectSub, cart.Total.String())
			assert.Equal(t, "RS-001", cart.Items[0].SKU)
		})
	}
}

func TestCart_AddItem(t *testing.T) {
	type expected struct {
		err     error
		code    string
		details map[string]string
	}

	tests := map[string]struct {
		input     domain.AddItemInput
		mockSetup func(carts *MockCartRepository, products *MockProductReader)
		expected  expected
	}{
		"should add a new line when within stock": {
			input: domain.AddItemInput{ProductID: productID, Quantity: 2},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(emptyCart(), nil).Once()
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil).Once()
				carts.On("SaveItem", cartID, productID, 2).Return(nil)
				carts.On("FindByID", cartID).Return(cartWithItem(2), nil).Once()
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil).Once()
			},
		},
		"should increment the resulting quantity of an existing line": {
			input: domain.AddItemInput{ProductID: productID, Quantity: 3},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(cartWithItem(2), nil).Once()
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil).Once()
				carts.On("SaveItem", cartID, productID, 5).Return(nil)
				carts.On("FindByID", cartID).Return(cartWithItem(5), nil).Once()
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil).Once()
			},
		},
		"should reject a quantity below 1 with a validation error": {
			input:     domain.AddItemInput{ProductID: productID, Quantity: 0},
			mockSetup: func(_ *MockCartRepository, _ *MockProductReader) {},
			expected:  expected{err: usecase.ErrInvalidCartInput, code: "validation_error"},
		},
		"should return cart_not_found when the cart is missing": {
			input: domain.AddItemInput{ProductID: productID, Quantity: 1},
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			expected: expected{err: domain.ErrNotFound, code: "cart_not_found"},
		},
		"should return product_not_found when the product is missing": {
			input: domain.AddItemInput{ProductID: productID, Quantity: 1},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(emptyCart(), nil)
				products.On("FindByID", productID).
					Return(domain.Product{}, domain.WrapNotFound(assert.AnError, "product_not_found", "product not found"))
			},
			expected: expected{err: domain.ErrNotFound, code: "product_not_found"},
		},
		"should reject when the requested quantity exceeds stock": {
			input: domain.AddItemInput{ProductID: productID, Quantity: 5},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(emptyCart(), nil)
				products.On("FindByID", productID).Return(productWithStock(3, "10.00"), nil)
			},
			expected: expected{
				err:     usecase.ErrInsufficientStock,
				code:    "insufficient_stock",
				details: map[string]string{"product_id": productID.String(), "requested": "5", "available": "3"},
			},
		},
		"should reject when the incremented quantity exceeds stock": {
			input: domain.AddItemInput{ProductID: productID, Quantity: 2},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(cartWithItem(2), nil)
				products.On("FindByID", productID).Return(productWithStock(3, "10.00"), nil)
			},
			expected: expected{
				err:     usecase.ErrInsufficientStock,
				code:    "insufficient_stock",
				details: map[string]string{"product_id": productID.String(), "requested": "4", "available": "3"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockCarts := &MockCartRepository{}
			mockProducts := &MockProductReader{}
			defer mockCarts.AssertExpectations(t)
			defer mockProducts.AssertExpectations(t)
			tc.mockSetup(mockCarts, mockProducts)
			svc := usecase.NewCart(mockCarts, mockProducts)

			_, err := svc.AddItem(context.Background(), cartID, tc.input)

			assertCartError(t, err, tc.expected.err, tc.expected.code, tc.expected.details)
		})
	}
}

func TestCart_SetItem(t *testing.T) {
	type expected struct {
		err     error
		code    string
		details map[string]string
	}

	tests := map[string]struct {
		input     domain.SetItemInput
		mockSetup func(carts *MockCartRepository, products *MockProductReader)
		expected  expected
	}{
		"should set the absolute quantity when within stock": {
			input: domain.SetItemInput{Quantity: 5},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(cartWithItem(2), nil).Once()
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil).Once()
				carts.On("SaveItem", cartID, productID, 5).Return(nil)
				carts.On("FindByID", cartID).Return(cartWithItem(5), nil).Once()
				products.On("FindByID", productID).Return(productWithStock(10, "10.00"), nil).Once()
			},
		},
		"should reject a quantity below 1": {
			input:     domain.SetItemInput{Quantity: 0},
			mockSetup: func(_ *MockCartRepository, _ *MockProductReader) {},
			expected:  expected{err: usecase.ErrInvalidCartInput, code: "validation_error"},
		},
		"should return cart_not_found when the cart is missing": {
			input: domain.SetItemInput{Quantity: 1},
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			expected: expected{err: domain.ErrNotFound, code: "cart_not_found"},
		},
		"should return item_not_found when the line is not in the cart": {
			input: domain.SetItemInput{Quantity: 1},
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).Return(emptyCart(), nil)
			},
			expected: expected{err: usecase.ErrItemNotFound, code: "item_not_found"},
		},
		"should reject when the set quantity exceeds stock": {
			input: domain.SetItemInput{Quantity: 5},
			mockSetup: func(carts *MockCartRepository, products *MockProductReader) {
				carts.On("FindByID", cartID).Return(cartWithItem(2), nil)
				products.On("FindByID", productID).Return(productWithStock(3, "10.00"), nil)
			},
			expected: expected{
				err:     usecase.ErrInsufficientStock,
				code:    "insufficient_stock",
				details: map[string]string{"product_id": productID.String(), "requested": "5", "available": "3"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockCarts := &MockCartRepository{}
			mockProducts := &MockProductReader{}
			defer mockCarts.AssertExpectations(t)
			defer mockProducts.AssertExpectations(t)
			tc.mockSetup(mockCarts, mockProducts)
			svc := usecase.NewCart(mockCarts, mockProducts)

			_, err := svc.SetItem(context.Background(), cartID, productID, tc.input)

			assertCartError(t, err, tc.expected.err, tc.expected.code, tc.expected.details)
		})
	}
}

func TestCart_RemoveItem(t *testing.T) {
	type expected struct {
		err  error
		code string
	}

	tests := map[string]struct {
		mockSetup func(carts *MockCartRepository, products *MockProductReader)
		expected  expected
	}{
		"should remove the line and return the cart": {
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).Return(cartWithItem(2), nil).Once()
				carts.On("RemoveItem", cartID, productID).Return(nil)
				carts.On("FindByID", cartID).Return(emptyCart(), nil).Once()
			},
		},
		"should return cart_not_found when the cart is missing": {
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			expected: expected{err: domain.ErrNotFound, code: "cart_not_found"},
		},
		"should return item_not_found when the line is not in the cart": {
			mockSetup: func(carts *MockCartRepository, _ *MockProductReader) {
				carts.On("FindByID", cartID).Return(emptyCart(), nil)
				carts.On("RemoveItem", cartID, productID).
					Return(domain.WrapNotFound(assert.AnError, "item_not_found", "cart item not found"))
			},
			expected: expected{err: domain.ErrNotFound, code: "item_not_found"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockCarts := &MockCartRepository{}
			mockProducts := &MockProductReader{}
			defer mockCarts.AssertExpectations(t)
			defer mockProducts.AssertExpectations(t)
			tc.mockSetup(mockCarts, mockProducts)
			svc := usecase.NewCart(mockCarts, mockProducts)

			_, err := svc.RemoveItem(context.Background(), cartID, productID)

			assertCartError(t, err, tc.expected.err, tc.expected.code, nil)
		})
	}
}

func assertCartError(t *testing.T, err, wantErr error, wantCode string, wantDetails map[string]string) {
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
