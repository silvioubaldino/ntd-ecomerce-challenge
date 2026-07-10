package registry

import (
	"ntd-ecomerce-api/internal/infrastructure/repository"

	"gorm.io/gorm"
)

type Registry struct {
	db *gorm.DB

	productRepository *repository.ProductRepository
	cartRepository    *repository.CartRepository
	orderRepository   *repository.OrderRepository
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

func (r *Registry) GetOrderRepository() *repository.OrderRepository {
	if r.orderRepository == nil {
		r.orderRepository = repository.NewOrderRepository(r.db)
	}
	return r.orderRepository
}
