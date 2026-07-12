package usecase_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ntd-ecomerce-api/internal/domain"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Add(_ context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepository) AddBatch(_ context.Context, products []domain.Product) ([]domain.Product, []string, error) {
	args := m.Called(products)
	var inserted []domain.Product
	if v := args.Get(0); v != nil {
		inserted = v.([]domain.Product)
	}
	var duplicateSKUs []string
	if v := args.Get(1); v != nil {
		duplicateSKUs = v.([]string)
	}
	return inserted, duplicateSKUs, args.Error(2)
}

func (m *MockProductRepository) FindAll(_ context.Context, filter domain.ProductFilter, page domain.PageRequest) (domain.ProductList, error) {
	args := m.Called(filter, page)
	return args.Get(0).(domain.ProductList), args.Error(1)
}

func (m *MockProductRepository) FindCategories(_ context.Context) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProductRepository) FindByID(_ context.Context, id uuid.UUID) (domain.Product, error) {
	args := m.Called(id)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(_ context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(_ context.Context, id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
