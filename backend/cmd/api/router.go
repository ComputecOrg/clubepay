package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/clubepay/backend/internal/config"
)

func setupRouter(_ *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		// POST /api/auth/register
		// POST /api/auth/login
		// POST /api/auth/register-subscriber
	})

	// Business routes (owner auth required)
	r.Route("/api/business", func(r chi.Router) {
		// GET  /api/business
		// PUT  /api/business
	})

	// Plan routes (owner auth required)
	r.Route("/api/plans", func(r chi.Router) {
		// GET    /api/plans
		// POST   /api/plans
		// PUT    /api/plans/{id}
		// DELETE /api/plans/{id}
	})

	// Subscription routes
	r.Route("/api", func(r chi.Router) {
		// POST /api/subscribe
		// GET  /api/subscriptions (owner)
		// DELETE /api/subscriptions/{id} (owner)
	})

	// Usage routes
	r.Route("/api", func(r chi.Router) {
		// POST /api/validate-usage (subscriber)
		// GET  /api/my-usage (subscriber)
	})

	// Subscriber routes
	r.Route("/api", func(r chi.Router) {
		// GET  /api/my-plan (subscriber)
		// POST /api/cancel (subscriber)
	})

	// Referral routes
	r.Route("/api", func(r chi.Router) {
		// GET  /api/my-referral-code (subscriber)
		// POST /api/referrals/apply
	})

	// Webhook
	r.Post("/api/psp/webhook", func(w http.ResponseWriter, r *http.Request) {
		// PSP webhook handler (HMAC validation)
		w.WriteHeader(http.StatusOK)
	})

	// Cron
	r.Post("/api/cron/reconcile", func(w http.ResponseWriter, r *http.Request) {
		// Reconciliation cron (secret auth)
		w.WriteHeader(http.StatusOK)
	})

	// Public routes (no auth)
	r.Route("/api/public", func(r chi.Router) {
		// GET /api/public/business/{slug}
		// GET /api/public/plans/{slug}
	})

	return r
}
