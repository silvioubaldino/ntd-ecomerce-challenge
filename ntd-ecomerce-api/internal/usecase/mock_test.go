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

// ctx is ignored in m.Called.
func (m *MockProductRepository) Add(_ context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepository) FindAll(_ context.Context, page domain.Page) (domain.ProductList, error) {
	args := m.Called(page)
	return args.Get(0).(domain.ProductList), args.Error(1)
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
