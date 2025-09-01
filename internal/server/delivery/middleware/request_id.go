package middleware

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID creates middleware that assigns a unique request ID to each HTTP request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Request.Header.Set(consts.HeaderXRequestID, requestID)
		c.Next()
	}
}
