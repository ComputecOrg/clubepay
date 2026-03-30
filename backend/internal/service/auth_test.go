package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/repository"
	"github.com/clubepay/backend/internal/service"
	"github.com/clubepay/backend/internal/testutil"
)

// newAuthService creates an AuthService backed by a real test DB.
func newAuthService(t *testing.T) (*service.AuthService, *repository.Queries) {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	q := repository.New(pool)
	cfg := &config.Config{
		JWTSecret:   "test-secret",
		FrontendURL: "http://localhost:3000",
	}
	svc := service.NewAuthService(q, cfg, &email.MockSender{})
	return svc, q
}

// ---- RegisterOwner ----

func TestAuthService_RegisterOwner_Success(t *testing.T) {
	svc, _ := newAuthService(t)

	resp, err := svc.RegisterOwner(context.Background(), domain.RegisterOwnerInput{
		Email:        "owner@test.com",
		Password:     "password123",
		Name:         "João Dono",
		BusinessName: "Café do João",
		Segment:      "cafeteria",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "owner@test.com", resp.User.Email)
	assert.Equal(t, "João Dono", resp.User.Name)
	assert.Equal(t, domain.RoleOwner, resp.User.Role)
	require.NotNil(t, resp.Business)
	assert.Equal(t, "Café do João", resp.Business.Name)
	assert.Equal(t, "cafe-do-joao", resp.Business.Slug)
}

func TestAuthService_RegisterOwner_DefaultSegment(t *testing.T) {
	svc, _ := newAuthService(t)

	resp, err := svc.RegisterOwner(context.Background(), domain.RegisterOwnerInput{
		Email:        "owner_seg@test.com",
		Password:     "password123",
		Name:         "Maria Dona",
		BusinessName: "Padaria Maria",
		// Segment intentionally omitted
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Business)
	assert.Equal(t, "cafeteria", resp.Business.Segment)
}

func TestAuthService_RegisterOwner_DuplicateEmail(t *testing.T) {
	svc, _ := newAuthService(t)

	input := domain.RegisterOwnerInput{
		Email:        "dup_owner@test.com",
		Password:     "password123",
		Name:         "Dup Dono",
		BusinessName: "Café Dup",
	}

	_, err := svc.RegisterOwner(context.Background(), input)
	require.NoError(t, err)

	// Second registration with same email
	input.BusinessName = "Outro Café"
	_, err = svc.RegisterOwner(context.Background(), input)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 409, svcErr.Code)
}

// ---- RegisterSubscriber ----

func TestAuthService_RegisterSubscriber_Success(t *testing.T) {
	svc, _ := newAuthService(t)

	resp, err := svc.RegisterSubscriber(context.Background(), domain.RegisterSubscriberInput{
		Email:    "sub_auth@test.com",
		Password: "password123",
		Name:     "Ana Assinante",
		Phone:    "11999990000",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "sub_auth@test.com", resp.User.Email)
	assert.Equal(t, "Ana Assinante", resp.User.Name)
	assert.Equal(t, domain.RoleSubscriber, resp.User.Role)
	assert.Equal(t, "11999990000", resp.User.Phone)
	assert.Nil(t, resp.Business)
}

func TestAuthService_RegisterSubscriber_DuplicateEmail(t *testing.T) {
	svc, _ := newAuthService(t)

	input := domain.RegisterSubscriberInput{
		Email:    "dup_sub_auth@test.com",
		Password: "password123",
		Name:     "Dup Sub",
	}

	_, err := svc.RegisterSubscriber(context.Background(), input)
	require.NoError(t, err)

	_, err = svc.RegisterSubscriber(context.Background(), input)
	require.Error(t, err)

	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 409, svcErr.Code)
}

// ---- Login ----

func TestAuthService_Login_Success(t *testing.T) {
	svc, q := newAuthService(t)

	owner := testutil.SeedOwner(t, q, "login_owner@test.com", "Login Owner")
	_ = owner

	resp, err := svc.Login(context.Background(), domain.LoginInput{
		Email:    "login_owner@test.com",
		Password: "password123",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "login_owner@test.com", resp.User.Email)
	assert.Equal(t, domain.RoleOwner, resp.User.Role)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, q := newAuthService(t)

	testutil.SeedOwner(t, q, "login_wrong_pw@test.com", "WrongPW Owner")

	_, err := svc.Login(context.Background(), domain.LoginInput{
		Email:    "login_wrong_pw@test.com",
		Password: "wrongpassword",
	})

	require.Error(t, err)
	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 400, svcErr.Code)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, _ := newAuthService(t)

	_, err := svc.Login(context.Background(), domain.LoginInput{
		Email:    "nonexistent@test.com",
		Password: "password123",
	})

	require.Error(t, err)
	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 400, svcErr.Code)
}

func TestAuthService_Login_SubscriberRole(t *testing.T) {
	svc, q := newAuthService(t)

	testutil.SeedSubscriber(t, q, "login_sub@test.com", "Login Sub", "11999990001")

	resp, err := svc.Login(context.Background(), domain.LoginInput{
		Email:    "login_sub@test.com",
		Password: "password123",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, domain.RoleSubscriber, resp.User.Role)
}

// ---- GetProfile ----

func TestAuthService_GetProfile_Success(t *testing.T) {
	svc, q := newAuthService(t)

	owner := testutil.SeedOwner(t, q, "profile_owner@test.com", "Profile Owner")

	resp, err := svc.GetProfile(context.Background(), owner.ID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, owner.ID, resp.User.ID)
	assert.Equal(t, "profile_owner@test.com", resp.User.Email)
	assert.Equal(t, "Profile Owner", resp.User.Name)
}

func TestAuthService_GetProfile_NotFound(t *testing.T) {
	svc, _ := newAuthService(t)

	_, err := svc.GetProfile(context.Background(), 999999)

	require.Error(t, err)
	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 404, svcErr.Code)
}

// ---- UpdateProfile ----

func TestAuthService_UpdateProfile_Success(t *testing.T) {
	svc, q := newAuthService(t)

	owner := testutil.SeedOwner(t, q, "update_profile@test.com", "Old Name")

	resp, err := svc.UpdateProfile(context.Background(), owner.ID, domain.UpdateProfileInput{
		Name:  "New Name",
		Phone: "11988887777",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "New Name", resp.User.Name)
	assert.Equal(t, "11988887777", resp.User.Phone)
}

// ---- ChangePassword ----

func TestAuthService_ChangePassword_Success(t *testing.T) {
	svc, q := newAuthService(t)

	owner := testutil.SeedOwner(t, q, "change_pw@test.com", "ChangePass Owner")

	err := svc.ChangePassword(context.Background(), owner.ID, domain.ChangePasswordInput{
		CurrentPassword: "password123",
		NewPassword:     "newpassword456",
	})

	require.NoError(t, err)

	// Verify new password works for login
	resp, err := svc.Login(context.Background(), domain.LoginInput{
		Email:    "change_pw@test.com",
		Password: "newpassword456",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	svc, q := newAuthService(t)

	owner := testutil.SeedOwner(t, q, "wrong_pw@test.com", "WrongPass Owner")

	err := svc.ChangePassword(context.Background(), owner.ID, domain.ChangePasswordInput{
		CurrentPassword: "wrongcurrentpassword",
		NewPassword:     "newpassword456",
	})

	require.Error(t, err)
	var svcErr *domain.ServiceError
	require.ErrorAs(t, err, &svcErr)
	assert.Equal(t, 400, svcErr.Code)
}
