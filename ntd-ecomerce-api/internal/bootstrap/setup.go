package bootstrap

import (
	"ntd-ecomerce-api/internal/bootstrap/cart"
	"ntd-ecomerce-api/internal/bootstrap/order"
	"ntd-ecomerce-api/internal/bootstrap/product"
	"ntd-ecomerce-api/internal/bootstrap/registry"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupComponents(r *gin.Engine, db *gorm.DB) {
	reg := registry.NewRegistry(db)

	product.Setup(r, reg)
	cart.Setup(r, reg)
	order.Setup(r, reg)
}
