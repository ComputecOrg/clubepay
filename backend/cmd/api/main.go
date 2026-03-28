package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to PostgreSQL
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	// Create repository queries
	queries := repository.New(pool)

	// Create PSP client (real Asaas or mock for dev)
	var pspClient psp.PSP
	if cfg.AsaasAPIKey != "" {
		pspClient = psp.NewAsaas(cfg.AsaasURL, cfg.AsaasAPIKey, cfg.AsaasWebhookSecret)
		slog.Info("using Asaas PSP client")
	} else {
		pspClient = &psp.MockPSP{}
		slog.Warn("ASAAS_API_KEY not set, using mock PSP client")
	}

	// Create email sender
	var emailSender email.Sender
	smtpSender := email.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	if smtpSender != nil {
		emailSender = smtpSender
		slog.Info("using SMTP email sender")
	} else {
		emailSender = &email.MockSender{}
		slog.Warn("SMTP not configured, using mock email sender")
	}

	// Wire handler
	h := handler.New(queries, cfg, pspClient, emailSender)

	// Setup router
	router := setupRouter(cfg, h)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
