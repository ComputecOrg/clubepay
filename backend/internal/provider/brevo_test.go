package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/provider"
)

func TestBrevoProvider_GetMonthlyCost_FreeTier(t *testing.T) {
	cfg := &config.Config{
		BrevoEmailCostCents: 0, // Free tier
	}
	b := provider.NewBrevoProvider(cfg)

	ctx := context.Background()
	cost, err := b.GetMonthlyCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), cost.CostCents)
	assert.Equal(t, "brevo", cost.Provider)
	assert.Contains(t, cost.Description, "free")
}

func TestBrevoProvider_GetMonthlyCost_Paid(t *testing.T) {
	cfg := &config.Config{
		BrevoEmailCostCents: 14500, // €25/month
	}
	b := provider.NewBrevoProvider(cfg)

	ctx := context.Background()
	cost, err := b.GetMonthlyCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(14500), cost.CostCents)
	assert.Equal(t, "brevo", cost.Provider)
}
