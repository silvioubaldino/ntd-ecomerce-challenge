package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type cartModel struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (cartModel) TableName() string { return "carts" }

type cartItemModel struct {
	CartID    uuid.UUID `gorm:"column:cart_id;type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"column:product_id;type:uuid;primaryKey"`
	Quantity  int       `gorm:"column:quantity"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (cartItemModel) TableName() string { return "cart_items" }

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) Create(ctx context.Context) (domain.Cart, error) {
	now := time.Now().UTC()
	model := cartModel{CreatedAt: now, UpdatedAt: now}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Cart{}, fmt.Errorf("creating cart: %w", err)
	}

	return domain.Cart{
		ID:        model.ID,
		Items:     []domain.CartItem{},
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *CartRepository) FindByID(ctx context.Context, cartID uuid.UUID) (domain.Cart, error) {
	var cart cartModel

	err := r.db.WithContext(ctx).Where("id = ?", cartID).First(&cart).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Cart{}, domain.WrapNotFound(err, "cart_not_found", "cart not found")
	}
	if err != nil {
		return domain.Cart{}, fmt.Errorf("finding cart: %w", err)
	}

	var items []cartItemModel
	if err := r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Order("created_at asc").
		Find(&items).Error; err != nil {
		return domain.Cart{}, fmt.Errorf("finding cart items: %w", err)
	}

	domainItems := make([]domain.CartItem, 0, len(items))
	for _, item := range items {
		domainItems = append(domainItems, domain.CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return domain.Cart{
		ID:        cart.ID,
		Items:     domainItems,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}, nil
}

// SaveItem upserts the absolute quantity for a (cart_id, product_id) line and bumps the
// cart's updated_at, atomically.
func (r *CartRepository) SaveItem(ctx context.Context, cartID, productID uuid.UUID, quantity int) error {
	now := time.Now().UTC()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item := cartItemModel{
			CartID:    cartID,
			ProductID: productID,
			Quantity:  quantity,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "cart_id"}, {Name: "product_id"}},
			DoUpdates: clause.Assignments(map[string]any{"quantity": quantity, "updated_at": now}),
		}).Create(&item).Error; err != nil {
			return fmt.Errorf("saving cart item: %w", err)
		}

		if err := tx.Model(&cartModel{}).Where("id = ?", cartID).Update("updated_at", now).Error; err != nil {
			return fmt.Errorf("touching cart: %w", err)
		}

		return nil
	})
}

// RemoveItem deletes a line and bumps the cart's updated_at, atomically. It reports
// item_not_found when the line does not exist.
func (r *CartRepository) RemoveItem(ctx context.Context, cartID, productID uuid.UUID) error {
	now := time.Now().UTC()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&cartItemModel{})
		if result.Error != nil {
			return fmt.Errorf("removing cart item: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return domain.WrapNotFound(gorm.ErrRecordNotFound, "item_not_found", "cart item not found")
		}

		if err := tx.Model(&cartModel{}).Where("id = ?", cartID).Update("updated_at", now).Error; err != nil {
			return fmt.Errorf("touching cart: %w", err)
		}

		return nil
	})
}
