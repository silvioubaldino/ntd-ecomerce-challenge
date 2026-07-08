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
)

type (
	ProductUsecase interface {
		Add(ctx context.Context, input domain.ProductInput) (domain.Product, error)
		FindAll(ctx context.Context, page domain.Page) (domain.ProductList, error)
		FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
		Update(ctx context.Context, id uuid.UUID, input domain.ProductInput) (domain.Product, error)
		DeleteOne(ctx context.Context, id uuid.UUID) error
		Import(ctx context.Context, r io.Reader) (domain.ImportReport, error)
	}

	ProductHandler struct {
		usecase ProductUsecase
	}
)

func NewProductHandlers(r *gin.Engine, srv ProductUsecase) {
	handler := ProductHandler{usecase: srv}

	group := r.Group("/products")
	group.POST("", handler.Add())
	group.GET("", handler.FindAll())
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

		page, err := h.parsePage(c)
		if err != nil {
			HandleErr(c, err)
			return
		}

		list, err := h.usecase.FindAll(ctx, page)
		if err != nil {
			HandleErr(c, err)
			return
		}

		c.JSON(http.StatusOK, list)
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

func (h ProductHandler) parsePage(c *gin.Context) (domain.Page, error) {
	page := domain.DefaultPage()

	if pageString := c.Query("page"); pageString != "" {
		n, err := strconv.Atoi(pageString)
		if err != nil {
			return domain.Page{}, domain.WrapInvalidInput(err, "page must be a positive integer")
		}
		page.Number = n
	}

	if sizeString := c.Query("page_size"); sizeString != "" {
		n, err := strconv.Atoi(sizeString)
		if err != nil {
			return domain.Page{}, domain.WrapInvalidInput(err, "page_size must be a positive integer")
		}
		page.Size = n
	}

	page.Query = strings.TrimSpace(c.Query("q"))

	if err := page.Validate(); err != nil {
		return domain.Page{}, domain.WrapInvalidInput(err, "invalid pagination")
	}

	return page, nil
}
