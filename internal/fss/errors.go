package fss

import (
	"fmt"
)

// NotFoundError implements error interface.
type NotFoundError struct {
	Err error
}

func (err NotFoundError) Error() string {
	return err.Err.Error()
}

func NewNotFoundError(format string, a ...any) NotFoundError {
	return NotFoundError{fmt.Errorf(format, a...)}
}

// ValidationError implements error interface.
type ValidationError struct {
	Err error
}

func (err ValidationError) Error() string {
	return err.Err.Error()
}

func NewValidationError(format string, a ...any) ValidationError {
	return ValidationError{fmt.Errorf(format, a...)}
}

// InternalError implements error interface.
type InternalError struct {
	Err error
}

func (err InternalError) Error() string {
	return err.Err.Error()
}

func NewInternalError(format string, a ...any) InternalError {
	return InternalError{fmt.Errorf(format, a...)}
}

// ConflictError implements error interface.
type ConflictError struct {
	Err error
}

func (err ConflictError) Error() string {
	return err.Err.Error()
}

func NewConflictError(format string, a ...any) ConflictError {
	return ConflictError{fmt.Errorf(format, a...)}
}

// BadRequestError implements error interface.
type BadRequestError struct {
	Err error
}

func (err BadRequestError) Error() string {
	return err.Err.Error()
}

func NewBadRequestError(format string, a ...any) BadRequestError {
	return BadRequestError{fmt.Errorf(format, a...)}
}

// ErrPair contains deferred and returned error.
type ErrPair struct {
	Def error
	Ret error
}

// Error returns concatenated error.
func (errPair ErrPair) Error() string {
	return fmt.Sprintf("returned: %s; deferred: %s", errPair.Def, errPair.Ret)
}

// HandleErrPair contains deferred and returned errors.
func HandleErrPair(def, ret error) error {
	if ret == nil {
		return def
	}

	if def == nil {
		return ret
	}

	return ErrPair{
		Def: def,
		Ret: ret,
	}
}
