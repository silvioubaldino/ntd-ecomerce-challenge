package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Cart struct {
	ID        uuid.UUID       `json:"id"`
	Items     []CartItem      `json:"items"`
	Total     decimal.Decimal `json:"total"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CartItem struct {
	ProductID uuid.UUID       `json:"product_id"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int             `json:"quantity"`
	Subtotal  decimal.Decimal `json:"subtotal"`
}

type AddItemInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

func (in AddItemInput) Validate() map[string]string {
	return validateQuantity(in.Quantity)
}

type SetItemInput struct {
	Quantity int `json:"quantity"`
}

func (in SetItemInput) Validate() map[string]string {
	return validateQuantity(in.Quantity)
}

func validateQuantity(quantity int) map[string]string {
	if quantity < 1 {
		return map[string]string{"quantity": "must_be_positive_integer"}
	}
	return nil
}

func (c *Cart) Recalculate() {
	total := decimal.Zero
	for i := range c.Items {
		subtotal := c.Items[i].UnitPrice.Mul(decimal.NewFromInt(int64(c.Items[i].Quantity)))
		c.Items[i].Subtotal = subtotal
		total = total.Add(subtotal)
	}
	c.Total = total
}
