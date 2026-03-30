package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
)

type Handler struct {
	Auth          *service.AuthService
	Business      *service.BusinessService
	Plans         *service.PlanService
	Subscriptions *service.SubscriptionService
	Usage         *service.UsageService
	Referrals     *service.ReferralService
	Config        *config.Config
	// Keep these for webhook, cron, public, and search handlers that need direct access
	Queries *repository.Queries
	PSP     psp.PSP
	Email   email.Sender
}

func New(q *repository.Queries, cfg *config.Config, p psp.PSP, e email.Sender) *Handler {
	return &Handler{
		Auth:          service.NewAuthService(q, cfg, e),
		Business:      service.NewBusinessService(q),
		Plans:         service.NewPlanService(q),
		Subscriptions: service.NewSubscriptionService(q, p, e),
		Usage:         service.NewUsageService(q),
		Referrals:     service.NewReferralService(q),
		Config:        cfg,
		Queries:       q,
		PSP:           p,
		Email:         e,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func handleServiceError(w http.ResponseWriter, err error) {
	var svcErr *domain.ServiceError
	if errors.As(err, &svcErr) {
		writeError(w, svcErr.Code, svcErr.Message)
		return
	}
	slog.Error("unhandled service error", "error", err)
	writeError(w, http.StatusInternalServerError, "erro interno")
}
