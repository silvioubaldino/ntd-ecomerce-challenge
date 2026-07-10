package cart

import (
	"ntd-ecomerce-api/internal/bootstrap/registry"
	"ntd-ecomerce-api/internal/infrastructure/api"
	"ntd-ecomerce-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, reg *registry.Registry) {
	cartRepo := reg.GetCartRepository()
	productRepo := reg.GetProductRepository()
	svc := usecase.NewCart(cartRepo, productRepo)
	api.NewCartHandlers(r, &svc)
}
