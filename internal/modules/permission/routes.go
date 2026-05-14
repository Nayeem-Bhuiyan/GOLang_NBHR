package permission

import (
	"nbhr/internal/middleware"
	"nbhr/internal/shared/jwt"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts permission routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwt.Manager) {
	auth := rg.Group("", middleware.Authenticate(jwtManager))
	{
		perms := auth.Group("/permissions")
		{
			perms.POST("", middleware.RequirePermission("permission:create"), h.Create)
			perms.GET("", middleware.RequirePermission("permission:list"), h.GetAll)
			perms.GET("/:id", middleware.RequirePermission("permission:read"), h.GetByID)
			perms.PUT("/:id", middleware.RequirePermission("permission:update"), h.Update)
			perms.DELETE("/:id", middleware.RequirePermission("permission:delete"), h.Delete)
		}
	}
}