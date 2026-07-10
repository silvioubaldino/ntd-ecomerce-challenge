package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ntd-ecomerce-api/internal/domain"
	"ntd-ecomerce-api/internal/infrastructure/repository"
)

var orderID = uuid.New()

func newTestOrderRepo(t *testing.T) (*repository.OrderRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	return repository.NewOrderRepository(gormDB), mock
}

func validCustomer() domain.Customer {
	return domain.Customer{Name: "Ada Lovelace", Email: "ada@example.com"}
}

func TestOrderRepository_Checkout(t *testing.T) {
	t.Run("should create the order, decrement stock and consume the cart", func(t *testing.T) {
		repo, mock := newTestOrderRepo(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "carts"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(cartID, time.Now(), time.Now()))
		mock.ExpectQuery(`SELECT \* FROM "cart_items"`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "product_id", "quantity", "created_at"}).
				AddRow(cartID, productID, 3, time.Now()))
		mock.ExpectQuery(`SELECT \* FROM "products"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "sku", "name", "price", "stock"}).
				AddRow(productID, "RS-001", "Running Shoes", "10.00", 10))
		mock.ExpectExec(`UPDATE "products" SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(`INSERT INTO "orders"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(`INSERT INTO "order_items"`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`DELETE FROM "carts"`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		order, err := repo.Checkout(context.Background(), cartID, validCustomer())

		assert.NoError(t, err)
		require.Len(t, order.Items, 1)
		assert.Equal(t, domain.OrderStatusConfirmed, order.Status)
		assert.Equal(t, "RS-001", order.Items[0].SKU)
		assert.Equal(t, "10", order.Items[0].UnitPrice.String())
		assert.Equal(t, "30", order.Items[0].Subtotal.String())
		assert.Equal(t, "30", order.Total.String())
		assert.Equal(t, "approved", order.Payment.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return cart_not_found when the cart is missing", func(t *testing.T) {
		repo, mock := newTestOrderRepo(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "carts"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectRollback()

		_, err := repo.Checkout(context.Background(), cartID, validCustomer())

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assertDomainCode(t, err, "cart_not_found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return cart_empty when the cart has no items", func(t *testing.T) {
		repo, mock := newTestOrderRepo(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "carts"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(cartID, time.Now(), time.Now()))
		mock.ExpectQuery(`SELECT \* FROM "cart_items"`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "product_id", "quantity", "created_at"}))
		mock.ExpectRollback()

		_, err := repo.Checkout(context.Background(), cartID, validCustomer())

		assert.ErrorIs(t, err, domain.ErrValidation)
		assertDomainCode(t, err, "cart_empty")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should roll back with insufficient_stock when a line exceeds current stock", func(t *testing.T) {
		repo, mock := newTestOrderRepo(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT \* FROM "carts"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(cartID, time.Now(), time.Now()))
		mock.ExpectQuery(`SELECT \* FROM "cart_items"`).
			WithArgs(cartID).
			WillReturnRows(sqlmock.NewRows([]string{"cart_id", "product_id", "quantity", "created_at"}).
				AddRow(cartID, productID, 5, time.Now()))
		mock.ExpectQuery(`SELECT \* FROM "products"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "sku", "name", "price", "stock"}).
				AddRow(productID, "RS-001", "Running Shoes", "10.00", 3))
		mock.ExpectRollback()

		_, err := repo.Checkout(context.Background(), cartID, validCustomer())

		assert.ErrorIs(t, err, domain.ErrConflict)
		assertDomainCode(t, err, "insufficient_stock")
		assertDomainDetails(t, err, map[string]string{productID.String(): "requested=5, available=3"})
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrderRepository_FindByID(t *testing.T) {
	t.Run("should return the order with its items", func(t *testing.T) {
		repo, mock := newTestOrderRepo(t)
		mock.ExpectQuery(`SELECT \* FROM "orders"`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "status", "customer_name", "customer_email", "total", "payment_method", "payment_status", "created_at"}).
				AddRow(orderID, "confirmed", "Ada", "ada@example.com", "30.00", "simulated", "approved", time.Now()))
		mock.ExpectQuery(`SELECT \* FROM "order_items"`).
			WithArgs(orderID).
			WillReturnRows(sqlmock.NewRows([]string{"order_id", "product_id", "sku", "name", "unit_price", "quantity", "subtotal", "created_at"}).
				AddRow(orderID, productID, "RS-001", "Running Shoes", "10.00", 3, "30.00", time.Now()))

		order, err := repo.FindByID(context.Background(), orderID)

		assert.NoError(t, err)
		require.Len(t, order.Items, 1)
		assert.Equal(t, "confirmed", order.Status)
		assert.Equal(t, "30", order.Total.String())
		assert.Equal(t, "RS-001", order.Items[0].SKU)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return order_not_found when the order is missing", func(t *testing.T) {
		repo, mock := newTestOrderRepo(t)
		mock.ExpectQuery(`SELECT \* FROM "orders"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := repo.FindByID(context.Background(), orderID)

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assertDomainCode(t, err, "order_not_found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func assertDomainCode(t *testing.T, err error, code string) {
	t.Helper()
	var domainErr *domain.Error
	if assert.ErrorAs(t, err, &domainErr) {
		assert.Equal(t, code, domainErr.Code)
	}
}

func assertDomainDetails(t *testing.T, err error, details map[string]string) {
	t.Helper()
	var domainErr *domain.Error
	if assert.ErrorAs(t, err, &domainErr) {
		assert.Equal(t, details, domainErr.Details)
	}
}
