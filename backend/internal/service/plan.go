package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type PlanService struct {
	Queries *repository.Queries
}

func NewPlanService(q *repository.Queries) *PlanService {
	return &PlanService{Queries: q}
}

// getOwnerBusiness retrieves the business owned by the given ownerID.
func (s *PlanService) getOwnerBusiness(ctx context.Context, ownerID int64) (*repository.Business, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}
	return &biz, nil
}

// Create creates a new plan for the owner's business, respecting the free tier limit.
func (s *PlanService) Create(ctx context.Context, ownerID int64, input domain.CreatePlanInput) (*domain.PlanDetailResponse, error) {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	count, err := s.Queries.CountPlansByBusinessID(ctx, biz.ID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao contar planos", err)
	}
	if count >= domain.FreeTierPlanLimit {
		return nil, domain.NewErrForbidden("limite de planos do free tier atingido")
	}

	plan, err := s.Queries.CreatePlan(ctx, repository.CreatePlanParams{
		BusinessID:  biz.ID,
		Name:        input.Name,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		PriceCents:  input.PriceCents,
		LimitType:   input.LimitType,
		LimitCount:  input.LimitCount,
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao criar plano", err)
	}
	return toPlanResponse(&plan), nil
}

// List lists all active plans for the owner's business.
func (s *PlanService) List(ctx context.Context, ownerID int64) (*domain.PlanListResponse, error) {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	plans, err := s.Queries.ListPlansByBusinessID(ctx, biz.ID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao listar planos", err)
	}

	resp := &domain.PlanListResponse{Plans: make([]domain.PlanDetailResponse, 0, len(plans))}
	for i := range plans {
		resp.Plans = append(resp.Plans, *toPlanResponse(&plans[i]))
	}
	return resp, nil
}

// Update updates an existing plan, verifying ownership.
func (s *PlanService) Update(ctx context.Context, ownerID, planID int64, input domain.UpdatePlanInput) (*domain.PlanDetailResponse, error) {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	existing, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNotFound("plano não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar plano", err)
	}

	if existing.BusinessID != biz.ID {
		return nil, domain.NewErrForbidden("plano não pertence ao seu negócio")
	}

	updated, err := s.Queries.UpdatePlan(ctx, repository.UpdatePlanParams{
		ID:          planID,
		Name:        input.Name,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		PriceCents:  input.PriceCents,
		LimitType:   input.LimitType,
		LimitCount:  input.LimitCount,
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao atualizar plano", err)
	}
	return toPlanResponse(&updated), nil
}

// Delete deactivates a plan, verifying ownership.
func (s *PlanService) Delete(ctx context.Context, ownerID, planID int64) error {
	biz, err := s.getOwnerBusiness(ctx, ownerID)
	if err != nil {
		return err
	}

	existing, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NewErrNotFound("plano não encontrado")
		}
		return domain.NewErrInternal("erro ao buscar plano", err)
	}

	if existing.BusinessID != biz.ID {
		return domain.NewErrForbidden("plano não pertence ao seu negócio")
	}

	if err := s.Queries.DeactivatePlan(ctx, planID); err != nil {
		return domain.NewErrInternal("erro ao desativar plano", err)
	}
	return nil
}

func toPlanResponse(p *repository.Plan) *domain.PlanDetailResponse {
	return &domain.PlanDetailResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description.String,
		PriceCents:  p.PriceCents,
		LimitType:   p.LimitType,
		LimitCount:  p.LimitCount,
		Active:      p.Active,
	}
}
