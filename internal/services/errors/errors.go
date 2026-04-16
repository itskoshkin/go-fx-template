package svcErr

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrRefreshTokenInvalidOrExpired = errors.New("refresh token is invalid or expired")
	ErrRefreshTokenCompromised      = errors.New("refresh token is compromised")
)

type NotFoundError struct {
	Operation string
	Entity    string
	Field     string
	Value     string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with %s '%s' not found", e.Entity, e.Field, e.Value)
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

type ForbiddenError struct {
	Message string
}

func (e ForbiddenError) Error() string {
	return e.Message
}

// DuplicateFieldError is returned by storage when a unique constraint is violated.
// ReturnMappedStorageError maps it to ConflictError for the API layer.
type DuplicateFieldError struct {
	Field string
}

func (e DuplicateFieldError) Error() string {
	return fmt.Sprintf("%s is already taken", e.Field)
}

func ReturnMappedStorageError(err error, operation, entity, field, value string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFoundError{Operation: operation, Entity: entity, Field: field, Value: value}
	}

	if dupErr, ok := errors.AsType[DuplicateFieldError](err); ok {
		return ConflictError{Message: dupErr.Error()}
	}

	return fmt.Errorf("%s: %w", operation, err)
}
