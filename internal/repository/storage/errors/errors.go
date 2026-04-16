package dbErr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"go-fx-template/internal/services/errors"
)

var constraintToField = map[string]string{
	"idx_users_username_unique": "username",
	"idx_users_email_unique":    "email",
}

func DuplicateFieldError(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
		var field string
		if field, ok = constraintToField[pgErr.ConstraintName]; ok {
			return svcErr.DuplicateFieldError{Field: field}
		}
	}
	return nil
}
