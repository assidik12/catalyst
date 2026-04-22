package pkg_test

import (
	"testing"

	"github.com/assidik12/go-restfull-api/internal/pkg/hash"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_Success(t *testing.T) {
	hasher := hash.NewCryptoHasher(bcrypt.MinCost)

	hashed, err := hasher.HashPassword("mysecretpassword")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
	// Hashed value should differ from plain-text password
	assert.NotEqual(t, "mysecretpassword", hashed)
}

func TestHashPassword_EmptyString(t *testing.T) {
	hasher := hash.NewCryptoHasher(bcrypt.MinCost)

	hashed, err := hasher.HashPassword("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
}

func TestComparePassword_Success(t *testing.T) {
	hasher := hash.NewCryptoHasher(bcrypt.MinCost)

	password := "correctpassword"
	hashed, err := hasher.HashPassword(password)
	assert.NoError(t, err)

	err = hasher.ComparePassword(hashed, password)

	assert.NoError(t, err)
}

func TestComparePassword_WrongPassword(t *testing.T) {
	hasher := hash.NewCryptoHasher(bcrypt.MinCost)

	hashed, err := hasher.HashPassword("correctpassword")
	assert.NoError(t, err)

	err = hasher.ComparePassword(hashed, "wrongpassword")

	assert.Error(t, err)
}

func TestComparePassword_InvalidHash(t *testing.T) {
	hasher := hash.NewCryptoHasher(bcrypt.MinCost)

	err := hasher.ComparePassword("not-a-valid-bcrypt-hash", "somepassword")

	assert.Error(t, err)
}
