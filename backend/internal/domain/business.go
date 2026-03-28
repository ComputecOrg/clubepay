package domain

import "time"

type Business struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Segment   string    `json:"segment"`
	Address   string    `json:"address,omitempty"`
	LogoURL   string    `json:"logo_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Plan struct {
	ID          int64     `json:"id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	LimitType   string    `json:"limit_type"`
	LimitCount  int       `json:"limit_count"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Subscription struct {
	ID              int64     `json:"id"`
	PlanID          int64     `json:"plan_id"`
	SubscriberID    int64     `json:"subscriber_id"`
	BusinessID      int64     `json:"business_id"`
	PSPSubscriptionID string  `json:"psp_subscription_id"`
	Status          string    `json:"status"`
	PeriodEnd       time.Time `json:"period_end"`
	GraceDeadline   *time.Time `json:"grace_deadline,omitempty"`
	ReferredBy      *int64    `json:"referred_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Usage struct {
	ID             int64     `json:"id"`
	SubscriptionID int64     `json:"subscription_id"`
	ValidatedAt    time.Time `json:"validated_at"`
}

type Referral struct {
	ID           int64     `json:"id"`
	ReferrerID   int64     `json:"referrer_id"`
	ReferredID   int64     `json:"referred_id"`
	BusinessID   int64     `json:"business_id"`
	Code         string    `json:"code"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Subscription statuses
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPending   = "pending"
	SubscriptionStatusGrace     = "grace"
	SubscriptionStatusBlocked   = "blocked"
	SubscriptionStatusCancelled = "cancelled"
)

// Plan limit types
const (
	LimitTypeDaily   = "daily"
	LimitTypeMonthly = "monthly"
)
