package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ntd-ecomerce-api/internal/domain"
	"ntd-ecomerce-api/internal/infrastructure/api"
)

var orderID = uuid.New()

func newOrderEngine(usecase api.OrderUsecase) *gin.Engine {
	r := gin.New()
	api.NewOrderHandlers(r, usecase)
	return r
}

func confirmedOrder() domain.Order {
	order := domain.Order{
		ID:       orderID,
		Status:   domain.OrderStatusConfirmed,
		Customer: domain.Customer{Name: "Ada Lovelace", Email: "ada@example.com"},
		Items: []domain.OrderItem{{
			ProductID: productID,
			SKU:       "RS-001",
			Name:      "Running Shoes",
			UnitPrice: decimal.RequireFromString("10.00"),
			Quantity:  3,
		}},
		Payment: domain.ApprovedPayment(),
	}
	order.Recalculate()
	return order
}

func TestOrderHandler_Checkout(t *testing.T) {
	validBody := `{"cart_id":"` + cartID.String() + `","customer":{"name":"Ada Lovelace","email":"ada@example.com"}}`
	validInput := domain.CheckoutInput{CartID: cartID, Customer: domain.Customer{Name: "Ada Lovelace", Email: "ada@example.com"}}

	tests := map[string]struct {
		body      string
		mockSetup func(mockUsecase *MockOrderUsecase)
		status    int
		code      string
		details   map[string]string
	}{
		"should respond 201 with the confirmed order": {
			body: validBody,
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Checkout", validInput).Return(confirmedOrder(), nil)
			},
			status: http.StatusCreated,
		},
		"should respond 404 when the cart does not exist": {
			body: validBody,
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Checkout", validInput).
					Return(domain.Order{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			status: http.StatusNotFound,
			code:   "cart_not_found",
		},
		"should respond 422 when the cart is empty": {
			body: validBody,
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Checkout", validInput).
					Return(domain.Order{}, domain.WrapValidationCode(assert.AnError, "cart_empty", "cart is empty"))
			},
			status: http.StatusUnprocessableEntity,
			code:   "cart_empty",
		},
		"should respond 422 with details when the customer is invalid": {
			body: `{"cart_id":"` + cartID.String() + `","customer":{"name":"","email":"bad"}}`,
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Checkout", domain.CheckoutInput{CartID: cartID, Customer: domain.Customer{Name: "", Email: "bad"}}).
					Return(domain.Order{}, domain.WrapValidation(assert.AnError, map[string]string{"name": "required", "email": "invalid"}))
			},
			status:  http.StatusUnprocessableEntity,
			code:    "validation_error",
			details: map[string]string{"name": "required", "email": "invalid"},
		},
		"should respond 409 with details when stock is insufficient": {
			body: validBody,
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Checkout", validInput).
					Return(domain.Order{}, domain.WrapConflictDetails(assert.AnError, "insufficient_stock", "insufficient stock",
						map[string]string{productID.String(): "requested=5, available=3"}))
			},
			status:  http.StatusConflict,
			code:    "insufficient_stock",
			details: map[string]string{productID.String(): "requested=5, available=3"},
		},
		"should respond 422 when the json body is malformed": {
			body:      `{"cart_id":`,
			mockSetup: func(_ *MockOrderUsecase) {},
			status:    http.StatusUnprocessableEntity,
			code:      "validation_error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUsecase := &MockOrderUsecase{}
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)
			engine := newOrderEngine(mockUsecase)

			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assertStatusAndCode(t, rec, tc.status, tc.code, tc.details)
		})
	}
}

func TestOrderHandler_Checkout_ResponseShape(t *testing.T) {
	t.Run("should marshal decimals as strings and expose the simulated payment", func(t *testing.T) {
		mockUsecase := &MockOrderUsecase{}
		defer mockUsecase.AssertExpectations(t)
		mockUsecase.On("Checkout", domain.CheckoutInput{CartID: cartID, Customer: domain.Customer{Name: "Ada Lovelace", Email: "ada@example.com"}}).
			Return(confirmedOrder(), nil)
		engine := newOrderEngine(mockUsecase)

		body := `{"cart_id":"` + cartID.String() + `","customer":{"name":"Ada Lovelace","email":"ada@example.com"}}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
		assert.Equal(t, "confirmed", payload["status"])
		assert.Equal(t, "30", payload["total"])
		payment := payload["payment"].(map[string]any)
		assert.Equal(t, "simulated", payment["method"])
		assert.Equal(t, "approved", payment["status"])
	})
}

func TestOrderHandler_Get(t *testing.T) {
	tests := map[string]struct {
		id        string
		mockSetup func(mockUsecase *MockOrderUsecase)
		status    int
		code      string
	}{
		"should respond 200 with the order when it exists": {
			id: orderID.String(),
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Get", orderID).Return(confirmedOrder(), nil)
			},
			status: http.StatusOK,
		},
		"should respond 404 when the order does not exist": {
			id: orderID.String(),
			mockSetup: func(mockUsecase *MockOrderUsecase) {
				mockUsecase.On("Get", orderID).
					Return(domain.Order{}, domain.WrapNotFound(assert.AnError, "order_not_found", "order not found"))
			},
			status: http.StatusNotFound,
			code:   "order_not_found",
		},
		"should respond 422 when the order id is not a valid uuid": {
			id:        "not-a-uuid",
			mockSetup: func(_ *MockOrderUsecase) {},
			status:    http.StatusUnprocessableEntity,
			code:      "validation_error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUsecase := &MockOrderUsecase{}
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)
			engine := newOrderEngine(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/orders/"+tc.id, nil)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assertStatusAndCode(t, rec, tc.status, tc.code, nil)
		})
	}
}
