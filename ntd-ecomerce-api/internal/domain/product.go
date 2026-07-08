package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	maxNameLength     = 255
	maxSKULength      = 64
	maxCategoryLength = 100
)

type Product struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	SKU         string          `json:"sku"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	WeightKg    decimal.Decimal `json:"weight_kg"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ProductList struct {
	Data       []Product  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type ProductInput struct {
	Name        string          `json:"name"`
	SKU         string          `json:"sku"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	WeightKg    decimal.Decimal `json:"weight_kg"`
}

func (in ProductInput) Validate() map[string]string {
	problems := map[string]string{}

	switch {
	case in.Name == "":
		problems["name"] = "required"
	case len(in.Name) > maxNameLength:
		problems["name"] = "too_long"
	}

	switch {
	case in.SKU == "":
		problems["sku"] = "required"
	case len(in.SKU) > maxSKULength:
		problems["sku"] = "too_long"
	}

	switch {
	case in.Category == "":
		problems["category"] = "required"
	case len(in.Category) > maxCategoryLength:
		problems["category"] = "too_long"
	}

	if in.Price.IsNegative() {
		problems["price"] = "must_be_non_negative_decimal"
	}

	if in.Stock < 0 {
		problems["stock"] = "must_be_non_negative_integer"
	}

	if in.WeightKg.IsNegative() {
		problems["weight_kg"] = "must_be_non_negative_decimal"
	}

	return problems
}

func (in ProductInput) ToProduct() Product {
	return Product{
		Name:        in.Name,
		SKU:         in.SKU,
		Description: in.Description,
		Category:    in.Category,
		Price:       in.Price,
		Stock:       in.Stock,
		WeightKg:    in.WeightKg,
	}
}
