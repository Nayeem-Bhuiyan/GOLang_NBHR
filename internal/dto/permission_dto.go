package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreatePermissionRequest is used for creating a permission.
type CreatePermissionRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=150"`
	Slug        string `json:"slug"        validate:"required,slug"`
	Resource    string `json:"resource"    validate:"required,min=2,max=100"`
	Action      string `json:"action"      validate:"required,oneof=create read update delete list"`
	Description string `json:"description" validate:"omitempty,max=500"`
	IsActive    *bool  `json:"is_active"   validate:"omitempty"`
}

// UpdatePermissionRequest is used for updating a permission.
type UpdatePermissionRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=2,max=150"`
	Description string `json:"description" validate:"omitempty,max=500"`
	IsActive    *bool  `json:"is_active"   validate:"omitempty"`
}

// PermissionResponse is the public permission representation.
type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}