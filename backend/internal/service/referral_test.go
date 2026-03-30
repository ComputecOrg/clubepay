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

func TestReferralService_GetOrCreateCode(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewReferralService(q)
	ctx := context.Background()

	t.Run("GetOrCreateCode_creates_new_code", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "ref_create_owner@test.com", "Ref Create Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Ref Create", "cafe-ref-create")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Ref Create", 2990, "daily", 1)
		subscriber := testutil.SeedSubscriber(t, q, "ref_create_sub@test.com", "Ref Create Sub", "11999992000")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		resp, err := svc.GetOrCreateCode(ctx, subscriber.ID)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Code, 8)
	})

	t.Run("GetOrCreateCode_returns_existing_code", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "ref_existing_owner@test.com", "Ref Existing Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Ref Existing", "cafe-ref-existing")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Ref Existing", 2990, "daily", 1)
		subscriber := testutil.SeedSubscriber(t, q, "ref_existing_sub@test.com", "Ref Existing Sub", "11999992001")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		resp1, err := svc.GetOrCreateCode(ctx, subscriber.ID)
		require.NoError(t, err)

		resp2, err := svc.GetOrCreateCode(ctx, subscriber.ID)
		require.NoError(t, err)

		// Should return the same code both times
		assert.Equal(t, resp1.Code, resp2.Code)
	})

	t.Run("GetOrCreateCode_no_subscription", func(t *testing.T) {
		subscriber := testutil.SeedSubscriber(t, q, "ref_nosub_sub@test.com", "Ref NoSub Sub", "11999992002")

		resp, err := svc.GetOrCreateCode(ctx, subscriber.ID)
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})
}

func TestReferralService_Apply(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewReferralService(q)
	ctx := context.Background()

	t.Run("Apply_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "ref_apply_owner@test.com", "Ref Apply Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Ref Apply", "cafe-ref-apply")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Ref Apply", 2990, "daily", 1)

		referrer := testutil.SeedSubscriber(t, q, "ref_apply_referrer@test.com", "Ref Apply Referrer", "11999992010")
		testutil.SeedSubscription(t, q, plan.ID, referrer.ID, biz.ID)

		// Get the referrer's code
		codeResp, err := svc.GetOrCreateCode(ctx, referrer.ID)
		require.NoError(t, err)

		// Another subscriber applies the code
		referred := testutil.SeedSubscriber(t, q, "ref_apply_referred@test.com", "Ref Apply Referred", "11999992011")
		testutil.SeedSubscription(t, q, plan.ID, referred.ID, biz.ID)

		err = svc.Apply(ctx, referred.ID, domain.ApplyReferralInput{Code: codeResp.Code})
		require.NoError(t, err)
	})

	t.Run("Apply_code_not_found", func(t *testing.T) {
		subscriber := testutil.SeedSubscriber(t, q, "ref_nf_sub@test.com", "Ref NF Sub", "11999992020")

		err := svc.Apply(ctx, subscriber.ID, domain.ApplyReferralInput{Code: "00000000"})
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})

	t.Run("Apply_self_referral", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "ref_self_owner@test.com", "Ref Self Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Ref Self", "cafe-ref-self")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Ref Self", 2990, "daily", 1)
		subscriber := testutil.SeedSubscriber(t, q, "ref_self_sub@test.com", "Ref Self Sub", "11999992030")
		testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

		codeResp, err := svc.GetOrCreateCode(ctx, subscriber.ID)
		require.NoError(t, err)

		// Try to apply own code
		err = svc.Apply(ctx, subscriber.ID, domain.ApplyReferralInput{Code: codeResp.Code})
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 400, svcErr.Code)
	})

	t.Run("Apply_referral_limit_reached", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "ref_lim_owner@test.com", "Ref Lim Owner")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Ref Lim", "cafe-ref-lim")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Ref Lim", 2990, "daily", 1)

		referrer := testutil.SeedSubscriber(t, q, "ref_lim_referrer@test.com", "Ref Lim Referrer", "11999992040")
		testutil.SeedSubscription(t, q, plan.ID, referrer.ID, biz.ID)

		codeResp, err := svc.GetOrCreateCode(ctx, referrer.ID)
		require.NoError(t, err)

		// Apply code 3 times (ReferralLimit = 3)
		for i := 0; i < domain.ReferralLimit; i++ {
			sub := testutil.SeedSubscriber(t, q, "ref_lim_referred_"+string(rune('a'+i))+"@test.com", "Referred", "1199999"+string(rune('5'+i))+"040")
			testutil.SeedSubscription(t, q, plan.ID, sub.ID, biz.ID)
			err = svc.Apply(ctx, sub.ID, domain.ApplyReferralInput{Code: codeResp.Code})
			require.NoError(t, err, "apply #%d should succeed", i+1)
		}

		// 4th application should fail
		extra := testutil.SeedSubscriber(t, q, "ref_lim_extra@test.com", "Extra Referred", "11999992099")
		testutil.SeedSubscription(t, q, plan.ID, extra.ID, biz.ID)
		err = svc.Apply(ctx, extra.ID, domain.ApplyReferralInput{Code: codeResp.Code})
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 403, svcErr.Code)
	})
}
