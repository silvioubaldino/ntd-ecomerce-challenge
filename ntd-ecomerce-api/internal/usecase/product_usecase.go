package usecase

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"ntd-ecomerce-api/internal/domain"

	"github.com/google/uuid"
)

type (
	ProductRepository interface {
		Add(ctx context.Context, product domain.Product) (domain.Product, error)
		FindAll(ctx context.Context, filter domain.ProductFilter, page domain.Page) (domain.ProductList, error)
		FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
		Update(ctx context.Context, product domain.Product) (domain.Product, error)
		Delete(ctx context.Context, id uuid.UUID) error
		FindCategories(ctx context.Context) ([]string, error)
	}

	Product struct {
		repo ProductRepository
	}
)

func NewProduct(repo ProductRepository) Product {
	return Product{repo: repo}
}

func (u *Product) Add(ctx context.Context, input domain.ProductInput) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Product{}, domain.WrapValidation(ErrInvalidProductInput, problems)
	}

	created, err := u.repo.Add(ctx, input.ToProduct())
	if err != nil {
		return domain.Product{}, fmt.Errorf("adding product: %w", err)
	}

	return created, nil
}

func (u *Product) FindAll(ctx context.Context, filter domain.ProductFilter, page domain.Page) (domain.ProductList, error) {
	list, err := u.repo.FindAll(ctx, filter, page)
	if err != nil {
		return domain.ProductList{}, fmt.Errorf("listing products: %w", err)
	}

	return list, nil
}

func (u *Product) FindCategories(ctx context.Context) ([]string, error) {
	categories, err := u.repo.FindCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing product categories: %w", err)
	}

	return categories, nil
}

func (u *Product) FindByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	product, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Product{}, fmt.Errorf("finding product: %w", err)
	}

	return product, nil
}

func (u *Product) Update(ctx context.Context, id uuid.UUID, input domain.ProductInput) (domain.Product, error) {
	input.SKU = strings.TrimSpace(input.SKU)
	if problems := input.Validate(); len(problems) > 0 {
		return domain.Product{}, domain.WrapValidation(ErrInvalidProductInput, problems)
	}

	product := input.ToProduct()
	product.ID = id

	updated, err := u.repo.Update(ctx, product)
	if err != nil {
		return domain.Product{}, fmt.Errorf("updating product: %w", err)
	}

	return updated, nil
}

func (u *Product) DeleteOne(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}

	return nil
}

func (u *Product) Import(ctx context.Context, r io.Reader) (domain.ImportReport, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return domain.ImportReport{}, domain.WrapValidationCode(err, "invalid_header", "csv header is missing or unreadable")
	}
	if err := domain.ValidateCSVHeader(header); err != nil {
		return domain.ImportReport{}, domain.WrapValidationCode(err, "invalid_header", "csv header does not match the expected columns")
	}

	report := domain.ImportReport{Rejected: []domain.RejectedRow{}}

	rowNum := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return domain.ImportReport{}, fmt.Errorf("reading csv row %d: %w", rowNum+1, err)
		}
		rowNum++
		report.Summary.Total++

		input, problems := domain.ParseProductCSVRecord(record)
		if len(problems) > 0 {
			report.Summary.Rejected++
			report.Rejected = append(report.Rejected, domain.RejectedRow{Row: rowNum, SKU: input.SKU, Errors: problems})
			continue
		}

		if _, err := u.repo.Add(ctx, input.ToProduct()); err != nil {
			if errors.Is(err, domain.ErrConflict) {
				report.Summary.Rejected++
				report.Rejected = append(report.Rejected, domain.RejectedRow{
					Row:    rowNum,
					SKU:    input.SKU,
					Errors: map[string]string{"sku": "duplicate_sku"},
				})
				continue
			}
			return domain.ImportReport{}, fmt.Errorf("importing csv row %d: %w", rowNum, err)
		}

		report.Summary.Imported++
	}

	return report, nil
}
