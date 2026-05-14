package router

import (
	"net/http"
	"time"

	"nbhr/config"
	"nbhr/internal/middleware"
	"nbhr/internal/modules/auth"
	"nbhr/internal/modules/permission"
	"nbhr/internal/modules/role"
	"nbhr/internal/modules/user"
	"nbhr/internal/shared/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Dependencies holds all module handlers and shared infrastructure.
type Dependencies struct {
	AuthHandler       *auth.Handler
	UserHandler       *user.Handler
	RoleHandler       *role.Handler
	PermissionHandler *permission.Handler
	JWTManager        *jwt.Manager
	Config            *config.Config
	Logger            *zap.Logger
}

// Setup configures and returns the main Gin engine with all middleware and routes.
func Setup(deps *Dependencies) *gin.Engine {
	if deps.Config.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// ---------------------------------------------------------------------------
	// Global middleware stack — order matters
	// ---------------------------------------------------------------------------
	engine.Use(
		middleware.CORS(deps.Config.CORS.AllowedOrigins),
		middleware.RequestID(),
		middleware.Logger(deps.Logger),
		middleware.Timeout(deps.Config.App.RequestTimeout),
		// Global IP-based rate limit applied to every route
		middleware.RateLimitByIP(
			deps.Config.RateLimit.Requests,
			deps.Config.RateLimit.Period,
			deps.Logger,
		),
		gin.Recovery(),
	)

	// ---------------------------------------------------------------------------
	// System routes — no auth, no extra rate limiting
	// ---------------------------------------------------------------------------
	engine.GET("/health", healthCheck())
	engine.GET("/ready", readinessCheck())

	// ---------------------------------------------------------------------------
	// API v1 routes
	// ---------------------------------------------------------------------------
	v1 := engine.Group("/api/" + deps.Config.App.Version)
	{
		registerAuthRoutes(v1, deps)
		user.RegisterRoutes(v1, deps.UserHandler, deps.JWTManager)
		role.RegisterRoutes(v1, deps.RoleHandler, deps.JWTManager)
		permission.RegisterRoutes(v1, deps.PermissionHandler, deps.JWTManager)
	}

	// ---------------------------------------------------------------------------
	// 404 catch-all
	// ---------------------------------------------------------------------------
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "route not found",
		})
	})

	return engine
}

// registerAuthRoutes mounts auth routes and applies tighter per-route limits
// on sensitive endpoints (login, register) before delegating to the handler.
func registerAuthRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authGroup := rg.Group("/auth")
	{
		// Sensitive public endpoints get their own stricter rate limits
		// applied on top of the global IP limiter.
		authGroup.POST("/login",
			middleware.RateLimitByRoute(10, time.Minute, deps.Logger),
			deps.AuthHandler.Login,
		)
		authGroup.POST("/register",
			middleware.RateLimitByRoute(5, time.Minute, deps.Logger),
			deps.AuthHandler.Register,
		)
		authGroup.POST("/refresh",
			middleware.RateLimitByRoute(30, time.Minute, deps.Logger),
			deps.AuthHandler.RefreshToken,
		)

		// Authenticated auth endpoints — standard global limit applies
		protected := authGroup.Group("", middleware.Authenticate(deps.JWTManager))
		{
			protected.POST("/logout", deps.AuthHandler.Logout)
			protected.POST("/logout-all", deps.AuthHandler.LogoutAll)
			protected.POST("/change-password", deps.AuthHandler.ChangePassword)
		}
	}
}

func healthCheck() gin.HandlerFunc {
	start := time.Now()
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"uptime":  time.Since(start).String(),
			"service": "nbhr",
		})
	}
}

func readinessCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}