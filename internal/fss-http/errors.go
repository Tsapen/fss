package fsshttp

import (
	"errors"
	"net/http"

	"github.com/Tsapen/fss/internal/fss"
)

func httpStatus(err error) int {
	switch {
	case errors.As(err, &fss.ValidationError{}):
		return http.StatusBadRequest

	case errors.As(err, &fss.NotFoundError{}):
		return http.StatusNotFound

	case errors.As(err, &fss.ConflictError{}):
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}
