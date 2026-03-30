package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

type BusinessService struct {
	Queries *repository.Queries
}

func NewBusinessService(q *repository.Queries) *BusinessService {
	return &BusinessService{Queries: q}
}

func (s *BusinessService) GetByOwner(ctx context.Context, ownerID int64) (*domain.BusinessResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}
	return toBizResponse(&biz), nil
}

func (s *BusinessService) Update(ctx context.Context, ownerID int64, input domain.UpdateBusinessInput) (*domain.BusinessResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}

	updated, err := s.Queries.UpdateBusiness(ctx, repository.UpdateBusinessParams{
		ID:      biz.ID,
		Name:    input.Name,
		Segment: input.Segment,
		Address: pgText(input.Address),
		LogoUrl: pgText(input.LogoURL),
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao atualizar negócio", err)
	}
	return toBizResponse(&updated), nil
}

func toBizResponse(b *repository.Business) *domain.BusinessResponse {
	return &domain.BusinessResponse{
		ID:      b.ID,
		Name:    b.Name,
		Slug:    b.Slug,
		Segment: b.Segment,
		Address: b.Address.String,
		LogoURL: b.LogoUrl.String,
	}
}

func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
