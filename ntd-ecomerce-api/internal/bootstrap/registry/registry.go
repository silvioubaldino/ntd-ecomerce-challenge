package registry

import (
	"ntd-ecomerce-api/internal/infrastructure/repository"

	"gorm.io/gorm"
)

// Registry is process-lifetime: it holds lazily-constructed repositories,
// never per-request state.
type Registry struct {
	db *gorm.DB

	productRepository *repository.ProductRepository
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
