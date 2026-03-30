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

func TestBusinessService(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	svc := service.NewBusinessService(q)
	ctx := context.Background()

	t.Run("GetByOwner_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "biz_get@example.com", "Owner Get")
		_ = testutil.SeedBusiness(t, q, owner.ID, "Cafe Get", "cafe-get")

		resp, err := svc.GetByOwner(ctx, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, "Cafe Get", resp.Name)
		assert.Equal(t, "cafe-get", resp.Slug)
		assert.Equal(t, "cafeteria", resp.Segment)
	})

	t.Run("GetByOwner_not_found", func(t *testing.T) {
		resp, err := svc.GetByOwner(ctx, 999999)
		assert.Nil(t, resp)
		require.Error(t, err)
		var svcErr *domain.ServiceError
		require.True(t, errors.As(err, &svcErr), "expected *domain.ServiceError")
		assert.Equal(t, 404, svcErr.Code)
	})

	t.Run("Update_success", func(t *testing.T) {
		owner := testutil.SeedOwner(t, q, "biz_upd@example.com", "Owner Update")
		_ = testutil.SeedBusiness(t, q, owner.ID, "Cafe Update", "cafe-update")

		input := domain.UpdateBusinessInput{
			Name:    "Cafe Atualizado",
			Segment: "padaria",
			Address: "Rua Nova, 123",
			LogoURL: "https://example.com/logo.png",
		}

		resp, err := svc.Update(ctx, owner.ID, input)
		require.NoError(t, err)
		assert.Equal(t, "Cafe Atualizado", resp.Name)
		assert.Equal(t, "padaria", resp.Segment)
		assert.Equal(t, "Rua Nova, 123", resp.Address)
		assert.Equal(t, "https://example.com/logo.png", resp.LogoURL)
	})
}
