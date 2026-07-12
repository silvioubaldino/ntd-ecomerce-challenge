package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ntd-ecomerce-api/internal/domain"
)

func TestDefaultPageRequest(t *testing.T) {
	page := domain.DefaultPageRequest()

	assert.Equal(t, domain.DefaultLimit, page.Limit)
	assert.Nil(t, page.Cursor)
}

func TestValidateLimit(t *testing.T) {
	type (
		input struct {
			limit int
		}
		expected struct {
			code string
			ok   bool
		}
	)

	tests := map[string]struct {
		input    input
		expected expected
	}{
		"should reject 0": {
			input:    input{limit: 0},
			expected: expected{code: "must_be_between_1_and_100", ok: false},
		},
		"should accept 1": {
			input:    input{limit: 1},
			expected: expected{code: "", ok: true},
		},
		"should accept 100": {
			input:    input{limit: 100},
			expected: expected{code: "", ok: true},
		},
		"should reject 101": {
			input:    input{limit: 101},
			expected: expected{code: "must_be_between_1_and_100", ok: false},
		},
		"should reject a negative limit": {
			input:    input{limit: -1},
			expected: expected{code: "must_be_between_1_and_100", ok: false},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			code, ok := domain.ValidateLimit(tc.input.limit)

			assert.Equal(t, tc.expected.code, code)
			assert.Equal(t, tc.expected.ok, ok)
		})
	}
}
