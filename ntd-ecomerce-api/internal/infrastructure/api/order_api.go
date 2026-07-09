package api

import (
	"context"
	"net/http"

	"ntd-ecomerce-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	OrderUsecase interface {
		Checkout(ctx context.Context, input domain.CheckoutInput) (domain.Order, error)
		Get(ctx context.Context, orderID uuid.UUID) (domain.Order, error)
	}

	OrderHandler struct {
		usecase OrderUsecase
	}
)

func NewOrderHandlers(r *gin.Engine, srv OrderUsecase) {
	handler := OrderHandler{usecase: srv}

	group := r.Group("/orders")
	group.POST("", handler.Checkout())
	group.GET("/:order_id", handler.Get())
}

func (h OrderHandler) Checkout() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input domain.CheckoutInput
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "invalid json body"))
			return
		}

		order, err := h.usecase.Checkout(c.Request.Context(), input)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusCreated, order)
	}
}

func (h OrderHandler) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID, err := parseUUIDParam(c, "order_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		order, err := h.usecase.Get(c.Request.Context(), orderID)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, order)
	}
}
