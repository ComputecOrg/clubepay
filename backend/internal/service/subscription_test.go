package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/psp"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
	"github.com/clubepay/backend/internal/testutil"
)

// newService creates a SubscriptionService backed by a real test DB.
func newService(t *testing.T) (*service.SubscriptionService, *repository.Queries) {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := &service.SubscriptionService{
		Queries: q,
		PSP:     &psp.MockPSP{},
		Email:   &email.MockSender{},
	}
	return svc, q
}

// ---- Subscribe ----

func TestSubscriptionService_Subscribe_Success(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_sub_owner@test.com", "Owner Sub")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café Svc", "cafe-svc-sub")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano Svc", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_subscriber@test.com", "Svc Sub", "11999990000")

	sub, err := svc.Subscribe(context.Background(), subscriber.ID, service.SubscribeInput{PlanID: plan.ID})

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, plan.ID, sub.PlanID)
	assert.Equal(t, subscriber.ID, sub.SubscriberID)
	assert.Equal(t, biz.ID, sub.BusinessID)
	assert.Equal(t, domain.SubscriptionStatusActive, sub.Status)
	assert.Equal(t, int32(0), sub.DiscountPercent)
}

func TestSubscriptionService_Subscribe_AlreadySubscribed(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_dup_owner@test.com", "Owner Dup")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café Dup Svc", "cafe-svc-dup")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano Dup", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_dup_sub@test.com", "Dup Sub", "11999990001")

	// First subscription — should succeed
	_, err := svc.Subscribe(context.Background(), subscriber.ID, service.SubscribeInput{PlanID: plan.ID})
	require.NoError(t, err)

	// Second subscription — should fail with conflict
	_, err = svc.Subscribe(context.Background(), subscriber.ID, service.SubscribeInput{PlanID: plan.ID})
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 409, svcErr.Code)
}

func TestSubscriptionService_Subscribe_PlanNotFound(t *testing.T) {
	svc, _ := newService(t)

	subscriber := testutil.SeedSubscriber(t, svc.Queries, "svc_notfound_sub@test.com", "NotFound Sub", "11999990002")

	_, err := svc.Subscribe(context.Background(), subscriber.ID, service.SubscribeInput{PlanID: 999999})
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 404, svcErr.Code)
}

func TestSubscriptionService_Subscribe_PlanInactive(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_inactive_owner@test.com", "Inactive Owner")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café Inactive", "cafe-svc-inactive")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano Inativo", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_inactive_sub@test.com", "Inactive Sub", "11999990003")

	err := q.DeactivatePlan(context.Background(), plan.ID)
	require.NoError(t, err)

	_, err = svc.Subscribe(context.Background(), subscriber.ID, service.SubscribeInput{PlanID: plan.ID})
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 400, svcErr.Code)
}

func TestSubscriptionService_Subscribe_FreeTierLimit(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_freetier_owner@test.com", "FreeTier Owner")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café FreeTier", "cafe-svc-freetier")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano Free", 2990, "daily", 1)

	// Seed 15 subscriptions to hit the limit
	for i := 0; i < 15; i++ {
		sub := testutil.SeedSubscriber(t, q, "svc_free_"+string(rune('a'+i))+"@test.com", "Free Sub", "1199999"+string(rune('0'+i))+"000")
		testutil.SeedSubscription(t, q, plan.ID, sub.ID, biz.ID)
	}

	extra := testutil.SeedSubscriber(t, q, "svc_free_extra@test.com", "Extra Sub", "11999990099")
	_, err := svc.Subscribe(context.Background(), extra.ID, service.SubscribeInput{PlanID: plan.ID})
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 403, svcErr.Code)
}

// ---- ListByBusiness ----

func TestSubscriptionService_ListByBusiness_Success(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_list_owner@test.com", "List Owner")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café List", "cafe-svc-list")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano List", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_list_sub@test.com", "List Sub", "11999990010")
	testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

	resp, err := svc.ListByBusiness(context.Background(), owner.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Subscriptions, 1)
	assert.Equal(t, "List Sub", resp.Subscriptions[0].SubscriberName)
	assert.Equal(t, "Plano List", resp.Subscriptions[0].PlanName)
}

func TestSubscriptionService_ListByBusiness_NoBusiness(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_list_nobiz@test.com", "NoBiz Owner")

	_, err := svc.ListByBusiness(context.Background(), owner.ID)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 404, svcErr.Code)
}

// ---- CancelByOwner ----

func TestSubscriptionService_CancelByOwner_Success(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_cancel_owner@test.com", "Cancel Owner")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café Cancel", "cafe-svc-cancel")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano Cancel", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_cancel_sub@test.com", "Cancel Sub", "11999990020")
	sub := testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

	err := svc.CancelByOwner(context.Background(), owner.ID, sub.ID)
	require.NoError(t, err)

	// Verify subscription is cancelled in DB
	updated, err := q.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", updated.Status)
}

func TestSubscriptionService_CancelByOwner_NotFound(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_cancel_notfound@test.com", "Cancel NF Owner")
	testutil.SeedBusiness(t, q, owner.ID, "Café Cancel NF", "cafe-svc-cancel-nf")

	err := svc.CancelByOwner(context.Background(), owner.ID, 999999)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 404, svcErr.Code)
}

func TestSubscriptionService_CancelByOwner_WrongBusiness(t *testing.T) {
	svc, q := newService(t)

	owner1 := testutil.SeedOwner(t, q, "svc_cancelown1@test.com", "Owner1")
	biz1 := testutil.SeedBusiness(t, q, owner1.ID, "Café Owner1", "cafe-svc-own1")
	plan1 := testutil.SeedPlan(t, q, biz1.ID, "Plano Owner1", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_cancelown_sub@test.com", "Sub", "11999990021")
	sub := testutil.SeedSubscription(t, q, plan1.ID, subscriber.ID, biz1.ID)

	owner2 := testutil.SeedOwner(t, q, "svc_cancelown2@test.com", "Owner2")
	testutil.SeedBusiness(t, q, owner2.ID, "Café Owner2", "cafe-svc-own2")

	err := svc.CancelByOwner(context.Background(), owner2.ID, sub.ID)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 403, svcErr.Code)
}

// ---- CancelBySubscriber ----

func TestSubscriptionService_CancelBySubscriber_Success(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_subcanceler_owner@test.com", "SubCancel Owner")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café SubCancel", "cafe-svc-subcancel")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano SubCancel", 2990, "daily", 1)
	subscriber := testutil.SeedSubscriber(t, q, "svc_subcanceler@test.com", "SubCancel Sub", "11999990030")
	sub := testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

	resp, err := svc.CancelBySubscriber(context.Background(), subscriber.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "cancelled", resp.Status)
	assert.NotEmpty(t, resp.Message)

	// Verify subscription cancelled in DB
	updated, err := q.GetSubscriptionByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", updated.Status)
}

func TestSubscriptionService_CancelBySubscriber_NoActiveSubscription(t *testing.T) {
	svc, q := newService(t)

	subscriber := testutil.SeedSubscriber(t, q, "svc_subcancel_none@test.com", "NoSub Sub", "11999990031")

	_, err := svc.CancelBySubscriber(context.Background(), subscriber.ID)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 404, svcErr.Code)
}

// ---- GetMyPlan ----

func TestSubscriptionService_GetMyPlan_Success(t *testing.T) {
	svc, q := newService(t)

	owner := testutil.SeedOwner(t, q, "svc_myplan_owner@test.com", "MyPlan Owner")
	biz := testutil.SeedBusiness(t, q, owner.ID, "Café MyPlan", "cafe-svc-myplan")
	plan := testutil.SeedPlan(t, q, biz.ID, "Plano MyPlan", 4990, "monthly", 4)
	subscriber := testutil.SeedSubscriber(t, q, "svc_myplan_sub@test.com", "MyPlan Sub", "11999990040")
	testutil.SeedSubscription(t, q, plan.ID, subscriber.ID, biz.ID)

	resp, err := svc.GetMyPlan(context.Background(), subscriber.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, plan.ID, resp.Plan.ID)
	assert.Equal(t, "Plano MyPlan", resp.Plan.Name)
	assert.Equal(t, int64(4990), resp.Plan.PriceCents)
	assert.Equal(t, "monthly", resp.Plan.LimitType)
	assert.Equal(t, int32(4), resp.Plan.LimitCount)

	assert.Equal(t, biz.ID, resp.Business.ID)
	assert.Equal(t, "Café MyPlan", resp.Business.Name)
	assert.Equal(t, "cafe-svc-myplan", resp.Business.Slug)

	assert.Equal(t, domain.SubscriptionStatusActive, resp.Subscription.Status)
	assert.NotNil(t, resp.Subscription.PeriodEnd)
}

func TestSubscriptionService_GetMyPlan_NoSubscription(t *testing.T) {
	svc, q := newService(t)

	subscriber := testutil.SeedSubscriber(t, q, "svc_myplan_none@test.com", "None Sub", "11999990041")

	_, err := svc.GetMyPlan(context.Background(), subscriber.ID)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 404, svcErr.Code)
}
