package bootstrap

import (
	"ntd-ecomerce-api/internal/bootstrap/product"
	"ntd-ecomerce-api/internal/bootstrap/registry"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupComponents(r *gin.Engine, db *gorm.DB) {
	reg := registry.NewRegistry(db)

	// here should be configured the authentication middleware
	product.Setup(r, reg)
}
