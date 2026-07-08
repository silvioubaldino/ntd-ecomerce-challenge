package api_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ntd-ecomerce-api/internal/domain"
)

type MockProductUsecase struct {
	mock.Mock
}

func (m *MockProductUsecase) Add(_ context.Context, input domain.ProductInput) (domain.Product, error) {
	args := m.Called(input)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductUsecase) FindAll(_ context.Context, page domain.Page) (domain.ProductList, error) {
	args := m.Called(page)
	return args.Get(0).(domain.ProductList), args.Error(1)
}

func (m *MockProductUsecase) FindByID(_ context.Context, id uuid.UUID) (domain.Product, error) {
	args := m.Called(id)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductUsecase) Update(_ context.Context, id uuid.UUID, input domain.ProductInput) (domain.Product, error) {
	args := m.Called(id, input)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductUsecase) DeleteOne(_ context.Context, id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
