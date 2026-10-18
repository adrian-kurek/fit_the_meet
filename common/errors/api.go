// Package commonerrors holds whole logic of APIs and DB errors
package commonerrors

import (
	"fmt"
	"net/http"
)

type Category string

const (
	CategoryValidation       Category = "VALIDATION"
	CategoryNotFound         Category = "NOT_FOUND"
	CategoryUnauthorized     Category = "UNAUTHORIZED"
	CategoryInternal         Category = "INTERNAL"
	CategoryRequestTimeout   Category = "REQUEST_TIMEOUT"
	CategoryMethodNotAllowed Category = "METHOD_NOT_ALLOWED"
)

type API struct {
	Category      Category
	StatusCode    int
	Message       string
	IsOperational bool
}

func NewAPI(statusCode int, category Category, message string, isOperational bool) *API {
	return &API{
		Category:      category,
		StatusCode:    statusCode,
		Message:       message,
		IsOperational: isOperational,
	}
}

func (api *API) Error() string {
	return fmt.Sprintf("api error: %s", api.Message)
}

func NotFound(msg string) *API {
	return &API{Category: CategoryNotFound, StatusCode: http.StatusNotFound, IsOperational: true, Message: msg}
}

func Validation(msg string) *API {
	return &API{
		Category: CategoryValidation, StatusCode: http.StatusUnprocessableEntity, IsOperational: true, Message: msg,
	}
}

func InvalidJSONFormat() *API {
	return &API{
		Category:      CategoryValidation,
		StatusCode:    http.StatusUnprocessableEntity,
		IsOperational: true,
		Message:       "provided invalid json format",
	}
}

func Unauthorized(msg string) *API {
	return &API{
		Category: CategoryUnauthorized, StatusCode: http.StatusUnauthorized, IsOperational: true, Message: msg,
	}
}

func RequestTimeout() *API {
	return &API{
		Category: CategoryUnauthorized, StatusCode: http.StatusUnauthorized, IsOperational: true, Message: "",
	}
}
