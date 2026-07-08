package api

import (
	"errors"
	"net/http"

	"ntd-ecomerce-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type (
	errorBody struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details,omitempty"`
	}

	errorEnvelope struct {
		Error errorBody `json:"error"`
	}
)

func HandleErr(c *gin.Context, err error) {
	status, body := toAPIError(err)
	c.JSON(status, body)
}

func toAPIError(err error) (int, errorEnvelope) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return statusForKind(domainErr.Kind), errorEnvelope{
			Error: errorBody{
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Details: domainErr.Details,
			},
		}
	}

	return http.StatusInternalServerError, errorEnvelope{
		Error: errorBody{Code: "internal_error", Message: "internal server error"},
	}
}

func statusForKind(kind domain.Kind) int {
	switch kind {
	case domain.KindValidation:
		return http.StatusUnprocessableEntity
	case domain.KindNotFound:
		return http.StatusNotFound
	case domain.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
