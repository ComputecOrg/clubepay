package domain

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RegisterOwnerInput struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	Name         string `json:"name" validate:"required"`
	BusinessName string `json:"business_name" validate:"required"`
	Segment      string `json:"segment"`
	Phone        string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterSubscriberInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone"`
}

type SubscribeInput struct {
	PlanID int64 `json:"plan_id" validate:"required,gt=0"`
}

type ValidateUsageInput struct {
	BusinessSlug string `json:"business_slug" validate:"required"`
}

type CreatePlanInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents" validate:"required,gt=0"`
	LimitType   string `json:"limit_type" validate:"required,oneof=daily monthly"`
	LimitCount  int32  `json:"limit_count" validate:"required,gt=0"`
}

type UpdatePlanInput struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents" validate:"required,gt=0"`
	LimitType   string `json:"limit_type" validate:"required,oneof=daily monthly"`
	LimitCount  int32  `json:"limit_count" validate:"required,gt=0"`
}

type UpdateBusinessInput struct {
	Name    string `json:"name" validate:"required"`
	Segment string `json:"segment"`
	Address string `json:"address"`
	LogoURL string `json:"logo_url"`
}

type ApplyReferralInput struct {
	Code string `json:"code" validate:"required"`
}

type UpdateProfileInput struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type RequestPasswordResetInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ConfirmPasswordResetInput struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type ValidateUsageOwnerInput struct {
	SubscriberID int64 `json:"subscriber_id" validate:"gt=0"`
}

// Validate validates a struct using validator/v10 tags.
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if vErrors, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("validacao: campo '%s' falhou na regra '%s'", vErrors[0].Field(), vErrors[0].Tag())
		}
		return fmt.Errorf("validacao: %w", err)
	}
	return nil
}
