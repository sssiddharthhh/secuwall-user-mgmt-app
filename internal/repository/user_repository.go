package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"user-management-api/internal/model"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already in use")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID.String(), u.Name, u.Email, u.PasswordHash,
		u.CreatedAt.UTC().Format(time.RFC3339),
		u.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE id = ?`,
		id.String(),
	)
	return scanOne(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE email = ?`,
		email,
	)
	return scanOne(row)
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?`,
		u.Name, u.Email,
		u.UpdatedAt.UTC().Format(time.RFC3339),
		u.ID.String(),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("repository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// List returns users whose email contains emailFilter (case-insensitive).
// An empty emailFilter matches all users.
func (r *UserRepository) List(ctx context.Context, emailFilter string, limit, offset int) ([]*model.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE LOWER(email) LIKE LOWER(?)
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		"%"+emailFilter+"%", limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// --- helpers ---

func scanOne(row *sql.Row) (*model.User, error) {
	var (
		u                             model.User
		idStr, createdStr, updatedStr string
	)
	err := row.Scan(&idStr, &u.Name, &u.Email, &u.PasswordHash, &createdStr, &updatedStr)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.scanOne: %w", err)
	}
	u.ID, _ = uuid.Parse(idStr)
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &u, nil
}

func scanRow(rows *sql.Rows) (*model.User, error) {
	var (
		u                             model.User
		idStr, createdStr, updatedStr string
	)
	err := rows.Scan(&idStr, &u.Name, &u.Email, &u.PasswordHash, &createdStr, &updatedStr)
	if err != nil {
		return nil, fmt.Errorf("repository.scanRow: %w", err)
	}
	u.ID, _ = uuid.Parse(idStr)
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &u, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
