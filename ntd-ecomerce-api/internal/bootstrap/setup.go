package bootstrap

import (
	"ntd-ecomerce-api/internal/bootstrap/product"
	"ntd-ecomerce-api/internal/bootstrap/registry"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupComponents wires every feature's handler/usecase/repository chain onto
// the given engine. There is no authentication in the MVP — every route below
// is public (guest-only scope, see requirements.md).
func SetupComponents(r *gin.Engine, db *gorm.DB) {
	reg := registry.NewRegistry(db)

	product.Setup(r, reg)
}
