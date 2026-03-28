package domain

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone,omitempty"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// User roles
const (
	RoleOwner      = "owner"
	RoleSubscriber = "subscriber"
)
