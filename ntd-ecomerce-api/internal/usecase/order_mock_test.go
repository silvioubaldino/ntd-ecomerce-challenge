package usecase_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ntd-ecomerce-api/internal/domain"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Checkout(_ context.Context, cartID uuid.UUID, customer domain.Customer) (domain.Order, error) {
	args := m.Called(cartID, customer)
	return args.Get(0).(domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByID(_ context.Context, orderID uuid.UUID) (domain.Order, error) {
	args := m.Called(orderID)
	return args.Get(0).(domain.Order), args.Error(1)
}
