package swagger

import "github.com/gin-gonic/gin"

// RegisterRoutes configures Swagger documentation endpoint.
// Exposes OpenAPI documentation at /swagger/* endpoints.
func RegisterRoutes(r *gin.RouterGroup, h gin.HandlerFunc) {
	r.GET("/swagger/*any", h)
}
