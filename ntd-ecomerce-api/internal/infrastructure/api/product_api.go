package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ntd-ecomerce-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var errInvalidProductFilter = errors.New("invalid product filter")

type (
	ProductUsecase interface {
		Add(ctx context.Context, input domain.ProductInput) (domain.Product, error)
		FindAll(ctx context.Context, filter domain.ProductFilter, page domain.PageRequest) (domain.ProductList, error)
		FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
		Update(ctx context.Context, id uuid.UUID, input domain.ProductInput) (domain.Product, error)
		DeleteOne(ctx context.Context, id uuid.UUID) error
		Import(ctx context.Context, r io.Reader) (domain.ImportReport, error)
		FindCategories(ctx context.Context) ([]string, error)
	}

	ProductHandler struct {
		usecase ProductUsecase
	}

	categoryListResponse struct {
		Data []string `json:"data"`
	}
)

func NewProductHandlers(r *gin.Engine, srv ProductUsecase) {
	handler := ProductHandler{usecase: srv}

	group := r.Group("/products")
	group.POST("", handler.Add())
	group.GET("", handler.FindAll())
	group.GET("/categories", handler.Categories())
	group.GET("/:id", handler.FindByID())
	group.PUT("/:id", handler.Update())
	group.DELETE("/:id", handler.DeleteOne())
	group.POST("/import", handler.Import())
}

func (h ProductHandler) Add() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var input domain.ProductInput
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "invalid json body"))
			return
		}

		created, err := h.usecase.Add(ctx, input)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusCreated, created)
	}
}

func (h ProductHandler) FindAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		filter, problems := h.parseFilter(c)

		page, pageProblems := h.parsePagination(c, filter)
		for field, code := range pageProblems {
			problems[field] = code
		}

		if len(problems) > 0 {
			HandleErr(c, domain.WrapValidation(errInvalidProductFilter, problems))
			return
		}

		list, err := h.usecase.FindAll(ctx, filter, page)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, list)
	}
}

func (h ProductHandler) Categories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		categories, err := h.usecase.FindCategories(ctx)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, categoryListResponse{Data: categories})
	}
}

func (h ProductHandler) FindByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "id must be valid"))
			return
		}

		product, err := h.usecase.FindByID(ctx, id)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

func (h ProductHandler) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "id must be valid"))
			return
		}

		var input domain.ProductInput
		if err := c.ShouldBindJSON(&input); err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "invalid json body"))
			return
		}

		updated, err := h.usecase.Update(ctx, id, input)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, updated)
	}
}

func (h ProductHandler) DeleteOne() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			HandleErr(c, domain.WrapInvalidInput(err, "id must be valid"))
			return
		}

		if err := h.usecase.DeleteOne(ctx, id); err != nil {
			HandleErr(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func (h ProductHandler) Import() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		fileHeader, err := c.FormFile("file")
		if err != nil {
			HandleErr(c, domain.WrapBadRequest(err, "invalid_file", "file is required"))
			return
		}
		if fileHeader.Size == 0 {
			HandleErr(c, domain.WrapBadRequest(errors.New("empty file"), "invalid_file", "file must not be empty"))
			return
		}
		if fileHeader.Size > domain.MaxImportFileBytes {
			HandleErr(c, domain.WrapPayloadTooLarge(errors.New("file too large"), "file_too_large", "file exceeds the maximum allowed size"))
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			HandleErr(c, domain.WrapBadRequest(err, "invalid_file", "file could not be read"))
			return
		}
		defer file.Close()

		report, err := h.usecase.Import(ctx, file)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

func (h ProductHandler) parseFilter(c *gin.Context) (domain.ProductFilter, map[string]string) {
	filter := domain.ProductFilter{
		Query:    strings.TrimSpace(c.Query("q")),
		Category: strings.TrimSpace(c.Query("category")),
		Sort:     domain.ProductSort(strings.TrimSpace(c.Query("sort"))),
	}

	problems := map[string]string{}

	if raw := strings.TrimSpace(c.Query("price_min")); raw != "" {
		if v, err := decimal.NewFromString(raw); err != nil {
			problems["price_min"] = "must_be_non_negative_decimal"
		} else {
			filter.PriceMin = &v
		}
	}

	if raw := strings.TrimSpace(c.Query("price_max")); raw != "" {
		if v, err := decimal.NewFromString(raw); err != nil {
			problems["price_max"] = "must_be_non_negative_decimal"
		} else {
			filter.PriceMax = &v
		}
	}

	for field, code := range filter.Validate() {
		problems[field] = code
	}

	return filter, problems
}

func (h ProductHandler) parsePagination(c *gin.Context, filter domain.ProductFilter) (domain.PageRequest, map[string]string) {
	page := domain.DefaultPageRequest()
	problems := map[string]string{}

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			problems["limit"] = "must_be_between_1_and_100"
		} else if code, ok := domain.ValidateLimit(n); !ok {
			problems["limit"] = code
		} else {
			page.Limit = n
		}
	}

	if raw := strings.TrimSpace(c.Query("cursor")); raw != "" {
		cursor, err := domain.DecodeCursor(raw, filter.EffectiveSort())
		if err != nil {
			problems["cursor"] = "invalid_cursor"
		} else {
			page.Cursor = &cursor
		}
	}

	return page, problems
}
