package datasync

import "github.com/gin-gonic/gin"

// RegisterRoutes registers data synchronization routes with the provided router group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	syncGroup := r.Group("/sync")
	syncGroup.POST("", h.Push)
	syncGroup.GET("", h.Pull)
}
