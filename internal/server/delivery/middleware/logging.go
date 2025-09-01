package middleware

import (
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogging creates middleware that logs HTTP request and response details.
func RequestLogging(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.Request.Header.Get(consts.HeaderXRequestID)

		logger.Infof("HTTP request_id=%s | request: method=%s uri=%s headers=%v",
			requestID,
			c.Request.Method,
			c.Request.RequestURI,
		)

		c.Next()
		for _, err := range c.Errors {
			logger.Errorf("HTTP request_id=%s | error occurred='%s'", requestID, err.Error())
		}

		processingTime := time.Since(start)
		logger.Infof("HTTP request_id=%s | response (processingTime: %s): status=%d size=%d headers=%v",
			requestID,
			processingTime,
			c.Writer.Status(),
			c.Writer.Size(),
		)
	}
}
