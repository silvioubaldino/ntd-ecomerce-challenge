package usecase

import (
	"context"
	"fmt"
	"strconv"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
)

type (
	CartRepository interface {
		Create(ctx context.Context) (domain.Cart, error)
		FindByID(ctx context.Context, cartID uuid.UUID) (domain.Cart, error)
		SaveItem(ctx context.Context, cartID, productID uuid.UUID, quantity int) error
		RemoveItem(ctx context.Context, cartID, productID uuid.UUID) error
	}

	ProductReader interface {
		FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
	}

	Cart struct {
		carts    CartRepository
		products ProductReader
	}
)

func NewCart(carts CartRepository, products ProductReader) Cart {
	return Cart{carts: carts, products: products}
}

func (u *Cart) Create(ctx context.Context) (domain.Cart, error) {
	cart, err := u.carts.Create(ctx)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("creating cart: %w", err)
	}

	cart.Recalculate()
	return cart, nil
}

func (u *Cart) Get(ctx context.Context, cartID uuid.UUID) (domain.Cart, error) {
	cart, err := u.carts.FindByID(ctx, cartID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("finding cart: %w", err)
	}

	if err := u.enrich(ctx, &cart); err != nil {
		return domain.Cart{}, err
	}

	return cart, nil
}

func (u *Cart) AddItem(ctx context.Context, cartID uuid.UUID, input domain.AddItemInput) (domain.Cart, error) {
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Cart{}, domain.WrapValidation(ErrInvalidCartInput, problems)
	}

	cart, err := u.carts.FindByID(ctx, cartID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("finding cart: %w", err)
	}

	product, err := u.products.FindByID(ctx, input.ProductID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("finding product: %w", err)
	}

	resulting := existingQuantity(cart, input.ProductID) + input.Quantity
	if resulting > product.Stock {
		return domain.Cart{}, insufficientStock(product.ID, resulting, product.Stock)
	}

	if err := u.carts.SaveItem(ctx, cartID, input.ProductID, resulting); err != nil {
		return domain.Cart{}, fmt.Errorf("saving cart item: %w", err)
	}

	return u.Get(ctx, cartID)
}

func (u *Cart) SetItem(ctx context.Context, cartID, productID uuid.UUID, input domain.SetItemInput) (domain.Cart, error) {
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Cart{}, domain.WrapValidation(ErrInvalidCartInput, problems)
	}

	cart, err := u.carts.FindByID(ctx, cartID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("finding cart: %w", err)
	}

	if !hasItem(cart, productID) {
		return domain.Cart{}, domain.WrapNotFound(ErrItemNotFound, "item_not_found", "cart item not found")
	}

	product, err := u.products.FindByID(ctx, productID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("finding product: %w", err)
	}

	if input.Quantity > product.Stock {
		return domain.Cart{}, insufficientStock(product.ID, input.Quantity, product.Stock)
	}

	if err := u.carts.SaveItem(ctx, cartID, productID, input.Quantity); err != nil {
		return domain.Cart{}, fmt.Errorf("saving cart item: %w", err)
	}

	return u.Get(ctx, cartID)
}

func (u *Cart) RemoveItem(ctx context.Context, cartID, productID uuid.UUID) (domain.Cart, error) {
	if _, err := u.carts.FindByID(ctx, cartID); err != nil {
		return domain.Cart{}, fmt.Errorf("finding cart: %w", err)
	}

	if err := u.carts.RemoveItem(ctx, cartID, productID); err != nil {
		return domain.Cart{}, fmt.Errorf("removing cart item: %w", err)
	}

	return u.Get(ctx, cartID)
}

func (u *Cart) enrich(ctx context.Context, cart *domain.Cart) error {
	for i := range cart.Items {
		product, err := u.products.FindByID(ctx, cart.Items[i].ProductID)
		if err != nil {
			return fmt.Errorf("loading product for cart item: %w", err)
		}
		cart.Items[i].SKU = product.SKU
		cart.Items[i].Name = product.Name
		cart.Items[i].UnitPrice = product.Price
	}

	cart.Recalculate()
	return nil
}

func existingQuantity(cart domain.Cart, productID uuid.UUID) int {
	for _, item := range cart.Items {
		if item.ProductID == productID {
			return item.Quantity
		}
	}
	return 0
}

func hasItem(cart domain.Cart, productID uuid.UUID) bool {
	for _, item := range cart.Items {
		if item.ProductID == productID {
			return true
		}
	}
	return false
}

func insufficientStock(productID uuid.UUID, requested, available int) error {
	return domain.WrapConflictDetails(ErrInsufficientStock, "insufficient_stock", "insufficient stock", map[string]string{
		"product_id": productID.String(),
		"requested":  strconv.Itoa(requested),
		"available":  strconv.Itoa(available),
	})
}
