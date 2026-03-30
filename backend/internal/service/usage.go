package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

// UsageService handles usage validation business logic.
type UsageService struct {
	Queries *repository.Queries
}

// NewUsageService creates a new UsageService.
func NewUsageService(q *repository.Queries) *UsageService {
	return &UsageService{Queries: q}
}

// Validate validates a usage for a subscriber, looking up the business by slug.
func (s *UsageService) Validate(ctx context.Context, subscriberID int64, input domain.ValidateUsageInput) (*domain.ValidateUsageResponse, error) {
	biz, err := s.Queries.GetBusinessBySlug(ctx, input.BusinessSlug)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}
	return s.validateUsage(ctx, subscriberID, biz.ID)
}

// ValidateByOwner validates a usage for a subscriber, looking up the business by owner.
func (s *UsageService) ValidateByOwner(ctx context.Context, ownerID int64, input domain.ValidateUsageOwnerInput) (*domain.ValidateUsageResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}
	return s.validateUsage(ctx, input.SubscriberID, biz.ID)
}

// GetMyUsage returns the usage list for the current billing period for a subscriber.
func (s *UsageService) GetMyUsage(ctx context.Context, subscriberID int64, businessSlug string) (*domain.UsageListResponse, error) {
	biz, err := s.Queries.GetBusinessBySlug(ctx, businessSlug)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}

	sub, err := s.Queries.GetActiveSubscription(ctx, repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   biz.ID,
	})
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("assinatura ativa não encontrada")
		}
		return nil, domain.NewErrInternal("erro ao buscar assinatura", err)
	}

	plan, err := s.Queries.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao buscar plano", err)
	}

	now := time.Now()
	var start, end time.Time
	if plan.LimitType == "daily" {
		start, end = domain.DailyRange(now)
	} else {
		start, end = domain.MonthlyRange(now)
	}

	usages, err := s.Queries.ListUsagesBySubscription(ctx, repository.ListUsagesBySubscriptionParams{
		SubscriptionID: sub.ID,
		Column2:        pgtype.Timestamptz{Time: start, Valid: true},
		Column3:        pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao listar usos", err)
	}

	return &domain.UsageListResponse{
		Used:     len(usages),
		Limit:    plan.LimitCount,
		PlanName: plan.Name,
		Usages:   usages,
	}, nil
}

// validateUsage is a private helper that does the core usage validation and recording.
func (s *UsageService) validateUsage(ctx context.Context, subscriberID, businessID int64) (*domain.ValidateUsageResponse, error) {
	sub, err := s.Queries.GetActiveSubscription(ctx, repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   businessID,
	})
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("assinatura ativa não encontrada")
		}
		return nil, domain.NewErrInternal("erro ao buscar assinatura", err)
	}

	plan, err := s.Queries.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao buscar plano", err)
	}

	now := time.Now()
	var start, end time.Time
	if plan.LimitType == "daily" {
		start, end = domain.DailyRange(now)
	} else {
		start, end = domain.MonthlyRange(now)
	}

	var count int64
	if plan.LimitType == "daily" {
		count, err = s.Queries.CountDailyUsage(ctx, repository.CountDailyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgtype.Timestamptz{Time: start, Valid: true},
			Column3:        pgtype.Timestamptz{Time: end, Valid: true},
		})
	} else {
		count, err = s.Queries.CountMonthlyUsage(ctx, repository.CountMonthlyUsageParams{
			SubscriptionID: sub.ID,
			Column2:        pgtype.Timestamptz{Time: start, Valid: true},
			Column3:        pgtype.Timestamptz{Time: end, Valid: true},
		})
	}
	if err != nil {
		return nil, domain.NewErrInternal("erro ao contar usos", err)
	}

	if err := domain.CheckUsageLimit(plan.LimitType, int(plan.LimitCount), count); err != nil {
		return nil, domain.NewErrForbidden("limite de uso atingido para este período")
	}

	if _, err := s.Queries.CreateUsage(ctx, sub.ID); err != nil {
		return nil, domain.NewErrInternal("erro ao registrar uso", err)
	}

	return &domain.ValidateUsageResponse{
		Status:   "validated",
		Used:     count + 1,
		Limit:    plan.LimitCount,
		PlanName: plan.Name,
	}, nil
}
