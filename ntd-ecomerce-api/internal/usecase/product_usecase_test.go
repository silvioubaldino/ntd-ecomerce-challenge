package usecase_test

import (
	"context"
	"fmt"
	"strings"
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
		input     input
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
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
		input     input
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
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
			filter domain.ProductFilter
			page   domain.PageRequest
		}
		expected struct {
			output domain.ProductList
			err    error
		}
	)

	page := domain.DefaultPageRequest()
	filter := domain.ProductFilter{Category: "Footwear"}
	list := domain.ProductList{
		Data:       []domain.Product{{ID: fixtureID}},
		Pagination: domain.Pagination{Limit: 20, NextCursor: nil},
	}

	tests := map[string]struct {
		input     input
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
	}{
		"should return the page of products": {
			input: input{filter: filter, page: page},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindAll", filter, page).Return(list, nil)
			},
			expected: expected{output: list, err: nil},
		},
		"should return error when repository fails": {
			input: input{filter: filter, page: page},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindAll", filter, page).Return(domain.ProductList{}, assert.AnError)
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

			output, err := svc.FindAll(context.Background(), tc.input.filter, tc.input.page)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.output, output)
		})
	}
}

func TestProduct_FindCategories(t *testing.T) {
	type expected struct {
		output []string
		err    error
	}

	tests := map[string]struct {
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
	}{
		"should return distinct categories": {
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindCategories").Return([]string{"Apparel", "Footwear"}, nil)
			},
			expected: expected{output: []string{"Apparel", "Footwear"}, err: nil},
		},
		"should return error when repository fails": {
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("FindCategories").Return([]string(nil), assert.AnError)
			},
			expected: expected{output: nil, err: assert.AnError},
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

			output, err := svc.FindCategories(context.Background())

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
		input     input
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
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

const csvHeader = "name,sku,description,category,price,stock,weight_kg\n"

func productInput(name, sku, category, price string, stock int, weightKg string) domain.ProductInput {
	return domain.ProductInput{
		Name:        name,
		SKU:         sku,
		Description: "desc",
		Category:    category,
		Price:       decimal.RequireFromString(price),
		Stock:       stock,
		WeightKg:    decimal.RequireFromString(weightKg),
	}
}

// batchImportRows builds n CSV data rows plus the domain.Product each row parses into,
// in file order, so a test can slice them into the batches the usecase is expected to send.
func batchImportRows(n int) (rows string, products []domain.Product) {
	var sb strings.Builder
	products = make([]domain.Product, n)
	for i := range n {
		sku := fmt.Sprintf("BATCH-%05d", i)
		fmt.Fprintf(&sb, "Item %d,%s,desc,Category,9.99,10,1.00\n", i, sku)
		products[i] = productInput(fmt.Sprintf("Item %d", i), sku, "Category", "9.99", 10, "1.00").ToProduct()
	}
	return sb.String(), products
}

func TestProduct_Import(t *testing.T) {
	type (
		input struct {
			csv string
		}
		expected struct {
			report domain.ImportReport
			err    error
		}
	)

	runningShoes := productInput("Running Shoes", "RS-001", "Footwear", "89.99", 150, "0.35")
	wirelessMouse := productInput("Wireless Mouse", "WM-042", "Electronics", "29.99", 75, "0.12")

	batchRows, batchProducts := batchImportRows(domain.ImportBatchSize + 1)

	tests := map[string]struct {
		input     input
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
	}{
		"should import every valid row in a single batch and report zero rejections": {
			input: input{csv: csvHeader +
				"Running Shoes,RS-001,desc,Footwear,89.99,150,0.35\n" +
				"Wireless Mouse,WM-042,desc,Electronics,29.99,75,0.12\n"},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("AddBatch", []domain.Product{runningShoes.ToProduct(), wirelessMouse.ToProduct()}).
					Return([]domain.Product{{SKU: "RS-001"}, {SKU: "WM-042"}}, []string(nil), nil)
			},
			expected: expected{
				report: domain.ImportReport{
					Summary:  domain.ImportSummary{Total: 2, Imported: 2, Rejected: 0},
					Rejected: []domain.RejectedRow{},
				},
				err: nil,
			},
		},
		"should reject an invalid row without buffering it and still import the valid one": {
			input: input{csv: csvHeader +
				",HD-099,desc,Electronics,149.99,30,0.25\n" +
				"Running Shoes,RS-001,desc,Footwear,89.99,150,0.35\n"},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("AddBatch", []domain.Product{runningShoes.ToProduct()}).
					Return([]domain.Product{{SKU: "RS-001"}}, []string(nil), nil)
			},
			expected: expected{
				report: domain.ImportReport{
					Summary: domain.ImportSummary{Total: 2, Imported: 1, Rejected: 1},
					Rejected: []domain.RejectedRow{
						{Row: 1, SKU: "HD-099", Errors: map[string]string{"name": "required"}},
					},
				},
				err: nil,
			},
		},
		"should reject an in-file duplicate sku before it reaches the batch": {
			input: input{csv: csvHeader +
				"Running Shoes,RS-001,desc,Footwear,89.99,150,0.35\n" +
				"Running Shoes 2,RS-001,desc,Footwear,79.99,50,0.30\n"},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("AddBatch", []domain.Product{runningShoes.ToProduct()}).
					Return([]domain.Product{{SKU: "RS-001"}}, []string(nil), nil)
			},
			expected: expected{
				report: domain.ImportReport{
					Summary: domain.ImportSummary{Total: 2, Imported: 1, Rejected: 1},
					Rejected: []domain.RejectedRow{
						{Row: 2, SKU: "RS-001", Errors: map[string]string{"sku": "duplicate_sku"}},
					},
				},
				err: nil,
			},
		},
		"should report duplicate_sku when the repository finds the sku already in the database": {
			input: input{csv: csvHeader + "Running Shoes,RS-001,desc,Footwear,89.99,150,0.35\n"},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("AddBatch", []domain.Product{runningShoes.ToProduct()}).
					Return(nil, []string{"RS-001"}, nil)
			},
			expected: expected{
				report: domain.ImportReport{
					Summary: domain.ImportSummary{Total: 1, Imported: 0, Rejected: 1},
					Rejected: []domain.RejectedRow{
						{Row: 1, SKU: "RS-001", Errors: map[string]string{"sku": "duplicate_sku"}},
					},
				},
				err: nil,
			},
		},
		"should return the error when the repository fails for a batch": {
			input: input{csv: csvHeader + "Running Shoes,RS-001,desc,Footwear,89.99,150,0.35\n"},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("AddBatch", []domain.Product{runningShoes.ToProduct()}).
					Return(nil, nil, assert.AnError)
			},
			expected: expected{
				report: domain.ImportReport{},
				err:    assert.AnError,
			},
		},
		"should return invalid_header error without reading any row": {
			input:     input{csv: "title,sku,description,category,price,stock,weight_kg\n"},
			mockSetup: func(_ *MockProductRepository) {},
			expected: expected{
				report: domain.ImportReport{},
				err:    domain.ErrInvalidCSVHeader,
			},
		},
		"should split rows into multiple batches when the file exceeds the batch size": {
			input: input{csv: csvHeader + batchRows},
			mockSetup: func(mockRepo *MockProductRepository) {
				mockRepo.On("AddBatch", batchProducts[:domain.ImportBatchSize]).
					Return(batchProducts[:domain.ImportBatchSize], []string(nil), nil)
				mockRepo.On("AddBatch", batchProducts[domain.ImportBatchSize:]).
					Return(batchProducts[domain.ImportBatchSize:], []string(nil), nil)
			},
			expected: expected{
				report: domain.ImportReport{
					Summary: domain.ImportSummary{
						Total:    len(batchProducts),
						Imported: len(batchProducts),
						Rejected: 0,
					},
					Rejected: []domain.RejectedRow{},
				},
				err: nil,
			},
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

			report, err := svc.Import(context.Background(), strings.NewReader(tc.input.csv))

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.report, report)
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
		input     input
		mockSetup func(mockRepo *MockProductRepository)
		expected  expected
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
