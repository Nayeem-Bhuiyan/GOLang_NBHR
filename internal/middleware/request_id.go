package middleware

import (
	"nbhr/internal/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID injects a unique request ID into every request context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(constants.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(constants.ContextKeyRequestID, requestID)
		c.Header(constants.HeaderRequestID, requestID)
		c.Next()
	}
}