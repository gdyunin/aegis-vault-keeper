package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for application health checking.
type Handler struct{}

// NewHandler creates a new health check handler instance.
func NewHandler() *Handler {
	return &Handler{}
}

// HealthCheck performs application health check.
// @Summary      Health check
// @Description  Returns HTTP 200 if the application is healthy and running
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 "Application is healthy"
// @Router       /health [get]
// .
func (h *Handler) HealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}
