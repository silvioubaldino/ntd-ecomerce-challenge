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

var fixtureID = uuid.New()

func validProductInput(sku string) domain.ProductInput {
	return domain.ProductInput{
		Name:        "Running Shoes",
		SKU:         sku,
		Description: "Lightweight running shoes",
		Category:    "Footwear",
		Price:       decimal.NewFromFloat(89.99),
		Stock:       150,
		WeightKg:    decimal.NewFromFloat(0.35),
	}
}

func TestProduct_Add(t *testing.T) {
	type (
		input struct {
			productInput domain.ProductInput
		}
		expected struct {
			output domain.Product
			err    error
		}
	)

	trimmedProduct := validProductInput("RS-001").ToProduct()
	createdProduct := trimmedProduct
	createdProduct.ID = fixtureID

	tests := map[string]struct {
		input input
		mockSetup func(mockRepo *MockProductRepository)
		expected expected
	}{
		"should trim sku and create product when input is valid": {
			input: input{productInput: validProductInput(" RS-001 ")},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("Add", trimmedProduct).Return(createdProduct, nil)
			},
			expected: expected{output: createdProduct, err: nil},
		},
		"should return validation error when input is invalid": {
			input:     input{productInput: domain.ProductInput{SKU: "RS-001"}},
			mockSetup: func(_ *MockProductRepository) {},
			expected:  expected{output: domain.Product{}, err: usecase.ErrInvalidProductInput},
		},
		"should return error when repository fails": {
			input: input{productInput: validProductInput("RS-001")},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("Add", trimmedProduct).Return(domain.Product{}, assert.AnError)
			},
			expected: expected{output: domain.Product{}, err: assert.AnError},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockRepo = &MockProductRepository{}
				svc      = usecase.NewProduct(mockRepo)
			)
			defer mockRepo.AssertExpectations(t)
			tc.mockSetup(mockRepo)

			output, err := svc.Add(context.Background(), tc.input.productInput)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.output, output)
		})
	}
}

func TestProduct_FindByID(t *testing.T) {
	type (
		input struct {
			id uuid.UUID
		}
		expected struct {
			output domain.Product
			err    error
		}
	)

	tests := map[string]struct {
		input input
		mockSetup func(mockRepo *MockProductRepository)
		expected expected
	}{
		"should return product when it exists": {
			input: input{id: fixtureID},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindByID", fixtureID).Return(domain.Product{ID: fixtureID}, nil)
			},
			expected: expected{output: domain.Product{ID: fixtureID}, err: nil},
		},
		"should return error when repository fails": {
			input: input{id: fixtureID},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindByID", fixtureID).Return(domain.Product{}, domain.ErrNotFound)
			},
			expected: expected{output: domain.Product{}, err: domain.ErrNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockRepo = &MockProductRepository{}
				svc      = usecase.NewProduct(mockRepo)
			)
			defer mockRepo.AssertExpectations(t)
			tc.mockSetup(mockRepo)

			output, err := svc.FindByID(context.Background(), tc.input.id)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.output, output)
		})
	}
}

func TestProduct_FindAll(t *testing.T) {
	type (
		input struct {
			page domain.Page
		}
		expected struct {
			output domain.ProductList
			err    error
		}
	)

	page := domain.DefaultPage()
	list := domain.ProductList{
		Data:       []domain.Product{{ID: fixtureID}},
		Pagination: domain.Pagination{Page: 1, PageSize: 20, Total: 1},
	}

	tests := map[string]struct {
		input input
		mockSetup func(mockRepo *MockProductRepository)
		expected expected
	}{
		"should return the page of products": {
			input: input{page: page},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindAll", page).Return(list, nil)
			},
			expected: expected{output: list, err: nil},
		},
		"should return error when repository fails": {
			input: input{page: page},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindAll", page).Return(domain.ProductList{}, assert.AnError)
			},
			expected: expected{output: domain.ProductList{}, err: assert.AnError},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockRepo = &MockProductRepository{}
				svc      = usecase.NewProduct(mockRepo)
			)
			defer mockRepo.AssertExpectations(t)
			tc.mockSetup(mockRepo)

			output, err := svc.FindAll(context.Background(), tc.input.page)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.output, output)
		})
	}
}

func TestProduct_Update(t *testing.T) {
	type (
		input struct {
			id           uuid.UUID
			productInput domain.ProductInput
		}
		expected struct {
			output domain.Product
			err    error
		}
	)

	productWithID := validProductInput("RS-001").ToProduct()
	productWithID.ID = fixtureID
	updatedProduct := productWithID

	tests := map[string]struct {
		input input
		mockSetup func(mockRepo *MockProductRepository)
		expected expected
	}{
		"should update product when input is valid": {
			input: input{id: fixtureID, productInput: validProductInput("RS-001")},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("Update", productWithID).Return(updatedProduct, nil)
			},
			expected: expected{output: updatedProduct, err: nil},
		},
		"should return validation error when input is invalid": {
			input:     input{id: fixtureID, productInput: domain.ProductInput{}},
			mockSetup: func(_ *MockProductRepository) {},
			expected:  expected{output: domain.Product{}, err: usecase.ErrInvalidProductInput},
		},
		"should return not found error when product does not exist": {
			input: input{id: fixtureID, productInput: validProductInput("RS-001")},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("Update", productWithID).Return(domain.Product{}, domain.ErrNotFound)
			},
			expected: expected{output: domain.Product{}, err: domain.ErrNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockRepo = &MockProductRepository{}
				svc      = usecase.NewProduct(mockRepo)
			)
			defer mockRepo.AssertExpectations(t)
			tc.mockSetup(mockRepo)

			output, err := svc.Update(context.Background(), tc.input.id, tc.input.productInput)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.output, output)
		})
	}
}

func TestProduct_DeleteOne(t *testing.T) {
	type (
		input struct {
			id uuid.UUID
		}
		expected struct {
			err error
		}
	)

	tests := map[string]struct {
		input input
		mockSetup func(mockRepo *MockProductRepository)
		expected expected
	}{
		"should delete product when it exists": {
			input: input{id: fixtureID},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("Delete", fixtureID).Return(nil)
			},
			expected: expected{err: nil},
		},
		"should return not found error when product does not exist": {
			input: input{id: fixtureID},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("Delete", fixtureID).Return(domain.ErrNotFound)
			},
			expected: expected{err: domain.ErrNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mockRepo = &MockProductRepository{}
				svc      = usecase.NewProduct(mockRepo)
			)
			defer mockRepo.AssertExpectations(t)
			tc.mockSetup(mockRepo)

			err := svc.DeleteOne(context.Background(), tc.input.id)

			assert.ErrorIs(t, err, tc.expected.err)
		})
	}
}
