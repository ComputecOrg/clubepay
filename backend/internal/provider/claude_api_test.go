package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/provider"
)

func TestClaudeAPIProvider_GetMonthlyCost(t *testing.T) {
	cfg := &config.Config{
		ClaudeAPICostCents: 20000, // $200
	}
	c := provider.NewClaudeAPIProvider(cfg)

	ctx := context.Background()
	cost, err := c.GetMonthlyCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(20000), cost.CostCents)
	assert.Equal(t, "claude_api", cost.Provider)
	assert.Contains(t, cost.Description, "$200")
}

func TestClaudeAPIProvider_GetMonthlyCost_ZeroCost(t *testing.T) {
	cfg := &config.Config{
		ClaudeAPICostCents: 0,
	}
	c := provider.NewClaudeAPIProvider(cfg)

	ctx := context.Background()
	cost, err := c.GetMonthlyCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), cost.CostCents)
	assert.Equal(t, "claude_api", cost.Provider)
}
