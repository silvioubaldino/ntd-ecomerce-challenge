package domain

import "errors"

// Kind classifies a domain error so the API layer can map it to an HTTP status
// without importing the usecase or repository packages.
type Kind int

const (
	KindValidation Kind = iota
	KindNotFound
	KindConflict
)

// Base sentinels. Every domain error wraps one of these, so callers can classify
// with errors.Is(err, domain.ErrNotFound) regardless of the concrete cause.
var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

// Error is the single typed error the API envelope is built from. Kind drives the
// HTTP status; Code is the machine-readable value returned to clients; Details
// carries per-field problems for validation errors.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Details map[string]string
	causes  []error
}

func (e *Error) Error() string { return e.Message }

// Unwrap exposes both the base sentinel and the original cause so errors.Is
// matches either.
func (e *Error) Unwrap() []error { return e.causes }

// WrapValidation reports invalid input with a field->problem map (AYD-001
// validation_error envelope). err is the usecase sentinel that identifies the rule.
func WrapValidation(err error, details map[string]string) error {
	return &Error{
		Kind:    KindValidation,
		Code:    "validation_error",
		Message: "validation failed",
		Details: details,
		causes:  []error{ErrValidation, err},
	}
}

// WrapInvalidInput reports malformed input that has no per-field breakdown (bad
// JSON body, unparsable id or pagination). Maps to validation_error / 422.
func WrapInvalidInput(err error, message string) error {
	return &Error{
		Kind:    KindValidation,
		Code:    "validation_error",
		Message: message,
		causes:  []error{ErrValidation, err},
	}
}

// WrapNotFound reports a missing resource with an explicit machine code
// (e.g. "product_not_found").
func WrapNotFound(err error, code, message string) error {
	return &Error{
		Kind:    KindNotFound,
		Code:    code,
		Message: message,
		causes:  []error{ErrNotFound, err},
	}
}

// WrapConflict reports a uniqueness/constraint conflict with an explicit machine
// code (e.g. "sku_already_exists").
func WrapConflict(err error, code, message string) error {
	return &Error{
		Kind:    KindConflict,
		Code:    code,
		Message: message,
		causes:  []error{ErrConflict, err},
	}
}
