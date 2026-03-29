package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type publicBusinessResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Segment         string `json:"segment"`
	SubscriberCount int64  `json:"subscriber_count"`
}

// GetPublicBusiness returns business info by slug for the public landing page. No auth required.
func (h *Handler) GetPublicBusiness(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	biz, err := h.Queries.GetBusinessBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	count, err := h.Queries.CountActiveSubscriptionsByBusinessID(r.Context(), biz.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao contar assinantes")
		return
	}

	resp := publicBusinessResponse{
		ID:              biz.ID,
		Name:            biz.Name,
		Slug:            biz.Slug,
		Segment:         biz.Segment,
		SubscriberCount: count,
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetPublicPlans returns active plans for a business by slug. No auth required.
func (h *Handler) GetPublicPlans(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	biz, err := h.Queries.GetBusinessBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	plans, err := h.Queries.ListPlansByBusinessID(r.Context(), biz.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar planos")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plans": plans})
}
