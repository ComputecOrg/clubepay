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
