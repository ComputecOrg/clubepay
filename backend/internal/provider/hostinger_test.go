package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/provider"
)

func TestHostingerProvider_GetMonthlyCost(t *testing.T) {
	cfg := &config.Config{
		HostingerVPSCostCents: 6495, // R$12.99
	}
	h := provider.NewHostingerProvider(cfg)

	ctx := context.Background()
	cost, err := h.GetMonthlyCost(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(6495), cost.CostCents)
	assert.Equal(t, "hostinger", cost.Provider)
	assert.Contains(t, cost.Description, "KVM1")
}
