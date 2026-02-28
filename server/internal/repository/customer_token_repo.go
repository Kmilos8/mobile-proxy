package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type CustomerTokenRepository struct {
	db *DB
}

func NewCustomerTokenRepository(db *DB) *CustomerTokenRepository {
	return &CustomerTokenRepository{db: db}
}

// Create inserts a new customer auth token (email verification or password reset).
func (r *CustomerTokenRepository) Create(ctx context.Context, token *domain.CustomerAuthToken) error {
	query := `INSERT INTO customer_auth_tokens (id, customer_id, token_hash, type, expires_at, used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Pool.Exec(ctx, query,
		token.ID, token.CustomerID, token.TokenHash, token.Type,
		token.ExpiresAt, token.UsedAt, token.CreatedAt)
	return err
}

// GetByHash retrieves a valid (unused, non-expired) token by its hash.
func (r *CustomerTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*domain.CustomerAuthToken, error) {
	query := `SELECT id, customer_id, token_hash, type, expires_at, used_at, created_at
		FROM customer_auth_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()`
	var t domain.CustomerAuthToken
	err := r.db.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.CustomerID, &t.TokenHash, &t.Type,
		&t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get token by hash: %w", err)
	}
	return &t, nil
}

// MarkUsed records the time a token was consumed so it cannot be reused.
func (r *CustomerTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE customer_auth_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// DeleteByCustomerAndType removes all tokens of a given type for a customer.
// Used to invalidate old tokens before issuing a new one (e.g., on resend verification).
func (r *CustomerTokenRepository) DeleteByCustomerAndType(ctx context.Context, customerID uuid.UUID, tokenType string) error {
	query := `DELETE FROM customer_auth_tokens WHERE customer_id = $1 AND type = $2`
	_, err := r.db.Pool.Exec(ctx, query, customerID, tokenType)
	return err
}
