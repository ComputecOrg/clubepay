package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
)

type Handler struct {
	Queries *repository.Queries
	Config  *config.Config
	PSP     psp.PSP
	Email   email.Sender
}

func New(q *repository.Queries, cfg *config.Config, p psp.PSP, e email.Sender) *Handler {
	return &Handler{Queries: q, Config: cfg, PSP: p, Email: e}
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
