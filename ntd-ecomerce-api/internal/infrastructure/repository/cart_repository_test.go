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

var (
	cartID    = uuid.New()
	productID = uuid.New()
)

func newTestCartRepo(t *testing.T) (*repository.CartRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	return repository.NewCartRepository(gormDB), mock
}

func TestCartRepository_Create(t *testing.T) {
	t.Run("should return the created empty cart", func(t *testing.T) {
		repo, mock := newTestCartRepo(t)
		mock.ExpectQuery(`INSERT INTO "carts"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cartID))

		cart, err := repo.Create(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, cartID, cart.ID)
		assert.Empty(t, cart.Items)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepository_FindByID(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(mock sqlmock.Sqlmock)
		expected  error
		items     int
	}{
		"should return the cart with its items": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "carts"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(cartID, time.Now(), time.Now()))
				mock.ExpectQuery(`SELECT \* FROM "cart_items" WHERE cart_id = \$1 ORDER BY created_at asc`).
					WithArgs(cartID).
					WillReturnRows(sqlmock.NewRows([]string{"cart_id", "product_id", "quantity"}).
						AddRow(cartID, productID, 3))
			},
			items: 1,
		},
		"should return cart_not_found when the cart is missing": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "carts"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			expected: domain.ErrNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestCartRepo(t)
			tc.mockSetup(mock)

			cart, err := repo.FindByID(context.Background(), cartID)

			assert.ErrorIs(t, err, tc.expected)
			if tc.expected == nil {
				assert.Len(t, cart.Items, tc.items)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepository_SaveItem(t *testing.T) {
	t.Run("should upsert the item and touch the cart", func(t *testing.T) {
		repo, mock := newTestCartRepo(t)
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "cart_items" .* ON CONFLICT`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`UPDATE "carts" SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.SaveItem(context.Background(), cartID, productID, 5)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepository_RemoveItem(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(mock sqlmock.Sqlmock)
		expected  error
	}{
		"should delete the item and touch the cart": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "cart_items"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(`UPDATE "carts" SET`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expected: nil,
		},
		"should return item_not_found when no line is deleted": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "cart_items"`).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			expected: domain.ErrNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestCartRepo(t)
			tc.mockSetup(mock)

			err := repo.RemoveItem(context.Background(), cartID, productID)

			assert.ErrorIs(t, err, tc.expected)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
