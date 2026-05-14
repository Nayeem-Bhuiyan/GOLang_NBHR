package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest is used for admin user creation.
type CreateUserRequest struct {
	FirstName string      `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string      `json:"last_name"  validate:"required,min=2,max=100"`
	Email     string      `json:"email"      validate:"required,email"`
	Password  string      `json:"password"   validate:"required,min=8,max=72"`
	Phone     string      `json:"phone"      validate:"omitempty,min=7,max=20"`
	RoleIDs   []uuid.UUID `json:"role_ids"   validate:"omitempty"`
}

// UpdateUserRequest is used for updating a user's profile.
type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  string `json:"last_name"  validate:"omitempty,min=2,max=100"`
	Phone     string `json:"phone"      validate:"omitempty,min=7,max=20"`
}

// AssignRolesRequest assigns roles to a user.
type AssignRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" validate:"required,min=1"`
}

// UserResponse is the public user representation.
type UserResponse struct {
	ID          uuid.UUID      `json:"id"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Email       string         `json:"email"`
	Phone       string         `json:"phone"`
	IsActive    bool           `json:"is_active"`
	IsVerified  bool           `json:"is_verified"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	Roles       []RoleResponse `json:"roles,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}