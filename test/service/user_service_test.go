package service_test

import (
	"context"
	"testing"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupUserServiceTesting() (*MockUserRepo, service.UserService) {
	mockRepo := new(MockUserRepo)
	validate := validator.New()
	userService := service.NewUserService(mockRepo, nil, validate, "test_secret_key")
	return mockRepo, userService
}

func TestRegisterUser_Success(t *testing.T) {
	mockRepo, userService := setupUserServiceTesting()

	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
		return u.Email == "test@example.com" && u.Name == "Test User" && u.Role == "user"
	})).Return(domain.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}, nil)

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := userService.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "Test User", resp.Name)
	assert.Equal(t, "test@example.com", resp.Email)
	mockRepo.AssertExpectations(t)
}

func TestRegisterUser_ConflictEmail(t *testing.T) {
	mockRepo, userService := setupUserServiceTesting()

	mockRepo.On("Save", mock.Anything, mock.Anything).Return(domain.User{}, domain.ErrConflict)

	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "existing@example.com",
		Password: "password123",
	}

	resp, err := userService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Empty(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	mockRepo, userService := setupUserServiceTesting()

	hashedPw, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	assert.NoError(t, err)

	mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(domain.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedPw),
		Role:     "user",
	}, nil)

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := userService.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_NotFound(t *testing.T) {
	mockRepo, userService := setupUserServiceTesting()

	mockRepo.On("FindByEmail", mock.Anything, "notexist@example.com").Return(domain.User{}, domain.ErrNotFound)

	req := dto.LoginRequest{
		Email:    "notexist@example.com",
		Password: "password123",
	}

	resp, err := userService.Login(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	mockRepo, userService := setupUserServiceTesting()

	hashedPw, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	assert.NoError(t, err)

	mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(domain.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPw),
		Role:     "user",
	}, nil)

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	resp, err := userService.Login(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.Empty(t, resp)
	mockRepo.AssertExpectations(t)
}
