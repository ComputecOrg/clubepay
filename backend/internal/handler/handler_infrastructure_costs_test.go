package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/handler"
	"github.com/clubepay/backend/internal/provider"
)

type mockCostProviderWithError struct {
	cost int64
	err  error
}

func (m *mockCostProviderWithError) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	if m.err != nil {
		return provider.Cost{}, m.err
	}
	return provider.Cost{CostCents: m.cost, Provider: "mock"}, nil
}

// TestHandler_CalculateTotalInfrastructureCost_WithAggregator verifies handler calculates total cost
func TestHandler_CalculateTotalInfrastructureCost_WithAggregator(t *testing.T) {
	providers := []provider.CostProvider{
		&mockCostProviderWithError{cost: 6495},   // Hostinger
		&mockCostProviderWithError{cost: 20000},  // Claude API
		&mockCostProviderWithError{cost: 0},      // Brevo free tier
	}
	agg := provider.NewAggregator(providers)

	h := &handler.Handler{
		CostAggregator: agg,
	}

	ctx := context.Background()
	total := h.CalculateTotalInfrastructureCost(ctx)

	assert.Equal(t, int64(26495), total)
}

// TestHandler_CalculateTotalInfrastructureCost_NoAggregator verifies safe handling when nil
func TestHandler_CalculateTotalInfrastructureCost_NoAggregator(t *testing.T) {
	h := &handler.Handler{
		CostAggregator: nil,
	}

	ctx := context.Background()
	total := h.CalculateTotalInfrastructureCost(ctx)

	// Should return 0 safely when aggregator is not configured
	assert.Equal(t, int64(0), total)
}

// TestHandler_CalculateTotalInfrastructureCost_ProviderErrors verifies non-blocking behavior
func TestHandler_CalculateTotalInfrastructureCost_ProviderErrors(t *testing.T) {
	providers := []provider.CostProvider{
		&mockCostProviderWithError{cost: 6495},
		&mockCostProviderWithError{err: errors.New("provider error")},
		&mockCostProviderWithError{cost: 20000},
	}
	agg := provider.NewAggregator(providers)

	h := &handler.Handler{
		CostAggregator: agg,
	}

	ctx := context.Background()
	total := h.CalculateTotalInfrastructureCost(ctx)

	// Should still return sum of successful providers even with error
	assert.Equal(t, int64(26495), total)
}

// TestHandler_New_InitializesAggregator verifies New() creates aggregator with real providers
func TestHandler_New_InitializesAggregator(t *testing.T) {
	cfg := &config.Config{
		HostingerVPSCostCents: 6495,
		ClaudeAPICostCents:    20000,
		BrevoEmailCostCents:   0,
	}

	h := handler.New(nil, cfg, nil, nil)

	require.NotNil(t, h.CostAggregator)

	// Verify aggregator calculates total from all three providers
	ctx := context.Background()
	total := h.CalculateTotalInfrastructureCost(ctx)

	assert.Equal(t, int64(26495), total) // 6495 + 20000 + 0
}

// TestHandler_New_InitializesAggregator_WithBrevoPayment verifies paid Brevo plan
func TestHandler_New_InitializesAggregator_WithBrevoPayment(t *testing.T) {
	cfg := &config.Config{
		HostingerVPSCostCents: 6495,
		ClaudeAPICostCents:    20000,
		BrevoEmailCostCents:   14500, // €25/month
	}

	h := handler.New(nil, cfg, nil, nil)

	require.NotNil(t, h.CostAggregator)

	ctx := context.Background()
	total := h.CalculateTotalInfrastructureCost(ctx)

	assert.Equal(t, int64(40995), total) // 6495 + 20000 + 14500
}
