package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE email = $1`
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Pool.Exec(ctx, query, u.ID, u.Email, u.PasswordHash, u.Name, u.Role)
	return err
}
