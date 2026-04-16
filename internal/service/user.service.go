package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/assidik12/go-restfull-api/config"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/pkg/hash"
	"github.com/assidik12/go-restfull-api/internal/pkg/jwt"
	"github.com/go-playground/validator/v10"

	"golang.org/x/crypto/bcrypt"
)

// UserService defines the business-logic contract for users.
type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
}

type userService struct {
	repo     domain.UserRepository // ← domain interface, not mysql package
	DB       *sql.DB
	validate *validator.Validate
}

// NewUserService constructs a UserService.
func NewUserService(repo domain.UserRepository, DB *sql.DB, validate *validator.Validate) UserService {
	return &userService{
		repo:     repo,
		DB:       DB,
		validate: validate,
	}
}

// Register implements UserService.
func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return dto.UserResponse{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	hashedPassword, err := hash.NewCryptoHasher(bcrypt.DefaultCost).HashPassword(req.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	newUser := domain.User{
		Name:      req.Name,
		Email:     req.Email,
		Role:      "user",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := s.repo.Save(ctx, newUser)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

// Login implements UserService.
// NOTE: config.GetConfig() is still called here intentionally — this is a
// Phase 2 item (inject jwtSecret via constructor). It is NOT changed in
// this Phase 1 refactor so as not to alter behaviour unintentionally.
func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	cfg := config.GetConfig()

	if err := s.validate.Struct(req); err != nil {
		return dto.LoginResponse{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Repository maps sql.ErrNoRows → domain.ErrNotFound
		return dto.LoginResponse{}, fmt.Errorf("%w: email %s", domain.ErrNotFound, req.Email)
	}

	if err := hash.NewCryptoHasher(bcrypt.DefaultCost).ComparePassword(user.Password, req.Password); err != nil {
		return dto.LoginResponse{}, fmt.Errorf("%w: invalid email or password", domain.ErrUnauthorized)
	}

	token, err := jwt.NewJWTService(cfg.JWTSecret).GenerateJWT(user)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}
