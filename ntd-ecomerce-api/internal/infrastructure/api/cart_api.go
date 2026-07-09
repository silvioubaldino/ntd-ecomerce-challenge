package api

import (
	"context"
	"net/http"

	"ntd-ecomerce-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	CartUsecase interface {
		Create(ctx context.Context) (domain.Cart, error)
		Get(ctx context.Context, cartID uuid.UUID) (domain.Cart, error)
		AddItem(ctx context.Context, cartID uuid.UUID, input domain.AddItemInput) (domain.Cart, error)
		SetItem(ctx context.Context, cartID, productID uuid.UUID, input domain.SetItemInput) (domain.Cart, error)
		RemoveItem(ctx context.Context, cartID, productID uuid.UUID) (domain.Cart, error)
	}

	CartHandler struct {
		usecase CartUsecase
	}
)

func NewCartHandlers(r *gin.Engine, srv CartUsecase) {
	handler := CartHandler{usecase: srv}

	group := r.Group("/carts")
	group.POST("", handler.Create())
	group.GET("/:cart_id", handler.Get())
	group.POST("/:cart_id/items", handler.AddItem())
	group.PUT("/:cart_id/items/:product_id", handler.SetItem())
	group.DELETE("/:cart_id/items/:product_id", handler.RemoveItem())
}

func (h CartHandler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		cart, err := h.usecase.Create(c.Request.Context())
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusCreated, cart)
	}
}

func (h CartHandler) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID, err := parseUUIDParam(c, "cart_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		cart, err := h.usecase.Get(c.Request.Context(), cartID)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, cart)
	}
}

func (h CartHandler) AddItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID, err := parseUUIDParam(c, "cart_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		var input domain.AddItemInput
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "invalid json body"))
			return
		}

		cart, err := h.usecase.AddItem(c.Request.Context(), cartID, input)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, cart)
	}
}

func (h CartHandler) SetItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID, err := parseUUIDParam(c, "cart_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		productID, err := parseUUIDParam(c, "product_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		var input domain.SetItemInput
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "invalid json body"))
			return
		}

		cart, err := h.usecase.SetItem(c.Request.Context(), cartID, productID, input)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, cart)
	}
}

func (h CartHandler) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID, err := parseUUIDParam(c, "cart_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		productID, err := parseUUIDParam(c, "product_id")
		if err != nil {
			HandleErr(c, err)
			return
		}

		cart, err := h.usecase.RemoveItem(c.Request.Context(), cartID, productID)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, cart)
	}
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, domain.WrapInvalidInput(err, name+" must be valid")
	}
	return id, nil
}
