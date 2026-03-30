package config

import (
	"fmt"
	"os"
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
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUsername:       os.Getenv("SMTP_USERNAME"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "*"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
