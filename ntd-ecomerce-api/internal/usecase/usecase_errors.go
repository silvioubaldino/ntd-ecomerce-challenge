package usecase

import "errors"

var (
	ErrInvalidProductInput  = errors.New("product input is invalid")
	ErrInvalidCartInput     = errors.New("cart input is invalid")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrItemNotFound         = errors.New("cart item not found")
	ErrInvalidCustomerInput = errors.New("customer input is invalid")
	ErrCartEmpty            = errors.New("cart is empty")
	ErrOrderNotFound        = errors.New("order not found")
)
