package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")

	ErrBusinessNotFound  = errors.New("business not found")
	ErrSlugAlreadyExists = errors.New("slug already exists")

	ErrPlanNotFound      = errors.New("plan not found")
	ErrPlanLimitReached  = errors.New("plan limit reached for free tier")
	ErrPlanNotActive     = errors.New("plan is not active")

	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrAlreadySubscribed      = errors.New("already subscribed to this business")
	ErrSubscriptionNotActive  = errors.New("subscription is not active")
	ErrSubscriberLimitReached = errors.New("subscriber limit reached for free tier")

	ErrUsageLimitReached = errors.New("usage limit reached for this period")

	ErrReferralNotFound     = errors.New("referral code not found")
	ErrReferralLimitReached = errors.New("referral limit reached (max 3)")
	ErrSelfReferral         = errors.New("cannot refer yourself")

	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
)
