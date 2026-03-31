package provider

import (
	"context"

	"github.com/clubepay/backend/internal/config"
)

// ClaudeAPIProvider returns Claude API costs
// TODO: Integrate with actual Claude billing API in future
type ClaudeAPIProvider struct {
	costCents int64
}

// NewClaudeAPIProvider creates a Claude API cost provider
func NewClaudeAPIProvider(cfg *config.Config) CostProvider {
	return &ClaudeAPIProvider{
		costCents: cfg.ClaudeAPICostCents,
	}
}

// GetMonthlyCost returns the monthly Claude API cost
func (c *ClaudeAPIProvider) GetMonthlyCost(ctx context.Context) (Cost, error) {
	return Cost{
		CostCents:   c.costCents,
		Provider:    "claude_api",
		Description: "Claude API Max 20x Plan ($200/month)",
	}, nil
}
