package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

// SearchSubscribers searches for active subscribers by name or phone.
// GET /api/subscribers/search?q=xxx
func (h *Handler) SearchSubscribers(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "parametro 'q' e obrigatorio")
		return
	}

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negocio nao encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negocio")
		return
	}

	results, err := h.Queries.SearchSubscribersByBusiness(r.Context(), repository.SearchSubscribersByBusinessParams{
		BusinessID: biz.ID,
		Column2:    pgtype.Text{String: query, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar assinantes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
	})
}
