package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ntd-ecomerce-api/internal/domain"
	"ntd-ecomerce-api/internal/infrastructure/api"
)

type testErrorEnvelope struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

var fixtureID = uuid.New()

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestEngine(usecase api.ProductUsecase) *gin.Engine {
	r := gin.New()
	api.NewProductHandlers(r, usecase)
	return r
}

func fixtureProductInput() domain.ProductInput {
	return domain.ProductInput{
		Name:        "Running Shoes",
		SKU:         "RS-001",
		Description: "Lightweight running shoes",
		Category:    "Footwear",
		Price:       decimal.RequireFromString("89.99"),
		Stock:       150,
		WeightKg:    decimal.RequireFromString("0.35"),
	}
}

func fixtureProductInputJSON() []byte {
	body, _ := json.Marshal(fixtureProductInput())
	return body
}

func TestProductHandler_Add(t *testing.T) {
	type expected struct {
		status  int
		code    string
		details map[string]string
	}

	tests := map[string]struct {
		body      []byte
		mockSetup func(mockUsecase *MockProductUsecase)
		expected  expected
	}{
		"should respond 201 with string decimals when product is created": {
			body: fixtureProductInputJSON(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Add", fixtureProductInput()).Return(domain.Product{
					ID:       fixtureID,
					Name:     "Running Shoes",
					SKU:      "RS-001",
					Category: "Footwear",
					Price:    decimal.RequireFromString("89.99"),
					Stock:    150,
					WeightKg: decimal.RequireFromString("0.35"),
				}, nil)
			},
			expected: expected{status: http.StatusCreated},
		},
		"should respond 409 when sku already exists": {
			body: fixtureProductInputJSON(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Add", fixtureProductInput()).
					Return(domain.Product{}, domain.WrapConflict(errors.New("dup"), "sku_already_exists", "sku already exists"))
			},
			expected: expected{status: http.StatusConflict, code: "sku_already_exists"},
		},
		"should respond 422 with details when usecase reports invalid fields": {
			body: fixtureProductInputJSON(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Add", fixtureProductInput()).
					Return(domain.Product{}, domain.WrapValidation(errors.New("invalid"), map[string]string{"price": "must_be_non_negative_decimal"}))
			},
			expected: expected{
				status:  http.StatusUnprocessableEntity,
				code:    "validation_error",
				details: map[string]string{"price": "must_be_non_negative_decimal"},
			},
		},
		"should respond 422 when json body is malformed": {
			body:      []byte(`{"name":`),
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusUnprocessableEntity, code: "validation_error"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockUsecase = &MockProductUsecase{}
				engine      = newTestEngine(mockUsecase)
			)
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Code)
			if tc.expected.code != "" {
				var envelope testErrorEnvelope
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
				assert.Equal(t, tc.expected.code, envelope.Error.Code)
				if tc.expected.details != nil {
					assert.Equal(t, tc.expected.details, envelope.Error.Details)
				}
			} else if tc.expected.status == http.StatusCreated {
				var product domain.Product
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &product))
				assert.Equal(t, "89.99", product.Price.String())
				assert.Equal(t, "0.35", product.WeightKg.String())
			}
		})
	}
}

func TestProductHandler_FindAll(t *testing.T) {
	type expected struct {
		status int
		code   string
	}

	tests := map[string]struct {
		query     string
		mockSetup func(mockUsecase *MockProductUsecase)
		expected  expected
	}{
		"should default to page 1 size 20 when no query params are given": {
			query: "",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 1, Size: 20}).
					Return(domain.ProductList{Pagination: domain.Pagination{Page: 1, PageSize: 20}}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should list the requested page": {
			query: "?page=2&page_size=20",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 2, Size: 20}).
					Return(domain.ProductList{
						Data:       make([]domain.Product, 5),
						Pagination: domain.Pagination{Page: 2, PageSize: 20, Total: 25},
					}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should respond 422 when page_size exceeds the max bound": {
			query:     "?page_size=101",
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusUnprocessableEntity},
		},
		"should behave as unfiltered list when q is absent": {
			query: "",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 1, Size: 20, Query: ""}).
					Return(domain.ProductList{Pagination: domain.Pagination{Page: 1, PageSize: 20}}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should behave as unfiltered list when q is blank": {
			query: "?q=%20%20%20",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 1, Size: 20, Query: ""}).
					Return(domain.ProductList{Pagination: domain.Pagination{Page: 1, PageSize: 20}}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should pass the trimmed q to the usecase": {
			query: "?q=%20blue%20",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 1, Size: 20, Query: "blue"}).
					Return(domain.ProductList{
						Data:       []domain.Product{{Name: "Blue Shirt"}},
						Pagination: domain.Pagination{Page: 1, PageSize: 20, Total: 1},
					}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should respond 200 with empty data when q has no matches": {
			query: "?q=zzz-nomatch",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 1, Size: 20, Query: "zzz-nomatch"}).
					Return(domain.ProductList{
						Data:       []domain.Product{},
						Pagination: domain.Pagination{Page: 1, PageSize: 20, Total: 0},
					}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should paginate the filtered set when q and page are given": {
			query: "?q=shirt&page=2&page_size=20",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 2, Size: 20, Query: "shirt"}).
					Return(domain.ProductList{
						Data:       make([]domain.Product, 5),
						Pagination: domain.Pagination{Page: 2, PageSize: 20, Total: 25},
					}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should be case-insensitive by passing q through unchanged": {
			query: "?q=BLUE",
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindAll", domain.Page{Number: 1, Size: 20, Query: "BLUE"}).
					Return(domain.ProductList{
						Data:       []domain.Product{{Name: "Blue Shirt"}},
						Pagination: domain.Pagination{Page: 1, PageSize: 20, Total: 1},
					}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should respond 422 when page is invalid regardless of q": {
			query:     "?q=blue&page=0",
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusUnprocessableEntity, code: "validation_error"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockUsecase = &MockProductUsecase{}
				engine      = newTestEngine(mockUsecase)
			)
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/products"+tc.query, nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			if tc.expected.code != "" {
				var envelope testErrorEnvelope
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
				assert.Equal(t, tc.expected.code, envelope.Error.Code)
			}

			assert.Equal(t, tc.expected.status, rec.Code)
		})
	}
}

func TestProductHandler_FindByID(t *testing.T) {
	type expected struct {
		status int
	}

	tests := map[string]struct {
		id        string
		mockSetup func(mockUsecase *MockProductUsecase)
		expected  expected
	}{
		"should respond 200 with the product when it exists": {
			id: fixtureID.String(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindByID", fixtureID).Return(domain.Product{ID: fixtureID}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should respond 404 when the product does not exist": {
			id: fixtureID.String(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("FindByID", fixtureID).
					Return(domain.Product{}, domain.WrapNotFound(errors.New("no rows"), "product_not_found", "product not found"))
			},
			expected: expected{status: http.StatusNotFound},
		},
		"should respond 422 when the id is not a valid uuid": {
			id:        "not-a-uuid",
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusUnprocessableEntity},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockUsecase = &MockProductUsecase{}
				engine      = newTestEngine(mockUsecase)
			)
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/products/"+tc.id, nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Code)
		})
	}
}

func TestProductHandler_Update(t *testing.T) {
	type expected struct {
		status int
	}

	tests := map[string]struct {
		body      []byte
		mockSetup func(mockUsecase *MockProductUsecase)
		expected  expected
	}{
		"should respond 200 with the updated product": {
			body: fixtureProductInputJSON(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Update", fixtureID, fixtureProductInput()).
					Return(domain.Product{ID: fixtureID}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should respond 404 when the product does not exist": {
			body: fixtureProductInputJSON(),
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Update", fixtureID, fixtureProductInput()).
					Return(domain.Product{}, domain.WrapNotFound(errors.New("no rows"), "product_not_found", "product not found"))
			},
			expected: expected{status: http.StatusNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockUsecase = &MockProductUsecase{}
				engine      = newTestEngine(mockUsecase)
			)
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)

			req := httptest.NewRequest(http.MethodPut, "/products/"+fixtureID.String(), bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Code)
		})
	}
}

func multipartCSVRequest(t *testing.T, content []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "products.csv")
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/products/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func matchesContent(expected []byte) any {
	return mock.MatchedBy(func(r io.Reader) bool {
		data, err := io.ReadAll(r)
		return err == nil && bytes.Equal(data, expected)
	})
}

func TestProductHandler_Import(t *testing.T) {
	type expected struct {
		status int
		code   string
	}

	validCSV := []byte("name,sku,description,category,price,stock,weight_kg\nRunning Shoes,RS-001,desc,Footwear,89.99,150,0.35\n")
	oversizedCSV := bytes.Repeat([]byte("a"), int(domain.MaxImportFileBytes)+1)

	tests := map[string]struct {
		req       func(t *testing.T) *http.Request
		mockSetup func(mockUsecase *MockProductUsecase)
		expected  expected
	}{
		"should respond 200 with the import report for a valid file": {
			req: func(t *testing.T) *http.Request { return multipartCSVRequest(t, validCSV) },
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Import", matchesContent(validCSV)).Return(domain.ImportReport{
					Summary:  domain.ImportSummary{Total: 1, Imported: 1, Rejected: 0},
					Rejected: []domain.RejectedRow{},
				}, nil)
			},
			expected: expected{status: http.StatusOK},
		},
		"should respond 400 when no file is sent": {
			req: func(_ *testing.T) *http.Request {
				return httptest.NewRequest(http.MethodPost, "/products/import", bytes.NewReader(nil))
			},
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusBadRequest, code: "invalid_file"},
		},
		"should respond 400 when the file is empty": {
			req:       func(t *testing.T) *http.Request { return multipartCSVRequest(t, []byte{}) },
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusBadRequest, code: "invalid_file"},
		},
		"should respond 413 when the file exceeds the size cap": {
			req:       func(t *testing.T) *http.Request { return multipartCSVRequest(t, oversizedCSV) },
			mockSetup: func(_ *MockProductUsecase) {},
			expected:  expected{status: http.StatusRequestEntityTooLarge, code: "file_too_large"},
		},
		"should respond 422 when the usecase reports an invalid header": {
			req: func(t *testing.T) *http.Request { return multipartCSVRequest(t, validCSV) },
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("Import", matchesContent(validCSV)).
					Return(domain.ImportReport{}, domain.WrapValidationCode(domain.ErrInvalidCSVHeader, "invalid_header", "csv header does not match"))
			},
			expected: expected{status: http.StatusUnprocessableEntity, code: "invalid_header"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockUsecase = &MockProductUsecase{}
				engine      = newTestEngine(mockUsecase)
			)
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, tc.req(t))

			assert.Equal(t, tc.expected.status, rec.Code)
			if tc.expected.code != "" {
				var envelope testErrorEnvelope
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
				assert.Equal(t, tc.expected.code, envelope.Error.Code)
			}
		})
	}
}

func TestProductHandler_DeleteOne(t *testing.T) {
	type expected struct {
		status int
	}

	tests := map[string]struct {
		mockSetup func(mockUsecase *MockProductUsecase)
		expected  expected
	}{
		"should respond 204 when the product is deleted": {
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("DeleteOne", fixtureID).Return(nil)
			},
			expected: expected{status: http.StatusNoContent},
		},
		"should respond 404 when the product does not exist": {
			mockSetup: func(mockUsecase *MockProductUsecase) {
				mockUsecase.On("DeleteOne", fixtureID).
					Return(domain.WrapNotFound(errors.New("no rows"), "product_not_found", "product not found"))
			},
			expected: expected{status: http.StatusNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockUsecase = &MockProductUsecase{}
				engine      = newTestEngine(mockUsecase)
			)
			defer mockUsecase.AssertExpectations(t)
			tc.mockSetup(mockUsecase)

			req := httptest.NewRequest(http.MethodDelete, "/products/"+fixtureID.String(), nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Code)
		})
	}
}
