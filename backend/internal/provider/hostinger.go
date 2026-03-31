package provider

import (
	"context"

	"github.com/clubepay/backend/internal/config"
)

// HostingerProvider returns static VPS cost
type HostingerProvider struct {
	costCents int64
}

// NewHostingerProvider creates a Hostinger cost provider
func NewHostingerProvider(cfg *config.Config) CostProvider {
	return &HostingerProvider{
		costCents: cfg.HostingerVPSCostCents,
	}
}

// GetMonthlyCost returns the monthly VPS cost
func (h *HostingerProvider) GetMonthlyCost(ctx context.Context) (Cost, error) {
	return Cost{
		CostCents:   h.costCents,
		Provider:    "hostinger",
		Description: "Hostinger VPS KVM1 (R$12.99/month)",
	}, nil
}
