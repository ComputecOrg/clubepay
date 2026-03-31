package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/clubepay/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// SeedOwner creates an owner user for test fixtures.
func SeedOwner(t *testing.T, q *repository.Queries, email, name string) repository.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	user, err := q.CreateUser(ctx, repository.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Phone:        pgtype.Text{Valid: false},
		Role:         "owner",
	})
	if err != nil {
		t.Fatalf("failed to seed owner: %v", err)
	}
	return user
}

// SeedSubscriber creates a subscriber user for test fixtures.
func SeedSubscriber(t *testing.T, q *repository.Queries, email, name, phone string) repository.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	user, err := q.CreateUser(ctx, repository.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Phone:        pgtype.Text{String: phone, Valid: phone != ""},
		Role:         "subscriber",
	})
	if err != nil {
		t.Fatalf("failed to seed subscriber: %v", err)
	}
	return user
}

// SeedBusiness creates a business for test fixtures.
func SeedBusiness(t *testing.T, q *repository.Queries, ownerID int64, name, slug string) repository.Business {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	biz, err := q.CreateBusiness(ctx, repository.CreateBusinessParams{
		OwnerID: ownerID,
		Name:    name,
		Slug:    slug,
		Segment: "cafeteria",
		Address: pgtype.Text{Valid: false},
		LogoUrl: pgtype.Text{Valid: false},
	})
	if err != nil {
		t.Fatalf("failed to seed business: %v", err)
	}
	return biz
}

// SeedPlan creates a plan for test fixtures.
func SeedPlan(t *testing.T, q *repository.Queries, businessID int64, name string, priceCents int64, limitType string, limitCount int32) repository.Plan {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	plan, err := q.CreatePlan(ctx, repository.CreatePlanParams{
		BusinessID:  businessID,
		Name:        name,
		Description: pgtype.Text{Valid: false},
		PriceCents:  priceCents,
		LimitType:   limitType,
		LimitCount:  limitCount,
	})
	if err != nil {
		t.Fatalf("failed to seed plan: %v", err)
	}
	return plan
}

// SeedSubscription creates an active subscription for test fixtures.
func SeedSubscription(t *testing.T, q *repository.Queries, planID, subscriberID, businessID int64) repository.Subscription {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sub, err := q.CreateSubscription(ctx, repository.CreateSubscriptionParams{
		PlanID:            planID,
		SubscriberID:      subscriberID,
		BusinessID:        businessID,
		PspSubscriptionID: pgtype.Text{String: "mock_sub_test", Valid: true},
		Status:            "active",
		PeriodEnd:         pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
		ReferredBy:        pgtype.Int8{Valid: false},
	})
	if err != nil {
		t.Fatalf("failed to seed subscription: %v", err)
	}
	return sub
}
