package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/config"
)

func TestLoad_RequiredFields(t *testing.T) {
	t.Run("missing DATABASE_URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("JWT_SECRET", "test-secret")
		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL")
	})

	t.Run("missing JWT_SECRET", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "")
		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET")
	})
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("PORT", "")
	t.Setenv("ASAAS_URL", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("FRONTEND_URL", "")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "https://sandbox.asaas.com/api/v3", cfg.AsaasURL)
	assert.Equal(t, "587", cfg.SMTPPort)
	assert.Equal(t, "*", cfg.CORSOrigins)
	assert.Equal(t, "http://localhost:3000", cfg.FrontendURL)
}

func TestLoad_AllVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://prod/db")
	t.Setenv("JWT_SECRET", "prod-secret")
	t.Setenv("PORT", "9090")
	t.Setenv("ASAAS_API_KEY", "key123")
	t.Setenv("ASAAS_URL", "https://api.asaas.com/api/v3")
	t.Setenv("CRON_SECRET", "cron123")
	t.Setenv("ASAAS_WEBHOOK_SECRET", "whsec123")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "user@example.com")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("CORS_ORIGINS", "https://example.com")
	t.Setenv("FRONTEND_URL", "https://app.example.com")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "key123", cfg.AsaasAPIKey)
	assert.Equal(t, "https://api.asaas.com/api/v3", cfg.AsaasURL)
	assert.Equal(t, "smtp.example.com", cfg.SMTPHost)
	assert.Equal(t, "465", cfg.SMTPPort)
	assert.Equal(t, "https://example.com", cfg.CORSOrigins)
	assert.Equal(t, "https://app.example.com", cfg.FrontendURL)
}
