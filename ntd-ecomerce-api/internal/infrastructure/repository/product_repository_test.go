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
		expected  expected
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

func TestProductRepository_FindAll(t *testing.T) {
	type (
		input struct {
			page domain.Page
		}
		expected struct {
			total int
			err   error
		}
	)

	tests := map[string]struct {
		input     input
		mockSetup func(mock sqlmock.Sqlmock)
		expected  expected
	}{
		"should not add a WHERE clause and order by created_at desc when query is empty": {
			input: input{page: domain.Page{Number: 1, Size: 20, Query: ""}},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "products"`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				mock.ExpectQuery(`SELECT \* FROM "products" ORDER BY created_at desc`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sku"}).
						AddRow(fixtureID, "Running Shoes", "RS-001").
						AddRow(fixtureID, "Blue Shirt", "BS-021"))
			},
			expected: expected{total: 2, err: nil},
		},
		"should filter case-insensitively over name/sku/description/category and order by name asc when query is present": {
			input: input{page: domain.Page{Number: 1, Size: 20, Query: "blue"}},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "products" WHERE LOWER\(name\) LIKE \$1 OR LOWER\(sku\) LIKE \$2 OR LOWER\(description\) LIKE \$3 OR LOWER\(category\) LIKE \$4`).
					WithArgs("%blue%", "%blue%", "%blue%", "%blue%").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "products" WHERE LOWER\(name\) LIKE \$1 OR LOWER\(sku\) LIKE \$2 OR LOWER\(description\) LIKE \$3 OR LOWER\(category\) LIKE \$4.*ORDER BY name asc`).
					WithArgs("%blue%", "%blue%", "%blue%", "%blue%", 20).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sku"}).
						AddRow(fixtureID, "Blue Shirt", "BS-021"))
			},
			expected: expected{total: 1, err: nil},
		},
		"should return zero total and no rows when query has no matches": {
			input: input{page: domain.Page{Number: 1, Size: 20, Query: "zzz-nomatch"}},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "products" WHERE`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery(`SELECT \* FROM "products" WHERE`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sku"}))
			},
			expected: expected{total: 0, err: nil},
		},
		"should return error when counting fails": {
			input: input{page: domain.Page{Number: 1, Size: 20, Query: ""}},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "products"`).
					WillReturnError(assert.AnError)
			},
			expected: expected{total: 0, err: assert.AnError},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			list, err := repo.FindAll(context.Background(), tc.input.page)

			assert.ErrorIs(t, err, tc.expected.err)
			assert.Equal(t, tc.expected.total, list.Pagination.Total)
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
		expected  expected
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
		expected  expected
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
		expected  expected
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
