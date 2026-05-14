package middleware

import (
	"time"

	"nbhr/internal/constants"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a structured request logging middleware using zap.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get(constants.ContextKeyRequestID)

		fields := []zap.Field{
			zap.String("request_id", toString(requestID)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if status >= 500 {
			log.Error("request completed with server error", fields...)
		} else if status >= 400 {
			log.Warn("request completed with client error", fields...)
		} else {
			log.Info("request completed", fields...)
		}
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}