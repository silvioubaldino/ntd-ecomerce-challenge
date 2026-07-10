package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	OrderStatusConfirmed = "confirmed"

	PaymentMethodSimulated = "simulated"
	PaymentStatusApproved  = "approved"
)

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type Order struct {
	ID        uuid.UUID       `json:"id"`
	Status    string          `json:"status"`
	Customer  Customer        `json:"customer"`
	Items     []OrderItem     `json:"items"`
	Total     decimal.Decimal `json:"total"`
	Payment   Payment         `json:"payment"`
	CreatedAt time.Time       `json:"created_at"`
}

type OrderItem struct {
	ProductID uuid.UUID       `json:"product_id"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int             `json:"quantity"`
	Subtotal  decimal.Decimal `json:"subtotal"`
}

type Customer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Payment struct {
	Method string `json:"method"`
	Status string `json:"status"`
}

type CheckoutInput struct {
	CartID   uuid.UUID `json:"cart_id"`
	Customer Customer  `json:"customer"`
}

func (c Customer) Validate() map[string]string {
	problems := map[string]string{}

	if c.Name == "" {
		problems["name"] = "required"
	}

	switch {
	case c.Email == "":
		problems["email"] = "required"
	case !emailPattern.MatchString(c.Email):
		problems["email"] = "invalid"
	}

	if len(problems) == 0 {
		return nil
	}
	return problems
}

func ApprovedPayment() Payment {
	return Payment{Method: PaymentMethodSimulated, Status: PaymentStatusApproved}
}

func (o *Order) Recalculate() {
	total := decimal.Zero
	for i := range o.Items {
		subtotal := o.Items[i].UnitPrice.Mul(decimal.NewFromInt(int64(o.Items[i].Quantity)))
		o.Items[i].Subtotal = subtotal
		total = total.Add(subtotal)
	}
	o.Total = total
}
