package delivery

import (
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MiddlewareRegistry manages HTTP middleware registration for the Gin router.
type MiddlewareRegistry struct {
	// logger provides logging functionality for middleware operations.
	logger *zap.SugaredLogger
}

// NewMiddlewareRegistry creates a new middleware registry with the provided logger.
func NewMiddlewareRegistry(logger *zap.SugaredLogger) *MiddlewareRegistry {
	return &MiddlewareRegistry{
		logger: logger,
	}
}

// RegisterMiddlewares configures standard middleware for the Gin router.
func (mr *MiddlewareRegistry) RegisterMiddlewares(router *gin.Engine) {
	router.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.RequestLogging(mr.logger.Named("http-request")),
	)
}
