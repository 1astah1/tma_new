package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrOrderStatusInvalid  = errors.New("order status transition not allowed")
	ErrKeyNotAvailable     = errors.New("no available keys")
	ErrTaskAlreadyAssigned = errors.New("task already assigned")
	ErrRateLimited         = errors.New("rate limited")
	ErrDuplicate           = errors.New("duplicate entry")
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func NewAPIError(code, message string, details interface{}) *APIError {
	return &APIError{Code: code, Message: message, Details: details}
}
