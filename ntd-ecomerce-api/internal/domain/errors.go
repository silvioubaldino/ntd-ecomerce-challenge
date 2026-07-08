package domain

import "errors"

type Kind int

const (
	KindValidation Kind = iota
	KindNotFound
	KindConflict
	KindBadRequest
	KindPayloadTooLarge
)

var (
	ErrValidation      = errors.New("validation error")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrBadRequest      = errors.New("bad request")
	ErrPayloadTooLarge = errors.New("payload too large")
)

type Error struct {
	Kind    Kind
	Code    string
	Message string
	Details map[string]string
	causes  []error
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() []error { return e.causes }

func WrapValidation(err error, details map[string]string) error {
	return &Error{
		Kind:    KindValidation,
		Code:    "validation_error",
		Message: "validation failed",
		Details: details,
		causes:  []error{ErrValidation, err},
	}
}

func WrapInvalidInput(err error, message string) error {
	return &Error{
		Kind:    KindValidation,
		Code:    "validation_error",
		Message: message,
		causes:  []error{ErrValidation, err},
	}
}

func WrapValidationCode(err error, code, message string) error {
	return &Error{
		Kind:    KindValidation,
		Code:    code,
		Message: message,
		causes:  []error{ErrValidation, err},
	}
}

func WrapNotFound(err error, code, message string) error {
	return &Error{
		Kind:    KindNotFound,
		Code:    code,
		Message: message,
		causes:  []error{ErrNotFound, err},
	}
}

func WrapConflict(err error, code, message string) error {
	return &Error{
		Kind:    KindConflict,
		Code:    code,
		Message: message,
		causes:  []error{ErrConflict, err},
	}
}

func WrapBadRequest(err error, code, message string) error {
	return &Error{
		Kind:    KindBadRequest,
		Code:    code,
		Message: message,
		causes:  []error{ErrBadRequest, err},
	}
}

func WrapPayloadTooLarge(err error, code, message string) error {
	return &Error{
		Kind:    KindPayloadTooLarge,
		Code:    code,
		Message: message,
		causes:  []error{ErrPayloadTooLarge, err},
	}
}
