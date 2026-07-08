package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const uniqueViolationCode = "23505"

// productModel is the GORM row for the products table. It never leaves this
// package — domain.Product is the only shape other layers see.
type productModel struct {
	ID          uuid.UUID       `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string          `gorm:"column:name"`
	SKU         string          `gorm:"column:sku"`
	Description string          `gorm:"column:description"`
	Category    string          `gorm:"column:category"`
	Price       decimal.Decimal `gorm:"column:price"`
	Stock       int             `gorm:"column:stock"`
	WeightKg    decimal.Decimal `gorm:"column:weight_kg"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

func (productModel) TableName() string { return "products" }

func fromDomain(p domain.Product) productModel {
	return productModel{
		ID:          p.ID,
		Name:        p.Name,
		SKU:         p.SKU,
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Stock:       p.Stock,
		WeightKg:    p.WeightKg,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (m productModel) toDomain() domain.Product {
	return domain.Product{
		ID:          m.ID,
		Name:        m.Name,
		SKU:         m.SKU,
		Description: m.Description,
		Category:    m.Category,
		Price:       m.Price,
		Stock:       m.Stock,
		WeightKg:    m.WeightKg,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Add(ctx context.Context, product domain.Product) (domain.Product, error) {
	now := time.Now().UTC()
	model := fromDomain(product)
	model.CreatedAt = now
	model.UpdatedAt = now

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Product{}, translateWriteErr(err)
	}

	return model.toDomain(), nil
}

func (r *ProductRepository) FindAll(ctx context.Context, page domain.Page) (domain.ProductList, error) {
	var (
		models []productModel
		total  int64
	)

	if err := r.db.WithContext(ctx).Model(&productModel{}).Count(&total).Error; err != nil {
		return domain.ProductList{}, fmt.Errorf("counting products: %w", err)
	}

	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Offset(page.Offset()).
		Limit(page.Size).
		Find(&models).Error
	if err != nil {
		return domain.ProductList{}, fmt.Errorf("listing products: %w", err)
	}

	products := make([]domain.Product, 0, len(models))
	for _, m := range models {
		products = append(products, m.toDomain())
	}

	return domain.ProductList{
		Data: products,
		Pagination: domain.Pagination{
			Page:     page.Number,
			PageSize: page.Size,
			Total:    int(total),
		},
	}, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	var model productModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Product{}, domain.WrapNotFound(err, "product_not_found", "product not found")
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("finding product: %w", err)
	}

	return model.toDomain(), nil
}

func (r *ProductRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	model := fromDomain(product)
	model.UpdatedAt = time.Now().UTC()

	result := r.db.WithContext(ctx).Model(&productModel{}).Where("id = ?", product.ID).Updates(map[string]any{
		"name":        model.Name,
		"sku":         model.SKU,
		"description": model.Description,
		"category":    model.Category,
		"price":       model.Price,
		"stock":       model.Stock,
		"weight_kg":   model.WeightKg,
		"updated_at":  model.UpdatedAt,
	})
	if result.Error != nil {
		return domain.Product{}, translateWriteErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.Product{}, domain.WrapNotFound(gorm.ErrRecordNotFound, "product_not_found", "product not found")
	}

	return r.FindByID(ctx, product.ID)
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&productModel{})
	if result.Error != nil {
		return fmt.Errorf("deleting product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.WrapNotFound(gorm.ErrRecordNotFound, "product_not_found", "product not found")
	}
	return nil
}

// translateWriteErr maps a unique-constraint violation on sku to a domain
// conflict; every other error bubbles up wrapped with context.
func translateWriteErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return domain.WrapConflict(err, "sku_already_exists", "sku already exists")
	}
	return fmt.Errorf("writing product: %w", err)
}
