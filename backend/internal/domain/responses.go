package domain

import "time"

// AuthResponse is returned by register and login endpoints.
// Token is set via HttpOnly cookie (Set-Cookie) and omitted from JSON to prevent XSS exposure.
type AuthResponse struct {
	Token    string            `json:"-"`
	User     UserResponse      `json:"user"`
	Business *BusinessResponse `json:"business,omitempty"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
	Role  string `json:"role"`
}

// BusinessResponse is the public representation of a business.
type BusinessResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Segment string `json:"segment"`
	Address string `json:"address,omitempty"`
	LogoURL string `json:"logo_url,omitempty"`
}

// PlanDetailResponse includes full plan info.
type PlanDetailResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	LimitType   string `json:"limit_type"`
	LimitCount  int32  `json:"limit_count"`
	Active      bool   `json:"active"`
}

// PlanListResponse wraps a list of plans.
type PlanListResponse struct {
	Plans []PlanDetailResponse `json:"plans"`
}

// SubscriptionListResponse wraps a list of subscriptions.
type SubscriptionListResponse struct {
	Subscriptions interface{} `json:"subscriptions"`
}

// SubscriptionInfo is a lightweight subscription summary.
type SubscriptionInfo struct {
	ID        int64      `json:"id"`
	Status    string     `json:"status"`
	PeriodEnd *time.Time `json:"period_end,omitempty"`
}

// MyPlanResponse combines plan, business, and subscription info for the subscriber.
type MyPlanResponse struct {
	Plan         PlanDetailResponse `json:"plan"`
	Business     BusinessResponse   `json:"business"`
	Subscription SubscriptionInfo   `json:"subscription"`
}

// ValidateUsageResponse is returned after validating a usage.
type ValidateUsageResponse struct {
	Status   string `json:"status"`
	Used     int64  `json:"used"`
	Limit    int32  `json:"limit"`
	PlanName string `json:"plan_name"`
}

// UsageListResponse is returned by my-usage.
type UsageListResponse struct {
	Used     int         `json:"used"`
	Limit    int32       `json:"limit"`
	PlanName string      `json:"plan_name"`
	Usages   interface{} `json:"usages"`
}

// CancelResponse is returned after cancelling a subscription.
type CancelResponse struct {
	Status    string     `json:"status"`
	PeriodEnd *time.Time `json:"period_end,omitempty"`
	Message   string     `json:"message"`
}

// ReconcileResponse is returned by the cron reconcile endpoint.
type ReconcileResponse struct {
	Blocked int `json:"blocked"`
	Synced  int `json:"synced"`
}

// ProfileResponse wraps a user profile.
type ProfileResponse struct {
	User UserResponse `json:"user"`
}

// SearchResponse wraps search results.
type SearchResponse struct {
	Results interface{} `json:"results"`
}

// ReferralCodeResponse returns the referral code.
type ReferralCodeResponse struct {
	Code string `json:"code"`
}

// PublicBusinessResponse includes subscriber count.
type PublicBusinessResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Segment         string `json:"segment"`
	SubscriberCount int64  `json:"subscriber_count"`
}
