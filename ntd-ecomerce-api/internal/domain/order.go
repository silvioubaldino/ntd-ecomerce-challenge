package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	// OrderStatusConfirmed is the single terminal Order status in the MVP.
	OrderStatusConfirmed = "confirmed"
	// PaymentMethodSimulated / PaymentStatusApproved describe the always-approved
	// simulated payment (no real provider, no decline path in the MVP).
	PaymentMethodSimulated = "simulated"
	PaymentStatusApproved  = "approved"
)

// emailPattern is a deliberately basic email check (RF-04 captures minimal contact only).
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

// CheckoutInput is the POST /orders request body.
type CheckoutInput struct {
	CartID   uuid.UUID `json:"cart_id"`
	Customer Customer  `json:"customer"`
}

// Validate reports missing/invalid customer contact fields (RF-04: minimal guest contact).
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

// ApprovedPayment returns the simulated, always-approved payment (RF-04).
func ApprovedPayment() Payment {
	return Payment{Method: PaymentMethodSimulated, Status: PaymentStatusApproved}
}

// Recalculate fills each item's subtotal from its snapshot unit_price and quantity, then
// sets the order total as the (immutable) sum of the subtotals.
func (o *Order) Recalculate() {
	total := decimal.Zero
	for i := range o.Items {
		subtotal := o.Items[i].UnitPrice.Mul(decimal.NewFromInt(int64(o.Items[i].Quantity)))
		o.Items[i].Subtotal = subtotal
		total = total.Add(subtotal)
	}
	o.Total = total
}
