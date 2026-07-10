package usecase_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ntd-ecomerce-api/internal/domain"
)

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) Create(_ context.Context) (domain.Cart, error) {
	args := m.Called()
	return args.Get(0).(domain.Cart), args.Error(1)
}

func (m *MockCartRepository) FindByID(_ context.Context, cartID uuid.UUID) (domain.Cart, error) {
	args := m.Called(cartID)
	return args.Get(0).(domain.Cart), args.Error(1)
}

func (m *MockCartRepository) SaveItem(_ context.Context, cartID, productID uuid.UUID, quantity int) error {
	args := m.Called(cartID, productID, quantity)
	return args.Error(0)
}

func (m *MockCartRepository) RemoveItem(_ context.Context, cartID, productID uuid.UUID) error {
	args := m.Called(cartID, productID)
	return args.Error(0)
}

type MockProductReader struct {
	mock.Mock
}

func (m *MockProductReader) FindByID(_ context.Context, id uuid.UUID) (domain.Product, error) {
	args := m.Called(id)
	return args.Get(0).(domain.Product), args.Error(1)
}
