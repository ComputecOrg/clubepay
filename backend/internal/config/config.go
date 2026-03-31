package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	AsaasAPIKey        string
	AsaasURL           string
	CronSecret         string
	AsaasWebhookSecret string
	SMTPHost           string
	SMTPPort           string
	SMTPUsername       string
	SMTPPassword       string
	CORSOrigins        string
	FrontendURL        string
	// New spending config
	MonthlyBudgetCents   int64
	SpendingAlertEmail   string
	WarnThresholdPct     int
	CriticalThresholdPct int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		AsaasAPIKey: os.Getenv("ASAAS_API_KEY"),
		AsaasURL:    getEnv("ASAAS_URL", "https://sandbox.asaas.com/api/v3"),
		CronSecret:         os.Getenv("CRON_SECRET"),
		AsaasWebhookSecret: os.Getenv("ASAAS_WEBHOOK_SECRET"),
		SMTPHost:           getEnv("SMTP_HOST", "localhost"),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUsername:       os.Getenv("SMTP_USERNAME"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "*"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		MonthlyBudgetCents: parseInt64Env("MONTHLY_BUDGET_CENTS", 500000),
		SpendingAlertEmail: getEnv("SPENDING_ALERT_EMAIL", "ceo@clubepay.com"),
		WarnThresholdPct:   parseIntEnv("WARN_THRESHOLD_PCT", 80),
		CriticalThresholdPct: parseIntEnv("CRITICAL_THRESHOLD_PCT", 95),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.WarnThresholdPct < 0 || cfg.WarnThresholdPct > 100 {
		return nil, fmt.Errorf("WARN_THRESHOLD_PCT must be between 0 and 100, got %d", cfg.WarnThresholdPct)
	}
	if cfg.CriticalThresholdPct < 0 || cfg.CriticalThresholdPct > 100 {
		return nil, fmt.Errorf("CRITICAL_THRESHOLD_PCT must be between 0 and 100, got %d", cfg.CriticalThresholdPct)
	}
	if cfg.WarnThresholdPct >= cfg.CriticalThresholdPct {
		return nil, fmt.Errorf("WARN_THRESHOLD_PCT (%d) must be less than CRITICAL_THRESHOLD_PCT (%d)", cfg.WarnThresholdPct, cfg.CriticalThresholdPct)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt64Env(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return val
}

func parseIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return val
}
