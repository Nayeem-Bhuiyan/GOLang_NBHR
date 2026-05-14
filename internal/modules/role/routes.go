package role

import (
	"nbhr/internal/middleware"
	"nbhr/internal/shared/jwt"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts role routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwt.Manager) {
	auth := rg.Group("", middleware.Authenticate(jwtManager))
	{
		roles := auth.Group("/roles")
		{
			roles.POST("", middleware.RequirePermission("role:create"), h.Create)
			roles.GET("", middleware.RequirePermission("role:list"), h.GetAll)
			roles.GET("/:id", middleware.RequirePermission("role:read"), h.GetByID)
			roles.PUT("/:id", middleware.RequirePermission("role:update"), h.Update)
			roles.DELETE("/:id", middleware.RequirePermission("role:delete"), h.Delete)
			roles.POST("/:id/permissions", middleware.RequirePermission("role:update"), h.AssignPermissions)
			roles.DELETE("/:id/permissions", middleware.RequirePermission("role:update"), h.RemovePermissions)
		}
	}
}