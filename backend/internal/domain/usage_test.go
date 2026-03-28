package domain_test

import (
	"testing"
	"time"

	"github.com/clubepay/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCheckUsageLimit_DailyAllowed(t *testing.T) {
	err := domain.CheckUsageLimit("daily", 1, 0)
	assert.NoError(t, err)
}

func TestCheckUsageLimit_DailyExceeded(t *testing.T) {
	err := domain.CheckUsageLimit("daily", 1, 1)
	assert.ErrorIs(t, err, domain.ErrUsageLimitReached)
}

func TestCheckUsageLimit_MonthlyAllowed(t *testing.T) {
	err := domain.CheckUsageLimit("monthly", 4, 3)
	assert.NoError(t, err)
}

func TestCheckUsageLimit_MonthlyExceeded(t *testing.T) {
	err := domain.CheckUsageLimit("monthly", 4, 4)
	assert.ErrorIs(t, err, domain.ErrUsageLimitReached)
}

func TestDailyRange(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)
	start, end := domain.DailyRange(now)

	assert.Equal(t, time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), start)
	assert.Equal(t, time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC), end)
}

func TestMonthlyRange(t *testing.T) {
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)
	start, end := domain.MonthlyRange(now)

	assert.Equal(t, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), start)
	assert.Equal(t, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), end)
}

func TestMonthlyRange_December(t *testing.T) {
	now := time.Date(2025, 12, 25, 10, 0, 0, 0, time.UTC)
	start, end := domain.MonthlyRange(now)

	assert.Equal(t, time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC), start)
	assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), end)
}
