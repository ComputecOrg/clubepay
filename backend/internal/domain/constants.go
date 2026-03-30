package domain

import "time"

// Business tier limits (free tier).
const (
	FreeTierPlanLimit       = 1
	FreeTierSubscriberLimit = 15
)

// Referral limits and discounts.
const (
	ReferralLimit           = 3
	ReferralDiscountPercent = 10
)

// Grace period for overdue payments.
const GracePeriodDays = 3

// JWT expiry durations.
const (
	OwnerJWTExpiry      = 24 * time.Hour
	SubscriberJWTExpiry = 30 * 24 * time.Hour
)
