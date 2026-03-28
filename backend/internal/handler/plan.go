package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/middleware"
	"github.com/clubepay/backend/internal/repository"
)

const freeTierPlanLimit = 1

// CreatePlan creates a new plan for the authenticated owner's business.
func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PriceCents  int64  `json:"price_cents"`
		LimitType   string `json:"limit_type"`
		LimitCount  int32  `json:"limit_count"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// Validate required fields
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}
	if input.PriceCents <= 0 {
		writeError(w, http.StatusBadRequest, "price_cents deve ser maior que zero")
		return
	}
	if input.LimitType != "daily" && input.LimitType != "monthly" {
		writeError(w, http.StatusBadRequest, "limit_type deve ser 'daily' ou 'monthly'")
		return
	}
	if input.LimitCount <= 0 {
		writeError(w, http.StatusBadRequest, "limit_count deve ser maior que zero")
		return
	}

	// Free tier: max 1 active plan
	count, err := h.Queries.CountPlansByBusinessID(r.Context(), biz.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao verificar planos")
		return
	}
	if count >= int64(freeTierPlanLimit) {
		writeError(w, http.StatusForbidden, "limite de plano atingido no plano gratuito")
		return
	}

	plan, err := h.Queries.CreatePlan(r.Context(), repository.CreatePlanParams{
		BusinessID:  biz.ID,
		Name:        input.Name,
		Description: pgText(input.Description),
		PriceCents:  input.PriceCents,
		LimitType:   input.LimitType,
		LimitCount:  input.LimitCount,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar plano")
		return
	}

	writeJSON(w, http.StatusCreated, plan)
}

// ListPlans returns all active plans for the authenticated owner's business.
func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
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

	writeJSON(w, http.StatusOK, plans)
}

// UpdatePlan updates an existing plan by ID. Verifies the plan belongs to the owner's business.
func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	planID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	// Verify plan belongs to this business
	existing, err := h.Queries.GetPlanByID(r.Context(), planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "plano não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar plano")
		return
	}
	if existing.BusinessID != biz.ID {
		writeError(w, http.StatusForbidden, "plano não pertence ao seu negócio")
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PriceCents  int64  `json:"price_cents"`
		LimitType   string `json:"limit_type"`
		LimitCount  int32  `json:"limit_count"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}
	if input.PriceCents <= 0 {
		writeError(w, http.StatusBadRequest, "price_cents deve ser maior que zero")
		return
	}
	if input.LimitType != "daily" && input.LimitType != "monthly" {
		writeError(w, http.StatusBadRequest, "limit_type deve ser 'daily' ou 'monthly'")
		return
	}
	if input.LimitCount <= 0 {
		writeError(w, http.StatusBadRequest, "limit_count deve ser maior que zero")
		return
	}

	updated, err := h.Queries.UpdatePlan(r.Context(), repository.UpdatePlanParams{
		ID:          planID,
		Name:        input.Name,
		Description: pgText(input.Description),
		PriceCents:  input.PriceCents,
		LimitType:   input.LimitType,
		LimitCount:  input.LimitCount,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar plano")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeletePlan deactivates a plan by ID (soft delete). Verifies ownership.
func (h *Handler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	planID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	biz, err := h.Queries.GetBusinessByOwnerID(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "negócio não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar negócio")
		return
	}

	// Verify plan belongs to this business
	existing, err := h.Queries.GetPlanByID(r.Context(), planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "plano não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar plano")
		return
	}
	if existing.BusinessID != biz.ID {
		writeError(w, http.StatusForbidden, "plano não pertence ao seu negócio")
		return
	}

	if err := h.Queries.DeactivatePlan(r.Context(), planID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao desativar plano")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
