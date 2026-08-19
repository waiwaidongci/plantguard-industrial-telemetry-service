package infrastructure

import (
	"database/sql"
	"errors"
	"strings"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

func MapSQLError(err error, fallback *shareddomain.Error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return shareddomain.ErrNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique constraint") ||
		strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
		return shareddomain.ErrConflict
	}
	if fallback != nil {
		return shareddomain.Wrap(fallback.Code, fallback.Message, fallback.Status, err)
	}
	return shareddomain.Wrap("database_error", "database operation failed", 500, err)
}

func ExecAffected(ctx interface {
	Exec(string, ...any) (sql.Result, error)
}, query string, args ...any) (bool, error) {
	res, err := ctx.Exec(query, args...)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
