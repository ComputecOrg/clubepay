package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
	"github.com/clubepay/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageService_Validate(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewUsageService(q)
	ctx := context.Background()

	t.Run("Validate_success_daily", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_val_owner@test.com", "Usage Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Usage", "cafe-usage-val")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Diário", 2990, "daily", 2)
		subscriber := testutil.SeedSubscriber(t, q, "usage_val_sub@test.com", "Usage Sub", "11999991000")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		resp, err := svc.Validate(ctx, subscriber.ID, domain.ValidateUsageInput{BusinessSlug: "cafe-usage-val"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "validated", resp.Status)
		assert.Equal(t, int64(1), resp.Used)
		assert.Equal(t, int32(2), resp.Limit)
		assert.Equal(t, "Plano Diário", resp.PlanName)
	})

	t.Run("Validate_limit_reached", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_limit_owner@test.com", "Usage Limit Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Limit", "cafe-usage-limit")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Limite", 2990, "daily", 1)
		subscriber := testutil.SeedSubscriber(t, q, "usage_limit_sub@test.com", "Usage Limit Sub", "11999991001")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		// First usage — should succeed
		_, err := svc.Validate(ctx, subscriber.ID, domain.ValidateUsageInput{BusinessSlug: "cafe-usage-limit"})
		require.NoError(t, err)

		// Second usage — should fail (limit = 1)
		resp, err := svc.Validate(ctx, subscriber.ID, domain.ValidateUsageInput{BusinessSlug: "cafe-usage-limit"})
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 403, svcErr.Code)
	})

	t.Run("Validate_business_not_found", func(t *testing.T) {
		subscriber := testutil.SeedSubscriber(t, q, "usage_nob_sub@test.com", "NoBiz Sub", "11999991002")

		resp, err := svc.Validate(ctx, subscriber.ID, domain.ValidateUsageInput{BusinessSlug: "nonexistent-slug"})
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})

	t.Run("Validate_no_subscription", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_nosub_owner@test.com", "NoSub Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe NoSub", "cafe-usage-nosub")
		_ = testutil.SeedPlan(t, q, biz.ID, "Plano NoSub", 2990, "daily", 1)
		subscriber := testutil.SeedSubscriber(t, q, "usage_nosub_sub@test.com", "NoSub Sub", "11999991003")

		resp, err := svc.Validate(ctx, subscriber.ID, domain.ValidateUsageInput{BusinessSlug: "cafe-usage-nosub"})
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})
}

func TestUsageService_ValidateByOwner(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewUsageService(q)
	ctx := context.Background()

	t.Run("ValidateByOwner_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_vbo_owner@test.com", "VBO Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe VBO", "cafe-usage-vbo")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano VBO", 2990, "daily", 3)
		subscriber := testutil.SeedSubscriber(t, q, "usage_vbo_sub@test.com", "VBO Sub", "11999991010")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		resp, err := svc.ValidateByOwner(ctx, owner.ID, domain.ValidateUsageOwnerInput{SubscriberID: subscriber.ID})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "validated", resp.Status)
		assert.Equal(t, int64(1), resp.Used)
	})

	t.Run("ValidateByOwner_owner_no_business", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_vbo_nobiz@test.com", "VBO NoBiz Owner")
		subscriber := testutil.SeedSubscriber(t, q, "usage_vbo_nobiz_sub@test.com", "VBO NoBiz Sub", "11999991011")

		resp, err := svc.ValidateByOwner(ctx, owner.ID, domain.ValidateUsageOwnerInput{SubscriberID: subscriber.ID})
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})
}

func TestUsageService_GetMyUsage(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewUsageService(q)
	ctx := context.Background()

	t.Run("GetMyUsage_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_my_owner@test.com", "MyUsage Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe MyUsage", "cafe-usage-my")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano MyUsage", 2990, "monthly", 4)
		subscriber := testutil.SeedSubscriber(t, q, "usage_my_sub@test.com", "MyUsage Sub", "11999991020")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		resp, err := svc.GetMyUsage(ctx, subscriber.ID, "cafe-usage-my")
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 0, resp.Used)
		assert.Equal(t, int32(4), resp.Limit)
		assert.Equal(t, "Plano MyUsage", resp.PlanName)
	})

	t.Run("GetMyUsage_after_validate", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_myval_owner@test.com", "MyUsageVal Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe MyUsageVal", "cafe-usage-myval")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano MyUsageVal", 2990, "daily", 3)
		subscriber := testutil.SeedSubscriber(t, q, "usage_myval_sub@test.com", "MyUsageVal Sub", "11999991021")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		valSvc := service.NewUsageService(q)
		_, err := valSvc.Validate(ctx, subscriber.ID, domain.ValidateUsageInput{BusinessSlug: "cafe-usage-myval"})
		require.NoError(t, err)

		resp, err := svc.GetMyUsage(ctx, subscriber.ID, "cafe-usage-myval")
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Used)
	})

	t.Run("GetMyUsage_business_not_found", func(t *testing.T) {
		subscriber := testutil.SeedSubscriber(t, q, "usage_my_nob@test.com", "MyUsage NoBiz", "11999991022")

		resp, err := svc.GetMyUsage(ctx, subscriber.ID, "slug-does-not-exist")
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})

	t.Run("GetMyUsage_no_subscription", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "usage_my_nosub_own@test.com", "MyUsage NoSub Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe NoSub My", "cafe-usage-my-nosub")
		_ = testutil.SeedPlan(t, q, biz.ID, "Plano", 2990, "daily", 1)
		subscriber := testutil.SeedSubscriber(t, q, "usage_my_nosub_sub@test.com", "NoSub My Sub", "11999991023")

		resp, err := svc.GetMyUsage(ctx, subscriber.ID, "cafe-usage-my-nosub")
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})
}
