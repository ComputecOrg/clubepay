package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/provider"
)

type mockProvider struct {
	cost int64
	err  error
}

func (m *mockProvider) GetMonthlyCost(ctx context.Context) (provider.Cost, error) {
	if m.err != nil {
		return provider.Cost{}, m.err
	}
	return provider.Cost{CostCents: m.cost, Provider: "mock"}, nil
}

// TestAggregator_GetTotalInfrastructureCost_Success verifies basic aggregation
func TestAggregator_GetTotalInfrastructureCost_Success(t *testing.T) {
	providers := []provider.CostProvider{
		&mockProvider{cost: 6495},   // Hostinger: R$12.99
		&mockProvider{cost: 20000},  // Claude: $200
		&mockProvider{cost: 0},      // Brevo: free tier
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(26495), total) // 6495 + 20000 + 0
}

// TestAggregator_GetTotalInfrastructureCost_SingleProvider verifies single provider
func TestAggregator_GetTotalInfrastructureCost_SingleProvider(t *testing.T) {
	providers := []provider.CostProvider{
		&mockProvider{cost: 20000},
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(20000), total)
}

// TestAggregator_GetTotalInfrastructureCost_NonBlockingOnError verifies non-blocking behavior
func TestAggregator_GetTotalInfrastructureCost_NonBlockingOnError(t *testing.T) {
	// One provider fails, but aggregator continues with others (non-blocking)
	providers := []provider.CostProvider{
		&mockProvider{cost: 6495},
		&mockProvider{err: errors.New("provider temporarily unavailable")},
		&mockProvider{cost: 20000},
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	// Should succeed even with one provider failing
	require.NoError(t, err)
	// Should sum only the successful providers
	assert.Equal(t, int64(26495), total) // 6495 + 20000, skipping the error
}

// TestAggregator_GetTotalInfrastructureCost_AllProvidersError verifies graceful handling
func TestAggregator_GetTotalInfrastructureCost_AllProvidersError(t *testing.T) {
	providers := []provider.CostProvider{
		&mockProvider{err: errors.New("api error")},
		&mockProvider{err: errors.New("timeout")},
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	// Should return zero, no error (non-blocking behavior)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

// TestAggregator_GetTotalInfrastructureCost_EmptyProviders verifies empty case
func TestAggregator_GetTotalInfrastructureCost_EmptyProviders(t *testing.T) {
	providers := []provider.CostProvider{}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

// TestAggregator_GetTotalCost is the original test from the plan
func TestAggregator_GetTotalCost(t *testing.T) {
	providers := []provider.CostProvider{
		&mockProvider{cost: 10000}, // $100
		&mockProvider{cost: 5000},  // $50
		&mockProvider{cost: 2500},  // $25
	}
	agg := provider.NewAggregator(providers)

	ctx := context.Background()
	total, err := agg.GetTotalInfrastructureCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(17500), total) // $175 total
}
