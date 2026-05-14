package middleware

import (
	"strings"

	"nbhr/internal/constants"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/shared/jwt"
	"nbhr/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Authenticate validates the Bearer JWT access token and injects claims into context.
func Authenticate(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(constants.HeaderAuthorization)
		if authHeader == "" {
			response.Error(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, constants.BearerPrefix) {
			response.Error(c, apperrors.ErrTokenInvalid)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, constants.BearerPrefix)
		claims, err := jwtManager.ValidateAccessToken(tokenStr)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		// Inject user context
		c.Set(constants.ContextKeyUserID, claims.UserID)
		c.Set(constants.ContextKeyUserEmail, claims.Email)
		c.Set(constants.ContextKeyRoles, claims.Roles)

		c.Next()
	}
}