package filedata

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/google/uuid"
)

// rawSave creates a function that performs raw database save operations for file data.
func rawSave(db db.DBClient) saveFunc {
	return func(ctx context.Context, p SaveParams) error {
		e := p.Entity

		query := `
			INSERT INTO aegis_vault_keeper.files (id, user_id, storage_key, hash_sum, description, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
			  storage_key        = EXCLUDED.storage_key,
			  hash_sum     = EXCLUDED.hash_sum,
			  description  = EXCLUDED.description,
			  updated_at   = EXCLUDED.updated_at
		`

		if _, err := db.Exec(ctx, query, e.ID, e.UserID, e.StorageKey, e.HashSum, e.Description, e.UpdatedAt); err != nil {
			return fmt.Errorf("failed to save file: %w", err)
		}
		return nil
	}
}

// rawLoad creates a function that performs raw database load operations for file data.
func rawLoad(db db.DBClient) func(ctx context.Context, p LoadParams) ([]*filedata.FileData, error) {
	return func(ctx context.Context, p LoadParams) ([]*filedata.FileData, error) {
		var (
			queryBuilder strings.Builder
			args         []any
			conditions   []string
			argIdx       = 1
		)

		queryBuilder.WriteString(`
			SELECT id, user_id, storage_key, hash_sum, description, updated_at
			FROM aegis_vault_keeper.files
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

		// fds collects all file data entities retrieved from the database.
		var fds []*filedata.FileData
		for rows.Next() {
			// c holds a single file data entity during database row scanning.
			var c filedata.FileData
			if err := rows.Scan(
				&c.ID,
				&c.UserID,
				&c.StorageKey,
				&c.HashSum,
				&c.Description,
				&c.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			fds = append(fds, &c)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("row iteration error: %w", err)
		}
		return fds, nil
	}
}
