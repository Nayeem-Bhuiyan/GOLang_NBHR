package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateRoleRequest is used for creating a new role.
type CreateRoleRequest struct {
	Name        string      `json:"name"        validate:"required,min=2,max=100"`
	Slug        string      `json:"slug"        validate:"required,slug"`
	Description string      `json:"description" validate:"omitempty,max=500"`
	IsActive    *bool       `json:"is_active"   validate:"omitempty"`
	PermIDs     []uuid.UUID `json:"perm_ids"    validate:"omitempty"`
}

// UpdateRoleRequest is used for updating a role.
type UpdateRoleRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	IsActive    *bool  `json:"is_active"   validate:"omitempty"`
}

// AssignPermissionsRequest assigns permissions to a role.
type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required,min=1"`
}

// RoleResponse is the public role representation.
type RoleResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Description string               `json:"description"`
	IsActive    bool                 `json:"is_active"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}