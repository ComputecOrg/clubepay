package testutil

import (
	"context"
	"testing"

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
	user, err := q.CreateUser(context.Background(), repository.CreateUserParams{
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
	user, err := q.CreateUser(context.Background(), repository.CreateUserParams{
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
	biz, err := q.CreateBusiness(context.Background(), repository.CreateBusinessParams{
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
	plan, err := q.CreatePlan(context.Background(), repository.CreatePlanParams{
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
