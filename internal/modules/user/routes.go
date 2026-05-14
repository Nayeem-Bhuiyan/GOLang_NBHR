package user

import (
	"nbhr/internal/middleware"
	"nbhr/internal/shared/jwt"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts user routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwt.Manager) {
	auth := rg.Group("", middleware.Authenticate(jwtManager))
	{
		// Authenticated user self-service
		auth.GET("/me", h.Me)

		// Admin user management
		users := auth.Group("/users")
		{
			users.POST("", middleware.RequirePermission("user:create"), h.Create)
			users.GET("", middleware.RequirePermission("user:list"), h.GetAll)
			users.GET("/:id", middleware.RequirePermission("user:read"), h.GetByID)
			users.PUT("/:id", middleware.RequirePermission("user:update"), h.Update)
			users.DELETE("/:id", middleware.RequirePermission("user:delete"), h.Delete)
			users.POST("/:id/roles", middleware.RequirePermission("user:update"), h.AssignRoles)
			users.PATCH("/:id/toggle-active", middleware.RequirePermission("user:update"), h.ToggleActive)
		}
	}
}