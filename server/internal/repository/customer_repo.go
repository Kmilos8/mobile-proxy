package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type CustomerRepository struct {
	db *DB
}

func NewCustomerRepository(db *DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, c *domain.Customer) error {
	query := `INSERT INTO customers (id, name, email) VALUES ($1, $2, $3)`
	_, err := r.db.Pool.Exec(ctx, query, c.ID, c.Name, c.Email)
	return err
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `SELECT id, name, email, active, password_hash, email_verified, google_id, google_email, created_at, updated_at FROM customers WHERE id = $1`
	var c domain.Customer
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Email, &c.Active,
		&c.PasswordHash, &c.EmailVerified, &c.GoogleID, &c.GoogleEmail,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) List(ctx context.Context) ([]domain.Customer, error) {
	query := `SELECT id, name, email, active, password_hash, email_verified, google_id, google_email, created_at, updated_at FROM customers ORDER BY name ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var c domain.Customer
		err := rows.Scan(
			&c.ID, &c.Name, &c.Email, &c.Active,
			&c.PasswordHash, &c.EmailVerified, &c.GoogleID, &c.GoogleEmail,
			&c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan customer: %w", err)
		}
		customers = append(customers, c)
	}
	return customers, nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) error {
	query := `UPDATE customers SET name = $2, email = $3, active = $4, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, c.ID, c.Name, c.Email, c.Active)
	return err
}

// GetByEmail retrieves a customer by email address (case-insensitive).
func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `SELECT id, name, email, active, password_hash, email_verified, google_id, google_email, created_at, updated_at FROM customers WHERE LOWER(email) = LOWER($1)`
	var c domain.Customer
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&c.ID, &c.Name, &c.Email, &c.Active,
		&c.PasswordHash, &c.EmailVerified, &c.GoogleID, &c.GoogleEmail,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get customer by email: %w", err)
	}
	return &c, nil
}

// GetByGoogleID retrieves a customer by their Google OAuth account ID.
func (r *CustomerRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.Customer, error) {
	query := `SELECT id, name, email, active, password_hash, email_verified, google_id, google_email, created_at, updated_at FROM customers WHERE google_id = $1`
	var c domain.Customer
	err := r.db.Pool.QueryRow(ctx, query, googleID).Scan(
		&c.ID, &c.Name, &c.Email, &c.Active,
		&c.PasswordHash, &c.EmailVerified, &c.GoogleID, &c.GoogleEmail,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get customer by google id: %w", err)
	}
	return &c, nil
}

// CreateWithAuth inserts a new customer including auth fields (used for self-signup).
func (r *CustomerRepository) CreateWithAuth(ctx context.Context, c *domain.Customer) error {
	query := `INSERT INTO customers (id, name, email, active, password_hash, email_verified, google_id, google_email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Pool.Exec(ctx, query,
		c.ID, c.Name, c.Email, c.Active,
		c.PasswordHash, c.EmailVerified, c.GoogleID, c.GoogleEmail)
	return err
}

// UpdateEmailVerified sets the email_verified flag for a customer.
func (r *CustomerRepository) UpdateEmailVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	query := `UPDATE customers SET email_verified = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, verified)
	return err
}

// UpdatePasswordHash updates the bcrypt password hash for a customer.
func (r *CustomerRepository) UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error {
	query := `UPDATE customers SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, hash)
	return err
}

// LinkGoogleAccount associates a Google OAuth account with an existing customer.
// Also marks the email as verified since Google has confirmed ownership.
func (r *CustomerRepository) LinkGoogleAccount(ctx context.Context, id uuid.UUID, googleID, googleEmail string) error {
	query := `UPDATE customers SET google_id = $2, google_email = $3, email_verified = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, googleID, googleEmail)
	return err
}
