package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func TestPage_Validate(t *testing.T) {
	type (
		input struct {
			page domain.Page
		}
		expected struct {
			err error
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should accept the default page": {
			input:    input{page: domain.DefaultPage()},
			expected: expected{err: nil},
		},
		"should accept page_size at the max bound": {
			input:    input{page: domain.Page{Number: 1, Size: domain.MaxPageSize}},
			expected: expected{err: nil},
		},
		"should reject page number below 1": {
			input:    input{page: domain.Page{Number: 0, Size: 20}},
			expected: expected{err: domain.ErrInvalidPagination},
		},
		"should reject page_size below 1": {
			input:    input{page: domain.Page{Number: 1, Size: 0}},
			expected: expected{err: domain.ErrInvalidPagination},
		},
		"should reject page_size above the max bound": {
			input:    input{page: domain.Page{Number: 1, Size: domain.MaxPageSize + 1}},
			expected: expected{err: domain.ErrInvalidPagination},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			page := tc.input.page

			err := page.Validate()

			assert.ErrorIs(t, err, tc.expected.err)
		})
	}
}

func TestPage_Offset(t *testing.T) {
	type (
		input struct {
			page domain.Page
		}
		expected struct {
			offset int
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should return 0 for the first page": {
			input:    input{page: domain.Page{Number: 1, Size: 20}},
			expected: expected{offset: 0},
		},
		"should return page_size steps for later pages": {
			input:    input{page: domain.Page{Number: 2, Size: 20}},
			expected: expected{offset: 20},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			page := tc.input.page

			offset := page.Offset()

			assert.Equal(t, tc.expected.offset, offset)
		})
	}
}
