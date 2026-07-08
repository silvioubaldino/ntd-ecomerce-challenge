package product

import (
	"ntd-ecomerce-api/internal/bootstrap/registry"
	"ntd-ecomerce-api/internal/infrastructure/api"
	"ntd-ecomerce-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, reg *registry.Registry) {
	repo := reg.GetProductRepository()
	svc := usecase.NewProduct(repo)
	api.NewProductHandlers(r, &svc)
}
