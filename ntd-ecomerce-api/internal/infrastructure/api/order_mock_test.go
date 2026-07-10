package api_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ntd-ecomerce-api/internal/domain"
)

type MockOrderUsecase struct {
	mock.Mock
}

func (m *MockOrderUsecase) Checkout(_ context.Context, input domain.CheckoutInput) (domain.Order, error) {
	args := m.Called(input)
	return args.Get(0).(domain.Order), args.Error(1)
}

func (m *MockOrderUsecase) Get(_ context.Context, orderID uuid.UUID) (domain.Order, error) {
	args := m.Called(orderID)
	return args.Get(0).(domain.Order), args.Error(1)
}
