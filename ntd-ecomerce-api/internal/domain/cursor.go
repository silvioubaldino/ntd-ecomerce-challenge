package domain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrInvalidCursor = errors.New("invalid cursor")

type Cursor struct {
	Sort   ProductSort `json:"sort"`
	Key    string      `json:"key"`
	LastID uuid.UUID   `json:"last_id"`
}

func EncodeCursor(c Cursor) string {
	raw, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func DecodeCursor(token string, expectedSort ProductSort) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}

	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return Cursor{}, ErrInvalidCursor
	}

	if c.Sort != expectedSort {
		return Cursor{}, ErrInvalidCursor
	}

	if c.LastID == uuid.Nil {
		return Cursor{}, ErrInvalidCursor
	}

	if err := validateCursorKey(c.Sort, c.Key); err != nil {
		return Cursor{}, ErrInvalidCursor
	}

	return c, nil
}

func validateCursorKey(sort ProductSort, key string) error {
	switch sort {
	case ProductSortPriceAsc, ProductSortPriceDesc:
		_, err := decimal.NewFromString(key)
		return err
	case ProductSortNameAsc, ProductSortNameDesc:
		return nil
	case ProductSortNewest:
		_, err := time.Parse(time.RFC3339Nano, key)
		return err
	default:
		return ErrInvalidCursor
	}
}
