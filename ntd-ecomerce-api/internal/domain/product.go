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

// Product is the catalog resource, per AYD-001@context.
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

// ProductList is the paginated list response shape for GET /products.
type ProductList struct {
	Data       []Product  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// ProductInput is Product minus the server-generated fields (id, created_at,
// updated_at) — the payload for POST/PUT.
type ProductInput struct {
	Name        string          `json:"name"`
	SKU         string          `json:"sku"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	WeightKg    decimal.Decimal `json:"weight_kg"`
}

// Validate returns a map of field -> problem code for every invalid field, or an
// empty map when the input is valid. Codes feed the AYD-001 validation_error
// envelope's `details`.
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

// ToProduct builds a Product from the input. ID and timestamps are left zero —
// the repository is responsible for generating/stamping them.
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
