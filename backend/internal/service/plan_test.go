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

func TestPlanService(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewPlanService(q)
	ctx := context.Background()

	t.Run("Create_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "plan_create@example.com", "Owner Plan Create")
		_ = testutil.SeedBusiness(t, q, owner.ID, "Cafe Plan Create", "cafe-plan-create")

		input := domain.CreatePlanInput{
			Name:        "Plano Café Diário",
			Description: "Um café por dia",
			PriceCents:  2990,
			LimitType:   "daily",
			LimitCount:  1,
		}

		resp, err := svc.Create(ctx, owner.ID, input)
		require.NoError(t, err)
		assert.Equal(t, "Plano Café Diário", resp.Name)
		assert.Equal(t, "Um café por dia", resp.Description)
		assert.Equal(t, int64(2990), resp.PriceCents)
		assert.Equal(t, "daily", resp.LimitType)
		assert.Equal(t, int32(1), resp.LimitCount)
		assert.True(t, resp.Active)
		assert.NotZero(t, resp.ID)
	})

	t.Run("Create_free_tier_limit", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "plan_limit@example.com", "Owner Plan Limit")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Plan Limit", "cafe-plan-limit")
		// Seed the maximum number of plans allowed (FreeTierPlanLimit = 1)
		testutil.SeedPlan(t, q, biz.ID, "Plano Existente", 1990, "daily", 1)

		input := domain.CreatePlanInput{
			Name:       "Plano Extra",
			PriceCents: 3990,
			LimitType:  "monthly",
			LimitCount: 4,
		}

		resp, err := svc.Create(ctx, owner.ID, input)
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr), "expected *domain.ServiceError")
		assert.Equal(t, 403, svcErr.Code)
	})

	t.Run("Create_no_business", func(t *testing.T) {
		resp, err := svc.Create(ctx, 999999, domain.CreatePlanInput{
			Name:       "Plano",
			PriceCents: 1000,
			LimitType:  "daily",
			LimitCount: 1,
		})
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr), "expected *domain.ServiceError")
		assert.Equal(t, 404, svcErr.Code)
	})

	t.Run("List_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "plan_list@example.com", "Owner Plan List")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Plan List", "cafe-plan-list")
		testutil.SeedPlan(t, q, biz.ID, "Plano Lista", 1990, "daily", 1)

		resp, err := svc.List(ctx, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Plans, 1)
		assert.Equal(t, "Plano Lista", resp.Plans[0].Name)
	})

	t.Run("List_no_business", func(t *testing.T) {
		resp, err := svc.List(ctx, 888888)
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 404, svcErr.Code)
	})

	t.Run("Update_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "plan_update@example.com", "Owner Plan Update")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Plan Update", "cafe-plan-update")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Antigo", 1990, "daily", 1)

		input := domain.UpdatePlanInput{
			Name:        "Plano Novo",
			Description: "Descrição atualizada",
			PriceCents:  2490,
			LimitType:   "monthly",
			LimitCount:  4,
		}

		resp, err := svc.Update(ctx, owner.ID, plan.ID, input)
		require.NoError(t, err)
		assert.Equal(t, "Plano Novo", resp.Name)
		assert.Equal(t, "Descrição atualizada", resp.Description)
		assert.Equal(t, int64(2490), resp.PriceCents)
		assert.Equal(t, "monthly", resp.LimitType)
		assert.Equal(t, int32(4), resp.LimitCount)
	})

	t.Run("Update_not_owner", func(t *testing.T) {
		owner1 := testutil.SeedOwner(t, q, "plan_upd_own1@example.com", "Owner 1")
		biz1 := testutil.SeedBusiness(t, q, owner1.ID, "Cafe Owner 1", "cafe-owner-1-plan")
		plan := testutil.SeedPlan(t, q, biz1.ID, "Plano Owner 1", 1990, "daily", 1)

		owner2 := testutil.SeedOwner(t, q, "plan_upd_own2@example.com", "Owner 2")
		_ = testutil.SeedBusiness(t, q, owner2.ID, "Cafe Owner 2", "cafe-owner-2-plan")

		input := domain.UpdatePlanInput{
			Name:       "Tentativa",
			PriceCents: 999,
			LimitType:  "daily",
			LimitCount: 1,
		}

		resp, err := svc.Update(ctx, owner2.ID, plan.ID, input)
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 403, svcErr.Code)
	})

	t.Run("Delete_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "plan_delete@example.com", "Owner Plan Delete")
		biz := testutil.SeedBusiness(t, q, owner.ID, "Cafe Plan Delete", "cafe-plan-delete")
		plan := testutil.SeedPlan(t, q, biz.ID, "Plano Para Deletar", 1990, "daily", 1)

		err := svc.Delete(ctx, owner.ID, plan.ID)
		require.NoError(t, err)

		// After deletion the plan list should be empty (active=false is excluded)
		resp, err := svc.List(ctx, owner.ID)
		require.NoError(t, err)
		assert.Len(t, resp.Plans, 0)
	})

	t.Run("Delete_not_owner", func(t *testing.T) {
		owner1 := testutil.SeedOwner(t, q, "plan_del_own1@example.com", "Del Owner 1")
		biz1 := testutil.SeedBusiness(t, q, owner1.ID, "Cafe Del Owner 1", "cafe-del-owner-1")
		plan := testutil.SeedPlan(t, q, biz1.ID, "Plano Del Owner 1", 1990, "daily", 1)

		owner2 := testutil.SeedOwner(t, q, "plan_del_own2@example.com", "Del Owner 2")
		_ = testutil.SeedBusiness(t, q, owner2.ID, "Cafe Del Owner 2", "cafe-del-owner-2")

		err := svc.Delete(ctx, owner2.ID, plan.ID)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr))
		assert.Equal(t, 403, svcErr.Code)
	})
}
