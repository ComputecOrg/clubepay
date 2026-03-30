package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

// ReferralService handles referral code business logic.
type ReferralService struct {
	Queries *repository.Queries
}

// NewReferralService creates a new ReferralService.
func NewReferralService(q *repository.Queries) *ReferralService {
	return &ReferralService{Queries: q}
}

// GetOrCreateCode returns or creates a referral code for the subscriber's active subscription.
// The code is unique per business and stored as a self-referencing referral record.
func (s *ReferralService) GetOrCreateCode(ctx context.Context, subscriberID int64) (*domain.ReferralCodeResponse, error) {
	sub, err := s.Queries.GetActiveSubscriptionBySubscriber(ctx, subscriberID)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("assinatura ativa não encontrada")
		}
		return nil, domain.NewErrInternal("erro ao buscar assinatura", err)
	}

	biz, err := s.Queries.GetBusinessByID(ctx, sub.BusinessID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}

	// Check if a code already exists for this subscriber+business
	existing, err := s.Queries.GetReferralCodeBySubscriberAndBusiness(ctx, repository.GetReferralCodeBySubscriberAndBusinessParams{
		ReferrerID: subscriberID,
		BusinessID: biz.ID,
	})
	if err == nil {
		return &domain.ReferralCodeResponse{Code: existing}, nil
	}
	if !isNotFound(err) {
		return nil, domain.NewErrInternal("erro ao buscar código de indicação", err)
	}

	// Generate new 8-char hex code
	code, err := generateHexCode(8)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao gerar código de indicação", err)
	}

	// Create self-referencing referral record (referrer_id == referred_id)
	_, err = s.Queries.CreateReferral(ctx, repository.CreateReferralParams{
		ReferrerID: subscriberID,
		ReferredID: subscriberID,
		BusinessID: biz.ID,
		Code:       code,
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao criar código de indicação", err)
	}

	return &domain.ReferralCodeResponse{Code: code}, nil
}

// Apply applies a referral code for the subscriber, creating a referral record.
func (s *ReferralService) Apply(ctx context.Context, subscriberID int64, input domain.ApplyReferralInput) error {
	referral, err := s.Queries.GetReferralByCode(ctx, input.Code)
	if err != nil {
		if isNotFound(err) {
			return domain.NewErrNotFound("código de indicação não encontrado ou inativo")
		}
		return domain.NewErrInternal("erro ao buscar código de indicação", err)
	}

	// Prevent self-referral
	if referral.ReferrerID == subscriberID {
		return domain.NewErrBadRequest("você não pode usar seu próprio código de indicação")
	}

	// Check referral limit for the referrer.
	// CountActiveReferralsByReferrer includes the self-referencing code record, so we
	// subtract 1 to get the true number of referred users.
	count, err := s.Queries.CountActiveReferralsByReferrer(ctx, repository.CountActiveReferralsByReferrerParams{
		ReferrerID: referral.ReferrerID,
		BusinessID: referral.BusinessID,
	})
	if err != nil {
		return domain.NewErrInternal("erro ao contar indicações ativas", err)
	}
	// Subtract 1 for the self-referencing code record (referrer_id == referred_id).
	actualReferrals := count - 1
	if actualReferrals < 0 {
		actualReferrals = 0
	}
	if actualReferrals >= int64(domain.ReferralLimit) {
		return domain.NewErrForbidden("limite de indicações atingido para este indicador")
	}

	_, err = s.Queries.CreateReferral(ctx, repository.CreateReferralParams{
		ReferrerID: referral.ReferrerID,
		ReferredID: subscriberID,
		BusinessID: referral.BusinessID,
		Code:       input.Code,
	})
	if err != nil {
		return domain.NewErrInternal("erro ao registrar indicação", err)
	}

	return nil
}

// generateHexCode generates a random hex string of the given length (bytes → length chars).
// 8 chars = 4 bytes of random.
func generateHexCode(chars int) (string, error) {
	bytes := make([]byte, chars/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
