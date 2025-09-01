package credential

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/google/uuid"
)

// rawSave creates a function that performs raw database save operations for credentials.
func rawSave(db db.DBClient) saveFunc {
	return func(ctx context.Context, p SaveParams) error {
		e := p.Entity

		query := `
			INSERT INTO aegis_vault_keeper.credentials (id, user_id, login, password, description, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
			  login        = EXCLUDED.login,
			  password     = EXCLUDED.password,
			  description  = EXCLUDED.description,
			  updated_at   = EXCLUDED.updated_at
		`

		if _, err := db.Exec(ctx, query, e.ID, e.UserID, e.Login, e.Password, e.Description, e.UpdatedAt); err != nil {
			return fmt.Errorf("failed to save credential: %w", err)
		}
		return nil
	}
}

// RawLoad creates a function that performs raw database load operations for credentials.
// Supports filtering by user ID and specific credential ID.
func rawLoad(db db.DBClient) func(ctx context.Context, p LoadParams) ([]*credential.Credential, error) {
	return func(ctx context.Context, p LoadParams) ([]*credential.Credential, error) {
		var (
			queryBuilder strings.Builder
			args         []interface{}
			conditions   []string
			argIdx       = 1
		)

		queryBuilder.WriteString(`
			SELECT id, user_id, login, password, description, updated_at
			FROM aegis_vault_keeper.credentials
		`)

		if p.ID != uuid.Nil {
			conditions = append(conditions, fmt.Sprintf("id = $%d", argIdx))
			args = append(args, p.ID)
			argIdx++
		}
		if p.UserID != uuid.Nil {
			conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
			args = append(args, p.UserID)
			// argIdx++ // Last usage, no need to increment
		}
		if len(conditions) == 0 {
			return nil, errors.New("at least one of ID or UserID must be provided")
		}

		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))

		rows, err := db.Query(ctx, queryBuilder.String(), args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer func() { _ = rows.Close() }()

		// creds collects all credential entities retrieved from the database.
		var creds []*credential.Credential
		for rows.Next() {
			// c holds a single credential entity during database row scanning.
			var c credential.Credential
			if err := rows.Scan(
				&c.ID,
				&c.UserID,
				&c.Login,
				&c.Password,
				&c.Description,
				&c.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			creds = append(creds, &c)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("row iteration error: %w", err)
		}
		return creds, nil
	}
}
