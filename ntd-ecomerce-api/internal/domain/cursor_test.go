package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func TestEncodeDecodeCursor_RoundTrip(t *testing.T) {
	type input struct {
		cursor domain.Cursor
	}

	id := uuid.New()

	tests := map[string]struct {
		input input
	}{
		"should round-trip a price cursor": {
			input: input{cursor: domain.Cursor{Sort: domain.ProductSortPriceAsc, Key: "19.99", LastID: id}},
		},
		"should round-trip a name cursor": {
			input: input{cursor: domain.Cursor{Sort: domain.ProductSortNameDesc, Key: "Running Shoes", LastID: id}},
		},
		"should round-trip a newest cursor": {
			input: input{cursor: domain.Cursor{Sort: domain.ProductSortNewest, Key: time.Now().UTC().Format(time.RFC3339Nano), LastID: id}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			token := domain.EncodeCursor(tc.input.cursor)

			decoded, err := domain.DecodeCursor(token, tc.input.cursor.Sort)

			assert.NoError(t, err)
			assert.Equal(t, tc.input.cursor, decoded)
		})
	}
}

func TestDecodeCursor_Errors(t *testing.T) {
	id := uuid.New()

	tests := map[string]struct {
		token        string
		expectedSort domain.ProductSort
	}{
		"should reject a malformed token": {
			token:        "not-a-valid-token",
			expectedSort: domain.ProductSortNewest,
		},
		"should reject a cursor issued for a different sort": {
			token:        domain.EncodeCursor(domain.Cursor{Sort: domain.ProductSortPriceAsc, Key: "10.00", LastID: id}),
			expectedSort: domain.ProductSortNameAsc,
		},
		"should reject a non-decimal key for a price sort": {
			token:        domain.EncodeCursor(domain.Cursor{Sort: domain.ProductSortPriceAsc, Key: "not-a-decimal", LastID: id}),
			expectedSort: domain.ProductSortPriceAsc,
		},
		"should reject a non-timestamp key for the newest sort": {
			token:        domain.EncodeCursor(domain.Cursor{Sort: domain.ProductSortNewest, Key: "not-a-timestamp", LastID: id}),
			expectedSort: domain.ProductSortNewest,
		},
		"should reject a cursor with a nil last id": {
			token:        domain.EncodeCursor(domain.Cursor{Sort: domain.ProductSortNameAsc, Key: "shirt", LastID: uuid.Nil}),
			expectedSort: domain.ProductSortNameAsc,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.DecodeCursor(tc.token, tc.expectedSort)

			assert.ErrorIs(t, err, domain.ErrInvalidCursor)
		})
	}
}
