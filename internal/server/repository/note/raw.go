package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/google/uuid"
)

// rawSave creates a database save function that persists note data directly to PostgreSQL.
// Uses INSERT ON CONFLICT DO UPDATE for upsert behavior.
func rawSave(db db.DBClient) saveFunc {
	return func(ctx context.Context, p SaveParams) error {
		e := p.Entity

		query := `
			INSERT INTO aegis_vault_keeper.notes (id, user_id, note, description, updated_at)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id) DO UPDATE SET
			  note        = EXCLUDED.note,
			  description = EXCLUDED.description,
			  updated_at  = EXCLUDED.updated_at
		`

		if _, err := db.Exec(ctx, query, e.ID, e.UserID, e.Note, e.Description, e.UpdatedAt); err != nil {
			return fmt.Errorf("failed to save note: %w", err)
		}
		return nil
	}
}

// rawLoad creates a database load function that retrieves note data from PostgreSQL.
// Supports filtering by user ID and specific note ID.
func rawLoad(db db.DBClient) func(ctx context.Context, p LoadParams) ([]*note.Note, error) {
	return func(ctx context.Context, p LoadParams) ([]*note.Note, error) {
		var (
			queryBuilder strings.Builder
			args         []interface{}
			conditions   []string
			argIdx       = 1
		)

		queryBuilder.WriteString(`
			SELECT id, user_id, note, description, updated_at
			FROM aegis_vault_keeper.notes
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

		// notes collects all note entities retrieved from the database.
		var notes []*note.Note
		for rows.Next() {
			// n holds a single note entity during database row scanning.
			var n note.Note
			if err := rows.Scan(
				&n.ID,
				&n.UserID,
				&n.Note,
				&n.Description,
				&n.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			notes = append(notes, &n)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows iteration error: %w", err)
		}

		return notes, nil
	}
}
