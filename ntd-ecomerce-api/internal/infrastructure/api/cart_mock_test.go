package api_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ntd-ecomerce-api/internal/domain"
)

type MockCartUsecase struct {
	mock.Mock
}

func (m *MockCartUsecase) Create(_ context.Context) (domain.Cart, error) {
	args := m.Called()
	return args.Get(0).(domain.Cart), args.Error(1)
}

func (m *MockCartUsecase) Get(_ context.Context, cartID uuid.UUID) (domain.Cart, error) {
	args := m.Called(cartID)
	return args.Get(0).(domain.Cart), args.Error(1)
}

func (m *MockCartUsecase) AddItem(_ context.Context, cartID uuid.UUID, input domain.AddItemInput) (domain.Cart, error) {
	args := m.Called(cartID, input)
	return args.Get(0).(domain.Cart), args.Error(1)
}

func (m *MockCartUsecase) SetItem(_ context.Context, cartID, productID uuid.UUID, input domain.SetItemInput) (domain.Cart, error) {
	args := m.Called(cartID, productID, input)
	return args.Get(0).(domain.Cart), args.Error(1)
}

func (m *MockCartUsecase) RemoveItem(_ context.Context, cartID, productID uuid.UUID) (domain.Cart, error) {
	args := m.Called(cartID, productID)
	return args.Get(0).(domain.Cart), args.Error(1)
}
