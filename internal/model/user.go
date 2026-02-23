package model

import (
	"time"

	"github.com/google/uuid"
)

// User is the core domain type. PasswordHash is never serialised to JSON.
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// --- request DTOs ---

type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type SignInRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=2"`
	Email string `json:"email" validate:"omitempty,email"`
}

type ListUsersQuery struct {
	Email  string `form:"email"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// --- response DTOs ---

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
