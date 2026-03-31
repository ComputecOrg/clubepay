package provider

import (
	"context"
	"fmt"

	"github.com/clubepay/backend/internal/config"
)

// BrevoProvider returns email service costs
type BrevoProvider struct {
	costCents int64
}

// NewBrevoProvider creates a Brevo cost provider
func NewBrevoProvider(cfg *config.Config) CostProvider {
	return &BrevoProvider{
		costCents: cfg.BrevoEmailCostCents,
	}
}

// GetMonthlyCost returns the monthly Brevo cost
func (b *BrevoProvider) GetMonthlyCost(ctx context.Context) (Cost, error) {
	var desc string
	if b.costCents == 0 {
		desc = "Brevo SMTP (free tier, 300 emails/day)"
	} else {
		desc = fmt.Sprintf("Brevo SMTP (paid plan, €%.2f/month)", float64(b.costCents)/100)
	}

	return Cost{
		CostCents:   b.costCents,
		Provider:    "brevo",
		Description: desc,
	}, nil
}
