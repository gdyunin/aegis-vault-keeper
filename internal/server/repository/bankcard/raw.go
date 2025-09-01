package bankcard

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	"github.com/google/uuid"
)

// rawSave creates a database save function that persists bank card data directly to PostgreSQL.
// Uses INSERT ON CONFLICT DO UPDATE for upsert behavior.
func rawSave(db db.DBClient) saveFunc {
	return func(ctx context.Context, p SaveParams) error {
		e := p.Entity

		query := `
			INSERT INTO aegis_vault_keeper.bank_cards (
				id, user_id, card_number, card_holder, expiry_month, expiry_year, cvv, description, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET
			  card_number   = EXCLUDED.card_number,
			  card_holder   = EXCLUDED.card_holder,
			  expiry_month  = EXCLUDED.expiry_month,
			  expiry_year   = EXCLUDED.expiry_year,
			  cvv           = EXCLUDED.cvv,
			  description   = EXCLUDED.description,
			  updated_at    = EXCLUDED.updated_at
		`

		if _, err := db.Exec(
			ctx,
			query,
			e.ID,
			e.UserID,
			e.CardNumber,
			e.CardHolder,
			e.ExpiryMonth,
			e.ExpiryYear,
			e.CVV,
			e.Description,
			e.UpdatedAt,
		); err != nil {
			return fmt.Errorf("query execution failed: %w", err)
		}
		return nil
	}
}

// rawLoad creates a database load function that retrieves bank card data from PostgreSQL.
// Supports filtering by user ID and specific bank card ID.
func rawLoad(db db.DBClient) func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
	return func(ctx context.Context, p LoadParams) ([]*bankcard.BankCard, error) {
		var (
			queryBuilder strings.Builder
			args         []any
			conditions   []string
			argIdx       = 1
		)

		queryBuilder.WriteString(`
			SELECT id, user_id, card_number, card_holder, expiry_month,
				   expiry_year, cvv, description, updated_at
			FROM aegis_vault_keeper.bank_cards
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
			return nil, fmt.Errorf("query execution failed: %w", err)
		}
		defer func() { _ = rows.Close() }()

		// cards collects all bank card entities retrieved from the database.
		var cards []*bankcard.BankCard
		for rows.Next() {
			// bc holds a single bank card entity during database row scanning.
			var bc bankcard.BankCard
			if err := rows.Scan(
				&bc.ID,
				&bc.UserID,
				&bc.CardNumber,
				&bc.CardHolder,
				&bc.ExpiryMonth,
				&bc.ExpiryYear,
				&bc.CVV,
				&bc.Description,
				&bc.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("row scan failed: %w", err)
			}
			cards = append(cards, &bc)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows iteration failed: %w", err)
		}
		return cards, nil
	}
}
