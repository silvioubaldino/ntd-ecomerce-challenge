package order

import (
	"ntd-ecomerce-api/internal/bootstrap/registry"
	"ntd-ecomerce-api/internal/infrastructure/api"
	"ntd-ecomerce-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, reg *registry.Registry) {
	orderRepo := reg.GetOrderRepository()
	svc := usecase.NewOrder(orderRepo)
	api.NewOrderHandlers(r, &svc)
}
