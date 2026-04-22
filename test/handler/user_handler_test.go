package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/handler"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService mocks service.UserService interface.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(dto.UserResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(dto.LoginResponse), args.Error(1)
}

func TestRegisterHandler_OK(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	reqBody := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	expectedResp := dto.UserResponse{ID: 1, Name: "Test User", Email: "test@example.com"}
	mockService.On("Register", mock.Anything, reqBody).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Register(rec, req, nil)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockService.AssertExpectations(t)
}

func TestRegisterHandler_BadJSON(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Register(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "Register")
}

func TestRegisterHandler_Conflict(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	reqBody := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "existing@example.com",
		Password: "password123",
	}
	mockService.On("Register", mock.Anything, reqBody).Return(dto.UserResponse{}, domain.ErrConflict)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Register(rec, req, nil)

	assert.Equal(t, http.StatusConflict, rec.Code)
	mockService.AssertExpectations(t)
}

func TestLoginHandler_OK(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	expectedResp := dto.LoginResponse{AccessToken: "token-abc", TokenType: "Bearer"}
	mockService.On("Login", mock.Anything, reqBody).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Login(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestLoginHandler_BadJSON(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Login(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "Login")
}

func TestLoginHandler_Unauthorized(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	mockService.On("Login", mock.Anything, reqBody).Return(dto.LoginResponse{}, domain.ErrUnauthorized)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Login(rec, req, nil)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockService.AssertExpectations(t)
}

func TestLoginHandler_NotFound(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	reqBody := dto.LoginRequest{
		Email:    "nobody@example.com",
		Password: "somepassword",
	}
	mockService.On("Login", mock.Anything, reqBody).Return(dto.LoginResponse{}, domain.ErrNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	userHandler.Login(rec, req, nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}
