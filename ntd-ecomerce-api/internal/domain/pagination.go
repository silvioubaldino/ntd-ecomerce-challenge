package domain

import "errors"

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var ErrInvalidPagination = errors.New("invalid pagination")

type Page struct {
	Number int
	Size   int
	Query  string
}

func DefaultPage() Page {
	return Page{Number: 1, Size: DefaultPageSize}
}

func (p Page) Validate() error {
	if p.Number < 1 {
		return ErrInvalidPagination
	}
	if p.Size < 1 || p.Size > MaxPageSize {
		return ErrInvalidPagination
	}
	return nil
}

func (p Page) Offset() int {
	return (p.Number - 1) * p.Size
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}
