package middleware

import (
	"nbhr/internal/constants"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequirePermission checks if the authenticated user's token contains the required permission slug.
// Permissions are loaded into the JWT at login time from the DB.
func RequirePermission(permissionSlug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(constants.ContextKeyUserID)
		if !exists || userID == uuid.Nil {
			response.Error(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		roles, _ := c.Get(constants.ContextKeyRoles)
		roleList, ok := roles.([]string)
		if !ok {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		// Super admins bypass all permission checks
		for _, r := range roleList {
			if r == constants.RoleSuperAdmin {
				c.Next()
				return
			}
		}

		// Check permission slugs stored in token
		permissionsRaw, exists := c.Get("permissions")
		if !exists {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		perms, ok := permissionsRaw.([]string)
		if !ok {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		for _, p := range perms {
			if p == permissionSlug {
				c.Next()
				return
			}
		}

		response.Error(c, apperrors.ErrForbidden)
		c.Abort()
	}
}

// RequireRole checks if the authenticated user has at least one of the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRolesRaw, exists := c.Get(constants.ContextKeyRoles)
		if !exists {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		userRoles, ok := userRolesRaw.([]string)
		if !ok {
			response.Error(c, apperrors.ErrForbidden)
			c.Abort()
			return
		}

		roleSet := make(map[string]struct{}, len(roles))
		for _, r := range roles {
			roleSet[r] = struct{}{}
		}

		for _, ur := range userRoles {
			if _, found := roleSet[ur]; found {
				c.Next()
				return
			}
		}

		response.Error(c, apperrors.ErrForbidden)
		c.Abort()
	}
}

// InjectPermissions is used by Authenticate to also store permissions from the JWT claims.
// Call this after Authenticate; it reads the permissions field set by the JWT manager.
func InjectPermissions(jwtPerms []string, c *gin.Context) {
	c.Set("permissions", jwtPerms)
}