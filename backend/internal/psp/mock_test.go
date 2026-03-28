package psp_test

import (
	"context"
	"testing"

	"github.com/clubepay/backend/internal/psp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockPSP_DefaultBehavior(t *testing.T) {
	m := &psp.MockPSP{}
	ctx := context.Background()

	customer, err := m.CreateCustomer(ctx, psp.CreateCustomerRequest{Name: "Test"})
	require.NoError(t, err)
	assert.Equal(t, "cus_mock_123", customer.ID)

	sub, err := m.CreateSubscription(ctx, psp.CreateSubscriptionRequest{CustomerID: "cus_1"})
	require.NoError(t, err)
	assert.Equal(t, "sub_mock_123", sub.ID)
	assert.Equal(t, "ACTIVE", sub.Status)

	err = m.CancelSubscription(ctx, "sub_1")
	assert.NoError(t, err)

	got, err := m.GetSubscription(ctx, "sub_1")
	require.NoError(t, err)
	assert.Equal(t, "sub_1", got.ID)

	payments, err := m.GetPayments(ctx, "sub_1")
	require.NoError(t, err)
	assert.Len(t, payments, 1)
	assert.Equal(t, "CONFIRMED", payments[0].Status)

	assert.True(t, m.ValidateWebhook([]byte("data"), "sig"))
}

func TestMockPSP_CustomFunctions(t *testing.T) {
	m := &psp.MockPSP{
		CreateCustomerFn: func(ctx context.Context, req psp.CreateCustomerRequest) (*psp.Customer, error) {
			return &psp.Customer{ID: "custom_cus"}, nil
		},
		ValidateWebhookFn: func(payload []byte, signature string) bool {
			return false
		},
	}
	ctx := context.Background()

	customer, err := m.CreateCustomer(ctx, psp.CreateCustomerRequest{})
	require.NoError(t, err)
	assert.Equal(t, "custom_cus", customer.ID)

	assert.False(t, m.ValidateWebhook(nil, ""))
}
