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

var (
	cartID    = uuid.New()
	productID = uuid.New()
)

func newCartEngine(usecase api.CartUsecase) *gin.Engine {
	r := gin.New()
	api.NewCartHandlers(r, usecase)
	return r
}

func filledCart() domain.Cart {
	cart := domain.Cart{
		ID: cartID,
		Items: []domain.CartItem{{
			ProductID: productID,
			SKU:       "RS-001",
			Name:      "Running Shoes",
			UnitPrice: decimal.RequireFromString("10.00"),
			Quantity:  3,
		}},
	}
	cart.Recalculate()
	return cart
}

func insufficientStockErr() error {
	return domain.WrapConflictDetails(assert.AnError, "insufficient_stock", "insufficient stock", map[string]string{
		"product_id": productID.String(),
		"requested":  "5",
		"available":  "3",
	})
}

func TestCartHandler_Create(t *testing.T) {
	t.Run("should respond 201 with an empty cart and string total", func(t *testing.T) {
		mockUsecase := &MockCartUsecase{}
		defer mockUsecase.AssertExpectations(t)
		mockUsecase.On("Create").Return(domain.Cart{ID: cartID, Items: []domain.CartItem{}, Total: decimal.Zero}, nil)
		engine := newCartEngine(mockUsecase)

		req := httptest.NewRequest(http.MethodPost, "/carts", nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "0", body["total"])
	})
}

func TestCartHandler_Get(t *testing.T) {
	tests := map[string]struct {
		id        string
		mockSetup func(mockUsecase *MockCartUsecase)
		status    int
		code      string
	}{
		"should respond 200 with the cart when it exists": {
			id: cartID.String(),
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("Get", cartID).Return(filledCart(), nil)
			},
			status: http.StatusOK,
		},
		"should respond 404 when the cart does not exist": {
			id: cartID.String(),
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("Get", cartID).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "cart_not_found", "cart not found"))
			},
			status: http.StatusNotFound,
			code:   "cart_not_found",
		},
		"should respond 422 when the cart id is not a valid uuid": {
			id:        "not-a-uuid",
			mockSetup: func(_ *MockCartUsecase) {},
			status:    http.StatusUnprocessableEntity,
			code:      "validation_error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUsecase := &MockCartUsecase{}
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)
			engine := newCartEngine(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/carts/"+tc.id, nil)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assertStatusAndCode(t, rec, tc.status, tc.code, nil)
		})
	}
}

func TestCartHandler_AddItem(t *testing.T) {
	tests := map[string]struct {
		body      string
		mockSetup func(mockUsecase *MockCartUsecase)
		status    int
		code      string
		details   map[string]string
	}{
		"should respond 200 with the updated cart": {
			body: `{"product_id":"` + productID.String() + `","quantity":3}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("AddItem", cartID, domain.AddItemInput{ProductID: productID, Quantity: 3}).
					Return(filledCart(), nil)
			},
			status: http.StatusOK,
		},
		"should respond 422 with details when quantity is below 1": {
			body: `{"product_id":"` + productID.String() + `","quantity":0}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("AddItem", cartID, domain.AddItemInput{ProductID: productID, Quantity: 0}).
					Return(domain.Cart{}, domain.WrapValidation(assert.AnError, map[string]string{"quantity": "must_be_positive_integer"}))
			},
			status:  http.StatusUnprocessableEntity,
			code:    "validation_error",
			details: map[string]string{"quantity": "must_be_positive_integer"},
		},
		"should respond 404 when the product does not exist": {
			body: `{"product_id":"` + productID.String() + `","quantity":1}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("AddItem", cartID, domain.AddItemInput{ProductID: productID, Quantity: 1}).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "product_not_found", "product not found"))
			},
			status: http.StatusNotFound,
			code:   "product_not_found",
		},
		"should respond 409 with details when stock is insufficient": {
			body: `{"product_id":"` + productID.String() + `","quantity":5}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("AddItem", cartID, domain.AddItemInput{ProductID: productID, Quantity: 5}).
					Return(domain.Cart{}, insufficientStockErr())
			},
			status:  http.StatusConflict,
			code:    "insufficient_stock",
			details: map[string]string{"product_id": productID.String(), "requested": "5", "available": "3"},
		},
		"should respond 422 when the json body is malformed": {
			body:      `{"quantity":`,
			mockSetup: func(_ *MockCartUsecase) {},
			status:    http.StatusUnprocessableEntity,
			code:      "validation_error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUsecase := &MockCartUsecase{}
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)
			engine := newCartEngine(mockUsecase)

			req := httptest.NewRequest(http.MethodPost, "/carts/"+cartID.String()+"/items", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assertStatusAndCode(t, rec, tc.status, tc.code, tc.details)
		})
	}
}

func TestCartHandler_SetItem(t *testing.T) {
	tests := map[string]struct {
		body      string
		mockSetup func(mockUsecase *MockCartUsecase)
		status    int
		code      string
	}{
		"should respond 200 with the updated cart": {
			body: `{"quantity":5}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("SetItem", cartID, productID, domain.SetItemInput{Quantity: 5}).
					Return(filledCart(), nil)
			},
			status: http.StatusOK,
		},
		"should respond 404 when the line is not in the cart": {
			body: `{"quantity":5}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("SetItem", cartID, productID, domain.SetItemInput{Quantity: 5}).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "item_not_found", "cart item not found"))
			},
			status: http.StatusNotFound,
			code:   "item_not_found",
		},
		"should respond 409 when stock is insufficient": {
			body: `{"quantity":5}`,
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("SetItem", cartID, productID, domain.SetItemInput{Quantity: 5}).
					Return(domain.Cart{}, insufficientStockErr())
			},
			status: http.StatusConflict,
			code:   "insufficient_stock",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUsecase := &MockCartUsecase{}
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)
			engine := newCartEngine(mockUsecase)

			url := "/carts/" + cartID.String() + "/items/" + productID.String()
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assertStatusAndCode(t, rec, tc.status, tc.code, nil)
		})
	}
}

func TestCartHandler_RemoveItem(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(mockUsecase *MockCartUsecase)
		status    int
		code      string
	}{
		"should respond 200 with the cart after removal": {
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("RemoveItem", cartID, productID).
					Return(domain.Cart{ID: cartID, Items: []domain.CartItem{}, Total: decimal.Zero}, nil)
			},
			status: http.StatusOK,
		},
		"should respond 404 when the line is not in the cart": {
			mockSetup: func(mockUsecase *MockCartUsecase) {
				mockUsecase.On("RemoveItem", cartID, productID).
					Return(domain.Cart{}, domain.WrapNotFound(assert.AnError, "item_not_found", "cart item not found"))
			},
			status: http.StatusNotFound,
			code:   "item_not_found",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUsecase := &MockCartUsecase{}
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)
			engine := newCartEngine(mockUsecase)

			url := "/carts/" + cartID.String() + "/items/" + productID.String()
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			assertStatusAndCode(t, rec, tc.status, tc.code, nil)
		})
	}
}

func assertStatusAndCode(t *testing.T, rec *httptest.ResponseRecorder, status int, code string, details map[string]string) {
	t.Helper()

	assert.Equal(t, status, rec.Code)
	if code != "" {
		var envelope testErrorEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
		assert.Equal(t, code, envelope.Error.Code)
		if details != nil {
			assert.Equal(t, details, envelope.Error.Details)
		}
	}
}
