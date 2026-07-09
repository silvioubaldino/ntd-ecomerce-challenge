package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Local sentinels used only as the underlying cause of the wrapped domain errors below.
var (
	errCartEmpty         = errors.New("cart is empty")
	errInsufficientStock = errors.New("insufficient stock")
)

type orderModel struct {
	ID            uuid.UUID       `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Status        string          `gorm:"column:status"`
	CustomerName  string          `gorm:"column:customer_name"`
	CustomerEmail string          `gorm:"column:customer_email"`
	Total         decimal.Decimal `gorm:"column:total"`
	PaymentMethod string          `gorm:"column:payment_method"`
	PaymentStatus string          `gorm:"column:payment_status"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
}

func (orderModel) TableName() string { return "orders" }

type orderItemModel struct {
	OrderID   uuid.UUID       `gorm:"column:order_id;type:uuid;primaryKey"`
	ProductID uuid.UUID       `gorm:"column:product_id;type:uuid;primaryKey"`
	SKU       string          `gorm:"column:sku"`
	Name      string          `gorm:"column:name"`
	UnitPrice decimal.Decimal `gorm:"column:unit_price"`
	Quantity  int             `gorm:"column:quantity"`
	Subtotal  decimal.Decimal `gorm:"column:subtotal"`
	CreatedAt time.Time       `gorm:"column:created_at"`
}

func (orderItemModel) TableName() string { return "order_items" }

func (m orderModel) toDomain(items []orderItemModel) domain.Order {
	domainItems := make([]domain.OrderItem, 0, len(items))
	for _, item := range items {
		domainItems = append(domainItems, domain.OrderItem{
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			Subtotal:  item.Subtotal,
		})
	}

	return domain.Order{
		ID:        m.ID,
		Status:    m.Status,
		Customer:  domain.Customer{Name: m.CustomerName, Email: m.CustomerEmail},
		Items:     domainItems,
		Total:     m.Total,
		Payment:   domain.Payment{Method: m.PaymentMethod, Status: m.PaymentStatus},
		CreatedAt: m.CreatedAt,
	}
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Checkout runs the whole checkout atomically (RN-04): load the cart, re-check every
// line's stock under a row lock (RN-03), snapshot price/sku/name, decrement stock, insert
// the confirmed order and its items, and consume (delete) the cart. Any error rolls the
// whole transaction back, so nothing is committed on a shortfall.
func (r *OrderRepository) Checkout(ctx context.Context, cartID uuid.UUID, customer domain.Customer) (domain.Order, error) {
	var result domain.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart cartModel
		err := tx.Where("id = ?", cartID).First(&cart).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.WrapNotFound(err, "cart_not_found", "cart not found")
		}
		if err != nil {
			return fmt.Errorf("finding cart: %w", err)
		}

		var items []cartItemModel
		if err := tx.Where("cart_id = ?", cartID).Order("created_at asc").Find(&items).Error; err != nil {
			return fmt.Errorf("finding cart items: %w", err)
		}
		if len(items) == 0 {
			return domain.WrapValidationCode(errCartEmpty, "cart_empty", "cart is empty")
		}

		productIDs := make([]uuid.UUID, 0, len(items))
		for _, item := range items {
			productIDs = append(productIDs, item.ProductID)
		}

		var products []productModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", productIDs).
			Order("id").
			Find(&products).Error; err != nil {
			return fmt.Errorf("locking products: %w", err)
		}

		byID := make(map[uuid.UUID]productModel, len(products))
		for _, product := range products {
			byID[product.ID] = product
		}

		shortfalls := map[string]string{}
		for _, item := range items {
			product, ok := byID[item.ProductID]
			available := 0
			if ok {
				available = product.Stock
			}
			if !ok || item.Quantity > available {
				shortfalls[item.ProductID.String()] = fmt.Sprintf("requested=%d, available=%d", item.Quantity, available)
			}
		}
		if len(shortfalls) > 0 {
			return domain.WrapConflictDetails(errInsufficientStock, "insufficient_stock", "insufficient stock", shortfalls)
		}

		now := time.Now().UTC()
		orderID := uuid.New()
		total := decimal.Zero
		orderItems := make([]orderItemModel, 0, len(items))
		domainItems := make([]domain.OrderItem, 0, len(items))

		for _, item := range items {
			product := byID[item.ProductID]
			subtotal := product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			total = total.Add(subtotal)

			orderItems = append(orderItems, orderItemModel{
				OrderID:   orderID,
				ProductID: product.ID,
				SKU:       product.SKU,
				Name:      product.Name,
				UnitPrice: product.Price,
				Quantity:  item.Quantity,
				Subtotal:  subtotal,
				CreatedAt: now,
			})
			domainItems = append(domainItems, domain.OrderItem{
				ProductID: product.ID,
				SKU:       product.SKU,
				Name:      product.Name,
				UnitPrice: product.Price,
				Quantity:  item.Quantity,
				Subtotal:  subtotal,
			})

			if err := tx.Model(&productModel{}).
				Where("id = ?", product.ID).
				Update("stock", product.Stock-item.Quantity).Error; err != nil {
				return fmt.Errorf("decrementing stock: %w", err)
			}
		}

		payment := domain.ApprovedPayment()
		order := orderModel{
			ID:            orderID,
			Status:        domain.OrderStatusConfirmed,
			CustomerName:  customer.Name,
			CustomerEmail: customer.Email,
			Total:         total,
			PaymentMethod: payment.Method,
			PaymentStatus: payment.Status,
			CreatedAt:     now,
		}
		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("creating order: %w", err)
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return fmt.Errorf("creating order items: %w", err)
		}

		if err := tx.Where("id = ?", cartID).Delete(&cartModel{}).Error; err != nil {
			return fmt.Errorf("consuming cart: %w", err)
		}

		result = domain.Order{
			ID:        orderID,
			Status:    domain.OrderStatusConfirmed,
			Customer:  customer,
			Items:     domainItems,
			Total:     total,
			Payment:   payment,
			CreatedAt: now,
		}
		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}

	return result, nil
}

func (r *OrderRepository) FindByID(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	var order orderModel

	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Order{}, domain.WrapNotFound(err, "order_not_found", "order not found")
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("finding order: %w", err)
	}

	var items []orderItemModel
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at asc").
		Find(&items).Error; err != nil {
		return domain.Order{}, fmt.Errorf("finding order items: %w", err)
	}

	return order.toDomain(items), nil
}
