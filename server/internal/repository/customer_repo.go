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
	query := `SELECT id, name, email, active, created_at, updated_at FROM customers WHERE id = $1`
	var c domain.Customer
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Email, &c.Active, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) List(ctx context.Context) ([]domain.Customer, error) {
	query := `SELECT id, name, email, active, created_at, updated_at FROM customers ORDER BY name ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var c domain.Customer
		err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Active, &c.CreatedAt, &c.UpdatedAt)
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
