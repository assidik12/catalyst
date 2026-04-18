package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/middleware"
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
