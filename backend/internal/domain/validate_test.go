package domain_test

import (
	"testing"

	"github.com/clubepay/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidate_RegisterOwnerInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.RegisterOwnerInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: domain.RegisterOwnerInput{
				Email:        "test@example.com",
				Password:     "password123",
				Name:         "Joao",
				BusinessName: "Cafe do Joao",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			input: domain.RegisterOwnerInput{
				Password:     "password123",
				Name:         "Joao",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			input: domain.RegisterOwnerInput{
				Email:        "not-an-email",
				Password:     "password123",
				Name:         "Joao",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "short password",
			input: domain.RegisterOwnerInput{
				Email:        "test@example.com",
				Password:     "short",
				Name:         "Joao",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			input: domain.RegisterOwnerInput{
				Email:        "test@example.com",
				Password:     "password123",
				BusinessName: "Cafe",
			},
			wantErr: true,
		},
		{
			name: "missing business name",
			input: domain.RegisterOwnerInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Joao",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_LoginInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.LoginInput
		wantErr bool
	}{
		{
			name:    "valid",
			input:   domain.LoginInput{Email: "a@b.com", Password: "12345678"},
			wantErr: false,
		},
		{
			name:    "missing email",
			input:   domain.LoginInput{Password: "12345678"},
			wantErr: true,
		},
		{
			name:    "missing password",
			input:   domain.LoginInput{Email: "a@b.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_RegisterSubscriberInput(t *testing.T) {
	valid := domain.RegisterSubscriberInput{
		Email:    "sub@example.com",
		Password: "password123",
		Name:     "Maria",
	}
	assert.NoError(t, domain.Validate(valid))

	invalid := domain.RegisterSubscriberInput{Email: "bad"}
	assert.Error(t, domain.Validate(invalid))
}

func TestValidate_UpdateProfileInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.UpdateProfileInput
		wantErr bool
	}{
		{
			name:    "valid with name only",
			input:   domain.UpdateProfileInput{Name: "Maria"},
			wantErr: false,
		},
		{
			name:    "valid with name and phone",
			input:   domain.UpdateProfileInput{Name: "Maria", Phone: "11999990000"},
			wantErr: false,
		},
		{
			name:    "missing name",
			input:   domain.UpdateProfileInput{Phone: "11999990000"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ChangePasswordInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.ChangePasswordInput
		wantErr bool
	}{
		{
			name:    "valid",
			input:   domain.ChangePasswordInput{CurrentPassword: "oldpass1", NewPassword: "newpass1"},
			wantErr: false,
		},
		{
			name:    "missing current password",
			input:   domain.ChangePasswordInput{NewPassword: "newpass1"},
			wantErr: true,
		},
		{
			name:    "missing new password",
			input:   domain.ChangePasswordInput{CurrentPassword: "oldpass1"},
			wantErr: true,
		},
		{
			name:    "new password too short",
			input:   domain.ChangePasswordInput{CurrentPassword: "oldpass1", NewPassword: "short"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_RequestPasswordResetInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.RequestPasswordResetInput
		wantErr bool
	}{
		{
			name:    "valid email",
			input:   domain.RequestPasswordResetInput{Email: "user@example.com"},
			wantErr: false,
		},
		{
			name:    "missing email",
			input:   domain.RequestPasswordResetInput{},
			wantErr: true,
		},
		{
			name:    "invalid email format",
			input:   domain.RequestPasswordResetInput{Email: "not-an-email"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ConfirmPasswordResetInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.ConfirmPasswordResetInput
		wantErr bool
	}{
		{
			name:    "valid",
			input:   domain.ConfirmPasswordResetInput{Token: "abc123token", Password: "newpass99"},
			wantErr: false,
		},
		{
			name:    "missing token",
			input:   domain.ConfirmPasswordResetInput{Password: "newpass99"},
			wantErr: true,
		},
		{
			name:    "missing password",
			input:   domain.ConfirmPasswordResetInput{Token: "abc123token"},
			wantErr: true,
		},
		{
			name:    "password too short",
			input:   domain.ConfirmPasswordResetInput{Token: "abc123token", Password: "short"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ValidateUsageOwnerInput(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.ValidateUsageOwnerInput
		wantErr bool
	}{
		{
			name:    "valid subscriber id",
			input:   domain.ValidateUsageOwnerInput{SubscriberID: 1},
			wantErr: false,
		},
		{
			name:    "zero subscriber id",
			input:   domain.ValidateUsageOwnerInput{SubscriberID: 0},
			wantErr: true,
		},
		{
			name:    "negative subscriber id",
			input:   domain.ValidateUsageOwnerInput{SubscriberID: -5},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
