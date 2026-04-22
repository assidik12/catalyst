package pkg_test

import (
	"testing"
	"time"

	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT_Success(t *testing.T) {
	jwtSvc := jwt.NewJWTService("test_secret")

	user := domain.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "user",
	}

	token, err := jwtSvc.GenerateJWT(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken_Success(t *testing.T) {
	jwtSvc := jwt.NewJWTService("test_secret")

	user := domain.User{
		ID:    42,
		Email: "verify@example.com",
		Name:  "Verify User",
		Role:  "admin",
	}

	token, err := jwtSvc.GenerateJWT(user)
	assert.NoError(t, err)

	claims, err := jwtSvc.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, 42, claims.UserID)
	assert.Equal(t, "verify@example.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	jwtSvc := jwt.NewJWTService("correct_secret")

	user := domain.User{ID: 1, Email: "test@example.com", Role: "user"}
	token, err := jwtSvc.GenerateJWT(user)
	assert.NoError(t, err)

	// Validate with a different secret key
	wrongSvc := jwt.NewJWTService("wrong_secret")
	claims, err := wrongSvc.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	jwtSvc := jwt.NewJWTService("test_secret")

	claims, err := jwtSvc.ValidateToken("not.a.valid.jwt")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGenerateJWT_ContainsExpiry(t *testing.T) {
	jwtSvc := jwt.NewJWTService("test_secret")

	user := domain.User{ID: 1, Email: "expiry@example.com", Role: "user"}
	token, err := jwtSvc.GenerateJWT(user)
	assert.NoError(t, err)

	claims, err := jwtSvc.ValidateToken(token)
	assert.NoError(t, err)

	// Token should expire roughly 24 hours from now
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(25*time.Hour)))
}
