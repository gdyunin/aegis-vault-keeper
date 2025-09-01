package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// rawSave creates a function that performs raw database save operations for users.
func rawSave(db db.DBClient) saveFunc {
	return func(ctx context.Context, p SaveParams) error {
		e := p.Entity

		query := `
			INSERT INTO aegis_vault_keeper.auth_users (id, login, password_hash, crypto_key)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET
			  login = EXCLUDED.login,
			  password_hash = EXCLUDED.password_hash,
			  crypto_key = EXCLUDED.crypto_key
		`

		if _, err := db.Exec(ctx, query, e.ID, e.Login, e.PasswordHash, e.CryptoKey); err != nil {
			// pgErr holds the PostgreSQL error details for constraint violation checking.
			var pgErr *pgconn.PgError
			if ok := errors.As(err, &pgErr); ok && pgErr.Code == "23505" {
				return ErrUserAlreadyExists
			}
			return fmt.Errorf("failed to execute query: %w", err)
		}
		return nil
	}
}

// rawLoad creates a function that performs raw database load operations for users.
func rawLoad(db db.DBClient) func(ctx context.Context, p LoadParams) (*auth.User, error) {
	return func(ctx context.Context, p LoadParams) (*auth.User, error) {
		var (
			queryBuilder strings.Builder
			args         []interface{}
			conditions   []string
			argIdx       = 1
		)

		queryBuilder.WriteString(`
			SELECT id, login, password_hash, crypto_key
			FROM aegis_vault_keeper.auth_users
		`)

		if p.ID != uuid.Nil {
			conditions = append(conditions, fmt.Sprintf("id = $%d", argIdx))
			args = append(args, p.ID)
			argIdx++
		}
		if p.Login != "" {
			conditions = append(conditions, fmt.Sprintf("login = $%d", argIdx))
			args = append(args, p.Login)
			// argIdx++ // Last usage, no need to increment
		}
		if len(conditions) == 0 {
			return nil, errors.New("at least one of ID or Login must be provided")
		}

		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))

		// user holds the retrieved user entity from the database.
		var user auth.User
		if err := db.QueryRow(ctx, queryBuilder.String(), args...).Scan(
			&user.ID,
			&user.Login,
			&user.PasswordHash,
			&user.CryptoKey,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrUserNotFound
			}
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		return &user, nil
	}
}
