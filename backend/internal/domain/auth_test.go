package domain_test

import (
	"testing"
	"time"

	"github.com/clubepay/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	hash, err := domain.HashPassword("mypassword")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mypassword", hash)
}

func TestCheckPassword(t *testing.T) {
	hash, _ := domain.HashPassword("mypassword")
	assert.True(t, domain.CheckPassword("mypassword", hash))
	assert.False(t, domain.CheckPassword("wrongpassword", hash))
}

func TestGenerateAndParseJWT(t *testing.T) {
	secret := "test-secret-key"
	userID := int64(42)
	role := "owner"

	token, err := domain.GenerateJWT(userID, role, secret, 1*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := domain.ParseJWT(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
}

func TestParseJWT_InvalidToken(t *testing.T) {
	_, err := domain.ParseJWT("invalid-token", "secret")
	assert.Error(t, err)
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token, _ := domain.GenerateJWT(1, "owner", secret, -1*time.Hour)
	_, err := domain.ParseJWT(token, secret)
	assert.Error(t, err)
}
