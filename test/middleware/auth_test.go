package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/middleware"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/pkg/jwt"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_MissingAuthorization(t *testing.T) {
	// Setup Next handler to be wrapped
	nextHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}

	authMiddleware := middleware.AuthMiddleware{}
	// "user" is the role, "secret_key" is the test key
	wrappedHandler := authMiddleware.Middleware("user", nextHandler, "secret_key")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// Missing Authorization header intentionally

	rec := httptest.NewRecorder()
	wrappedHandler(rec, req, nil)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Authorization header is missing")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Setup Next handler to be wrapped
	nextHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}

	authMiddleware := middleware.AuthMiddleware{}
	wrappedHandler := authMiddleware.Middleware("user", nextHandler, "secret_key")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// Add an invalid header format
	req.Header.Add("Authorization", "InvalidFormatWithoutBearer 123456etc")

	rec := httptest.NewRecorder()
	wrappedHandler(rec, req, nil)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Authorization header format is invalid")
}

func TestAuthMiddleware_ValidToken_AllowedRole(t *testing.T) {
	nextCalled := false
	nextHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authMiddleware := middleware.AuthMiddleware{}
	secretKey := "test_secret_key"
	wrappedHandler := authMiddleware.Middleware("user", nextHandler, secretKey)

	// Generate a valid token using the jwt package
	jwtSvc := jwt.NewJWTService(secretKey)
	token, err := jwtSvc.GenerateJWT(domain.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "user",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	wrappedHandler(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, nextCalled)
}

func TestAuthMiddleware_ValidToken_ForbiddenRole(t *testing.T) {
	nextHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}

	authMiddleware := middleware.AuthMiddleware{}
	secretKey := "test_secret_key"
	// Route requires "admin" role
	wrappedHandler := authMiddleware.Middleware("admin", nextHandler, secretKey)

	// Token belongs to a "user" role
	jwtSvc := jwt.NewJWTService(secretKey)
	token, err := jwtSvc.GenerateJWT(domain.User{
		ID:    2,
		Email: "user@example.com",
		Name:  "Regular User",
		Role:  "user",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	wrappedHandler(rec, req, nil)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	nextHandler := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}

	authMiddleware := middleware.AuthMiddleware{}
	wrappedHandler := authMiddleware.Middleware("user", nextHandler, "correct_secret")

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.jwt")

	rec := httptest.NewRecorder()
	wrappedHandler(rec, req, nil)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
