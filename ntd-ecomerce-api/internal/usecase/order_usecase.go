package usecase

import (
	"context"
	"fmt"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
)

type (
	OrderRepository interface {
		Checkout(ctx context.Context, cartID uuid.UUID, customer domain.Customer) (domain.Order, error)
		FindByID(ctx context.Context, orderID uuid.UUID) (domain.Order, error)
	}

	Order struct {
		orders OrderRepository
	}
)

func NewOrder(orders OrderRepository) Order {
	return Order{orders: orders}
}

func (u *Order) Checkout(ctx context.Context, input domain.CheckoutInput) (domain.Order, error) {
	if problems := input.Customer.Validate(); len(problems) > 0 {
		return domain.Order{}, domain.WrapValidation(ErrInvalidCustomerInput, problems)
	}

	order, err := u.orders.Checkout(ctx, input.CartID, input.Customer)
	if err != nil {
		return domain.Order{}, fmt.Errorf("checking out cart: %w", err)
	}

	order.Recalculate()
	return order, nil
}

func (u *Order) Get(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	order, err := u.orders.FindByID(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("finding order: %w", err)
	}

	order.Recalculate()
	return order, nil
}
