package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/middleware"
)

func setupRouter(cfg *config.Config, db DBPinger, h *handler.Handler) http.Handler {
	reg := prometheus.NewRegistry()

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.SentryRecovery(cfg.SentryDSN))
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.Logging)
	r.Use(middleware.NewPrometheus(reg))

	// Health checks
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	})
	r.Handle("/healthz", healthzHandler(db))

	// Prometheus metrics (acesso interno)
	r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// Public (no auth)
	r.Get("/api/public/business/{slug}", h.GetPublicBusiness)
	r.Get("/api/public/plans/{slug}", h.GetPublicPlans)

	// Auth (no auth, rate limited)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(10, time.Minute))
		r.Post("/api/auth/register", h.RegisterOwner)
		r.Post("/api/auth/login", h.Login)
		r.Post("/api/auth/register-subscriber", h.RegisterSubscriber)
		r.Post("/api/auth/request-password-reset", h.RequestPasswordReset)
		r.Post("/api/auth/confirm-password-reset", h.ConfirmPasswordReset)
	})

	// Webhook + Cron (custom auth)
	r.Post("/api/psp/webhook", h.PSPWebhook)
	r.Post("/api/cron/reconcile", h.Reconcile)

	// Shared auth routes (any authenticated user)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Get("/api/profile", h.GetProfile)
		r.Put("/api/profile", h.UpdateProfile)
		r.Post("/api/profile/change-password", h.ChangePassword)
	})

	// Owner routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Use(middleware.RequireRole("owner"))

		r.Get("/api/business", h.GetBusiness)
		r.Put("/api/business", h.UpdateBusiness)

		r.Get("/api/plans", h.ListPlans)
		r.Post("/api/plans", h.CreatePlan)
		r.Put("/api/plans/{id}", h.UpdatePlan)
		r.Delete("/api/plans/{id}", h.DeletePlan)

		r.Get("/api/subscriptions", h.ListSubscriptions)
		r.Delete("/api/subscriptions/{id}", h.CancelSubscriptionByOwner)

		r.Get("/api/subscribers/search", h.SearchSubscribers)
		r.Post("/api/validate-usage-owner", h.ValidateUsageByOwner)

		// Spending alerts
		r.Get("/api/owner/spending/status", h.GetSpendingStatus)
		r.Get("/api/owner/spending/history", h.GetSpendingHistory)
		r.Get("/api/owner/spending/alerts", h.GetAlertHistory)
	})

	// Subscriber routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Use(middleware.RequireRole("subscriber"))

		r.Post("/api/subscribe", h.Subscribe)
		r.Post("/api/validate-usage", h.ValidateUsage)
		r.Get("/api/my-usage", h.MyUsage)
		r.Get("/api/my-plan", h.MyPlan)
		r.Post("/api/cancel", h.CancelBySubscriber)
		r.Get("/api/my-referral-code", h.MyReferralCode)
		r.Post("/api/referrals/apply", h.ApplyReferral)
	})

	return r
}
