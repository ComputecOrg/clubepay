package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	emailtmpl "github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
)

// SubscriptionService handles subscription business logic.
type SubscriptionService struct {
	Queries *repository.Queries
	PSP     psp.PSP
	Email   email.Sender
}

// SubscribeInput contains the parameters required to create a new subscription.
type SubscribeInput struct {
	PlanID int64
}

// SubscriptionListResponse wraps a list of subscriptions with subscriber info.
type SubscriptionListResponse struct {
	Subscriptions []repository.ListSubscriptionsByBusinessIDRow `json:"subscriptions"`
}

// CancelResponse is returned after cancelling a subscription.
type CancelResponse struct {
	Status    string     `json:"status"`
	PeriodEnd *time.Time `json:"period_end,omitempty"`
	Message   string     `json:"message"`
}

// MyPlanResponse combines plan, business, and subscription info for the subscriber.
type MyPlanResponse struct {
	Plan         domain.PlanDetailResponse `json:"plan"`
	Business     domain.BusinessResponse   `json:"business"`
	Subscription domain.SubscriptionInfo   `json:"subscription"`
}

// pgTimestamptz converts a time.Time to a pgtype.Timestamptz.
func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// Subscribe creates a new subscription for the given subscriber.
// It performs the full orchestration: plan check, duplicate check, free tier limit,
// PSP customer/subscription creation, referral discount, DB creation, and welcome email.
func (s *SubscriptionService) Subscribe(ctx context.Context, subscriberID int64, input SubscribeInput) (*repository.Subscription, error) {
	// Verify plan exists and is active
	plan, err := s.Queries.GetPlanByID(ctx, input.PlanID)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("plano não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar plano", err)
	}
	if !plan.Active {
		return nil, domain.NewErrBadRequest("plano não está ativo")
	}

	// Check for existing active subscription in this business
	_, err = s.Queries.GetActiveSubscription(ctx, repository.GetActiveSubscriptionParams{
		SubscriberID: subscriberID,
		BusinessID:   plan.BusinessID,
	})
	if err == nil {
		return nil, domain.NewErrConflict("já possui assinatura ativa neste negócio")
	}
	if !isNotFound(err) {
		return nil, domain.NewErrInternal("erro ao verificar assinatura existente", err)
	}

	// Check free tier subscriber limit
	count, err := s.Queries.CountActiveSubscriptionsByBusinessID(ctx, plan.BusinessID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao contar assinantes", err)
	}
	if count >= int64(domain.FreeTierSubscriberLimit) {
		return nil, domain.NewErrForbidden("limite de assinantes atingido no plano gratuito")
	}

	// Get subscriber data for PSP
	subscriber, err := s.Queries.GetUserByID(ctx, subscriberID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao buscar dados do assinante", err)
	}

	// Create PSP customer
	customer, err := s.PSP.CreateCustomer(ctx, psp.CreateCustomerRequest{
		Name:  subscriber.Name,
		Email: subscriber.Email,
		Phone: subscriber.Phone.String,
	})
	if err != nil {
		slog.Error("failed to create PSP customer", "error", err)
		return nil, domain.NewErrInternal("erro ao criar cliente no PSP", err)
	}

	// Check for active referral discount
	discountPercent := int32(0)
	referral, refErr := s.Queries.GetReferralByReferredAndBusiness(ctx, repository.GetReferralByReferredAndBusinessParams{
		ReferredID: subscriberID,
		BusinessID: plan.BusinessID,
	})
	if refErr == nil {
		discountPercent = domain.ReferralDiscountPercent
	}

	// Calculate discounted price
	priceCents := plan.PriceCents
	if discountPercent > 0 {
		priceCents = priceCents * int64(100-discountPercent) / 100
	}

	// Create PSP subscription
	pspSub, err := s.PSP.CreateSubscription(ctx, psp.CreateSubscriptionRequest{
		CustomerID:  customer.ID,
		PriceCents:  priceCents,
		Description: plan.Name,
		Cycle:       "MONTHLY",
	})
	if err != nil {
		slog.Error("failed to create PSP subscription", "error", err)
		return nil, domain.NewErrInternal("erro ao criar assinatura no PSP", err)
	}

	// Build DB params
	periodEnd := time.Now().AddDate(0, 1, 0)
	referredBy := pgtype.Int8{}
	if refErr == nil {
		referredBy = pgtype.Int8{Int64: referral.ReferrerID, Valid: true}
	}

	sub, err := s.Queries.CreateSubscriptionWithDiscount(ctx, repository.CreateSubscriptionWithDiscountParams{
		PlanID:            plan.ID,
		SubscriberID:      subscriberID,
		BusinessID:        plan.BusinessID,
		PspSubscriptionID: pgText(pspSub.ID),
		Status:            domain.SubscriptionStatusActive,
		PeriodEnd:         pgTimestamptz(periodEnd),
		ReferredBy:        referredBy,
		DiscountPercent:   discountPercent,
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao criar assinatura", err)
	}

	// Send welcome email asynchronously
	go s.sendWelcomeEmail(subscriberID, plan.ID, plan.BusinessID)

	return &sub, nil
}

// ListByBusiness returns all subscriptions for the given owner's business.
func (s *SubscriptionService) ListByBusiness(ctx context.Context, ownerID int64) (*SubscriptionListResponse, error) {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("negócio não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}

	subs, err := s.Queries.ListSubscriptionsByBusinessID(ctx, biz.ID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao listar assinaturas", err)
	}

	return &SubscriptionListResponse{Subscriptions: subs}, nil
}

// CancelByOwner cancels a subscription by the owner, verifying business ownership.
func (s *SubscriptionService) CancelByOwner(ctx context.Context, ownerID, subID int64) error {
	biz, err := s.Queries.GetBusinessByOwnerID(ctx, ownerID)
	if err != nil {
		if isNotFound(err) {
			return domain.NewErrNotFound("negócio não encontrado")
		}
		return domain.NewErrInternal("erro ao buscar negócio", err)
	}

	sub, err := s.Queries.GetSubscriptionByID(ctx, subID)
	if err != nil {
		if isNotFound(err) {
			return domain.NewErrNotFound("assinatura não encontrada")
		}
		return domain.NewErrInternal("erro ao buscar assinatura", err)
	}

	if sub.BusinessID != biz.ID {
		return domain.NewErrForbidden("assinatura não pertence ao seu negócio")
	}

	// Cancel in PSP (best effort)
	s.cancelPSPSubscription(ctx, sub.PspSubscriptionID)

	// Cancel in DB
	if err := s.Queries.CancelSubscription(ctx, subID); err != nil {
		return domain.NewErrInternal("erro ao cancelar assinatura", err)
	}

	return nil
}

// CancelBySubscriber allows a subscriber to cancel their own active subscription.
func (s *SubscriptionService) CancelBySubscriber(ctx context.Context, subscriberID int64) (*CancelResponse, error) {
	sub, err := s.Queries.GetActiveSubscriptionBySubscriber(ctx, subscriberID)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("assinatura não encontrada")
		}
		return nil, domain.NewErrInternal("erro ao buscar assinatura", err)
	}

	// Cancel in PSP (best effort)
	s.cancelPSPSubscription(ctx, sub.PspSubscriptionID)

	// Cancel in DB
	if err := s.Queries.CancelSubscription(ctx, sub.ID); err != nil {
		return nil, domain.NewErrInternal("erro ao cancelar assinatura", err)
	}

	// Send cancellation email asynchronously
	go s.sendCancellationEmail(subscriberID, sub.PlanID, sub.PeriodEnd)

	resp := &CancelResponse{
		Status:  "cancelled",
		Message: "Assinatura cancelada. Acesso continua até o fim do período pago.",
	}
	if sub.PeriodEnd.Valid {
		t := sub.PeriodEnd.Time
		resp.PeriodEnd = &t
	}
	return resp, nil
}

// GetMyPlan returns the subscription, plan and business info for the authenticated subscriber.
func (s *SubscriptionService) GetMyPlan(ctx context.Context, subscriberID int64) (*MyPlanResponse, error) {
	sub, err := s.Queries.GetActiveSubscriptionBySubscriber(ctx, subscriberID)
	if err != nil {
		if isNotFound(err) {
			return nil, domain.NewErrNotFound("assinatura não encontrada")
		}
		return nil, domain.NewErrInternal("erro ao buscar assinatura", err)
	}

	plan, err := s.Queries.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao buscar plano", err)
	}

	biz, err := s.Queries.GetBusinessByID(ctx, sub.BusinessID)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao buscar negócio", err)
	}

	planResp := domain.PlanDetailResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description.String,
		PriceCents:  plan.PriceCents,
		LimitType:   plan.LimitType,
		LimitCount:  plan.LimitCount,
		Active:      plan.Active,
	}

	bizResp := domain.BusinessResponse{
		ID:      biz.ID,
		Name:    biz.Name,
		Slug:    biz.Slug,
		Segment: biz.Segment,
		Address: biz.Address.String,
		LogoURL: biz.LogoUrl.String,
	}

	subInfo := domain.SubscriptionInfo{
		ID:     sub.ID,
		Status: sub.Status,
	}
	if sub.PeriodEnd.Valid {
		t := sub.PeriodEnd.Time
		subInfo.PeriodEnd = &t
	}

	return &MyPlanResponse{
		Plan:         planResp,
		Business:     bizResp,
		Subscription: subInfo,
	}, nil
}

// cancelPSPSubscription cancels the PSP subscription in a best-effort manner.
// Errors are logged but not returned.
func (s *SubscriptionService) cancelPSPSubscription(ctx context.Context, pspSubID pgtype.Text) {
	if !pspSubID.Valid || pspSubID.String == "" {
		return
	}
	if err := s.PSP.CancelSubscription(ctx, pspSubID.String); err != nil {
		slog.Error("failed to cancel PSP subscription", "error", err, "psp_id", pspSubID.String)
	}
}

// sendWelcomeEmail sends a welcome email to the subscriber asynchronously.
func (s *SubscriptionService) sendWelcomeEmail(subscriberID, planID, businessID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subscriber, err := s.Queries.GetUserByID(ctx, subscriberID)
	if err != nil {
		slog.Error("sendWelcomeEmail: failed to get subscriber", "error", err, "subscriber_id", subscriberID)
		return
	}

	plan, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		slog.Error("sendWelcomeEmail: failed to get plan", "error", err, "plan_id", planID)
		return
	}

	biz, err := s.Queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		slog.Error("sendWelcomeEmail: failed to get business", "error", err, "business_id", businessID)
		return
	}

	subject, body := emailtmpl.WelcomeEmail(subscriber.Name, plan.Name, biz.Name)
	if err := s.Email.Send(subscriber.Email, subject, body); err != nil {
		slog.Error("sendWelcomeEmail: failed to send email", "error", err, "subscriber_id", subscriberID)
	}
}

// sendCancellationEmail sends a cancellation email to the subscriber asynchronously.
func (s *SubscriptionService) sendCancellationEmail(subscriberID, planID int64, periodEnd pgtype.Timestamptz) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subscriber, err := s.Queries.GetUserByID(ctx, subscriberID)
	if err != nil {
		slog.Error("sendCancellationEmail: failed to get subscriber", "error", err, "subscriber_id", subscriberID)
		return
	}

	plan, err := s.Queries.GetPlanByID(ctx, planID)
	if err != nil {
		slog.Error("sendCancellationEmail: failed to get plan", "error", err, "plan_id", planID)
		return
	}

	validUntil := ""
	if periodEnd.Valid {
		validUntil = periodEnd.Time.Format("02/01/2006")
	}

	subject, body := emailtmpl.SubscriptionCancelledEmail(subscriber.Name, plan.Name, validUntil)
	if err := s.Email.Send(subscriber.Email, subject, body); err != nil {
		slog.Error("sendCancellationEmail: failed to send email", "error", err, "subscriber_id", subscriberID)
	}
}

// isNotFound returns true if err is a pgx ErrNoRows sentinel.
func isNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
