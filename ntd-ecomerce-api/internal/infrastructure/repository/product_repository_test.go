package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ntd-ecomerce-api/internal/domain"
	"ntd-ecomerce-api/internal/infrastructure/repository"
)

var fixtureID = uuid.New()

func newTestRepo(t *testing.T) (*repository.ProductRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	return repository.NewProductRepository(gormDB), mock
}

func fixtureProduct() domain.Product {
	return domain.Product{
		Name:        "Running Shoes",
		SKU:         "RS-001",
		Description: "Lightweight running shoes",
		Category:    "Footwear",
		Price:       decimal.NewFromFloat(89.99),
		Stock:       150,
		WeightKg:    decimal.NewFromFloat(0.35),
	}
}

func TestProductRepository_Add(t *testing.T) {
	type expected struct {
		err error
	}

	tests := map[string]struct {
		mockSetup func(mock sqlmock.Sqlmock)
		expected expected
	}{
		"should return created product when insert succeeds": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO "products"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(fixtureID, time.Now(), time.Now()))
			},
			expected: expected{err: nil},
		},
		"should return conflict when sku already exists": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO "products"`).
					WillReturnError(&pgconn.PgError{Code: "23505"})
			},
			expected: expected{err: domain.ErrConflict},
		},
		"should return error when insert fails": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO "products"`).
					WillReturnError(assert.AnError)
			},
			expected: expected{err: assert.AnError},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			_, err := repo.Add(context.Background(), fixtureProduct())

			assert.ErrorIs(t, err, tc.expected.err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductRepository_FindByID(t *testing.T) {
	type expected struct {
		err error
	}

	tests := map[string]struct {
		mockSetup func(mock sqlmock.Sqlmock)
		expected expected
	}{
		"should return product when it exists": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "products"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sku"}).
						AddRow(fixtureID, "Running Shoes", "RS-001"))
			},
			expected: expected{err: nil},
		},
		"should return not found when no row matches": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "products"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sku"}))
			},
			expected: expected{err: domain.ErrNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			_, err := repo.FindByID(context.Background(), fixtureID)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductRepository_Delete(t *testing.T) {
	type expected struct {
		err error
	}

	tests := map[string]struct {
		mockSetup func(mock sqlmock.Sqlmock)
		expected expected
	}{
		"should succeed when a row is deleted": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM "products"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expected: expected{err: nil},
		},
		"should return not found when no row is deleted": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM "products"`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expected: expected{err: domain.ErrNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			err := repo.Delete(context.Background(), fixtureID)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestProductRepository_Update(t *testing.T) {
	type expected struct {
		err error
	}

	tests := map[string]struct {
		mockSetup func(mock sqlmock.Sqlmock)
		expected expected
	}{
		"should return updated product when the row exists": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE "products"`).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery(`SELECT \* FROM "products"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sku"}).
						AddRow(fixtureID, "Running Shoes", "RS-001"))
			},
			expected: expected{err: nil},
		},
		"should return not found when no row is updated": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE "products"`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expected: expected{err: domain.ErrNotFound},
		},
		"should return conflict when the new sku collides": {
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE "products"`).
					WillReturnError(&pgconn.PgError{Code: "23505"})
			},
			expected: expected{err: domain.ErrConflict},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)
			product := fixtureProduct()
			product.ID = fixtureID

			_, err := repo.Update(context.Background(), product)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
