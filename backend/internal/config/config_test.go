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
	t.Setenv("SMTP_HOST", "")
	t.Setenv("MONTHLY_BUDGET_CENTS", "")
	t.Setenv("SPENDING_ALERT_EMAIL", "")
	t.Setenv("WARN_THRESHOLD_PCT", "")
	t.Setenv("CRITICAL_THRESHOLD_PCT", "")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "https://sandbox.asaas.com/api/v3", cfg.AsaasURL)
	assert.Equal(t, "587", cfg.SMTPPort)
	assert.Equal(t, "*", cfg.CORSOrigins)
	assert.Equal(t, "http://localhost:3000", cfg.FrontendURL)
	assert.Equal(t, "localhost", cfg.SMTPHost)
	assert.Equal(t, int64(500000), cfg.MonthlyBudgetCents)
	assert.Equal(t, "ceo@clubepay.com", cfg.SpendingAlertEmail)
	assert.Equal(t, 80, cfg.WarnThresholdPct)
	assert.Equal(t, 95, cfg.CriticalThresholdPct)
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
	t.Setenv("MONTHLY_BUDGET_CENTS", "1000000")
	t.Setenv("SPENDING_ALERT_EMAIL", "alerts@example.com")
	t.Setenv("WARN_THRESHOLD_PCT", "75")
	t.Setenv("CRITICAL_THRESHOLD_PCT", "90")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "key123", cfg.AsaasAPIKey)
	assert.Equal(t, "https://api.asaas.com/api/v3", cfg.AsaasURL)
	assert.Equal(t, "smtp.example.com", cfg.SMTPHost)
	assert.Equal(t, "465", cfg.SMTPPort)
	assert.Equal(t, "https://example.com", cfg.CORSOrigins)
	assert.Equal(t, "https://app.example.com", cfg.FrontendURL)
	assert.Equal(t, int64(1000000), cfg.MonthlyBudgetCents)
	assert.Equal(t, "alerts@example.com", cfg.SpendingAlertEmail)
	assert.Equal(t, 75, cfg.WarnThresholdPct)
	assert.Equal(t, 90, cfg.CriticalThresholdPct)
}

func TestParseIntEnv(t *testing.T) {
	t.Run("with valid value", func(t *testing.T) {
		t.Setenv("WARN_THRESHOLD_PCT", "75")
		// Call Load to test parseIntEnv indirectly
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("MONTHLY_BUDGET_CENTS", "")
		t.Setenv("SPENDING_ALERT_EMAIL", "")
		t.Setenv("CRITICAL_THRESHOLD_PCT", "")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, 75, cfg.WarnThresholdPct)
	})

	t.Run("with invalid value returns fallback", func(t *testing.T) {
		t.Setenv("WARN_THRESHOLD_PCT", "invalid")
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("MONTHLY_BUDGET_CENTS", "")
		t.Setenv("SPENDING_ALERT_EMAIL", "")
		t.Setenv("CRITICAL_THRESHOLD_PCT", "")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, 80, cfg.WarnThresholdPct) // fallback value
	})

	t.Run("with empty value returns fallback", func(t *testing.T) {
		t.Setenv("WARN_THRESHOLD_PCT", "")
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("MONTHLY_BUDGET_CENTS", "")
		t.Setenv("SPENDING_ALERT_EMAIL", "")
		t.Setenv("CRITICAL_THRESHOLD_PCT", "")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, 80, cfg.WarnThresholdPct) // fallback value
	})
}

func TestParseInt64Env(t *testing.T) {
	t.Run("with valid value", func(t *testing.T) {
		t.Setenv("MONTHLY_BUDGET_CENTS", "2500000")
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("SPENDING_ALERT_EMAIL", "")
		t.Setenv("WARN_THRESHOLD_PCT", "")
		t.Setenv("CRITICAL_THRESHOLD_PCT", "")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, int64(2500000), cfg.MonthlyBudgetCents)
	})

	t.Run("with invalid value returns fallback", func(t *testing.T) {
		t.Setenv("MONTHLY_BUDGET_CENTS", "not_a_number")
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("SPENDING_ALERT_EMAIL", "")
		t.Setenv("WARN_THRESHOLD_PCT", "")
		t.Setenv("CRITICAL_THRESHOLD_PCT", "")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, int64(500000), cfg.MonthlyBudgetCents) // fallback value
	})

	t.Run("with empty value returns fallback", func(t *testing.T) {
		t.Setenv("MONTHLY_BUDGET_CENTS", "")
		t.Setenv("DATABASE_URL", "postgres://localhost/test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("SMTP_HOST", "")
		t.Setenv("SPENDING_ALERT_EMAIL", "")
		t.Setenv("WARN_THRESHOLD_PCT", "")
		t.Setenv("CRITICAL_THRESHOLD_PCT", "")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, int64(500000), cfg.MonthlyBudgetCents) // fallback value
	})
}
