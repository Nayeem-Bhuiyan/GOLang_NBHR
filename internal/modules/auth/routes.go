package auth

import (
	"nbhr/internal/middleware"
	"nbhr/internal/shared/jwt"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts auth routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwt.Manager) {
	authGroup := rg.Group("/auth")
	{
		// Public routes
		authGroup.POST("/login", h.Login)
		authGroup.POST("/register", h.Register)
		authGroup.POST("/refresh", h.RefreshToken)

		// Authenticated routes
		protected := authGroup.Group("", middleware.Authenticate(jwtManager))
		{
			protected.POST("/logout", h.Logout)
			protected.POST("/logout-all", h.LogoutAll)
			protected.POST("/change-password", h.ChangePassword)
		}
	}
}