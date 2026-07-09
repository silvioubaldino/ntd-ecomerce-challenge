package usecase

import "errors"

var (
	ErrInvalidProductInput = errors.New("product input is invalid")
	ErrInvalidCartInput    = errors.New("cart input is invalid")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrItemNotFound        = errors.New("cart item not found")
)
