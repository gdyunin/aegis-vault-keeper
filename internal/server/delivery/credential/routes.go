package credential

import "github.com/gin-gonic/gin"

// RegisterRoutes configures credential endpoints in the router group.
// Sets up CRUD operations: POST/GET for collections, GET/PUT for individual items.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	credentialsGroup := r.Group("/credentials")
	credentialsGroup.POST("", h.Push)
	credentialsGroup.GET("", h.List)

	credentialsIDGroup := credentialsGroup.Group("/:id")
	credentialsIDGroup.GET("", h.Pull)
	credentialsIDGroup.PUT("", h.Push)
}
