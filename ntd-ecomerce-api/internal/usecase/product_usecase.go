package usecase

import (
	"context"
	"fmt"
	"strings"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
)

type (
	ProductRepository interface {
		Add(ctx context.Context, product domain.Product) (domain.Product, error)
		FindAll(ctx context.Context, page domain.Page) (domain.ProductList, error)
		FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
		Update(ctx context.Context, product domain.Product) (domain.Product, error)
		Delete(ctx context.Context, id uuid.UUID) error
	}

	Product struct {
		repo ProductRepository
	}
)

func NewProduct(repo ProductRepository) Product {
	return Product{repo: repo}
}

func (u *Product) Add(ctx context.Context, input domain.ProductInput) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Product{}, domain.WrapValidation(ErrInvalidProductInput, problems)
	}

	created, err := u.repo.Add(ctx, input.ToProduct())
	if err != nil {
		return domain.Product{}, fmt.Errorf("adding product: %w", err)
	}

	return created, nil
}

func (u *Product) FindAll(ctx context.Context, page domain.Page) (domain.ProductList, error) {
	list, err := u.repo.FindAll(ctx, page)
	if err != nil {
		return domain.ProductList{}, fmt.Errorf("listing products: %w", err)
	}

	return list, nil
}

func (u *Product) FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Product{}, fmt.Errorf("finding product: %w", err)
	}

	return product, nil
}

func (u *Product) Update(ctx context.Context, id uuid.UUID, input domain.ProductInput) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Product{}, domain.WrapValidation(ErrInvalidProductInput, problems)
	}

	product := input.ToProduct()
	product.ID = id

	updated, err := u.repo.Update(ctx, product)
	if err != nil {
		return domain.Product{}, fmt.Errorf("updating product: %w", err)
	}

	return updated, nil
}

func (u *Product) DeleteOne(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}

	return nil
}
