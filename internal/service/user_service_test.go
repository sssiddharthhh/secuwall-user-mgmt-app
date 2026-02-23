package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"user-management-api/internal/model"
	"user-management-api/internal/repository"
	"user-management-api/internal/service"
)

// setupService spins up a real in-memory SQLite DB and returns a wired UserService.
// The DB is closed automatically when the test ends.
func setupService(t *testing.T) *service.UserService {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
		CREATE TABLE users (
			id            TEXT PRIMARY KEY,
			name          TEXT NOT NULL,
			email         TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at    TEXT NOT NULL,
			updated_at    TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewUserRepository(db)
	return service.NewUserService(repo, "test-secret", 24*time.Hour)
}

func TestRegister_Success(t *testing.T) {
	svc := setupService(t)

	resp, err := svc.Register(context.Background(), &model.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected a JWT token, got empty string")
	}
	if resp.User.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", resp.User.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := setupService(t)

	req := &model.RegisterRequest{Name: "Alice", Email: "alice@example.com", Password: "secret123"}
	if _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err := svc.Register(context.Background(), req)
	if !errors.Is(err, repository.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestSignIn_Success(t *testing.T) {
	svc := setupService(t)

	if _, err := svc.Register(context.Background(), &model.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	resp, err := svc.SignIn(context.Background(), &model.SignInRequest{
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected a JWT token, got empty string")
	}
}

func TestSignIn_WrongPassword(t *testing.T) {
	svc := setupService(t)

	if _, err := svc.Register(context.Background(), &model.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err := svc.SignIn(context.Background(), &model.SignInRequest{
		Email:    "alice@example.com",
		Password: "wrongpassword",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestSignIn_UnknownEmail(t *testing.T) {
	svc := setupService(t)

	_, err := svc.SignIn(context.Background(), &model.SignInRequest{
		Email:    "nobody@example.com",
		Password: "secret123",
	})
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	svc := setupService(t)

	resp, err := svc.Register(context.Background(), &model.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	updated, err := svc.UpdateUser(context.Background(), resp.User.ID, &model.UpdateUserRequest{
		Name: "Alice Smith",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Alice Smith" {
		t.Errorf("expected name 'Alice Smith', got '%s'", updated.Name)
	}
	// Email should be unchanged
	if updated.Email != "alice@example.com" {
		t.Errorf("email should be unchanged, got %s", updated.Email)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc := setupService(t)

	_, err := svc.UpdateUser(context.Background(), uuid.New(), &model.UpdateUserRequest{
		Name: "Ghost",
	})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
