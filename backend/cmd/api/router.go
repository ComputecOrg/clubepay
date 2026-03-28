package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/middleware"
)

func setupRouter(cfg *config.Config, h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public (no auth)
	r.Get("/api/public/business/{slug}", h.GetPublicBusiness)
	r.Get("/api/public/plans/{slug}", h.GetPublicPlans)

	// Auth (no auth)
	r.Post("/api/auth/register", h.RegisterOwner)
	r.Post("/api/auth/login", h.Login)
	r.Post("/api/auth/register-subscriber", h.RegisterSubscriber)

	// Webhook + Cron (custom auth)
	r.Post("/api/psp/webhook", h.PSPWebhook)
	r.Post("/api/cron/reconcile", h.Reconcile)

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
