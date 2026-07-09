package registry

import (
	"ntd-ecomerce-api/internal/infrastructure/repository"

	"gorm.io/gorm"
)

type Registry struct {
	db *gorm.DB

	productRepository *repository.ProductRepository
	cartRepository    *repository.CartRepository
}

func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{db: db}
}

func (r *Registry) GetProductRepository() *repository.ProductRepository {
	if r.productRepository == nil {
		r.productRepository = repository.NewProductRepository(r.db)
	}
	return r.productRepository
}

func (r *Registry) GetCartRepository() *repository.CartRepository {
	if r.cartRepository == nil {
		r.cartRepository = repository.NewCartRepository(r.db)
	}
	return r.cartRepository
}
