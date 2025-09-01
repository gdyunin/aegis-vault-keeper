package health

import "github.com/gin-gonic/gin"

// RegisterRoutes configures health check endpoint in the router group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/health", h.HealthCheck)
}
