package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"user-management-api/internal/model"
	"user-management-api/internal/repository"
)

// ErrInvalidCredentials is returned when email/password don't match.
var ErrInvalidCredentials = errors.New("invalid credentials")

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewUserService(repo *repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("service.Register: %w", err)
	}

	now := time.Now().UTC()
	u := &model.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err // propagate ErrEmailTaken as-is
	}

	token, err := s.issueToken(u)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: u}, nil
}

func (s *UserService) SignIn(ctx context.Context, req *model.SignInRequest) (*model.AuthResponse, error) {
	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(u)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: u}, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	u.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) ListUsers(ctx context.Context, q *model.ListUsersQuery) ([]*model.User, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	return s.repo.List(ctx, q.Email, q.Limit, q.Offset)
}

func (s *UserService) issueToken(u *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   u.ID.String(),
		"email": u.Email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(s.jwtExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
